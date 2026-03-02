package infrastructure

import (
	"context"
	"fmt"
	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/internal/controller"
	shortenerpb "singkatin-api/shortener/pkg/api/v1/proto/shortener"
	"singkatin-api/shortener/pkg/logger"

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

	queues := []string{cfg.RabbitMQ.QueueCreateShortener, cfg.RabbitMQ.QueueUpdateVisitor, cfg.RabbitMQ.QueueUpdateShortener, cfg.RabbitMQ.QueueDeleteShortener}

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
func (r *RabbitMQConnectionProvider) ConsumeMessages(ctx context.Context, cfg *config.Configuration, shortController controller.ShortController, queueName string) {
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
			switch queueName {
			case cfg.RabbitMQ.QueueCreateShortener:
				req := &shortenerpb.CreateShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					logger.Errorf("Unmarshal proto CreateShortenerMessage ERROR, %v", err)
				}

				logger.Infof("[%s] Success Consume Message : %v", queueName, req)

				err = shortController.ProcessCreateShortUser(ctx, req)
				if err != nil {
					logger.Errorf("ProcessCreateShortUser ERROR, %v", err)
				}

				logger.Infof("[%s] Success Process Message : %v", queueName, req)
			case cfg.RabbitMQ.QueueUpdateVisitor:
				req := &shortenerpb.UpdateVisitorCountMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					logger.Errorf("Unmarshal proto UpdateVisitorCountMessage ERROR, %v", err)
				}

				logger.Infof("[%s] Success Consume Message : %v", queueName, req)

				err = shortController.ProcessUpdateVisitorCount(ctx, req)
				if err != nil {
					logger.Errorf("ProcessUpdateVisitorCount ERROR, %v", err)
				}

				logger.Infof("[%s] Success Process Message : %v", queueName, req)
			case cfg.RabbitMQ.QueueUpdateShortener:
				req := &shortenerpb.UpdateShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					logger.Errorf("Unmarshal proto UpdateShortenerMessage ERROR, %v", err)
				}

				logger.Infof("[%s] Success Consume Message : %v", queueName, req)

				err = shortController.ProcessUpdateShortUser(ctx, req)
				if err != nil {
					logger.Errorf("ProcessUpdateShortUser ERROR, %v", err)
				}

				logger.Info(fmt.Sprintf("[%s] Success Process Message :", queueName), req)
			case cfg.RabbitMQ.QueueDeleteShortener:
				req := &shortenerpb.DeleteShortenerMessage{}

				err := proto.Unmarshal(msg.Body, req)
				if err != nil {
					logger.Errorf("Unmarshal proto DeleteShortenerMessage ERROR, %v", err)
				}

				logger.Infof("[%s] Success Consume Message : %v", queueName, req)

				err = shortController.ProcessDeleteShortUser(ctx, req)
				if err != nil {
					logger.Errorf("ProcessDeleteShortUser ERROR, %v", err)
				}

				logger.Infof("[%s] Success Process Message : %v", queueName, req)
			}
		}
	}()
}
