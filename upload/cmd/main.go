package main

import (
	"context"
	"os"
	"runtime"

	"github.com/PickHD/singkatin-revamp/upload/internal/application"
	"github.com/PickHD/singkatin-revamp/upload/internal/infrastructure"
	"github.com/joho/godotenv"
)

const (
	consumerMode = "consumer"
)

func main() {
	err := godotenv.Load("./cmd/.env")
	if err != nil {
		panic(err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Checking command arguments
	var (
		args = os.Args[1:]
		mode = consumerMode
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
	case consumerMode:
		// Make a channel to receive messages into infinite loop.
		forever := make(chan bool)

		queues := []string{app.Config.RabbitMQ.QueueUploadAvatar}

		for _, q := range queues {
			go infrastructure.ConsumeMessages(app, q)
		}

		<-forever
	}
}
