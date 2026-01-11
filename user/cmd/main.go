package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/PickHD/singkatin-revamp/user/internal/application"
	"github.com/PickHD/singkatin-revamp/user/internal/infrastructure"
	"github.com/joho/godotenv"
)

const (
	localServerMode = "local"
	httpServerMode  = "http"
)

// @title           Singkatin Revamp API
// @version         1.0
// @description     Revamped URL Shortener API - User Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8082
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

	//create a channel for listening to OS signals and connecting OS interrupts to the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	serverShutdown := make(chan struct{})

	go func() {
		_ = <-c
		app.Logger.Info("SERVER SHUTDOWN GRACEFULLY")
		app.Close(app.Context)
		_ = app.Application.Shutdown()
		serverShutdown <- struct{}{}
	}()

	switch mode {
	case localServerMode, httpServerMode:
		var (
			httpServer = infrastructure.ServeHTTP(app)
		)

		if err := httpServer.Listen(fmt.Sprintf(":%d", app.Config.Server.AppPort)); err != nil {
			app.Logger.Error("Failed to to start server. Error: ", err)
			panic(err)
		}

	}
}
