// Package config
package config

import "github.com/spf13/viper"

type ConfVal struct {
	Mode     string         `mapstructure:"mode"`
	Grpc     GrpcConfig     `mapstructure:"grpc"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Logger   LogConfig      `mapstructure:"logger"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Mysql    MysqlConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Etcd     EtcdConfig     `mapstructure:"etcd"`
	Consul   ConsulConfig   `mapstructure:"consul"`
	Registry RegistryConfig `mapstructure:"registry"`
	// Discovery DiscoveryConfig `mapstructure:"discovery"`
}

func (c *ConfVal) Get(key string) any {
	return viper.Get(key)
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}

type GrpcConfig struct {
	Listen string `mapstructure:"listen"`
}

type HTTPConfig struct {
	Listen string `mapstructure:"listen"`
	SSL    struct {
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"ssl"`
}
type LogConfig struct {
	Level      int8   `mapstructure:"level"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type MetricsConfig struct {
	Enable bool   `mapstructure:"enable"`
	Listen string `mapstructure:"listen"`
	Prefix string `mapstructure:"prefix"`
}

type MysqlConfig struct {
	DSN  string `mapstructure:"dsn"`
	Pool struct {
		MaxOpenConns int `mapstructure:"max_open_conns"`
		MaxIdleConns int `mapstructure:"max_idle_conns"`
		MaxLifetime  int `mapstructure:"max_lifetime"`
	} `mapstructure:"pool"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ConsulConfig struct {
	Address string `mapstructure:"address"`
	Scheme  string `mapstructure:"scheme"`
}

type RegistryConfig struct {
	Name        string                  `mapstructure:"name"`
	Port        string                  `mapstructure:"port"`
	Tags        []string                `mapstructure:"tags"`
	Address     string                  `mapstructure:"address"`
	HealthCheck RegistryHealthCheckConf `mapstructure:"health_check"`
}

type RegistryHealthCheckConf struct {
	URI                            string `mapstructure:"uri"`
	Timeout                        string `mapstructure:"timeout"`
	Interval                       string `mapstructure:"interval"`
	DeregisterCriticalServiceAfter string `mapstructure:"dregister_critical_service_after"`
}
