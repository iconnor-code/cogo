package config

import (
	"fmt"
	"path"
	"sync"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/spf13/viper"
)

var loadOnce sync.Once

type FileConfig struct {
	filepath string
	value    core.IConfigValue
}

func WithFilepath(filepath string) core.ConfigOption {
	return func(c core.IConfig) error {
		if filepath == "" {
			return cerrs.New("config filepath is required")
		}
		conf, ok := c.(*FileConfig)
		if !ok {
			return cerrs.New("config is not a FileConfig")
		}
		conf.filepath = filepath
		return nil
	}
}

func WithConfigValue(value core.IConfigValue) core.ConfigOption {
	return func(c core.IConfig) error {
		conf, ok := c.(*FileConfig)
		if !ok {
			return cerrs.New("config is not a FileConfig")
		}
		conf.value = value
		return nil
	}
}

func NewConfig(opts ...core.ConfigOption) (*FileConfig, error) {
	conf := &FileConfig{}
	for _, opt := range opts {
		if err := opt(conf); err != nil {
			return nil, err
		}
	}
	if err := conf.LoadConfig(); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *FileConfig) LoadConfig() error {
	var loadErr error
	loadOnce.Do(func() {
		dir := path.Dir(c.filepath)
		fileNameWithoutExt := path.Base(c.filepath)
		ext := path.Ext(c.filepath)
		fileNameWithoutExt = fileNameWithoutExt[:len(fileNameWithoutExt)-len(ext)]

		viper.SetConfigType("yaml")
		viper.AddConfigPath(dir)
		viper.SetConfigName(fileNameWithoutExt)

		if err := viper.ReadInConfig(); err != nil {
			loadErr = cerrs.Wrap(err, fmt.Sprintf("reading config file error,filepath:%s", c.filepath))
			return
		}

		err := viper.Unmarshal(c.value)
		if err != nil {
			loadErr = cerrs.Wrap(err)
			return
		}
	})
	return loadErr
}

func (c *FileConfig) GetConfig() core.IConfigValue {
	return c.value
}

func (c *FileConfig) Get(key string) any {
	return c.value.Get(key)
}
