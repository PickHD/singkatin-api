package infrastructure

import (
	"context"

	"singkatin-api/upload/internal/config"
	"singkatin-api/upload/internal/controller"
	uploadpb "singkatin-api/upload/pkg/api/v1/proto/upload"
	"singkatin-api/upload/pkg/logger"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

type RabbitMQConnectionProvider struct {
	client *amqp.Channel
}

func NewRabbitMQConnection(ctx context.Context, cfg *config.Configuration) *RabbitMQConnectionProvider {
	amqpConn, err := amqp.Dial(cfg.RabbitMQ.ConnURL)
	if err != nil {
		logger.Errorf("failed dial RabbitMQ, error : %v", err)
		return nil
	}

	amqpClient, err := amqpConn.Channel()
	if err != nil {
		logger.Errorf("failed open RabbitMQ Channels, error : %v", err)
		return nil
	}

	queues := []string{cfg.RabbitMQ.QueueUploadAvatar}

	for _, q := range queues {
		_, err = amqpClient.QueueDeclare(
			q,     // queue name
			true,  // durable
			false, // auto delete
			false, // exclusive
			false, // no wait
			nil,   // arguments
		)
		if err != nil {
			logger.Errorf("failed queue declare Channels, error : %v", err)
			return nil
		}
	}

	return &RabbitMQConnectionProvider{client: amqpClient}
}

func (r *RabbitMQConnectionProvider) GetClient() *amqp.Channel {
	return r.client
}

func (r *RabbitMQConnectionProvider) Close() error {
	return r.client.Close()
}

// ConsumeMessages generic function to consume message from defined param queues
func (r *RabbitMQConnectionProvider) ConsumeMessages(ctx context.Context, cfg *config.Configuration, uploadController controller.UploadController, queueName string) {
	// Subscribing to queues for getting messages.
	messages, err := r.client.Consume(
		queueName, // queue name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		logger.Errorf("Failed consume message in queue %s, %v", queueName, err)
	}

	logger.Infof("Waiting Message in Queues %s.....", queueName)

	go func() {
		for msg := range messages {
			req := &uploadpb.UploadAvatarMessage{}

			err := proto.Unmarshal(msg.Body, req)
			if err != nil {
				logger.Errorf("Unmarshal proto UploadAvatarMessage ERROR, %v", err)
			}

			logger.Infof("[%s] Success Consume Message : %v", queueName, req)

			err = uploadController.ProcessUploadAvatarUser(ctx, req)
			if err != nil {
				logger.Errorf("ProcessUploadAvatarUser ERROR, %v", err)
			}

			logger.Infof("[%s] Success Process Message : %v", queueName, req)
		}
	}()
}
