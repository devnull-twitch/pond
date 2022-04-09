package auth

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"encoding/binary"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/credentials"
)

type pondJwtTC struct {
	token string
}

func NewPondJWT(token string) credentials.TransportCredentials {
	return &pondJwtTC{token}
}

func (auth *pondJwtTC) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		ProtocolVersion:  "1.0",
		SecurityProtocol: "none",
		ServerName:       "pond",
	}
}

func (auth *pondJwtTC) Clone() credentials.TransportCredentials {
	return &pondJwtTC{}
}

func (auth *pondJwtTC) OverrideServerName(string) error {
	return nil
}

type JWTAuthInfo interface {
	GetClaims() *CustomClaims
}

type jwtAuthType struct {
	claims *CustomClaims
}

func (jwtType *jwtAuthType) AuthType() string {
	return "pond_jwt"
}

func (jwtType *jwtAuthType) GetClaims() *CustomClaims {
	return jwtType.claims
}

func (auth *pondJwtTC) ClientHandshake(
	ctx context.Context,
	authority string,
	conn net.Conn,
) (net.Conn, credentials.AuthInfo, error) {
	buf := make([]byte, 4)
	tokenBytes := []byte(auth.token)
	binary.BigEndian.PutUint32(buf, uint32(len(tokenBytes)))

	_, err := conn.Write(buf)
	if err != nil {
		return conn, nil, fmt.Errorf("unable to write token length")
	}

	if len(tokenBytes) > 0 {
		conn.Write(tokenBytes)
		if err != nil {
			return conn, nil, fmt.Errorf("unable to write token")
		}
	}

	resBuf := make([]byte, 1)
	_, err = conn.Read(resBuf)
	if err != nil {
		return conn, nil, fmt.Errorf("unable to get ACK")
	}

	return conn, &jwtAuthType{}, nil
}

func (auth *pondJwtTC) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	rawBuffer := make([]byte, 4)
	_, err := conn.Read(rawBuffer)
	if err != nil {
		return conn, nil, fmt.Errorf("unable to read token length")
	}
	tokenLength := binary.BigEndian.Uint32(rawBuffer)

	// also allow non authed calls
	if tokenLength == 0 {
		// still send ACK on 0 length token
		_, err = conn.Write([]byte{1})
		if err != nil {
			return conn, nil, fmt.Errorf("unable to send ACK")
		}

		fmt.Println("empty but success")
		return conn, nil, nil
	}

	tokenBytes := make([]byte, tokenLength)
	_, err = conn.Read(tokenBytes)
	if err != nil {
		return conn, nil, fmt.Errorf("unable to read token")
	}

	claims := &CustomClaims{}
	_, err = jwt.ParseWithClaims(string(tokenBytes), claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return conn, nil, fmt.Errorf("unable to validate token")
	}

	_, err = conn.Write([]byte{1})
	if err != nil {
		return conn, nil, fmt.Errorf("unable to send ACK")
	}

	fmt.Println("success")

	return conn, &jwtAuthType{claims}, nil
}

type CustomClaims struct {
	TwitchUserAccess  string `json:"tw_user_access"`
	TwitchUserRefresh string `json:"tw_user_refresh"`
	*jwt.RegisteredClaims
}

func CreateToken(
	accountName string,
	twUserAccess string,
	twUserRefresh string,
) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("HS256"))
	t.Claims = &CustomClaims{
		TwitchUserAccess:  twUserAccess,
		TwitchUserRefresh: twUserRefresh,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			Subject:   accountName,
		},
	}

	return t.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
