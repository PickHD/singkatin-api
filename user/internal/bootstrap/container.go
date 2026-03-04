package bootstrap

import (
	"context"
	shortenerpb "singkatin-api/proto/api/v1/proto/shortener"
	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/controller"
	"singkatin-api/user/internal/infrastructure"
	"singkatin-api/user/internal/middleware"
	"singkatin-api/user/internal/repository"
	"singkatin-api/user/internal/service"
	"singkatin-api/user/pkg/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Container ...
type Container struct {
	Context  context.Context
	Config   *config.Config
	RabbitMQ *infrastructure.RabbitMQConnectionProvider
	Tracer   *infrastructure.TracerProvider
	Mongo    *infrastructure.MongoConnectionProvider
	GRPC     *grpc.ClientConn
	JWT      *infrastructure.JwtProvider

	HealthCheckController controller.HealthCheckController
	UserController        controller.UserController

	AuthMiddleware *middleware.AuthMiddleware
}

// NewContainer configuring dependencies app needed
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Load()

	db := infrastructure.NewMongoConnection(ctx, cfg)
	rabbitmq := infrastructure.NewRabbitMQConnection(ctx, cfg)
	tracer := infrastructure.NewTracerProvider(ctx, cfg)
	jwt := infrastructure.NewJWTProvider(cfg)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	grpcConn, err := grpc.NewClient(cfg.Common.GRPCPort, opts...)
	if err != nil {
		return nil, err
	}

	// GRPC Client
	shortClient := shortenerpb.NewShortenerServiceClient(grpcConn)

	// repository
	healthCheckRepo := repository.NewHealthCheckRepository(cfg, tracer.Tracer("Health Check Repository"), db.GetDatabase())
	userRepo := repository.NewUserRepository(cfg, tracer.Tracer("User Repository"), db.GetDatabase(), rabbitmq.GetClient())

	// service
	healthCheckSvc := service.NewHealthCheckService(cfg, tracer.Tracer("Health Check Service"), healthCheckRepo)
	userSvc := service.NewUserService(cfg, tracer.Tracer("User Service"), userRepo, shortClient)

	// controller
	healthCheckController := controller.NewHealthCheckController(cfg, tracer.Tracer("Health Check Controller"), healthCheckSvc)
	userController := controller.NewUserController(cfg, tracer.Tracer("User Controller"), userSvc)

	// middleware
	authMiddleware := middleware.NewAuthMiddleware(jwt)

	logger.Info("USER SERVICE RUN SUCCESSFULLY")

	return &Container{
		Context:  ctx,
		Config:   cfg,
		Mongo:    db,
		RabbitMQ: rabbitmq,
		Tracer:   tracer,
		GRPC:     grpcConn,
		JWT:      jwt,

		HealthCheckController: healthCheckController,
		UserController:        userController,

		AuthMiddleware: authMiddleware,
	}, nil
}

// Close method will close any instances before app terminated
func (c *Container) Close(ctx context.Context) {
	defer func(ctx context.Context) {
		// TRACER
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if c.Tracer != nil {
			if err := c.Tracer.Shutdown(ctx); err != nil {
				panic(err)
			}
		}

		// MONGO
		if c.Mongo != nil && c.Mongo.GetClient() != nil {
			if err := c.Mongo.GetClient().Disconnect(c.Context); err != nil {
				panic(err)
			}
		}

		// RABBITMQ
		if c.RabbitMQ != nil && c.RabbitMQ.GetClient() != nil {
			if err := c.RabbitMQ.Close(); err != nil {
				panic(err)
			}
		}

		// GRPC
		if c.GRPC != nil {
			if err := c.GRPC.Close(); err != nil {
				panic(err)
			}
		}

	}(ctx)
}
