package database

import (
	"github.com/iconnor-code/cogo/core"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
	conf map[string]any
}

type RedisClientOption func(client *RedisClient) error

func WithRedisConfig(config core.IConfig) RedisClientOption {
	return func(client *RedisClient) error {
		client.conf = config.Get("redis").(map[string]any)
		return nil
	}
}

func NewRedisClient(opts ...RedisClientOption) *RedisClient {
	client := &RedisClient{}
	for _, opt := range opts {
		opt(client)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     client.conf["addr"].(string),
		Password: client.conf["password"].(string),
		DB:       client.conf["db"].(int),
	})

	return &RedisClient{
		Client: redisClient,
	}
}
