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
	nacosClient  *client.Nacos
	logger       core.ILogger

	etcdClient  *client.EtcdClient
	leaseTTL    int64
	leaseCancel context.CancelFunc
	leaseID     clientv3.LeaseID
}

func NewRegistry(conf core.IConfig, logger core.ILogger, opts ...core.RegistryOption) (*Registry, error) {
	registry := &Registry{
		logger: logger,
		config: conf,
	}
	for _, opt := range opts {
		optErr := opt(registry)
		if optErr != nil {
			return nil, optErr
		}
	}
	return registry, nil
}

func (r *Registry) Register(ctx context.Context) error {
	if r.consulClient != nil {
		return r.consulRegister()
	}
	if r.nacosClient != nil {
		return r.nacosRegister(ctx)
	}
	if r.etcdClient != nil {
		return r.etcdRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithConsulClient, WithNacosClient or WithEtcdClient to configure a registry client")
}

func (r *Registry) DeRegister(ctx context.Context) error {
	if r.consulClient != nil {
		return r.consulDeRegister()
	}
	if r.nacosClient != nil {
		return r.nacosDeRegister(ctx)
	}
	if r.etcdClient != nil {
		return r.etcdDeRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithConsulClient, WithNacosClient or WithEtcdClient to configure a registry client")
}

func (r *Registry) getInstanceID() (string, error) {
	if r.instanceID != "" {
		return r.instanceID, nil
	}
	name, err := core.GetString(r.config, "registry.name")
	if err != nil {
		return "", cerrs.Wrap(err)
	}
	address, err := core.GetString(r.config, "registry.address")
	if err != nil {
		return "", cerrs.Wrap(err)
	}
	port, err := core.GetInt(r.config, "registry.port")
	if err != nil {
		return "", cerrs.Wrap(err)
	}
	r.instanceID = fmt.Sprintf("%s-%s:%d",
		name,
		address,
		port,
	)
	return r.instanceID, nil
}

func (r *Registry) serviceConfig() (name string, address string, port int, err error) {
	name, err = core.GetString(r.config, "registry.name")
	if err != nil {
		return "", "", 0, cerrs.Wrap(err)
	}
	address, err = core.GetString(r.config, "registry.address")
	if err != nil {
		return "", "", 0, cerrs.Wrap(err)
	}
	port, err = core.GetInt(r.config, "registry.port")
	if err != nil {
		return "", "", 0, cerrs.Wrap(err)
	}
	return name, address, port, nil
}
