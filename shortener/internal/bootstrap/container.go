package bootstrap

import (
	"context"
	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/controller"
	"singkatin-api/shortener/internal/infrastructure"
	"singkatin-api/shortener/internal/repository"
	"singkatin-api/shortener/internal/service"
	"singkatin-api/shortener/pkg/logger"
	"time"

	"google.golang.org/grpc"
)

// Container ...
type Container struct {
	Context  context.Context
	Config   *config.Config
	DB       *infrastructure.MongoConnectionProvider
	Redis    *infrastructure.RedisConnectionProvider
	RabbitMQ *infrastructure.RabbitMQConnectionProvider
	Tracer   *infrastructure.TracerProvider
	GRPC     *grpc.Server

	HealthCheckController controller.HealthCheckController
	ShortController       *controller.ShortControllerImpl
}

// NewContainer configuring dependencies app needed
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Load()

	db := infrastructure.NewMongoConnection(ctx, cfg)
	redis := infrastructure.NewRedisConnection(ctx, cfg)
	rabbitmq := infrastructure.NewRabbitMQConnection(ctx, cfg)
	tracer := infrastructure.NewTracerProvider(ctx, cfg)
	grpc := grpc.NewServer()

	// repository
	healthCheckRepo := repository.NewHealthCheckRepository(cfg, tracer.Tracer("Shortener Repository"), db.GetDatabase(), redis.GetClient())
	shortRepo := repository.NewShortRepository(cfg, tracer.Tracer("Shortener Repository"), db.GetDatabase(), redis.GetClient(), rabbitmq.GetClient())

	// service
	healthCheckSvc := service.NewHealthCheckService(cfg, tracer.Tracer("Shortener Service"), healthCheckRepo)
	shortSvc := service.NewShortService(cfg, tracer.Tracer("Shortener Service"), shortRepo)

	// controller
	healthCheckController := controller.NewHealthCheckController(cfg, tracer.Tracer("Shortener Controller"), healthCheckSvc)
	shortController := controller.NewShortController(cfg, tracer.Tracer("Shortener Controller"), shortSvc)

	logger.Info("SHORTENER SERVICE RUN SUCCESSFULLY")

	return &Container{
		Context:  ctx,
		Config:   cfg,
		DB:       db,
		Redis:    redis,
		RabbitMQ: rabbitmq,
		Tracer:   tracer,
		GRPC:     grpc,

		HealthCheckController: healthCheckController,
		ShortController:       shortController,
	}, nil
}

// Close method will close any instances before app terminated
func (c *Container) Close(ctx context.Context) {
	defer func(ctx context.Context) {
		// DB
		if c.DB != nil && c.DB.Client() != nil {
			if err := c.DB.Client().Disconnect(ctx); err != nil {
				panic(err)
			}
		}

		// TRACER
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if c.Tracer != nil {
			if err := c.Tracer.Shutdown(ctx); err != nil {
				panic(err)
			}
		}

		// REDIS
		if c.Redis != nil && c.Redis.GetClient() != nil {
			if err := c.Redis.Close(); err != nil {
				panic(err)
			}
		}

		// RABBITMQ
		if c.RabbitMQ != nil && c.RabbitMQ.GetClient() != nil {
			if err := c.RabbitMQ.Close(); err != nil {
				panic(err)
			}
		}

	}(ctx)
}
