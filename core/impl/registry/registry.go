// Package registry
package registry

import (
	"context"
	"fmt"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry struct {
	instanceID   string
	config       core.IConfig
	consulClient *client.Consul
	logger       core.ILogger

	etcdClient  *client.EtcdClient
	leaseTTL    int64
	leaseCancel context.CancelFunc
	leaseID     clientv3.LeaseID
}

type Option func(*Registry) error

func NewRegistry(conf core.IConfig, logger core.ILogger, opts ...Option) (*Registry, error) {
	registry := &Registry{
		logger: logger,
		config: conf,
	}
	for _, opt := range opts {
		if err := opt(registry); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// NewDefault builds the registry selected by the framework's current default
// policy. Keeping this policy in the registry package lets transport servers
// accept any IRegistry implementation without knowing how it is constructed.
func NewDefault(conf core.IConfig, logger core.ILogger) (core.IRegistry, error) {
	registryConf := conf.GetRegistry()
	if conf.GetConsul().Address == "" || registryConf.Name == "" ||
		registryConf.Address == "" || registryConf.Port == 0 {
		return nil, nil
	}
	consul, err := client.NewConsul(conf)
	if err != nil {
		return nil, err
	}
	return NewRegistry(conf, logger, WithConsulClient(consul))
}

func (r *Registry) Register(ctx context.Context) error {
	if r.consulClient != nil {
		return r.consulRegister(ctx)
	}
	if r.etcdClient != nil {
		return r.etcdRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithConsulClient or WithEtcdClient to configure a registry client")
}

func (r *Registry) DeRegister(ctx context.Context) error {
	if r.consulClient != nil {
		return r.consulDeRegister(ctx)
	}
	if r.etcdClient != nil {
		return r.etcdDeRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithConsulClient or WithEtcdClient to configure a registry client")
}

func (r *Registry) getInstanceID() (string, error) {
	if r.instanceID != "" {
		return r.instanceID, nil
	}
	registryConf := r.config.GetRegistry()
	name := registryConf.Name
	address := registryConf.Address
	port := registryConf.Port
	r.instanceID = fmt.Sprintf("%s-%s:%d",
		name,
		address,
		port,
	)
	return r.instanceID, nil
}

func (r *Registry) serviceConfig() (name string, address string, port int, err error) {
	registryConf := r.config.GetRegistry()
	return registryConf.Name, registryConf.Address, registryConf.Port, nil
}
