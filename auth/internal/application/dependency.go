package application

import (
	"github.com/PickHD/singkatin-revamp/auth/internal/controller"
	"github.com/PickHD/singkatin-revamp/auth/internal/repository"
	"github.com/PickHD/singkatin-revamp/auth/internal/service"
)

type Dependency struct {
	HealthCheckController controller.HealthCheckController
	AuthController        controller.AuthController
}

func SetupDependencyInjection(app *App) *Dependency {
	// repository
	healthCheckRepoImpl := repository.NewHealthCheckRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB, app.Redis)
	authRepoImpl := repository.NewAuthRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB, app.Redis)

	// service
	healthCheckSvcImpl := service.NewHealthCheckService(app.Context, app.Config, app.Tracer, healthCheckRepoImpl)
	authSvcImpl := service.NewAuthService(app.Context, app.Config, app.Logger, app.Tracer, app.Mailer, authRepoImpl)

	// controller
	healthCheckControllerImpl := controller.NewHealthCheckController(app.Context, app.Config, app.Tracer, healthCheckSvcImpl)
	authControllerImpl := controller.NewAuthController(app.Context, app.Config, app.Logger, app.Tracer, authSvcImpl)

	return &Dependency{
		HealthCheckController: healthCheckControllerImpl,
		AuthController:        authControllerImpl,
	}
}
