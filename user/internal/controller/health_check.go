package controller

import (
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
		Config         *config.Config
		Tracer         *trace.TracerProvider
		HealthCheckSvc service.HealthCheckService
	}
)

// NewHealthCheckController return new instances health check controller
func NewHealthCheckController(config *config.Config, tracer *trace.TracerProvider, healthCheckSvc service.HealthCheckService) HealthCheckController {
	return &healthCheckControllerImpl{
		Config:         config,
		Tracer:         tracer,
		HealthCheckSvc: healthCheckSvc,
	}
}

func (c *healthCheckControllerImpl) Check(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-Check Controller")

	userCtxValue := ctx.UserContext()
	userCtxValue, span := tr.Start(userCtxValue, "Start Check")
	defer span.End()

	ok, err := c.HealthCheckSvc.Check(userCtxValue)
	if err != nil || !ok {
		return response.NewResponses[any](ctx, http.StatusInternalServerError, "not OK", ok, err, nil)
	}

	return response.NewResponses[any](ctx, http.StatusOK, "OK", ok, nil, nil)
}
