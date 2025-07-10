package config

import "github.com/spf13/viper"

type ConfigValue struct {
	Mode     string         `mapstructure:"mode"`
	Grpc     GrpcConfig     `mapstructure:"grpc"`
	Logger   LogConfig      `mapstructure:"logger"`
	Http     HttpConfig     `mapstructure:"http"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Mysql    MysqlConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Smtp     SmtpConfig     `mapstructure:"smtp"`
	JwtToken JwtTokenConfig `mapstructure:"jwt_token"`
	Etcd     EtcdConfig     `mapstructure:"etcd"`
	Registry RegistryConfig `mapstructure:"registry"`
}

func (c *ConfigValue) Get(key string) any {
	return viper.Get(key)
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}

type RegistryConfig struct {
	Hostname string `mapstructure:"hostname"`
}

type GrpcConfig struct {
	Listen string `mapstructure:"listen"`
}

type HttpConfig struct {
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

type SmtpConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JwtTokenConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`
	AccessExpire  int    `mapstructure:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
}
