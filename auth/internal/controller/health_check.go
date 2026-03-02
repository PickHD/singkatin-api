package controller

import (
	"context"
	"net/http"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/service"
	"singkatin-api/auth/pkg/response"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// HealthCheckController is an interface that has all the function to be implemented inside health check controller
	HealthCheckController interface {
		Check(ctx *gin.Context)
	}

	// healthCheckControllerImpl is an app health check struct that consists of all the dependencies needed for health check controller
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

func (c *healthCheckControllerImpl) Check(ctx *gin.Context) {
	tr := c.Tracer.Tracer("Auth-Check Controller")
	_, span := tr.Start(ctx, "Start Check")
	defer span.End()

	ok, err := c.HealthCheckSvc.Check()
	if err != nil || !ok {
		response.NewResponses[any](ctx, http.StatusInternalServerError, "Not OK", ok, err, nil)
	}

	response.NewResponses[any](ctx, http.StatusOK, "OK", ok, nil, nil)
}
