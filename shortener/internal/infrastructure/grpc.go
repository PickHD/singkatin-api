package infrastructure

import (
	"github.com/PickHD/singkatin-revamp/shortener/internal/application"
	shortenerpb "github.com/PickHD/singkatin-revamp/shortener/pkg/api/v1/proto/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func ServeGRPC(app *application.App) *grpc.Server {
	// call register
	return register(app)
}

func register(app *application.App) *grpc.Server {
	var dep = application.SetupDependencyInjection(app)

	reflection.Register(app.GRPC)

	shortenerpb.RegisterShortenerServiceServer(app.GRPC, dep.ShortController)

	return app.GRPC
}
