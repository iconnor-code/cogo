package client

import (
	"github.com/iconnor-code/cogo/core"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
	conf map[string]any
}

type RedisClientOption func(client *RedisClient) error

func NewRedisClient(config core.IConfig) *RedisClient {
	client := &RedisClient{
		conf: config.Get("redis").(map[string]any),
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
