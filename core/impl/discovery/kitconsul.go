// Package discovery idiscovery implement
package discovery

import (
	"context"

	kitconsul "github.com/go-kit/kit/sd/consul"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
)

type KitConsulDiscovery struct {
	conf        core.IConfig
	logger      core.ILogger
	consul      kitconsul.Client
	loadBalance core.IDiscoveryLoadBalance
}

func WithDLBOption(dlb core.IDiscoveryLoadBalance) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		d.(*KitConsulDiscovery).loadBalance = dlb
		return nil
	}
}

func NewKitConsulDiscovery(conf core.IConfig, logger core.ILogger) *KitConsulDiscovery {
	consulConf := consul.DefaultConfig()
	consulConf.Address = conf.Get("consul.address").(string)
	consul, err := consul.NewClient(consulConf)
	if err != nil {
		logger.Panic("consul discovery new consul client error", zap.Error(err))
	}
	return &KitConsulDiscovery{
		logger: logger,
		consul: kitconsul.NewClient(consul),
	}
}

func (kcd *KitConsulDiscovery) GetServer(ctx context.Context, serverName string) (core.ServerInstance, error) {
	instance, err := kcd.loadBalance.GetInstance(ctx, serverName)
	if err != nil {
		return nil, err
	}
	if instance != nil {
		return instance, nil
	}
	kcd.watchServersByName(ctx, serverName)

	servers, _, err := kcd.consul.Service(serverName, "", true, nil)
	if err != nil {
		return nil, err
	}
	serverInstances := make([]core.ServerInstance, len(servers))
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
	err = kcd.loadBalance.RefreshAll(ctx, serverName, serverInstances)
	if err != nil {
		return nil, err
	}
	return kcd.loadBalance.GetInstance(ctx, serverName)
}

func (kcd *KitConsulDiscovery) watchServersByName(ctx context.Context, serverName string) {
	// 监控实例变化
	go func() {
		params := make(map[string]any)
		params["type"] = "service"
		params["service"] = serverName
		plan, err := watch.Parse(params)
		if err != nil {
			kcd.logger.Error("discovery watch servers error", zap.String("serverName", serverName), zap.Error(err))
			return
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
				kcd.loadBalance.RefreshAll(ctx, serverName, nil)
				return
			}
			healthInstances := make([]core.ServerInstance, len(v))
			for _, service := range v {
				if service.Checks.AggregatedStatus() == consul.HealthPassing {
					healthInstances = append(healthInstances, &ServerInstance{
						Name:    service.Service.Service,
						Address: service.Service.Address,
						Port:    service.Service.Port,
					})
				}
			}
			kcd.loadBalance.RefreshAll(ctx, serverName, healthInstances)
		}
		defer plan.Stop()
		plan.Run(kcd.conf.Get("discovery.address").(string))
	}()
}
