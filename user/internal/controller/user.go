package controller

import (
	"context"
	"strings"

	"singkatin-api/user/pkg/response"

	"singkatin-api/user/internal/config"
	"singkatin-api/user/internal/middleware"
	"singkatin-api/user/internal/model"
	"singkatin-api/user/internal/service"

	"github.com/gofiber/fiber/v2"
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

	// userControllerImpl is an app user struct that consists of all the dependencies needed for user controller
	userControllerImpl struct {
		Context context.Context
		Config  *config.Config
		Tracer  *trace.TracerProvider
		UserSvc service.UserService
	}
)

// NewUserController return new instances user controller
func NewUserController(ctx context.Context, config *config.Config, tracer *trace.TracerProvider, userSvc service.UserService) UserController {
	return &userControllerImpl{
		Context: ctx,
		Config:  config,
		Tracer:  tracer,
		UserSvc: userSvc,
	}
}

func (c *userControllerImpl) Profile(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-Profile Controller")
	_, span := tr.Start(c.Context, "Start Profile")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	detail, err := c.UserSvc.GetUserDetail(extData.Email)
	if err != nil {
		if strings.Contains(err.Error(), string(model.NotFound)) {
			return response.NewResponses[any](ctx, fiber.StatusNotFound, err.Error(), nil, err, nil)
		}

		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success get Profiles", detail, nil, nil)
}

func (c *userControllerImpl) Dashboard(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-Dashboard Controller")
	_, span := tr.Start(c.Context, "Start Dashboard")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	detail, err := c.UserSvc.GetUserShorts(extData.UserID)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success get Dashboard", detail, nil, nil)
}

func (c *userControllerImpl) GenerateShort(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-GenerateShort Controller")
	_, span := tr.Start(c.Context, "Start GenerateShort")
	defer span.End()

	var req model.ShortUserRequest

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	if err := ctx.BodyParser(&req); err != nil {
		return response.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	newShort, err := c.UserSvc.GenerateUserShorts(extData.UserID, &req)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusCreated, "Success generate Short URL's", newShort, nil, nil)
}

func (c *userControllerImpl) EditProfile(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-EditProfile Controller")
	_, span := tr.Start(c.Context, "Start EditProfile")
	defer span.End()

	var req model.EditProfileRequest

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	if err := ctx.BodyParser(&req); err != nil {
		return response.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	err = c.UserSvc.UpdateUserProfile(extData.UserID, &req)
	if err != nil {
		if strings.Contains(err.Error(), string(model.Validation)) {
			return response.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
		}

		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success update profile", nil, nil, nil)
}

func (c *userControllerImpl) UploadAvatar(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-UploadAvatar Controller")
	_, span := tr.Start(c.Context, "Start UploadAvatar")
	defer span.End()

	data := ctx.Locals(model.KeyJWTValidAccess)
	extData, err := middleware.Extract(data)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	resp, err := c.UserSvc.UploadUserAvatar(ctx, extData.UserID)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success upload avatar users", resp, nil, nil)
}

func (c *userControllerImpl) UpdateShort(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-UpdateShort Controller")
	_, span := tr.Start(c.Context, "Start UpdateShort")
	defer span.End()

	var req model.ShortUserRequest

	if err := ctx.BodyParser(&req); err != nil {
		return response.NewResponses[any](ctx, fiber.StatusBadRequest, err.Error(), nil, err, nil)
	}

	shortID := ctx.Params("id", "")
	if shortID == "" {
		return response.NewResponses[any](ctx, fiber.StatusBadRequest, "id required", model.NewError(model.Validation, "ID Required"), nil, nil)
	}

	_, err := c.UserSvc.UpdateUserShorts(shortID, &req)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success update Short URL's", nil, nil, nil)
}

func (c *userControllerImpl) DeleteShort(ctx *fiber.Ctx) error {
	tr := c.Tracer.Tracer("User-DeleteShort Controller")
	_, span := tr.Start(c.Context, "Start DeleteShort")
	defer span.End()

	shortID := ctx.Params("id", "")
	if shortID == "" {
		return response.NewResponses[any](ctx, fiber.StatusBadRequest, "id required", model.NewError(model.Validation, "ID Required"), nil, nil)
	}

	_, err := c.UserSvc.DeleteUserShorts(shortID)
	if err != nil {
		return response.NewResponses[any](ctx, fiber.StatusInternalServerError, err.Error(), nil, err, nil)
	}

	return response.NewResponses[any](ctx, fiber.StatusOK, "Success delete Short URL's", nil, nil, nil)
}
