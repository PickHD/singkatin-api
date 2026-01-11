package infrastructure

import (
	"github.com/PickHD/singkatin-revamp/user/internal/application"
	"github.com/PickHD/singkatin-revamp/user/internal/helper"
	"github.com/PickHD/singkatin-revamp/user/internal/middleware"
	"github.com/gofiber/fiber/v2"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// ServeHTTP is wrapper function to start the apps infra in HTTP mode
func ServeHTTP(app *application.App) *fiber.App {
	// call setup router
	setupRouter(app)

	return app.Application
}

// setupRouter is function to manage all routings
func setupRouter(app *application.App) {
	var dep = application.SetupDependencyInjection(app)

	v1 := app.Application.Group("/v1")
	{
		v1.Get("/swagger/*any", fiberSwagger.WrapHandler)

		v1.Get("/health-check", dep.HealthCheckController.Check)

		v1.Get("/me", middleware.ValidateJWTMiddleware, dep.UserController.Profile)

		v1.Put("/me/edit", middleware.ValidateJWTMiddleware, dep.UserController.EditProfile)

		v1.Get("/dashboard", middleware.ValidateJWTMiddleware, dep.UserController.Dashboard)

		v1.Post("/short/generate", middleware.ValidateJWTMiddleware, dep.UserController.GenerateShort)

		v1.Post("/upload/avatar", middleware.ValidateJWTMiddleware, dep.UserController.UploadAvatar)

		v1.Put("/short/:id", middleware.ValidateJWTMiddleware, dep.UserController.UpdateShort)

		v1.Delete("/short/:id", middleware.ValidateJWTMiddleware, dep.UserController.DeleteShort)
	}

	// handler for route not found
	app.Application.Use(func(c *fiber.Ctx) error {
		return helper.NewResponses[any](c, fiber.StatusNotFound, "Route not found", nil, nil, nil)
	})

}
