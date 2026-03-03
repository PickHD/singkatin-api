package bootstrap

import (
	"context"
	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/controller"
	"singkatin-api/upload/internal/infrastructure"
	"singkatin-api/upload/internal/repository"
	"singkatin-api/upload/internal/service"
	"singkatin-api/upload/pkg/logger"
	"time"
)

// Container ...
type Container struct {
	Context  context.Context
	Config   *config.Config
	RabbitMQ *infrastructure.RabbitMQConnectionProvider
	Tracer   *infrastructure.TracerProvider
	MinIO    *infrastructure.MinIOConnectionProvider

	UploadController controller.UploadController
}

// NewContainer configuring dependencies app needed
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Load()

	rabbitmq := infrastructure.NewRabbitMQConnection(ctx, cfg)
	tracer := infrastructure.NewTracerProvider(ctx, cfg)
	minioClient := infrastructure.NewMinIOConnection(ctx, cfg)

	// repository
	uploadRepo := repository.NewUploadRepository(cfg, tracer.Tracer("Upload Repository"), minioClient.GetClient())

	// service
	uploadSvc := service.NewUploadService(cfg, tracer.Tracer("Upload Service"), uploadRepo)

	// controller
	uploadController := controller.NewUploadController(cfg, tracer.Tracer("Upload Controller"), uploadSvc)

	logger.Info("UPLOAD SERVICE RUN SUCCESSFULLY")

	return &Container{
		Context:  ctx,
		Config:   cfg,
		RabbitMQ: rabbitmq,
		Tracer:   tracer,
		MinIO:    minioClient,

		UploadController: uploadController,
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

		// RABBITMQ
		if c.RabbitMQ != nil && c.RabbitMQ.GetClient() != nil {
			if err := c.RabbitMQ.Close(); err != nil {
				panic(err)
			}
		}

	}(ctx)
}
