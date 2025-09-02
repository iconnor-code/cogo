// Package client
package client

import (
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/core"
)

type Consul struct {
	defaultClient kitconsul.Client
}

func NewConsul(config core.IConfig) *Consul {
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = config.Get("consul.address").(string)

	defaultConsul, err := api.NewClient(defaultConfig)
	if err != nil {
		panic(err)
	}
	defaultClient := kitconsul.NewClient(defaultConsul)
	return &Consul{
		defaultClient: defaultClient,
	}
}

func (c *Consul) DefaultClient() kitconsul.Client {
	return c.defaultClient
}
