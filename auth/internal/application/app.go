package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/PickHD/singkatin-revamp/auth/internal/config"
	"github.com/PickHD/singkatin-revamp/auth/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gopkg.in/gomail.v2"
)

// App ...
type App struct {
	Application *gin.Engine
	Context     context.Context
	Config      *config.Configuration
	Logger      *logrus.Logger
	DB          *mongo.Database
	Redis       *redis.Client
	Tracer      *trace.TracerProvider
	Mailer      *gomail.Dialer
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

	// initialize mailer
	app.Logger.Info(app.Config.Mailer)
	app.Mailer = initSMTPMailDialer(app.Config)

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", app.Config.Database.Host, app.Config.Database.Port)))
	if err != nil {
		app.Logger.Error("failed connect mongoDB, error :", err)
		return app, err
	}

	db := mongoClient.Database(app.Config.Database.Name)
	app.DB = db

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", app.Config.Redis.Host, app.Config.Redis.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	app.Redis = redisClient

	app.Application = gin.New()
	app.Application.Use(middleware.CORSMiddleware())

	app.Logger.Info("APP RUN SUCCESSFULLY")

	return app, nil
}

// Close method will close any instances before app terminated
func (a *App) Close(ctx context.Context) {
	a.Logger.Info("APP CLOSED SUCCESSFULLY")

	defer func(ctx context.Context) {
		// DB
		if err := a.DB.Client().Disconnect(ctx); err != nil {
			panic(err)
		}

		// TRACER
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

// initSMTPMailDialer returns an Dialer configured to use
func initSMTPMailDialer(cfg *config.Configuration) *gomail.Dialer {
	d := gomail.NewDialer(cfg.Mailer.Host, cfg.Mailer.Port, cfg.Mailer.Username, cfg.Mailer.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d
}
