package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigReturnsIndependentInstances(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.yaml")
	secondPath := filepath.Join(dir, "second.yaml")
	if err := os.WriteFile(firstPath, []byte("mode: first\n"), 0o600); err != nil {
		t.Fatalf("write first config: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("mode: second\n"), 0o600); err != nil {
		t.Fatalf("write second config: %v", err)
	}

	first, err := NewConfig(WithFilePath(firstPath))
	if err != nil {
		t.Fatalf("new first config: %v", err)
	}
	second, err := NewConfig(WithFilePath(secondPath))
	if err != nil {
		t.Fatalf("new second config: %v", err)
	}

	if got := first.Mode; got != "first" {
		t.Fatalf("first name = %v, want first", got)
	}
	if got := second.Mode; got != "second" {
		t.Fatalf("second name = %v, want second", got)
	}
}

func TestNewConfigUnmarshalsStructuredFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := []byte(`
mode: debug
biz_id: 101100
biz_name: account
grpc:
  listen: :10000
http:
  listen: :18080
logger:
  file_path: deploy/log
  max_size: 100
  max_backups: 7
  max_age: 7
metrics:
  enable: true
  listen: :10090
mysql:
  dsn: root:pass@tcp(localhost:3306)/app
  pool:
    max_open_conns: 10
    max_idle_conns: 5
    max_lifetime: 60
redis:
  addr: 127.0.0.1:6379
  db: 1
etcd:
  endpoints:
    - 127.0.0.1:2379
consul:
  address: 127.0.0.1:8500
registry:
  name: account
  address: 127.0.0.1
  port: 10000
  health_check:
    interval: 3s
    timeout: 5s
jwt:
  access_secret: secret
  access_expire: 24
  refresh_expire: 7
oss:
  endpoint: minio.local:9001
  access_key_id: access
  access_key_secret: secret
  bucket_name: mysite
  base_url: http://minio.local:9001/mysite
  presign_expire: 900
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := NewConfig(WithFilePath(configPath))
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	if conf.Mode != "debug" {
		t.Fatalf("mode = %q, want debug", conf.Mode)
	}
	if conf.GRPC.Listen != ":10000" {
		t.Fatalf("grpc listen = %q, want :10000", conf.GRPC.Listen)
	}
	if conf.MySQL.Pool.MaxOpenConns != 10 {
		t.Fatalf("mysql max open conns = %d, want 10", conf.MySQL.Pool.MaxOpenConns)
	}
	if conf.Registry.HealthCheck.Interval != "3s" {
		t.Fatalf("registry health check interval = %q, want 3s", conf.Registry.HealthCheck.Interval)
	}
	if conf.OSS.BucketName != "mysite" {
		t.Fatalf("oss bucket = %q, want mysite", conf.OSS.BucketName)
	}
}

func TestLoadSupportsBusinessConfigEmbedding(t *testing.T) {
	type businessConfig struct {
		Config `mapstructure:",squash"`

		Admin struct {
			UserIDs []int `mapstructure:"user_ids" yaml:"user_ids"`
		} `mapstructure:"admin" yaml:"admin"`
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "app.yaml")
	content := []byte(`
mode: debug
grpc:
  listen: :10000
admin:
  user_ids:
    - 1
    - 2
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := Load[businessConfig](WithFilePath(configPath))
	if err != nil {
		t.Fatalf("load business config: %v", err)
	}

	if conf.GRPC.Listen != ":10000" {
		t.Fatalf("grpc listen = %q, want :10000", conf.GRPC.Listen)
	}
	if len(conf.Admin.UserIDs) != 2 || conf.Admin.UserIDs[0] != 1 || conf.Admin.UserIDs[1] != 2 {
		t.Fatalf("admin user ids = %+v, want [1 2]", conf.Admin.UserIDs)
	}
}
