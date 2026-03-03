package service

import (
	"context"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/repository"

	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckService is an interface that has all the function to be implemented inside health check service
	HealthCheckService interface {
		Check(ctx context.Context) (bool, error)
	}

	// healthCheckServiceImpl is an app health check struct that consists of all the dependencies needed for health check service
	healthCheckServiceImpl struct {
		Config          *config.Config
		Tracer          *trace.TracerProvider
		HealthCheckRepo repository.HealthCheckRepository
	}
)

// NewHealthCheckService return new instances health check service
func NewHealthCheckService(config *config.Config, tracer *trace.TracerProvider, healthCheckRepo repository.HealthCheckRepository) HealthCheckService {
	return &healthCheckServiceImpl{
		Config:          config,
		Tracer:          tracer,
		HealthCheckRepo: healthCheckRepo,
	}
}

func (s *healthCheckServiceImpl) Check(ctx context.Context) (bool, error) {
	tr := s.Tracer.Tracer("Auth-Check Service")
	_, span := tr.Start(ctx, "Start Check")
	defer span.End()

	return s.HealthCheckRepo.Check(ctx)
}
