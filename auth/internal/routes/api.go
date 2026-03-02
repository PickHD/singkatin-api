package routes

import (
	"singkatin-api/auth/internal/bootstrap"
	"singkatin-api/auth/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	container *bootstrap.Container
	app       *gin.Engine
}

func newRouter(container *bootstrap.Container) *Router {
	app := gin.Default()

	return &Router{
		container: container,
		app:       app,
	}
}

func (r *Router) setupMiddleware() {
	// CORS
	r.app.Use(middleware.CORSMiddleware())

	// Recovery
	r.app.Use(gin.Recovery())
}

func (r *Router) setupRoutes() {
	v1 := r.app.Group("/v1")
	{
		v1.GET("/health-check", r.container.HealthCheckController.Check)

		v1.POST("/register", r.container.AuthController.Register)

		v1.GET("/register/verify", r.container.AuthController.VerifyRegister)

		v1.POST("/login", r.container.AuthController.Login)

		v1.POST("/forgot-password", r.container.AuthController.ForgotPassword)

		v1.GET("/forgot-password/verify", r.container.AuthController.VerifyForgotPassword)

		v1.PUT("/reset-password", r.container.AuthController.ResetPassword)
	}
}

func ServeHTTP(container *bootstrap.Container) *gin.Engine {
	router := newRouter(container)

	router.setupMiddleware()
	router.setupRoutes()

	return router.app
}
