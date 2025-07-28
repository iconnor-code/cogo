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

func LoadFileConfig(filepath string, value core.IConfig) error {
	var loadErr error
	loadOnce.Do(func() {
		dir := path.Dir(filepath)
		fileNameWithoutExt := path.Base(filepath)
		ext := path.Ext(filepath)
		fileNameWithoutExt = fileNameWithoutExt[:len(fileNameWithoutExt)-len(ext)]

		viper.SetConfigType("yaml")
		viper.AddConfigPath(dir)
		viper.SetConfigName(fileNameWithoutExt)

		if err := viper.ReadInConfig(); err != nil {
			loadErr = cerrs.Wrap(err, fmt.Sprintf("reading config file error,filepath:%s", filepath))
			return
		}

		err := viper.Unmarshal(value)
		if err != nil {
			loadErr = cerrs.Wrap(err)
			return
		}
	})
	return loadErr
}
