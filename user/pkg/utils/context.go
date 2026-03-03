package utils

import (
	"errors"
	"singkatin-api/user/internal/infrastructure"
	"singkatin-api/user/internal/model"

	"github.com/gofiber/fiber/v2"
)

func GetUserContext(ctx *fiber.Ctx) (*infrastructure.MyClaims, error) {
	userContext := ctx.Locals(model.KeyJWTValidAccess)
	if claims, ok := userContext.(*infrastructure.MyClaims); ok {
		return claims, nil
	}
	return nil, errors.New("failed to get user from context")
}
