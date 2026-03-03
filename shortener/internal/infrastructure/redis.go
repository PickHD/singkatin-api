package infrastructure

import (
	"context"
	"fmt"
	"singkatin-api/shortener/internal/config"
	"singkatin-api/shortener/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisConnectionProvider struct {
	client *redis.Client
}

func NewRedisConnection(ctx context.Context, cfg *config.Config) *RedisConnectionProvider {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Errorf("failed connect redis, error: %v", err)
	}
	return &RedisConnectionProvider{client: client}
}

func (r *RedisConnectionProvider) GetClient() *redis.Client {
	return r.client
}

func (r *RedisConnectionProvider) Close() error {
	return r.client.Close()
}
