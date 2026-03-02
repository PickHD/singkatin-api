package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"singkatin-api/user/internal/bootstrap"
	"singkatin-api/user/internal/routes"
	"singkatin-api/user/pkg/logger"

	"github.com/joho/godotenv"
)

const (
	localServerMode = "local"
	httpServerMode  = "http"
)

// @title           Singkatin API
// @version         1.0
// @description     URL Shortener API - User Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8082
// @BasePath        /v1
// @Schemes         http
func main() {
	envPaths := []string{
		"./.env", "./user/.env",
	}

	var loadErr error
	for _, path := range envPaths {
		if loadErr = godotenv.Load(path); loadErr == nil {
			break
		}
	}

	if loadErr != nil {
		panic(loadErr)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var (
		args = os.Args[1:]
		mode = localServerMode
	)

	if len(args) > 0 {
		mode = os.Args[1]
	}

	ctx := context.Background()
	appContainer, err := bootstrap.NewContainer(ctx)
	if err != nil {
		panic(err)
	}
	appContainer.Tracer.SetTracerProvider()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	switch mode {
	case localServerMode, httpServerMode:
		httpServer := routes.ServeHTTP(appContainer)

		go func() {
			port := fmt.Sprintf(":%d", appContainer.Config.Server.AppPort)
			if err := httpServer.Listen(port); err != nil {
				logger.Errorf("Server listen error/closed: %v", err)
			}
		}()

		<-c

		if err := httpServer.Shutdown(); err != nil {
			logger.Errorf("Failed to gracefully shutdown HTTP server: %v", err)
		}

		appContainer.Close(appContainer.Context)

		logger.Info("USER SERVICE CLOSED GRACEFULLY")
	}
}
