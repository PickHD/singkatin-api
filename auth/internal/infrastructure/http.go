package infrastructure

import (
	"github.com/PickHD/singkatin-revamp/auth/internal/application"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ServeHTTP is wrapper function to start the apps infra in HTTP mode
func ServeHTTP(app *application.App) *gin.Engine {
	// call setup router
	setupRouter(app)

	return app.Application
}

// setupRouter is function to manage all routings
func setupRouter(app *application.App) {
	var dep = application.SetupDependencyInjection(app)

	v1 := app.Application.Group("/v1")
	{
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		v1.GET("/health-check", dep.HealthCheckController.Check)

		v1.POST("/register", dep.AuthController.Register)

		v1.GET("/register/verify", dep.AuthController.VerifyRegister)

		v1.POST("/login", dep.AuthController.Login)

		v1.POST("/forgot-password", dep.AuthController.ForgotPassword)

		v1.GET("/forgot-password/verify", dep.AuthController.VerifyForgotPassword)

		v1.PUT("/reset-password", dep.AuthController.ResetPassword)
	}

}
