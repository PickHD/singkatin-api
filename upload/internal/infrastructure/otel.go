package infrastructure

import (
	"context"
	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type TracerProvider struct {
	tracer *trace.TracerProvider
}

func NewTracerProvider(ctx context.Context, cfg *config.Config) *TracerProvider {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Tracer.JaegerURL)))
	if err != nil {
		logger.Errorf("failed init Jaeger Tracer: %v", err)
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
	return &TracerProvider{tracer: tp}
}

func (t *TracerProvider) SetTracerProvider() {
	otel.SetTracerProvider(t.tracer)
}

func (t *TracerProvider) Shutdown(ctx context.Context) error {
	return t.tracer.Shutdown(ctx)
}

func (t *TracerProvider) Tracer(name string) *trace.TracerProvider {
	return t.tracer
}
