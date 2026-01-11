package infrastructure

import (
	"github.com/PickHD/singkatin-revamp/shortener/internal/application"
	"github.com/labstack/echo/v4"

	echoSwagger "github.com/swaggo/echo-swagger"
)

// ServeHTTP is wrapper function to start the apps infra in HTTP mode
func ServeHTTP(app *application.App) *echo.Echo {
	// call setup router
	setupRouter(app)

	return app.Application
}

// setupRouter is function to manage all routings
func setupRouter(app *application.App) {
	var dep = application.SetupDependencyInjection(app)

	v1 := app.Application.Group("/v1")
	{
		v1.GET("/swagger/*any", echoSwagger.WrapHandler)

		v1.GET("/health-check", dep.HealthCheckController.Check)

		v1.GET("/:short_url", dep.ShortController.ClickShortener)
	}

}
