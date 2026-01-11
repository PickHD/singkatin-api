package infrastructure

import (
	"fmt"

	"github.com/PickHD/singkatin-revamp/shortener/internal/application"
	shortenerpb "github.com/PickHD/singkatin-revamp/shortener/pkg/api/v1/proto/shortener"
	"google.golang.org/protobuf/proto"
)

// ConsumeMessages generic function to consume message from defined param queues
func ConsumeMessages(app *application.App, queueName string) {
	dep := application.SetupDependencyInjection(app)

	// Subscribing to queues for getting messages.
	messages, err := app.RabbitMQ.Consume(
		queueName, // queue name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		app.Logger.Error("Failed consume message in queue", queueName)
	}

	app.Logger.Info("Waiting Message in Queues ", queueName, ".....")

	go func() {
		for msg := range messages {
			switch queueName {
			case app.Config.RabbitMQ.QueueCreateShortener:
				req := &shortenerpb.CreateShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					app.Logger.Error("Unmarshal proto CreateShortenerMessage ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Consume Message :", queueName), req)

				err = dep.ShortController.ProcessCreateShortUser(app.Context, req)
				if err != nil {
					app.Logger.Error("ProcessCreateShortUser ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
			case app.Config.RabbitMQ.QueueUpdateVisitor:
				req := &shortenerpb.UpdateVisitorCountMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					app.Logger.Error("Unmarshal proto UpdateVisitorCountMessage ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Consume Message :", queueName), req)

				err = dep.ShortController.ProcessUpdateVisitorCount(app.Context, req)
				if err != nil {
					app.Logger.Error("ProcessUpdateVisitorCount ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
			case app.Config.RabbitMQ.QueueUpdateShortener:
				req := &shortenerpb.UpdateShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					app.Logger.Error("Unmarshal proto UpdateShortenerMessage ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Consume Message :", queueName), req)

				err = dep.ShortController.ProcessUpdateShortUser(app.Context, req)
				if err != nil {
					app.Logger.Error("ProcessUpdateShortUser ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
			case app.Config.RabbitMQ.QueueDeleteShortener:
				req := &shortenerpb.DeleteShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					app.Logger.Error("Unmarshal proto DeleteShortenerMessage ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Consume Message :", queueName), req)

				err = dep.ShortController.ProcessDeleteShortUser(app.Context, req)
				if err != nil {
					app.Logger.Error("ProcessDeleteShortUser ERROR, ", err)
				}

				app.Logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
			}
		}
	}()
}
