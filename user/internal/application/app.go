package application

import (
	"context"
	"fmt"
	"time"

	"github.com/PickHD/singkatin-revamp/user/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// App ...
type App struct {
	Application *fiber.App
	Context     context.Context
	Config      *config.Configuration
	Logger      *logrus.Logger
	DB          *mongo.Database
	RabbitMQ    *amqp.Channel
	GRPC        *grpc.ClientConn
	Tracer      *trace.TracerProvider
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

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", app.Config.Database.Host, app.Config.Database.Port)))
	if err != nil {
		app.Logger.Error("failed connect mongoDB, error :", err)
		return app, err
	}
	db := mongoClient.Database(app.Config.Database.Name)
	app.DB = db

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

	queues := []string{app.Config.RabbitMQ.QueueCreateShortener, app.Config.RabbitMQ.QueueUpdateVisitor, app.Config.RabbitMQ.QueueUpdateShortener, app.Config.RabbitMQ.QueueDeleteShortener}

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

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	app.GRPC, err = grpc.Dial(app.Config.Common.GRPCPort, opts...)
	if err != nil {
		app.Logger.Error("failed Dial GRPC, error :", err)
		return app, err
	}

	app.Application = fiber.New()
	app.Application.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Logger.Info("APP RUN SUCCESSFULLY")

	return app, nil
}

// Close method will close any instances before app terminated
func (a *App) Close(ctx context.Context) {
	a.Logger.Info("APP CLOSED SUCCESSFULLY")

	defer func(ctx context.Context) {
		if err := a.DB.Client().Disconnect(a.Context); err != nil {
			panic(err)
		}

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
