package client

import (
	"testing"

	"github.com/iconnor-code/cogo/core"
	configimpl "github.com/iconnor-code/cogo/core/impl/config"
)

func TestNewRedisClientUsesSingleNodeByDefault(t *testing.T) {
	conf := &configimpl.Config{Config: core.Config{Redis: core.RedisConfig{
		Addr:     "redis:6379",
		Username: "app",
		Password: "secret",
		DB:       2,
	}}}

	client, err := NewRedisClient(conf)
	if err != nil {
		t.Fatalf("new redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	options := client.Options()
	if options.Addr != "redis:6379" || options.Username != "app" || options.Password != "secret" || options.DB != 2 {
		t.Fatalf("single-node options = %+v", options)
	}
}

func TestNewRedisClientUsesSentinelWhenConfigured(t *testing.T) {
	conf := &configimpl.Config{Config: core.Config{Redis: core.RedisConfig{
		Addr:          "redis:6379",
		MasterName:    "mymaster",
		SentinelAddrs: []string{"sentinel-0:26379", "sentinel-1:26379"},
		Username:      "app",
		Password:      "secret",
		DB:            3,
	}}}

	client, err := NewRedisClient(conf)
	if err != nil {
		t.Fatalf("new redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	options := client.Options()
	if options.Addr != "FailoverClient" || options.Username != "app" || options.Password != "secret" || options.DB != 3 {
		t.Fatalf("sentinel options = %+v", options)
	}
}

func TestNewRedisClientFallsBackToSingleNodeWithIncompleteSentinelConfig(t *testing.T) {
	conf := &configimpl.Config{Config: core.Config{Redis: core.RedisConfig{
		Addr:       "redis:6379",
		MasterName: "mymaster",
	}}}

	client, err := NewRedisClient(conf)
	if err != nil {
		t.Fatalf("new redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if got := client.Options().Addr; got != "redis:6379" {
		t.Fatalf("redis addr = %q, want single-node fallback", got)
	}
}
