package registry

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
)

func WithConsulClient(c *client.Consul) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).consulClient = c
		return nil
	}
}

func (r *Registry) consulRegister() error {
	name, address, port, err := r.serviceConfig()
	if err != nil {
		return err
	}
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}
	serviceRegistration := &consul.AgentServiceRegistration{
		ID:      instanceID,
		Name:    name,
		Address: address, Port: port,
		Check: &consul.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s:%d/%s", address, port, name),
			Interval: r.config.GetRegistry().HealthCheck.Interval,
			Timeout:  r.config.GetRegistry().HealthCheck.Timeout,
			Status:   consul.HealthPassing,
		},
	}
	r.logger.Info("consul register", zap.String("id", serviceRegistration.ID), zap.String("name", serviceRegistration.Name), zap.String("address", serviceRegistration.Address), zap.Int("port", serviceRegistration.Port))
	return r.consulClient.DefaultClient().Agent().ServiceRegister(serviceRegistration)
}

func (r *Registry) consulDeRegister() error {
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}
	err = r.consulClient.DefaultClient().Agent().ServiceDeregister(instanceID)
	if err != nil {
		return err
	}
	r.logger.Info("consul deregister", zap.String("id", instanceID))
	return nil
}
