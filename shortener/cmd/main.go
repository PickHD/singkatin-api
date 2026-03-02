package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"singkatin-api/shortener/internal/bootstrap"
	"singkatin-api/shortener/internal/routes"
	"singkatin-api/shortener/pkg/logger"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	localServerMode = "local"
	httpServerMode  = "http"
	consumerMode    = "consumer"
	grpcMode        = "grpc"
)

// @title           Singkatin API
// @version         1.0
// @description     URL Shortener API - Shortener Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8081
// @BasePath        /v1
// @Schemes         http
func main() {
	envPaths := []string{
		"./.env", "./shortener/.env",
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

	// Checking command arguments
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
		logger.Errorf("Failed to initialize app. Error: %v", err)
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

		// Create a channel to receive OS signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		// Start the HTTP server in a separate Goroutine
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Errorf("Failed to to start server. Error: %v", err)
			}
		}()

		// Wait for a SIGINT or SIGTERM signal
		<-sigCh

		// Create a context with a timeout of 5 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		appContainer.Close(ctx)

		// Shutdown the server gracefully
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Failed to shutdown server. Error: %v", err)
		}

		logger.Info("SHORTENER SERVICE CLOSED GRACEFULLY")
	case grpcMode:
		var (
			grpcServer = routes.ServeGRPC(appContainer)
		)

		errC := make(chan error, 1)

		ctx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		go func() {
			addr := fmt.Sprintf("0.0.0.0:%d", appContainer.Config.Common.GrpcPort)

			lis, err := net.Listen("tcp", addr)
			if err != nil {
				logger.Errorf("cannot listen tcp grpc %v", err)
			}

			logger.Infof("Listening and serving GRPC server %s", lis.Addr().String())

			if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				errC <- err
			}
		}()

		go func() {
			<-ctx.Done()

			logger.Info("Shutdown signal received")

			defer func() {
				stop()
				close(errC)
			}()

			logger.Info("Shutdown completed")
		}()

		if err := <-errC; err != nil {
			logger.Errorf("Error received by channel %v", err)
		}
	case consumerMode:
		forever := make(chan bool)

		queues := []string{appContainer.Config.RabbitMQ.QueueCreateShortener, appContainer.Config.RabbitMQ.QueueUpdateVisitor, appContainer.Config.RabbitMQ.QueueUpdateShortener, appContainer.Config.RabbitMQ.QueueDeleteShortener}

		for _, q := range queues {
			go appContainer.RabbitMQ.ConsumeMessages(appContainer.Context, appContainer.Config, appContainer.ShortController, q)
		}

		<-forever
	}
}
