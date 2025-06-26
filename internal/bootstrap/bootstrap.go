package bootstrap

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(logger *slog.Logger, redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	client := redis.NewClient(opt)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("Redis connection successful")

	return client
}
