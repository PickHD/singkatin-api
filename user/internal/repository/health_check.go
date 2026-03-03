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
		Check(ctx context.Context) (bool, error)
	}

	// healthCheckRepositoryImpl is an app health check struct that consists of all the dependencies needed for health check repository
	healthCheckRepositoryImpl struct {
		Config *config.Config
		Tracer *trace.TracerProvider
		DB     *mongo.Database
	}
)

// NewHealthCheckRepository return new instances health check repository
func NewHealthCheckRepository(config *config.Config, tracer *trace.TracerProvider, db *mongo.Database) HealthCheckRepository {
	return &healthCheckRepositoryImpl{
		Config: config,
		Tracer: tracer,
		DB:     db,
	}
}

func (h *healthCheckRepositoryImpl) Check(ctx context.Context) (bool, error) {
	tr := h.Tracer.Tracer("User-Check Repository")
	_, span := tr.Start(ctx, "Start Check")
	defer span.End()

	if err := h.DB.Client().Ping(ctx, nil); err != nil {
		logger.Error("HealthCheckRepositoryImpl.Check() Ping DB ERROR, ", err)
		return false, nil
	}
	return true, nil
}
