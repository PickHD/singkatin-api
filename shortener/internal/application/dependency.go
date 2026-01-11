package application

import (
	"github.com/PickHD/singkatin-revamp/shortener/internal/controller"
	"github.com/PickHD/singkatin-revamp/shortener/internal/repository"
	"github.com/PickHD/singkatin-revamp/shortener/internal/service"
)

type Dependency struct {
	HealthCheckController controller.HealthCheckController
	ShortController       *controller.ShortControllerImpl
}

func SetupDependencyInjection(app *App) *Dependency {
	// repository
	healthCheckRepoImpl := repository.NewHealthCheckRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB, app.Redis)
	shortRepoImpl := repository.NewShortRepository(app.Context, app.Config, app.Logger, app.Tracer, app.DB, app.Redis, app.RabbitMQ)

	// service
	healthCheckSvcImpl := service.NewHealthCheckService(app.Context, app.Config, app.Tracer, healthCheckRepoImpl)
	shortSvcImpl := service.NewShortService(app.Context, app.Config, app.Logger, app.Tracer, shortRepoImpl)

	// controller
	healthCheckControllerImpl := controller.NewHealthCheckController(app.Context, app.Config, app.Tracer, healthCheckSvcImpl)
	shortControllerImpl := controller.NewShortController(app.Context, app.Config, app.Logger, app.Tracer, shortSvcImpl)

	return &Dependency{
		HealthCheckController: healthCheckControllerImpl,
		ShortController:       shortControllerImpl,
	}
}
