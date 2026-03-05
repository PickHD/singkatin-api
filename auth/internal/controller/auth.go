package controller

import (
	"net/http"
	"strings"

	"singkatin-api/auth/internal/config"
	"singkatin-api/auth/internal/dto/request"
	"singkatin-api/auth/internal/model"
	"singkatin-api/auth/internal/service"
	"singkatin-api/auth/pkg/response"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// Authcontroller is an interface that has all the function to be implemented inside auth controller
	AuthController interface {
		Register(ctx *gin.Context)
		Login(ctx *gin.Context)
		VerifyRegister(ctx *gin.Context)
		ForgotPassword(ctx *gin.Context)
		VerifyForgotPassword(ctx *gin.Context)
		ResetPassword(ctx *gin.Context)
	}

	// authControllerImpl is an app auth struct that consists of all the dependencies needed for auth controller
	authControllerImpl struct {
		Config  *config.Config
		Tracer  *trace.TracerProvider
		AuthSvc service.AuthService
	}
)

// NewAuthController return new instances auth controller
func NewAuthController(config *config.Config, tracer *trace.TracerProvider, authSvc service.AuthService) AuthController {
	return &authControllerImpl{
		Config:  config,
		Tracer:  tracer,
		AuthSvc: authSvc,
	}
}

func (c *authControllerImpl) Register(ctx *gin.Context) {
	var req request.RegisterRequest

	tr := c.Tracer.Tracer("Auth-Register Controller")
	_, span := tr.Start(ctx, "Start Register")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	data, err := c.AuthSvc.RegisterUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req.Email, err, nil)
			return
		}

		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed register user", data, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusCreated, "Success register, please check email for further verification", data, nil, nil)
}

func (c *authControllerImpl) Login(ctx *gin.Context) {
	var req request.LoginRequest

	tr := c.Tracer.Tracer("Auth-Login Controller")
	_, span := tr.Start(ctx, "Start Login")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	data, err := c.AuthSvc.LoginUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			response.NewResponses[any](ctx, http.StatusNotFound, err.Error(), req, err, nil)
			return
		}

		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed login user", data, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusOK, "Success login user", data, nil, nil)
}

func (c *authControllerImpl) VerifyRegister(ctx *gin.Context) {
	tr := c.Tracer.Tracer("Auth-VerifyRegister Controller")
	_, span := tr.Start(ctx, "Start VerifyRegister")
	defer span.End()

	getCode := ctx.Query("code")

	if getCode == "" {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	data, err := c.AuthSvc.VerifyCode(ctx, getCode, model.RegisterVerification)
	if err != nil {
		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed Verify Code", nil, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusOK, "Verify success, Redirecting to Login Page....", data, err, nil)
}

func (c *authControllerImpl) ForgotPassword(ctx *gin.Context) {
	var req request.ForgotPasswordRequest

	tr := c.Tracer.Tracer("Auth-ForgotPassword Controller")
	_, span := tr.Start(ctx, "Start ForgotPassword")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	err := c.AuthSvc.ForgotPasswordUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req.Email, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			response.NewResponses[any](ctx, http.StatusNotFound, err.Error(), req.Email, err, nil)
			return
		}

		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed request forgot password", nil, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusOK, "Success sent verification to your email, please check your email", nil, nil, nil)
}

func (c *authControllerImpl) VerifyForgotPassword(ctx *gin.Context) {
	tr := c.Tracer.Tracer("Auth-VerifyForgotPassword Controller")
	_, span := tr.Start(ctx, "Start VerifyForgotPassword")
	defer span.End()

	getCode := ctx.Query("code")

	if getCode == "" {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	_, err := c.AuthSvc.VerifyCode(ctx, getCode, model.ForgotPasswordVerification)
	if err != nil {
		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed Verify Code", nil, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusOK, "Verify success, Redirecting to Reset Password Page...", nil, err, nil)
}

func (c *authControllerImpl) ResetPassword(ctx *gin.Context) {
	tr := c.Tracer.Tracer("Auth-ResetPassword Controller")
	_, span := tr.Start(ctx, "Start ResetPassword")
	defer span.End()

	var req request.ResetPasswordRequest

	getCode := ctx.Query("code")

	if getCode == "" {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	if err := ctx.BindJSON(&req); err != nil {
		response.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	err := c.AuthSvc.ResetPasswordUser(ctx, &req, getCode)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			response.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), nil, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			response.NewResponses[any](ctx, http.StatusNotFound, err.Error(), nil, err, nil)
			return
		}

		response.NewResponses[any](ctx, http.StatusInternalServerError, "Failed reset password", nil, err, nil)
		return
	}

	response.NewResponses[any](ctx, http.StatusOK, "Success reset password", nil, nil, nil)
}
