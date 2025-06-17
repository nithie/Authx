package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var RedisContext = context.Background()

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: RedisUrl,
	})
}
