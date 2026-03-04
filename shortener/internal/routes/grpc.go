package routes

import (
	shortenerpb "singkatin-api/proto/api/v1/proto/shortener"
	"singkatin-api/shortener/internal/bootstrap"

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
