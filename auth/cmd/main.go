package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/PickHD/singkatin-revamp/auth/internal/application"
	"github.com/PickHD/singkatin-revamp/auth/internal/infrastructure"
	"github.com/joho/godotenv"
)

const (
	localServerMode = "local"
	httpServerMode  = "http"
)

// @title           Singkatin Revamp API
// @version         1.0
// @description     Revamped URL Shortener API - Auth Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /v1
// @Schemes         http
func main() {
	err := godotenv.Load("./cmd/.env")
	if err != nil {
		panic(err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Checking command arguments
	var (
		args = os.Args[1:]
		mode = localServerMode
	)

	if len(args) > 0 {
		mode = os.Args[1]
	}

	// create a context with background for setup the application
	ctx := context.Background()
	app, err := application.SetupApplication(ctx)
	if err != nil {
		app.Logger.Error("Failed to initialize app. Error: ", err)
		panic(err)
	}

	switch mode {
	case localServerMode, httpServerMode:
		var (
			httpServer = infrastructure.ServeHTTP(app)
		)

		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", app.Config.Server.AppPort),
			Handler: httpServer,
		}

		// Create a channel to receive OS signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		// Start the HTTP server in a separate Goroutine
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				app.Logger.Error("Failed to to start server. Error: ", err)
			}
		}()

		// Wait for a SIGINT or SIGTERM signal
		<-sigCh

		// Create a context with a timeout of 5 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.Close(ctx)

		// Shutdown the server gracefully
		if err := server.Shutdown(ctx); err != nil {
			app.Logger.Error("Failed to shutdown server. Error: ", err)
		}

		app.Logger.Info("SERVER SHUTDOWN GRACEFULLY")
	}
}
