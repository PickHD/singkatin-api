package controller

import (
	"context"
	"strings"

	"github.com/PickHD/singkatin-revamp/user/internal/config"
	"github.com/PickHD/singkatin-revamp/user/internal/helper"
	"github.com/PickHD/singkatin-revamp/user/internal/middleware"
	"github.com/PickHD/singkatin-revamp/user/internal/model"
	"github.com/PickHD/singkatin-revamp/user/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/trace"
)

type (
	// UserController is an interface that has all the function to be implemented inside user controller
	UserController interface {
		Profile(ctx *fiber.Ctx) error
		Dashboard(ctx *fiber.Ctx) error
		GenerateShort(ctx *fiber.Ctx) error
		EditProfile(ctx *fiber.Ctx) error
		UploadAvatar(ctx *fiber.Ctx) error
		UpdateShort(ctx *fiber.Ctx) error
		DeleteShort(ctx *fiber.Ctx) error
	}

	// UserControllerImpl is an app user struct that consists of all the dependencies needed for user controller
	UserControllerImpl struct {
		Context context.Context
		Config  *config.Configuration
		Logger  *logrus.Logger
		Tracer  *trace.TracerProvider
		UserSvc service.UserService
	}
)

// NewUserController return new instances user controller
func NewUserController(ctx context.Context, config *config.Configuration, logger *logrus.Logger, tracer *trace.TracerProvider, userSvc service.UserService) *UserControllerImpl {
	return &UserControllerImpl{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		Tracer:  tracer,
		UserSvc: userSvc,
	}
}

// Check godoc
// @Summary      Get Profiles
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Success      200  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /me [get]
func (uc *UserControllerImpl) Profile(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-Profile Controller")
	_, span := tr.Start(uc.Context, "Start Profile")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	detail, err := uc.UserSvc.GetUserDetail(extData.Email)
	if err != nil {
		if strings.Contains(err.Error(), string(model.NotFound)) {
			return helper.NewResponses[any](ctx, fiber.StatusNotFound, err.Error(), nil, err, nil)
		}

		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success get Profiles", detail, nil, nil)
}

// Check godoc
// @Summary      Get Dashboard
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Success      200  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /dashboard [get]
func (uc *UserControllerImpl) Dashboard(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-Dashboard Controller")
	_, span := tr.Start(uc.Context, "Start Dashboard")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	detail, err := uc.UserSvc.GetUserShorts(extData.UserID)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success get Dashboard", detail, nil, nil)
}

// Check godoc
// @Summary      Generate Users Short URL
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Param        short body model.ShortUserRequest true "generate short user"
// @Success      201  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /short/generate [post]
func (uc *UserControllerImpl) GenerateShort(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-GenerateShort Controller")
	_, span := tr.Start(uc.Context, "Start GenerateShort")
	defer span.End()

	var req model.ShortUserRequest

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	if err := ctx.BodyParser(&req); err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	newShort, err := uc.UserSvc.GenerateUserShorts(extData.UserID, &req)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusCreated, "Success generate Short URL's", newShort, nil, nil)
}

// Check godoc
// @Summary      Update Users Profile
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Param        profile body model.EditProfileRequest true "generate short user"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /me/edit [put]
func (uc *UserControllerImpl) EditProfile(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-EditProfile Controller")
	_, span := tr.Start(uc.Context, "Start EditProfile")
	defer span.End()

	var req model.EditProfileRequest

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	if err := ctx.BodyParser(&req); err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	err = uc.UserSvc.UpdateUserProfile(extData.UserID, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			return helper.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
		}

		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success update profile", nil, nil, nil)
}

// Check godoc
// @Summary      Upload Users Avatar
// @Tags         User
// @Accept       mpfd
// @Produce      json
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Param        file formData file true "file avatar"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /upload/avatar [post]
func (uc *UserControllerImpl) UploadAvatar(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-UploadAvatar Controller")
	_, span := tr.Start(uc.Context, "Start UploadAvatar")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	resp, err := uc.UserSvc.UploadUserAvatar(ctx, extData.UserID)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success upload avatar users", resp, nil, nil)
}

// Check godoc
// @Summary      Update Users Short URL
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "id short urls"
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Param        short body model.ShortUserRequest true "update short user"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /short/{id} [put]
func (uc *UserControllerImpl) UpdateShort(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-UpdateShort Controller")
	_, span := tr.Start(uc.Context, "Start UpdateShort")
	defer span.End()

	var req model.ShortUserRequest

	if err := ctx.BodyParser(&req); err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	shortID := ctx.Params("id", "")
	if shortID == "" {
		return helper.NewResponses[any](ctx, fiber.StatusBadRequest, "id required", model.NewError(model.Validation, "ID Required"), nil, nil)
	}

	_, err := uc.UserSvc.UpdateUserShorts(shortID, &req)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success update Short URL's", nil, nil, nil)
}

// Check godoc
// @Summary      Delete Users Short URL
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id   path string  true  "id short urls"
// @Param        Authorization header string true "Authorization Bearer <Place Access Token Here>"
// @Success      200  {object}  helper.BaseResponse
// @Failure      400  {object}  helper.BaseResponse
// @Failure      404  {object}  helper.BaseResponse
// @Failure      500  {object}  helper.BaseResponse
// @Router       /short/{id} [delete]
func (uc *UserControllerImpl) DeleteShort(ctx *fiber.Ctx) error {
	tr := uc.Tracer.Tracer("User-DeleteShort Controller")
	_, span := tr.Start(uc.Context, "Start DeleteShort")
	defer span.End()

	shortID := ctx.Params("id", "")
	if shortID == "" {
		return helper.NewResponses[any](ctx, fiber.StatusBadRequest, "id required", model.NewError(model.Validation, "ID Required"), nil, nil)
	}

	_, err := uc.UserSvc.DeleteUserShorts(shortID)
	if err != nil {
		return helper.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return helper.NewResponses[any](ctx, fiber.StatusOK, "Success delete Short URL's", nil, nil, nil)
}
