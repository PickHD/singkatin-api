package bootstrap

import (
	"context"
	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/controller"
	"singkatin-api/auth/internal/infrastructure"
	"singkatin-api/auth/internal/repository"
	"singkatin-api/auth/internal/service"
	"singkatin-api/auth/pkg/logger"
	"time"
)

// Container ...
type Container struct {
	Context context.Context
	Config  *config.Config
	DB      *infrastructure.MongoConnectionProvider
	Redis   *infrastructure.RedisConnectionProvider
	Tracer  *infrastructure.TracerProvider
	Mailer  *infrastructure.EmailProvider

	HealthCheckController controller.HealthCheckController
	AuthController        controller.AuthController
}

// NewContainer configuring dependencies app needed
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Load()

	db := infrastructure.NewMongoConnection(ctx, cfg)
	redis := infrastructure.NewRedisConnection(ctx, cfg)
	tracer := infrastructure.NewTracerProvider(ctx, cfg)
	mailer := infrastructure.NewEmailProvider(cfg)

	// repository
	healthCheckRepo := repository.NewHealthCheckRepository(ctx, cfg, tracer.Tracer("Auth Repository"), db.GetDatabase(), redis.GetClient())
	authRepo := repository.NewAuthRepository(ctx, cfg, tracer.Tracer("Auth Repository"), db.GetDatabase(), redis.GetClient())

	// service
	healthCheckSvc := service.NewHealthCheckService(ctx, cfg, tracer.Tracer("Auth Service"), healthCheckRepo)
	authSvc := service.NewAuthService(ctx, cfg, tracer.Tracer("Auth Service"), mailer.GetDialer(), authRepo)

	// controller
	healthCheckController := controller.NewHealthCheckController(ctx, cfg, tracer.Tracer("Auth Controller"), healthCheckSvc)
	authController := controller.NewAuthController(ctx, cfg, tracer.Tracer("Auth Controller"), authSvc)

	logger.Info("AUTH SERVICE RUN SUCCESSFULLY")

	return &Container{
		Context: ctx,
		Config:  cfg,
		DB:      db,
		Redis:   redis,
		Tracer:  tracer,
		Mailer:  mailer,

		HealthCheckController: healthCheckController,
		AuthController:        authController,
	}, nil
}

// Close method will close any instances before app terminated
func (c *Container) Close(ctx context.Context) {
	defer func(ctx context.Context) {
		// DB
		if err := c.DB.Client().Disconnect(ctx); err != nil {
			panic(err)
		}

		// TRACER
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := c.Tracer.Shutdown(ctx); err != nil {
			panic(err)
		}

		// REDIS
		if err := c.Redis.Close(); err != nil {
			panic(err)
		}

		// MAILER
		if err := c.Mailer.Close(); err != nil {
			panic(err)
		}

	}(ctx)
}
