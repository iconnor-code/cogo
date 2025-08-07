// Package client
package client

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/core"
)

type Consul struct {
	config *consul.Config
	client *consul.Client
}

func NewConsul(config core.IConfig) *Consul {
	consulConfig := &consul.Config{
		Address: config.Get("consul.address").(string),
		Scheme:  config.Get("consul.scheme").(string),
	}
	consul, err := consul.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	return &Consul{
		config: consulConfig,
		client: consul,
	}
}
