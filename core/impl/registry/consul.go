package registry

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
)

func WithKitConsulClient(c *client.Consul) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).consulClient = c
		return nil
	}
}

func (r *Registry) kitconsulRegister() error {
	serviceRegistration := &consul.AgentServiceRegistration{
		ID:      r.id,
		Name:    r.config.Get("registry.name").(string),
		Address: r.config.Get("registry.address").(string), Port: r.config.Get("registry.port").(int),
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", r.config.Get("registry.address").(string), r.config.Get("registry.port").(int)),
			Interval: r.config.Get("registry.health_check.interval").(string),
			Timeout:  r.config.Get("registry.health_check.timeout").(string),
			Status:   "passing",
		},
	}
	r.logger.Info("consul register", zap.String("id", serviceRegistration.ID), zap.String("name", serviceRegistration.Name), zap.String("address", serviceRegistration.Address), zap.Int("port", serviceRegistration.Port))
	return r.consulClient.GetRegisterClient().Register(serviceRegistration)
}

func (r *Registry) kitconsulDeRegister() error {
	return r.consulClient.GetRegisterClient().Deregister(&consul.AgentServiceRegistration{
		ID: r.id,
	})
}
