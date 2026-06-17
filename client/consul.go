// Package client
package client

import (
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
)

type Consul struct {
	defaultClient kitconsul.Client
}

func NewConsul(config core.IConfig) (*Consul, error) {
	address, err := core.GetString(config, "consul.address")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = address

	defaultConsul, err := api.NewClient(defaultConfig)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	defaultClient := kitconsul.NewClient(defaultConsul)
	return &Consul{
		defaultClient: defaultClient,
	}, nil
}

func (c *Consul) DefaultClient() kitconsul.Client {
	return c.defaultClient
}
