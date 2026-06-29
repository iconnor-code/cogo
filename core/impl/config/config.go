// Package config provides a configuration management implementation for the Cogo framework.
package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/spf13/viper"
)

type Config struct {
	core.Config `mapstructure:",squash"`

	filepath string
	viper    *viper.Viper
}

type ConfigOption func(*Config) error

func WithFilePath(filepath string) ConfigOption {
	return func(c *Config) error {
		c.filepath = filepath
		return nil
	}
}

func NewConfig(opts ...ConfigOption) (*Config, error) {
	config := &Config{
		viper: viper.New(),
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, cerrs.Wrap(err, "applying config option error")
		}
	}
	if err := config.Reload(); err != nil {
		return nil, err
	}
	return config, nil
}

func Load[T any](opts ...ConfigOption) (*T, error) {
	config, err := NewConfig(opts...)
	if err != nil {
		return nil, err
	}

	var out T
	if err := config.Unmarshal(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (ct *Config) Unmarshal(out any) error {
	if ct.viper == nil {
		return cerrs.New("config viper not initialized")
	}
	if err := ct.viper.Unmarshal(out); err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("unmarshal config file error,filepath:%s", ct.filepath))
	}
	return nil
}

func (ct *Config) GetMode() string { return ct.Mode }

func (ct *Config) GetBizID() int { return ct.BizID }

func (ct *Config) GetBizName() string { return ct.BizName }

func (ct *Config) GetGRPC() core.GRPCConfig { return ct.GRPC }

func (ct *Config) GetHTTP() core.HTTPConfig { return ct.HTTP }

func (ct *Config) GetLogger() core.LoggerConfig { return ct.Logger }

func (ct *Config) GetMetrics() core.MetricsConfig { return ct.Metrics }

func (ct *Config) GetMySQL() core.MySQLConfig { return ct.MySQL }

func (ct *Config) GetRedis() core.RedisConfig { return ct.Redis }

func (ct *Config) GetEtcd() core.EtcdConfig { return ct.Etcd }

func (ct *Config) GetConsul() core.ConsulConfig { return ct.Consul }

func (ct *Config) GetRegistry() core.RegistryConfig { return ct.Registry }

func (ct *Config) GetSMTP() core.SMTPConfig { return ct.SMTP }

func (ct *Config) GetJWT() core.JWTConfig { return ct.JWT }

func (ct *Config) GetAdmin() core.AdminConfig { return ct.Admin }

func (ct *Config) Reload() error {
	if ct.filepath != "" {
		return ct.loadFromFile()
	}
	return nil
}

func (ct *Config) loadFromFile() error {
	dir := path.Dir(ct.filepath)
	fileNameWithoutExt := path.Base(ct.filepath)
	ext := path.Ext(ct.filepath)
	fileNameWithoutExt = fileNameWithoutExt[:len(fileNameWithoutExt)-len(ext)]
	configType := strings.TrimPrefix(ext, ".")
	if configType == "" {
		configType = "yaml"
	}

	ct.viper.SetConfigType(configType)
	ct.viper.AddConfigPath(dir)
	ct.viper.SetConfigName(fileNameWithoutExt)

	if err := ct.viper.ReadInConfig(); err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("reading config file error,filepath:%s", ct.filepath))
	}
	return ct.Unmarshal(ct)
}
