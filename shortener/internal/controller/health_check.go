package controller

import (
	"context"
	"net/http"

	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/service"
	"singkatin-api/shortener/pkg/response"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckController is an interface that has all the function to be implemented inside health check controller
	HealthCheckController interface {
		Check(ctx echo.Context) error
	}

	// healthCheckControllerImpl is an app health check struct that consists of all the dependencies needed for health check controller
	healthCheckControllerImpl struct {
		Context        context.Context
		Config         *config.Configuration
		Tracer         *trace.TracerProvider
		HealthCheckSvc service.HealthCheckService
	}
)

// NewHealthCheckController return new instances health check controller
func NewHealthCheckController(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, healthCheckSvc service.HealthCheckService) HealthCheckController {
	return &healthCheckControllerImpl{
		Context:        ctx,
		Config:         config,
		Tracer:         tracer,
		HealthCheckSvc: healthCheckSvc,
	}
}

func (c *healthCheckControllerImpl) Check(ctx echo.Context) error {
	tr := c.Tracer.Tracer("Shortener-Check Controller")
	_, span := tr.Start(c.Context, "Start Check")
	defer span.End()

	ok, err := c.HealthCheckSvc.Check()
	if err != nil || !ok {
		return response.NewResponses[any](ctx, http.StatusInternalServerError, "not OK", ok, err, nil)
	}

	return response.NewResponses[any](ctx, http.StatusOK, "OK", ok, nil, nil)
}
