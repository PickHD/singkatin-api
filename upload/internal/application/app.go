package application

import (
	"context"
	"time"

	"github.com/PickHD/singkatin-revamp/upload/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// App ...
type App struct {
	Context  context.Context
	Config   *config.Configuration
	Logger   *logrus.Logger
	RabbitMQ *amqp.Channel
	Tracer   *trace.TracerProvider
	MinIO    *minio.Client
}

// SetupApplication configuring dependencies app needed
func SetupApplication(ctx context.Context) (*App, error) {
	var err error

	app := &App{}
	app.Context = context.TODO()
	app.Config = config.NewConfig()

	// custom log app with logrus
	logWithLogrus := logrus.New()
	logWithLogrus.Formatter = &logrus.JSONFormatter{}
	logWithLogrus.ReportCaller = true
	app.Logger = logWithLogrus

	// initialize tracers
	app.Tracer, err = initJaegerTracerProvider(app.Config)
	if err != nil {
		app.Logger.Error("failed init Jaeger Tracer", err)
		return app, nil
	}

	otel.SetTracerProvider(app.Tracer)

	amqpConn, err := amqp.Dial(app.Config.RabbitMQ.ConnURL)
	if err != nil {
		app.Logger.Error("failed dial RabbitMQ, error :", err)
		return app, err
	}

	amqpClient, err := amqpConn.Channel()
	if err != nil {
		app.Logger.Error("failed open RabbitMQ Channels, error :", err)
		return app, err
	}

	queues := []string{app.Config.RabbitMQ.QueueUploadAvatar}

	for _, q := range queues {
		_, err = amqpClient.QueueDeclare(
			q,     // queue name
			true,  // durable
			false, // auto delete
			false, // exclusive
			false, // no wait
			nil,   // arguments
		)
		if err != nil {
			app.Logger.Error("failed queue declare Channels, error :", err)
			return nil, err
		}
	}

	app.RabbitMQ = amqpClient

	// initialize minIO client
	minioClient, errInit := minio.New(app.Config.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(app.Config.MinIO.AccessKey, app.Config.MinIO.SecretKey, ""),
		Secure: app.Config.MinIO.UseSSL,
	})
	if errInit != nil {
		app.Logger.Error("failed minio.New, error :", err)
		return nil, err
	}

	err = minioClient.MakeBucket(ctx, app.Config.MinIO.Bucket, minio.MakeBucketOptions{Region: app.Config.MinIO.Location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, app.Config.MinIO.Bucket)
		if errBucketExists == nil && exists {
			app.Logger.Info("Bucket already exists.")
		} else {
			app.Logger.Error("failed check BucketExists", errBucketExists)
			return nil, err
		}
	}

	app.MinIO = minioClient

	app.Logger.Info("APP RUN SUCCESSFULLY")

	return app, nil
}

// Close method will close any instances before app terminated
func (a *App) Close(ctx context.Context) {
	a.Logger.Info("APP CLOSED SUCCESSFULLY")

	defer func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := a.Tracer.Shutdown(ctx); err != nil {
			panic(err)
		}
	}(ctx)
}

// initJaegerTracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func initJaegerTracerProvider(cfg *config.Configuration) (*trace.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Tracer.JaegerURL)))
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(exp),
		// Record information about this application in a Resource.
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.Server.AppName),
			attribute.String("environment", cfg.Server.AppEnv),
			attribute.String("ID", cfg.Server.AppID),
		)),
	)
	return tp, nil
}
