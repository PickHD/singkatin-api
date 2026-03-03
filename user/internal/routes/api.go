package routes

import (
	"singkatin-api/user/internal/bootstrap"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	container *bootstrap.Container
	app       *fiber.App
}

func newRouter(container *bootstrap.Container) *Router {
	app := fiber.New()

	return &Router{
		container: container,
		app:       app,
	}
}

func (r *Router) setupMiddleware() {
	// CORS
	r.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "*",
	}))

	// Recovery
	r.app.Use(recover.New())
}

func (r *Router) setupRoutes() {
	v1 := r.app.Group("/v1")
	{
		v1.Get("/health-check", r.container.HealthCheckController.Check)

		v1.Get("/me", r.container.AuthMiddleware.VerifyToken,r.container.UserController.Profile)

		v1.Put("/me/edit", r.container.AuthMiddleware.VerifyToken, r.container.UserController.EditProfile)

		v1.Get("/dashboard", r.container.AuthMiddleware.VerifyToken, r.container.UserController.Dashboard)

		v1.Post("/short/generate", r.container.AuthMiddleware.VerifyToken, r.container.UserController.GenerateShort)

		v1.Post("/upload/avatar", r.container.AuthMiddleware.VerifyToken, r.container.UserController.UploadAvatar)

		v1.Put("/short/:id", r.container.AuthMiddleware.VerifyToken, r.container.UserController.UpdateShort)

		v1.Delete("/short/:id", r.container.AuthMiddleware.VerifyToken, r.container.UserController.DeleteShort)
	}
}

func ServeHTTP(container *bootstrap.Container) *fiber.App {
	router := newRouter(container)

	router.setupMiddleware()
	router.setupRoutes()

	return router.app
}
