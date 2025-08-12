// Package registry
package registry

import (
	"context"

	"github.com/google/uuid"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry struct {
	id           string
	config       core.IConfig
	consulClient *client.Consul
	logger       core.ILogger

	etcdClient  *client.EtcdClient
	leaseTTL    int64
	leaseCancel context.CancelFunc
	leaseID     clientv3.LeaseID
}

func NewRegistry(conf core.IConfig, logger core.ILogger, opts ...core.RegistryOption) (*Registry, error) {
	registry := &Registry{
		id:     uuid.New().String(),
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
		return r.kitconsulRegister()
	}
	if r.etcdClient != nil {
		return r.etcdRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithKitConsulClient or WithEtcdClient to configure a registry client")
}

func (r *Registry) DeRegister(ctx context.Context) error {
	if r.consulClient != nil {
		return r.kitconsulDeRegister()
	}
	if r.etcdClient != nil {
		return r.etcdDeRegister(ctx)
	}
	return cerrs.New("no registry client configured, please use WithKitConsulClient or WithEtcdClient to configure a registry client")
}
