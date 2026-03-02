package controller

import (
	"context"
	"net/http"

	"singkatin-api/user/pkg/response"

	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckController is an interface that has all the function to be implemented inside health check controller
	HealthCheckController interface {
		Check(ctx *fiber.Ctx) error
	}

	// HealthCheckControllerImpl is an app health check struct that consists of all the dependencies needed for health check controller
	healthCheckControllerImpl struct {
		Context        context.Context
		Config         *config.Config
		Tracer         *trace.TracerProvider
		HealthCheckSvc service.HealthCheckService
	}
)

// NewHealthCheckController return new instances health check controller
func NewHealthCheckController(ctx context.Context, config *config.Config, tracer *trace.TracerProvider, healthCheckSvc service.HealthCheckService) HealthCheckController {
	return &healthCheckControllerImpl{
		Context:        ctx,
		Config:         config,
		Tracer:         tracer,
		HealthCheckSvc: healthCheckSvc,
	}
}

func (c *healthCheckControllerImpl) Check(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-Check Controller")
	_, span := tr.Start(c.Context, "Start Check")
	defer span.End()

	ok, err := c.HealthCheckSvc.Check()
	if err != nil || !ok {
		return response.NewResponses[any](ctx, http.StatusInternalServerError, "not OK", ok, err, nil)
	}

	return response.NewResponses[any](ctx, http.StatusOK, "OK", ok, nil, nil)
}
