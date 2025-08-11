package registry

import (
	"context"
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
)

func WithKitConsulClient(c *client.Consul) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).consulClient = c
		return nil
	}
}

func (r *Registry) kitconsulRegister(ctx context.Context) error {
	conf := r.config.Get("registry").(core.IConfig)
	healthConf := conf.Get("health_check").(core.IConfig)
	serviceRegistration := &consul.AgentServiceRegistration{
		ID:   r.id,
		Name: conf.Get("name").(string),
		Tags: conf.Get("tags").([]string),
		Port: conf.Get("port").(int),
		Meta: conf.Get("meta").(map[string]string),
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", conf.Get("address"), conf.Get("port")),
			Interval: healthConf.Get("interval").(string),
			Timeout:  healthConf.Get("timeout").(string),
			Status:   "passing",
		},
	}
	return r.consulClient.getKitConsul().Register(serviceRegistration)
}

func (r *Registry) kitconsulDeRegister(ctx context.Context) error {
	return r.consulClient.Deregister(&consul.AgentServiceRegistration{
		ID: r.id,
	})
}
