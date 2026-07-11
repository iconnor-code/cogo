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

func NewRedisClient(config core.IConfig) (*RedisClient, error) {
	redisConf := config.GetRedis()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConf.Addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	})

	return &RedisClient{
		Client: redisClient,
	}, nil
}
