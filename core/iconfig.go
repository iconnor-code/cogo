// Package core core interface
package core

type ConfigOption func(IConfig) error

type IConfig interface {
	LoadConfig() error
	GetConfig() IConfigValue
	Get(key string) any
}

type IConfigValue interface {
	Get(key string) any
}
