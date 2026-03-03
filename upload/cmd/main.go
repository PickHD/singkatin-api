package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"singkatin-api/upload/internal/bootstrap"
	"singkatin-api/upload/pkg/logger"

	"github.com/joho/godotenv"
)

const (
	consumerMode = "consumer"
)

// @title           Singkatin API
// @version         1.0
// @description     URL Shortener API - Upload Services
// @contact.name    Taufik Januar
// @contact.email   taufikjanuar35@gmail.com
// @license.name    MIT
// @host            localhost:8083
func main() {
	envPaths := []string{
		"./.env", "./upload/.env",
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
		mode = consumerMode
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
	case consumerMode:
		ctx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		defer stop()

		queues := []string{appContainer.Config.RabbitMQ.QueueUploadAvatar}

		for _, q := range queues {
			appContainer.RabbitMQ.ConsumeMessages(ctx, appContainer.Config, appContainer.UploadController, q)
		}

		logger.Info("RabbitMQ Consumers started")

		<-ctx.Done()

		logger.Info("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		appContainer.Close(shutdownCtx)

		logger.Info("UPLOAD SERVICE CLOSED GRACEFULLY")
	}
}
