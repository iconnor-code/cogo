// Package registry
package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	if conf == nil {
		return nil, cerrs.New("registry config is required")
	}
	if logger == nil {
		return nil, cerrs.New("registry logger is required")
	}
	registry := &Registry{
		logger: logger,
		config: conf,
	}
	for _, opt := range opts {
		if err := opt(registry); err != nil {
			return nil, err
		}
	}
	if registry.consulClient == nil && registry.etcdClient == nil {
		return nil, cerrs.New("exactly one registry client is required")
	}
	if registry.consulClient != nil && registry.etcdClient != nil {
		return nil, cerrs.New("consul and etcd registry clients cannot be configured together")
	}
	if registry.etcdClient != nil && registry.leaseTTL <= 0 {
		return nil, cerrs.New("etcd registry lease ttl must be positive")
	}
	return registry, nil
}

// NewDefault builds the registry selected by the framework's current default
// policy. Keeping this policy in the registry package lets transport servers
// accept any IRegistry implementation without knowing how it is constructed.
func NewDefault(conf core.IConfig, logger core.ILogger) (core.IRegistry, error) {
	if conf == nil {
		return nil, cerrs.New("registry config is required")
	}
	if logger == nil {
		return nil, cerrs.New("registry logger is required")
	}
	registryConf := conf.GetRegistry()
	if strings.TrimSpace(conf.GetConsul().Address) == "" {
		return nil, nil
	}
	if strings.TrimSpace(registryConf.Name) == "" {
		return nil, cerrs.New("registry name is required when consul is configured")
	}
	if strings.TrimSpace(registryConf.Address) == "" {
		return nil, cerrs.New("registry address is required when consul is configured")
	}
	if registryConf.Port <= 0 || registryConf.Port > 65535 {
		return nil, cerrs.New("registry port must be between 1 and 65535 when consul is configured")
	}
	if err := validatePositiveDuration("registry health check interval", registryConf.HealthCheck.Interval); err != nil {
		return nil, err
	}
	if err := validatePositiveDuration("registry health check timeout", registryConf.HealthCheck.Timeout); err != nil {
		return nil, err
	}
	consul, err := client.NewConsul(conf)
	if err != nil {
		return nil, err
	}
	return NewRegistry(conf, logger, WithConsulClient(consul))
}

func validatePositiveDuration(name, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", name, err)
	}
	if duration <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
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
