package repository

import (
	"context"

	"singkatin-api/shortener/pkg/logger"
	"singkatin-api/user/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckRepository is an interface that has all the function to be implemented inside health check repository
	HealthCheckRepository interface {
		Check() (bool, error)
	}

	// healthCheckRepositoryImpl is an app health check struct that consists of all the dependencies needed for health check repository
	healthCheckRepositoryImpl struct {
		Context context.Context
		Config  *config.Config
		Tracer  *trace.TracerProvider
		DB      *mongo.Database
	}
)

// NewHealthCheckRepository return new instances health check repository
func NewHealthCheckRepository(ctx context.Context, config *config.Config, tracer *trace.TracerProvider, db *mongo.Database) HealthCheckRepository {
	return &healthCheckRepositoryImpl{
		Context: ctx,
		Config:  config,
		Tracer:  tracer,
		DB:      db,
	}
}

func (h *healthCheckRepositoryImpl) Check() (bool, error) {
	tr := h.Tracer.Tracer("User-Check Repository")
	_, span := tr.Start(h.Context, "Start Check")
	defer span.End()

	if err := h.DB.Client().Ping(h.Context, nil); err != nil {
		logger.Error("HealthCheckRepositoryImpl.Check() Ping DB ERROR, ", err)
		return false, nil
	}
	return true, nil
}
