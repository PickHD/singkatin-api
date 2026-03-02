package routes

import (
	"singkatin-api/shortener/internal/bootstrap"
	shortenerpb "singkatin-api/shortener/pkg/api/v1/proto/shortener"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func ServeGRPC(container *bootstrap.Container) *grpc.Server {
	// call register
	return register(container)
}

func register(container *bootstrap.Container) *grpc.Server {
	reflection.Register(container.GRPC)

	shortenerpb.RegisterShortenerServiceServer(container.GRPC, container.ShortController)

	return container.GRPC
}
