package infrastructure

import (
	"context"
	"singkatin-api/user/internal/config"
	"singkatin-api/user/pkg/logger"

	"github.com/streadway/amqp"
)

type RabbitMQConnectionProvider struct {
	client *amqp.Channel
}

func NewRabbitMQConnection(ctx context.Context, cfg *config.Config) *RabbitMQConnectionProvider {
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
