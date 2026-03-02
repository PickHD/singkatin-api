package repository

import (
	"context"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/pkg/logger"

	"github.com/redis/go-redis/v9"
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
		Config  *config.Configuration
		Tracer  *trace.TracerProvider
		DB      *mongo.Database
		Redis   *redis.Client
	}
)

// NewHealthCheckRepository return new instances health check repository
func NewHealthCheckRepository(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, db *mongo.Database, redis *redis.Client) HealthCheckRepository {
	return &healthCheckRepositoryImpl{
		Context: ctx,
		Config:  config,
		Tracer:  tracer,
		DB:      db,
		Redis:   redis,
	}
}

func (r *healthCheckRepositoryImpl) Check() (bool, error) {
	tr := r.Tracer.Tracer("Shortener-Check Repository")
	_, span := tr.Start(r.Context, "Start Check")
	defer span.End()

	if err := r.DB.Client().Ping(r.Context, nil); err != nil {
		logger.Error("HealthCheckRepositoryImpl.Check() Ping DB ERROR, ", err)
		return false, nil
	}

	if err := r.Redis.Ping(r.Context).Err(); err != nil {
		logger.Error("HealthCheckRepositoryImpl.Check() Ping Redis ERROR, ", err)
		return false, nil
	}

	return true, nil
}
