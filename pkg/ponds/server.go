package ponds

import (
	"context"

	rpc_ponds "github.com/devnull-twitch/pond-com/protobuf/com/v1"
	"github.com/devnull-twitch/pond/pkg/auth"
	"github.com/nicklaw5/helix"
	"github.com/thanhpk/randstr"
	"google.golang.org/grpc"
)

type server struct {
	rpc_ponds.UnimplementedLoginServiceServer
	twClient *helix.Client
}

func NewLoginServer(twClient *helix.Client) rpc_ponds.LoginServiceServer {
	return &server{
		twClient: twClient,
	}
}

func Register(grpcServer *grpc.Server, serverImpl rpc_ponds.LoginServiceServer) {
	rpc_ponds.RegisterLoginServiceServer(grpcServer, serverImpl)
}

func (s *server) Start(ctx context.Context, payload *rpc_ponds.StartRequest) (*rpc_ponds.StartResponse, error) {
	reqToken := randstr.Base62(10)
	auth.Add(reqToken)
	url := s.twClient.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		State:        reqToken,
		Scopes: []string{
			"channel:read:redemptions",
			"channel:manage:redemptions",
			"channel:read:subscriptions",
			"bits:read",
			"chat:read",
			"channel:read:polls",
			"channel:manage:polls",
			"channel:read:predictions",
			"channel:manage:predictions",
		},
	})

	return &rpc_ponds.StartResponse{
		AuthUrl:      url,
		RequestToken: reqToken,
	}, nil
}
