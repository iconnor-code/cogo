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
	value    map[string]any
	filepath string
	viper    *viper.Viper
}

func WithFilePath(filepath string) core.ConfigOption {
	return func(c core.IConfig) error {
		c.(*Config).filepath = filepath
		return nil
	}
}

func NewConfig(opts ...core.ConfigOption) (*Config, error) {
	config := &Config{
		viper: viper.New(),
	}
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, cerrs.Wrap(err, "applying config option error")
		}
	}
	if err := config.ReLoad(); err != nil {
		return nil, err
	}
	return config, nil
}

func (ct *Config) Get(key string) any {
	return ct.viper.Get(key)
}

func (ct *Config) ReLoad() error {
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

	err := ct.viper.Unmarshal(&ct.value)
	if err != nil {
		return cerrs.Wrap(err)
	}
	return nil
}
