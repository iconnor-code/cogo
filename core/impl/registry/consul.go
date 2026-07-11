package registry

import (
	"context"
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/client"
	"go.uber.org/zap"
)

func WithConsulClient(c *client.Consul) Option {
	return func(r *Registry) error {
		r.consulClient = c
		return nil
	}
}

func (r *Registry) consulRegister(ctx context.Context) error {
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
	opts := consul.ServiceRegisterOpts{}.WithContext(ctx)
	return r.consulClient.DefaultClient().Agent().ServiceRegisterOpts(serviceRegistration, opts)
}

func (r *Registry) consulDeRegister(ctx context.Context) error {
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}
	err = r.consulClient.DefaultClient().Agent().ServiceDeregisterOpts(instanceID, (&consul.QueryOptions{}).WithContext(ctx))
	if err != nil {
		return err
	}
	r.logger.Info("consul deregister", zap.String("id", instanceID))
	return nil
}
