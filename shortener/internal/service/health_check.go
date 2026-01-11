package service

import (
	"context"

	"github.com/PickHD/singkatin-revamp/shortener/internal/config"
	"github.com/PickHD/singkatin-revamp/shortener/internal/repository"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckService is an interface that has all the function to be implemented inside health check service
	HealthCheckService interface {
		Check() (bool, error)
	}

	// HealthCheckServiceImpl is an app health check struct that consists of all the dependencies needed for health check service
	HealthCheckServiceImpl struct {
		Context         context.Context
		Config          *config.Configuration
		Tracer          *trace.TracerProvider
		HealthCheckRepo repository.HealthCheckRepository
	}
)

// NewHealthCheckService return new instances health check service
func NewHealthCheckService(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, healthCheckRepo repository.HealthCheckRepository) *HealthCheckServiceImpl {
	return &HealthCheckServiceImpl{
		Context:         ctx,
		Config:          config,
		Tracer:          tracer,
		HealthCheckRepo: healthCheckRepo,
	}
}

func (hs *HealthCheckServiceImpl) Check() (bool, error) {
	tr := hs.Tracer.Tracer("Shortener-Check Service")
	_, span := tr.Start(hs.Context, "Start Check")
	defer span.End()

	return hs.HealthCheckRepo.Check()
}
