package db

import (
	"github.com/iconnor-code/cogo/pkg/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	*redis.Client
}

func NewRedisCache(config *config.Conf) *RedisCache {
	return &RedisCache{
		redis.NewClient(&redis.Options{
			Addr:     config.Redis.Addr,
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
		}),
	}
}
