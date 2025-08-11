package config

import (
	"fmt"
	"path"
	"sync"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/spf13/viper"
)

var (
	configOnce     sync.Once
	configInstance *Config
)

type Config struct {
	rwmutex  sync.RWMutex
	value    core.IConfVal
	filepath string
}

func WithFilePath(filepath string) core.ConfigOption {
	return func(c core.IConfig) error {
		c.(*Config).filepath = filepath
		return nil
	}
}

func NewConfig(val core.IConfVal, opts ...core.ConfigOption) (*Config, error) {
	var err error
	configOnce.Do(func() {
		configInstance = &Config{}
		for _, opt := range opts {
			err = opt(configInstance)
			if err != nil {
				err = cerrs.Wrap(err, "applying config option error")
			}
		}
	})
	return configInstance, err
}

func (ct *Config) Get(key string) any {
	ct.rwmutex.RLock()
	defer ct.rwmutex.RUnlock()

	return ct.value.Get(key)
}

func (ct *Config) GetVal() core.IConfVal {
	ct.rwmutex.RLock()
	defer ct.rwmutex.RUnlock()

	return ct.value
}

func (ct *Config) ReLoad() error {
	ct.rwmutex.Lock()
	defer ct.rwmutex.Unlock()
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

	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.SetConfigName(fileNameWithoutExt)

	if err := viper.ReadInConfig(); err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("reading config file error,filepath:%s", ct.filepath))
	}

	err := viper.Unmarshal(ct.value)
	if err != nil {
		return cerrs.Wrap(err)
	}
	return nil
}
