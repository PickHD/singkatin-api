package application

import (
	"github.com/PickHD/singkatin-revamp/user/internal/controller"
	"github.com/PickHD/singkatin-revamp/user/internal/repository"
	"github.com/PickHD/singkatin-revamp/user/internal/service"
	shortenerpb "github.com/PickHD/singkatin-revamp/user/pkg/api/v1/proto/shortener"
)

type Dependency struct {
	HealthCheckController controller.HealthCheckController
	UserController        controller.UserController
}

func SetupDependencyInjection(app *App) *Dependency {
	shortenerServiceClient := shortenerpb.NewShortenerServiceClient(app.GRPC)

	// repository
	healthCheckRepoImpl := repository.NewHealthCheckRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB)
	userRepoImpl := repository.NewUserRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB, app.RabbitMQ)

	// service
	healthCheckSvcImpl := service.NewHealthCheckService(app.Context, app.Config, app.Tracer, healthCheckRepoImpl)
	userSvcImpl := service.NewUserService(app.Context, app.Config, app.Logger, app.Tracer, userRepoImpl, shortenerServiceClient)

	// controller
	healthCheckControllerImpl := controller.NewHealthCheckController(app.Context, app.Config, app.Tracer, healthCheckSvcImpl)
	userControllerImpl := controller.NewUserController(app.Context, app.Config, app.Logger, app.Tracer, userSvcImpl)

	return &Dependency{
		HealthCheckController: healthCheckControllerImpl,
		UserController:        userControllerImpl,
	}
}
