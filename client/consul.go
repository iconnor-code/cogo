// Package client
package client

import (
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/core"
)

type Consul struct {
	registerClient kitconsul.Client
}

func NewConsul(config core.IConfig) *Consul {
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = config.Get("consul.address").(string)

	registerConfig := defaultConfig
	registerConfig.Scheme = config.Get("registry.schema").(string)
	registerAPIClient, err := api.NewClient(registerConfig)
	if err != nil {
		panic(err)
	}
	registerClient := kitconsul.NewClient(registerAPIClient)
	return &Consul{
		registerClient: registerClient,
	}
}

func (c *Consul) GetRegisterClient() kitconsul.Client {
	return c.registerClient
}
