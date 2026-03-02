package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"time"

	"singkatin-api/upload/internal/bootstrap"
	"singkatin-api/upload/pkg/logger"

	"github.com/joho/godotenv"
)

const (
	consumerMode = "consumer"
)

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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	switch mode {
	case consumerMode:
		forever := make(chan bool)

		queues := []string{appContainer.Config.RabbitMQ.QueueUploadAvatar}

		for _, q := range queues {
			go appContainer.RabbitMQ.ConsumeMessages(appContainer.Context, appContainer.Config, appContainer.UploadController, q)
		}

		go func() {
			<-sigCh
			logger.Info("Shutdown signal received")
			forever <- true
		}()

		<-forever

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		appContainer.Close(ctx)

		logger.Info("UPLOAD SERVICE CLOSED GRACEFULLY")
	}
}
