// Package captcha
package captcha

import (
	"context"
	"fmt"
	"time"

	"github.com/iconnor-code/cogo/client"
	"github.com/mojocn/base64Captcha"
)

type RedisStore struct {
	redisClient *client.RedisClient
	ctx         context.Context
	key         string
	expire      time.Duration
}

func NewRedisStore(ctx context.Context, redisClient *client.RedisClient, key string, expire time.Duration) base64Captcha.Store {
	return &RedisStore{redisClient: redisClient, ctx: ctx, key: key, expire: expire}
}

func (s *RedisStore) Set(id string, value string) error {
	return s.redisClient.Set(s.ctx, fmt.Sprintf(s.key, id), value, s.expire).Err()
}

func (s *RedisStore) Get(id string, clear bool) string {
	key := fmt.Sprintf(s.key, id)
	value, err := s.redisClient.Get(s.ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.redisClient.Del(s.ctx, key)
	}
	return value
}

func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	key := fmt.Sprintf(s.key, id)
	value, err := s.redisClient.Get(s.ctx, key).Result()
	if err != nil {
		return false
	}
	if clear {
		s.redisClient.Del(s.ctx, key)
	}
	return value == answer
}
