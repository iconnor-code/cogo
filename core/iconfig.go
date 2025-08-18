// Package core core interface
package core

type IConfVal interface {
	Get(key string) any
}

type (
	ConfigOption func(IConfig) error
	IConfig      interface {
		Get(key string) any
		ReLoad() error
	}
)
