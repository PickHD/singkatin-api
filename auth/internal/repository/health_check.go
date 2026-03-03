package repository

import (
	"context"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/pkg/logger"

	"github.com/redis/go-redis/v9"
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
		Redis  *redis.Client
	}
)

// NewHealthCheckRepository return new instances health check repository
func NewHealthCheckRepository(config *config.Config, tracer *trace.TracerProvider, db *mongo.Database, rds *redis.Client) HealthCheckRepository {
	return &healthCheckRepositoryImpl{
		Config: config,
		Tracer: tracer,
		DB:     db,
		Redis:  rds,
	}
}

func (r *healthCheckRepositoryImpl) Check(ctx context.Context) (bool, error) {
	tr := r.Tracer.Tracer("Auth-Check Repository")
	_, span := tr.Start(ctx, "Start Check")
	defer span.End()

	if err := r.DB.Client().Ping(ctx, nil); err != nil {
		logger.Errorf("HealthCheckRepositoryImpl.Check() Ping DB ERROR, %v", err)
		return false, nil
	}

	if err := r.Redis.Ping(ctx).Err(); err != nil {
		logger.Errorf("HealthCheckRepositoryImpl.Check() Ping Redis ERROR, %v", err)
		return false, nil
	}

	return true, nil
}
