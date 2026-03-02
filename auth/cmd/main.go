package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"singkatin-api/auth/internal/bootstrap"
	"singkatin-api/auth/internal/routes"
	"singkatin-api/auth/pkg/logger"

	"github.com/joho/godotenv"
)

const (
	localServerMode = "local"
	httpServerMode  = "http"
)

// @title           Singkatin API
// @version         1.0
// @description     URL Shortener API - Auth Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /v1
// @Schemes         http
func main() {
	envPaths := []string{
		"./.env", "./auth/.env",
	}

	var loadErr error
	for _, path := range envPaths {
		if loadErr = godotenv.Load(path); loadErr == nil {
			break
		}
	}

	if loadErr != nil {
		logger.Warn("Warning: .env file not found (this is OK in Docker, env vars will be used)")
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
		logger.Errorf("Failed to initialize application container: %v", err)
		os.Exit(1)
	}

	switch mode {
	case localServerMode, httpServerMode:
		var (
			httpServer = routes.ServeHTTP(appContainer)
		)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", appContainer.Config.Server.AppPort),
			Handler: httpServer,
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Errorf("Failed to to start server. Error: %v", err)
			}
		}()

		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		appContainer.Close(ctx)

		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Failed to shutdown server. Error: %v", err)
		}

		logger.Info("AUTH SERVICE CLOSED SUCCESSFULLY")
	}
}
