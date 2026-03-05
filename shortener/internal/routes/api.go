package routes

import (
	"singkatin-api/shortener/internal/bootstrap"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Router struct {
	container *bootstrap.Container
	app       *echo.Echo
}

func newRouter(container *bootstrap.Container) *Router {
	app := echo.New()

	return &Router{
		container: container,
		app:       app,
	}
}

func (r *Router) setupMiddleware() {
	// CORS
	r.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Recovery
	r.app.Use(middleware.Recover())
}

func (r *Router) setupRoutes() {
	v1 := r.app.Group("/v1")
	{
		v1.GET("/health-check", r.container.HealthCheckController.Check)
		v1.GET("/:short_url", r.container.ShortController.ClickShortener)
		v1.GET("/:short_url/qr", r.container.ShortController.GenerateQRCode)
	}
}

func ServeHTTP(container *bootstrap.Container) *echo.Echo {
	router := newRouter(container)

	router.setupMiddleware()
	router.setupRoutes()

	return router.app
}
