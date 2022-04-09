package main

import (
	"log"
	"net"
	"os"

	"github.com/devnull-twitch/pond/pkg/auth"
	"github.com/devnull-twitch/pond/pkg/handler"
	"github.com/devnull-twitch/pond/pkg/ponds"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nicklaw5/helix"
	"google.golang.org/grpc"
)

func main() {
	godotenv.Load(".env.yaml")

	/*
		conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal(fmt.Errorf("unable to connect to database: %w", err))
		}
	*/

	client, err := helix.NewClient(&helix.Options{
		ClientID:       os.Getenv("TW_CLIENTID"),
		AppAccessToken: os.Getenv("TW_APP_ACCESS"),
		RedirectURI:    "http://localhost:8080/pond/tw",
	})
	if err != nil {
		log.Fatal("unable to create twitch api client")
	}

	jwtCT := auth.NewPondJWT("")
	s := grpc.NewServer(grpc.Creds(jwtCT))

	loginServer := ponds.NewLoginServer(client)
	ponds.Register(s, loginServer)

	go func() {
		r := gin.Default()
		r.SetTrustedProxies(nil)

		h := handler.New(client)

		pondBaseGrp := r.Group("/pond")
		{
			pondBaseGrp.GET("/tw", h.AuthHandler())
		}

		r.Run(os.Getenv("WEBSERVER_BIND"))
	}()

	lis, err := net.Listen("tcp", ":50201")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s.Serve(lis)
}
