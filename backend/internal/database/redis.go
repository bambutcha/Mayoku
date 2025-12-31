package database

import (
	"context"
	"fmt"

	"github.com/Chelaran/mayoku/internal/config"
	"github.com/redis/go-redis/v9"
)

// ConnectRedis создает подключение к Redis
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверка подключения
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
