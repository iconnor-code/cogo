package config

import (
	"log"
	"path"

	"github.com/spf13/viper"
)

type Conf struct {
	Mode      string         `mapstructure:"mode"`
	BizID     int32          `mapstructure:"biz_id"`
	Grpc      GrpcConfig     `mapstructure:"grpc"`
	Log       LogConfig      `mapstructure:"log"`
	HttpProxy HttpConfig     `mapstructure:"http_proxy"`
	Metrics   MetricsConfig  `mapstructure:"metrics"`
	Mysql     MysqlConfig    `mapstructure:"mysql"`
	Redis     RedisConfig    `mapstructure:"redis"`
	Smtp      SmtpConfig     `mapstructure:"smtp"`
	JwtToken  JwtTokenConfig `mapstructure:"jwt_token"`
	Etcd      EtcdConfig     `mapstructure:"etcd"`
	Registry  RegistryConfig `mapstructure:"registry"`
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}

type RegistryConfig struct {
	Hostname string `mapstructure:"hostname"`
	Key      string `mapstructure:"key"`
}

type GrpcConfig struct {
	Listen string `mapstructure:"listen"`
}

type HttpConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Listen   string `mapstructure:"listen"`
	SSL      bool   `mapstructure:"ssl"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
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

func NewConfig(filepath string) *Conf {
	return loadConfigFromFile(filepath)
}

func loadConfigFromFile(filepath string) *Conf {
	if filepath == "" {
		return nil
	}

	dir := path.Dir(filepath)
	fileNameWithoutExt := path.Base(filepath)
	ext := path.Ext(filepath)
	fileNameWithoutExt = fileNameWithoutExt[:len(fileNameWithoutExt)-len(ext)]

	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.SetConfigName(fileNameWithoutExt)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: path:%s, err:%s", filepath, err)
	}

	conf := &Conf{}
	err := viper.Unmarshal(conf)
	if err != nil {
		log.Fatalf("Unable to decode config file into struct, %v", err)
	}

	return conf
}
