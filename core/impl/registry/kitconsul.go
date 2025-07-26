// Package registry
package registry

import (
	"context"
	"fmt"

	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
)

type KitConsulRegistry struct {
	id     string
	conf   core.IConfig
	consul kitconsul.Client
	logger core.ILogger
}

func NewKitConsulRegistry(conf core.IConfig, logger core.ILogger) *KitConsulRegistry {
	id := uuid.New().String()
	consulConf := consul.DefaultConfig()
	consulConf.Address = conf.Get("consul.address").(string)
	consul, err := consul.NewClient(consulConf)
	if err != nil {
		logger.Panic("new consul client error", zap.Error(err))
	}
	return &KitConsulRegistry{
		id:     id,
		logger: logger,
		consul: kitconsul.NewClient(consul),
	}
}

func (kcd *KitConsulRegistry) Register(ctx context.Context) error {
	conf := kcd.conf.Get("registry").(core.IConfigValue)
	healthConf := conf.Get("health_check").(core.IConfigValue)
	serviceRegistration := &consul.AgentServiceRegistration{
		ID:   kcd.id,
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
	return kcd.consul.Register(serviceRegistration)
}

func (kcd *KitConsulRegistry) DeRegister(ctx context.Context) error {
	return kcd.consul.Deregister(&consul.AgentServiceRegistration{
		ID: kcd.id,
	})
}
