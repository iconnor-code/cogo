package client

import (
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
	conf map[string]any
}

type RedisClientOption func(client *RedisClient) error

func NewRedisClient(config core.IConfig) (*RedisClient, error) {
	addr, err := core.GetString(config, "redis.addr")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	password, err := core.GetString(config, "redis.password")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	db, err := core.GetInt(config, "redis.db")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{
		Client: redisClient,
	}, nil
}
