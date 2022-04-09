package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/peer"
)

func TokensFromContext(ctx context.Context) (string, string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", "", fmt.Errorf("no peer data")
	}

	if p.AuthInfo == nil {
		return "", "", fmt.Errorf("no auth info")
	}

	jwtInfo, ok := p.AuthInfo.(JWTAuthInfo)
	if !ok {
		return "", "", fmt.Errorf("invalid auth info type")
	}

	cc := jwtInfo.GetClaims()
	if cc == nil {
		return "", "", fmt.Errorf("no claims on token")
	}

	return cc.TwitchUserAccess, cc.TwitchUserRefresh, nil
}

func TwitchLoginFromContext(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no peer data")
	}

	if p.AuthInfo == nil {
		return "", fmt.Errorf("no auth info")
	}

	jwtInfo, ok := p.AuthInfo.(JWTAuthInfo)
	if !ok {
		return "", fmt.Errorf("invalid auth info type")
	}

	return jwtInfo.GetClaims().Subject, nil
}
