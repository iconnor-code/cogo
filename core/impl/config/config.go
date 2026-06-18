// Package config provides a configuration management implementation for the Cogo framework.
package config

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

type Config struct {
	value           map[string]any
	filepath        string
	nacosClient     *client.Nacos
	nacosDataID     string
	nacosGroup      string
	nacosConfigType string
	viper           *viper.Viper
}

func WithFilePath(filepath string) core.ConfigOption {
	return func(c core.IConfig) error {
		c.(*Config).filepath = filepath
		return nil
	}
}

func WithNacosClient(nacos *client.Nacos) core.ConfigOption {
	return func(c core.IConfig) error {
		c.(*Config).nacosClient = nacos
		return nil
	}
}

func WithNacosConfig(dataID, group, configType string) core.ConfigOption {
	return func(c core.IConfig) error {
		config := c.(*Config)
		config.nacosDataID = dataID
		config.nacosGroup = group
		config.nacosConfigType = configType
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
	if ct.nacosClient != nil && ct.nacosDataID != "" {
		return ct.loadFromNacos()
	}
	if ct.filepath != "" {
		return ct.loadFromFile()
	}
	return nil
}

func (ct *Config) loadFromNacos() error {
	group := ct.nacosGroup
	if group == "" {
		group = ct.nacosClient.GroupName()
	}
	content, err := ct.nacosClient.ConfigClient().GetConfig(vo.ConfigParam{
		DataId: ct.nacosDataID,
		Group:  group,
	})
	if err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("reading nacos config error,data_id:%s,group:%s", ct.nacosDataID, group))
	}
	configType := ct.nacosConfigType
	if configType == "" {
		configType = configTypeFromPath(ct.nacosDataID)
	}

	ct.viper.SetConfigType(configType)
	if err := ct.viper.ReadConfig(bytes.NewBufferString(content)); err != nil {
		return cerrs.Wrap(err, fmt.Sprintf("parsing nacos config error,data_id:%s,group:%s", ct.nacosDataID, group))
	}

	err = ct.viper.Unmarshal(&ct.value)
	if err != nil {
		return cerrs.Wrap(err)
	}
	return nil
}

func configTypeFromPath(filepath string) string {
	ext := strings.TrimPrefix(path.Ext(filepath), ".")
	if ext == "" {
		return "yaml"
	}
	return ext
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
