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
		return
	}

	logger.Infof("Waiting Message in Queues %s.....", queueName)

	var handler func(context.Context, amqp.Delivery) error

	switch queueName {
	case cfg.RabbitMQ.QueueCreateShortener:
		handler = func(ctx context.Context, msg amqp.Delivery) error {
			req := &shortenerpb.CreateShortenerMessage{}
			if err := proto.Unmarshal(msg.Body, req); err != nil {
				return fmt.Errorf("Unmarshal proto CreateShortenerMessage ERROR, %w", err)
			}
			logger.Infof("[%s] Success Consume Message : %v", queueName, req)

			if err := shortController.ProcessCreateShortUser(ctx, req); err != nil {
				return fmt.Errorf("ProcessCreateShortUser ERROR, %w", err)
			}
			logger.Infof("[%s] Success Process Message : %v", queueName, req)
			return nil
		}
	case cfg.RabbitMQ.QueueUpdateVisitor:
		handler = func(ctx context.Context, msg amqp.Delivery) error {
			req := &shortenerpb.UpdateVisitorCountMessage{}
			if err := proto.Unmarshal(msg.Body, req); err != nil {
				return fmt.Errorf("Unmarshal proto UpdateVisitorCountMessage ERROR, %w", err)
			}
			logger.Infof("[%s] Success Consume Message : %v", queueName, req)

			if err := shortController.ProcessUpdateVisitorCount(ctx, req); err != nil {
				return fmt.Errorf("ProcessUpdateVisitorCount ERROR, %w", err)
			}
			logger.Infof("[%s] Success Process Message : %v", queueName, req)
			return nil
		}
	case cfg.RabbitMQ.QueueUpdateShortener:
		handler = func(ctx context.Context, msg amqp.Delivery) error {
			req := &shortenerpb.UpdateShortenerMessage{}
			if err := proto.Unmarshal(msg.Body, req); err != nil {
				return fmt.Errorf("Unmarshal proto UpdateShortenerMessage ERROR, %w", err)
			}
			logger.Infof("[%s] Success Consume Message : %v", queueName, req)

			if err := shortController.ProcessUpdateShortUser(ctx, req); err != nil {
				return fmt.Errorf("ProcessUpdateShortUser ERROR, %w", err)
			}
			logger.Infof("[%s] Success Process Message : %v", queueName, req)
			return nil
		}
	case cfg.RabbitMQ.QueueDeleteShortener:
		handler = func(ctx context.Context, msg amqp.Delivery) error {
			req := &shortenerpb.DeleteShortenerMessage{}
			if err := proto.Unmarshal(msg.Body, req); err != nil {
				return fmt.Errorf("Unmarshal proto DeleteShortenerMessage ERROR, %w", err)
			}
			logger.Infof("[%s] Success Consume Message : %v", queueName, req)

			if err := shortController.ProcessDeleteShortUser(ctx, req); err != nil {
				return fmt.Errorf("ProcessDeleteShortUser ERROR, %w", err)
			}
			logger.Infof("[%s] Success Process Message : %v", queueName, req)
			return nil
		}
	default:
		logger.Errorf("Unknown queue name to consume: %s", queueName)
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infof("Context cancelled, stopping consumer for queue: %s", queueName)
				return
			case msg, ok := <-messages:
				if !ok {
					logger.Infof("Message channel closed for queue: %s", queueName)
					return
				}
				if err := handler(ctx, msg); err != nil {
					logger.Errorf("Queue %s handling error: %v", queueName, err)
				}
			}
		}
	}()
}
