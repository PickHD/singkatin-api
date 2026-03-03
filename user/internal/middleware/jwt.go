package middleware

import (
	"net/http"
	"strings"

	"singkatin-api/user/internal/infrastructure"
	"singkatin-api/user/internal/model"
	"singkatin-api/user/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	jwtProvider *infrastructure.JwtProvider
}

func NewAuthMiddleware(jwtProvider *infrastructure.JwtProvider) *AuthMiddleware {
	return &AuthMiddleware{
		jwtProvider: jwtProvider,
	}
}

func (m *AuthMiddleware) VerifyToken(ctx *fiber.Ctx) error {
	var tokenString string

	authHeader := ctx.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	if tokenString == "" {
		tokenString = ctx.Query("token")
	}

	if tokenString == "" {
		return response.NewResponses[any](ctx, http.StatusUnauthorized, "missing or invalid authentication token", nil, nil, nil)
	}

	claims, err := m.jwtProvider.ValidateToken(tokenString)
	if err != nil {
		return response.NewResponses[any](ctx, http.StatusUnauthorized, "invalid or expired token", nil, err, nil)
	}

	ctx.Locals(model.KeyJWTValidAccess, claims)

	return ctx.Next()
}
