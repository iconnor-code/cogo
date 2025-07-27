// Package discovery idiscovery implement
package discovery

import (
	"context"
	"fmt"
	"sync"

	kitconsul "github.com/go-kit/kit/sd/consul"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
)

type KitConsulDiscovery struct {
	conf        core.IConfig
	consul      kitconsul.Client
	loadBalance core.DiscoveryLoadBalance
	instances   sync.Map
	lock        sync.RWMutex
}

func WithDLBOption(loadBalance core.DiscoveryLoadBalance) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		d.(*KitConsulDiscovery).loadBalance = loadBalance
		return nil
	}
}

func NewKitConsulDiscovery(conf core.IConfig) (*KitConsulDiscovery, error) {
	consulConf := consul.DefaultConfig()
	consulConf.Address = conf.Get("consul.address").(string)
	consul, err := consul.NewClient(consulConf)
	if err != nil {
		return nil, cerrs.Wrap("new consul client error", err)
	}
	return &KitConsulDiscovery{
		consul: kitconsul.NewClient(consul),
	}, nil
}

func (kcd *KitConsulDiscovery) GetServer(ctx context.Context, serverName string) (core.IServerInstance, error) {
	instances, ok := kcd.instances.Load(serverName)
	if ok {
		if dlb, ok := instances.(core.IDiscoveryLoadBalance); ok {
			return dlb.GetInstance(ctx, serverName)
		} else {
			return nil, cerrs.New(fmt.Sprintf("wrong type of service discovery instances, servername:%s", serverName))
		}
	}

	kcd.lock.Lock()
	defer kcd.lock.Unlock()

	kcd.watchServersByName(ctx, serverName)

	servers, _, err := kcd.consul.Service(serverName, "", true, nil)
	if err != nil {
		return nil, cerrs.Wrap("fail to find consul discovery service", err)
	}
	serverInstances := make([]core.IServerInstance, len(servers))
	for i, server := range servers {
		serverInstances[i] = &ServerInstance{
			ID:      server.Service.ID,
			Name:    server.Service.Service,
			Address: server.Service.Address,
			Port:    server.Service.Port,
			Tags:    server.Service.Tags,
			Meta:    server.Service.Meta,
		}
	}
	dlb := kcd.getLoadBalance(serverName)
	err = dlb.RefreshInstance(ctx, serverName, serverInstances)
	if err != nil {
		return nil, cerrs.Wrap("fail to refresh consul instances", err)
	}
	kcd.instances.Store(serverName, dlb)

	return dlb.GetInstance(ctx, serverName)
}

func (kcd *KitConsulDiscovery) watchServersByName(ctx context.Context, serverName string) {
	// 监控实例变化
	go func() error {
		params := make(map[string]any)
		params["type"] = "service"
		params["service"] = serverName
		plan, err := watch.Parse(params)
		if err != nil {
			return cerrs.Wrap("parsh error", err)
		}
		plan.Handler = func(idx uint64, data any) {
			if data == nil {
				return
			}
			v, ok := data.([]*consul.ServiceEntry)
			if !ok {
				return
			}
			if len(v) == 0 {
				kcd.instances.Store(serverName, nil)
				return
			}
			healthInstances := make([]core.IServerInstance, len(v))
			for _, service := range v {
				if service.Checks.AggregatedStatus() == consul.HealthPassing {
					healthInstances = append(healthInstances, &ServerInstance{
						Name:    service.Service.Service,
						Address: service.Service.Address,
						Port:    service.Service.Port,
					})
				}
			}
			dlb := kcd.getLoadBalance(serverName)
			err := dlb.RefreshInstance(ctx, serverName, healthInstances)
			if err != nil {
				return
			}
			kcd.instances.Store(serverName, dlb)
		}
		defer plan.Stop()
		plan.Run(kcd.conf.Get("discovery.address").(string))
		return nil
	}()
}

func (kcd *KitConsulDiscovery) getLoadBalance(serverName string) core.IDiscoveryLoadBalance {
	if dlb, ok := kcd.instances.Load(serverName); ok {
		return dlb.(core.IDiscoveryLoadBalance)
	}

	switch kcd.loadBalance {
	case core.Random:
		return NewRandomLoadBalance()
	case core.RoundRobin:
		return NewRoundRobinLoadBalance()
	default:
		return NewRandomLoadBalance()
	}
}
