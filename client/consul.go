// Package client
package client

import (
	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
)

type Consul struct {
	defaultClient *api.Client
}

func NewConsul(config core.IConfig) (*Consul, error) {
	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = config.GetConsul().Address

	defaultConsul, err := api.NewClient(defaultConfig)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	return &Consul{
		defaultClient: defaultConsul,
	}, nil
}

func (c *Consul) DefaultClient() *api.Client {
	return c.defaultClient
}
