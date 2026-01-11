package controller

import (
	"context"
	"net/http"

	"github.com/PickHD/singkatin-revamp/user/internal/config"
	"github.com/PickHD/singkatin-revamp/user/internal/helper"
	"github.com/PickHD/singkatin-revamp/user/internal/service"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckController is an interface that has all the function to be implemented inside health check controller
	HealthCheckController interface {
		Check(ctx *fiber.Ctx) error
	}

	// HealthCheckControllerImpl is an app health check struct that consists of all the dependencies needed for health check controller
	HealthCheckControllerImpl struct {
		Context        context.Context
		Config         *config.Configuration
		Tracer         *trace.TracerProvider
		HealthCheckSvc service.HealthCheckService
	}
)

// NewHealthCheckController return new instances health check controller
func NewHealthCheckController(ctx context.Context, config *config.Configuration, tracer *trace.TracerProvider, healthCheckSvc service.HealthCheckService) *HealthCheckControllerImpl {
	return &HealthCheckControllerImpl{
		Context:        ctx,
		Config:         config,
		Tracer:         tracer,
		HealthCheckSvc: healthCheckSvc,
	}
}

// Check godoc
// @Summary      Checking Health Services
// @Tags         Health Check
// @Accept       json
// @Produce      json
// @Success      200  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /health-check [get]
func (hc *HealthCheckControllerImpl) Check(ctx *fiber.Ctx) error {
	tr := hc.Tracer.Tracer("User-Check Controller")
	_, span := tr.Start(hc.Context, "Start Check")
	defer span.End()

	ok, err := hc.HealthCheckSvc.Check()
	if err != nil || !ok {
		return helper.NewResponses[any](ctx, http.StatusInternalServerError, "not OK", ok, err, nil)
	}

	return helper.NewResponses[any](ctx, http.StatusOK, "OK", ok, nil, nil)
}
