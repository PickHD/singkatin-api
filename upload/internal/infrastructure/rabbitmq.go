package infrastructure

import (
	"fmt"

	"github.com/PickHD/singkatin-revamp/upload/internal/application"
	uploadpb "github.com/PickHD/singkatin-revamp/upload/pkg/api/v1/proto/upload"
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
			req := &uploadpb.UploadAvatarMessage{}

			err := proto.Unmarshal(msg.Body, req)
			if err != nil {
				app.Logger.Error("Unmarshal proto UploadAvatarMessage ERROR, ", err)
			}

			app.Logger.Info(fmt.Sprintf("[%s] Success Consume Message :", queueName), req)

			err = dep.UploadController.ProcessUploadAvatarUser(app.Context, req)
			if err != nil {
				app.Logger.Error("ProcessUploadAvatarUser ERROR, ", err)
			}

			app.Logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
		}
	}()
}
