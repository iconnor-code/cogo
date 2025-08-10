package config

import (
	"errors"
	"fmt"
	"path"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/spf13/viper"
)

type IConfig interface {
	Get() IConfig
	ReLoad() error
}

type ConfigOption func(c *Config) error

func WithFilePath(filepath string) ConfigOption {
	return func(c *Config) error {
		c.filepath = filepath
		return c.ReLoad()
	}
}

func (c *Config) ReLoad() error {
	if c.filepath != "" {
		return c.loadFromFile(c.filepath)
	}
	return errors.New("no config source provided")
}

func (c *Config) Get() IConfig {
	c.rwmutex.RLock()
	defer c.rwmutex.RUnlock()
	return c
}

func (c *Config) loadFromFile(filepath string) error {
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()

	dir := path.Dir(filepath)
	fileNameWithoutExt := path.Base(filepath)
	ext := path.Ext(filepath)
	fileNameWithoutExt = fileNameWithoutExt[:len(fileNameWithoutExt)-len(ext)]

	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.SetConfigName(fileNameWithoutExt)

	if err := viper.ReadInConfig(); err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("reading config file error,filepath:%s", filepath))
	}

	err := viper.Unmarshal(c)
	if err != nil {
		return cerrs.Wrap(err)
	}
	return nil
}
