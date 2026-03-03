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
		ctx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		defer stop()

		queues := []string{appContainer.Config.RabbitMQ.QueueCreateShortener, appContainer.Config.RabbitMQ.QueueUpdateVisitor, appContainer.Config.RabbitMQ.QueueUpdateShortener, appContainer.Config.RabbitMQ.QueueDeleteShortener}

		for _, q := range queues {
			appContainer.RabbitMQ.ConsumeMessages(ctx, appContainer.Config, appContainer.ShortController, q)
		}

		logger.Info("RabbitMQ Consumers started")

		<-ctx.Done()

		logger.Info("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		appContainer.Close(shutdownCtx)

		logger.Info("SHORTENER CONSUMER SERVICE CLOSED GRACEFULLY")
	}
}
