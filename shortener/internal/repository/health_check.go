package repository

import (
	"context"

	"github.com/PickHD/singkatin-revamp/shortener/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckRepository is an interface that has all the function to be implemented inside health check repository
	HealthCheckRepository interface {
		Check() (bool, error)
	}

	// HealthCheckRepositoryImpl is an app health check struct that consists of all the dependencies needed for health check repository
	HealthCheckRepositoryImpl struct {
		Context context.Context
		Config  *config.Configuration
		Logger  *logrus.Logger
		Tracer  *trace.TracerProvider
		DB      *mongo.Database
		Redis   *redis.Client
	}
)

// NewHealthCheckRepository return new instances health check repository
func NewHealthCheckRepository(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, db *mongo.Database, redis *redis.Client) *HealthCheckRepositoryImpl {
	return &HealthCheckRepositoryImpl{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		Tracer:  tracer,
		DB:      db,
		Redis:   redis,
	}
}

func (hr *HealthCheckRepositoryImpl) Check() (bool, error) {
	tr := hr.Tracer.Tracer("Shortener-Check Repository")
	_, span := tr.Start(hr.Context, "Start Check")
	defer span.End()

	if err := hr.DB.Client().Ping(hr.Context, nil); err != nil {
		hr.Logger.Error("HealthCheckRepositoryImpl.Check() Ping DB ERROR, ", err)
		return false, nil
	}

	if err := hr.Redis.Ping(hr.Context).Err(); err != nil {
		hr.Logger.Error("HealthCheckRepositoryImpl.Check() Ping Redis ERROR, ", err)
		return false, nil
	}

	return true, nil
}
