package service

import (
	"context"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/repository"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckService is an interface that has all the function to be implemented inside health check service
	HealthCheckService interface {
		Check() (bool, error)
	}

	// healthCheckServiceImpl is an app health check struct that consists of all the dependencies needed for health check service
	healthCheckServiceImpl struct {
		Context         context.Context
		Config          *config.Configuration
		Tracer          *trace.TracerProvider
		HealthCheckRepo repository.HealthCheckRepository
	}
)

// NewHealthCheckService return new instances health check service
func NewHealthCheckService(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, healthCheckRepo repository.HealthCheckRepository) HealthCheckService {
	return &healthCheckServiceImpl{
		Context:         ctx,
		Config:          config,
		Tracer:          tracer,
		HealthCheckRepo: healthCheckRepo,
	}
}

func (s *healthCheckServiceImpl) Check() (bool, error) {
	tr := s.Tracer.Tracer("Shortener-Check Service")
	_, span := tr.Start(s.Context, "Start Check")
	defer span.End()

	return s.HealthCheckRepo.Check()
}
