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
	var redisClient *redis.Client
	if redisConf.MasterName != "" && len(redisConf.SentinelAddrs) > 0 {
		redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       redisConf.MasterName,
			SentinelAddrs:    redisConf.SentinelAddrs,
			SentinelUsername: redisConf.SentinelUsername,
			SentinelPassword: redisConf.SentinelPassword,
			Username:         redisConf.Username,
			Password:         redisConf.Password,
			DB:               redisConf.DB,
		})
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConf.Addr,
			Username: redisConf.Username,
			Password: redisConf.Password,
			DB:       redisConf.DB,
		})
	}

	return &RedisClient{
		Client: redisClient,
	}, nil
}
