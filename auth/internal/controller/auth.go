package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/PickHD/singkatin-revamp/auth/internal/config"
	"github.com/PickHD/singkatin-revamp/auth/internal/helper"
	"github.com/PickHD/singkatin-revamp/auth/internal/model"
	"github.com/PickHD/singkatin-revamp/auth/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

	// AuthcontrollerImpl is an app auth struct that consists of all the dependencies needed for auth controller
	AuthControllerImpl struct {
		Context context.Context
		Config  *config.Configuration
		Logger  *logrus.Logger
		Tracer  *trace.TracerProvider
		AuthSvc service.AuthService
	}
)

// NewAuthController return new instances auth controller
func NewAuthController(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, authSvc service.AuthService) *AuthControllerImpl {
	return &AuthControllerImpl{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		Tracer:  tracer,
		AuthSvc: authSvc,
	}
}

// Check godoc
// @Summary      Register users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user body model.RegisterRequest true "register user"
// @Success      201  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /register [post]
func (ac *AuthControllerImpl) Register(ctx *gin.Context) {
	var req model.RegisterRequest

	tr := ac.Tracer.Tracer("Auth-Register Controller")
	_, span := tr.Start(ctx, "Start Register")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	data, err := ac.AuthSvc.RegisterUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			helper.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req.Email, err, nil)
			return
		}

		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed register user", data, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusCreated, "Success register, please check email for further verification", data, nil, nil)
}

// Check godoc
// @Summary      Login users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user body model.LoginRequest true "login user"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /login [post]
func (ac *AuthControllerImpl) Login(ctx *gin.Context) {
	var req model.LoginRequest

	tr := ac.Tracer.Tracer("Auth-Login Controller")
	_, span := tr.Start(ctx, "Start Login")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	data, err := ac.AuthSvc.LoginUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			helper.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			helper.NewResponses[any](ctx, http.StatusNotFound, err.Error(), req, err, nil)
			return
		}

		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed login user", data, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusOK, "Success login user", data, nil, nil)
}

// Check godoc
// @Summary      Verify Register Users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        code  query   string  true  "Code Verification"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /register/verify [get]
func (ac *AuthControllerImpl) VerifyRegister(ctx *gin.Context) {
	tr := ac.Tracer.Tracer("Auth-VerifyRegister Controller")
	_, span := tr.Start(ctx, "Start VerifyRegister")
	defer span.End()

	getCode := ctx.Query("code")

	if getCode == "" {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	data, err := ac.AuthSvc.VerifyCode(ctx, getCode, model.RegisterVerification)
	if err != nil {
		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed Verify Code", nil, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusOK, "Verify success, Redirecting to Login Page....", data, err, nil)
}

// Check godoc
// @Summary      Forgot Password Users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        forgotPassword body model.ForgotPasswordRequest true "forgot password user"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /forgot-password [post]
func (ac *AuthControllerImpl) ForgotPassword(ctx *gin.Context) {
	var req model.ForgotPasswordRequest

	tr := ac.Tracer.Tracer("Auth-ForgotPassword Controller")
	_, span := tr.Start(ctx, "Start ForgotPassword")
	defer span.End()

	if err := ctx.BindJSON(&req); err != nil {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	err := ac.AuthSvc.ForgotPasswordUser(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			helper.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), req.Email, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			helper.NewResponses[any](ctx, http.StatusNotFound, err.Error(), req.Email, err, nil)
			return
		}

		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed request forgot password", nil, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusOK, "Success sent verification to your email, please check your email", nil, nil, nil)
}

// Check godoc
// @Summary      Verify Forgot Password Users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        code  query   string  true  "Code Verification"
// @Success      200  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /forgot-password/verify [get]
func (ac *AuthControllerImpl) VerifyForgotPassword(ctx *gin.Context) {
	tr := ac.Tracer.Tracer("Auth-VerifyForgotPassword Controller")
	_, span := tr.Start(ctx, "Start VerifyForgotPassword")
	defer span.End()

	getCode := ctx.Query("code")

	if getCode == "" {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	_, err := ac.AuthSvc.VerifyCode(ctx, getCode, model.ForgotPasswordVerification)
	if err != nil {
		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed Verify Code", nil, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusOK, "Verify success, Redirecting to Reset Password Page...", nil, err, nil)
}

// Check godoc
// @Summary      Reset Password Users
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        code  query   string  true  "Code Verification"
// @Param        forgotPassword body model.ResetPasswordRequest true "reset password user"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /reset-password [put]
func (ac *AuthControllerImpl) ResetPassword(ctx *gin.Context) {
	tr := ac.Tracer.Tracer("Auth-ResetPassword Controller")
	_, span := tr.Start(ctx, "Start ResetPassword")
	defer span.End()

	var req model.ResetPasswordRequest

	getCode := ctx.Query("code")

	if getCode == "" {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Code Required", nil, nil, nil)
		return
	}

	if err := ctx.BindJSON(&req); err != nil {
		helper.NewResponses[any](ctx, http.StatusBadRequest, "Invalid request", req, err, nil)
		return
	}

	err := ac.AuthSvc.ResetPasswordUser(ctx, &req, getCode)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			helper.NewResponses[any](ctx, http.StatusBadRequest, err.Error(), nil, err, nil)
			return
		}

		if strings.Contains(err.Error(), string(model.NotFound)) {
			helper.NewResponses[any](ctx, http.StatusNotFound, err.Error(), nil, err, nil)
			return
		}

		helper.NewResponses[any](ctx, http.StatusInternalServerError, "Failed reset password", nil, err, nil)
		return
	}

	helper.NewResponses[any](ctx, http.StatusOK, "Success reset password", nil, nil, nil)
}
