// Package core core interface
package core

type ConfigOption func(IConfig) error

type IConfig interface {
	Get(key string) any
}
