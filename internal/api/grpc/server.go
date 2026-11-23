package grpcapi

import (
	"google.golang.org/grpc"

	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	middleware "github.com/vishenosik/notify/internal/api/grpc/middleware"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
)

func NewServer(
	adminKey string,
	push interfaces.PushUseCase,
	clients interfaces.ClientUseCase,
	keys interfaces.APIKeyUseCase,
) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryAuth(adminKey, keys),
			middleware.UnaryStatus(),
		),
	)
	notifyv1.RegisterWebPushServer(server, newWebPushServer(push))
	notifyv1.RegisterClientRegistryServer(server, newClientRegistryServer(clients, keys))
	return server
}
