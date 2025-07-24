package discovery

import (
	"context"
	"fmt"
	"sync"

	kitconsul "github.com/go-kit/kit/sd/consul"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/iconnor-code/cogo/core"
)

// 服务发现
type ConsulDiscovery struct {
	consulConfig *consul.Config
	kitConsul    kitconsul.Client
	instanceMap  sync.Map
	mutex        sync.Mutex
}

func NewConsulDiscovery(consulConfig *consul.Config) *ConsulDiscovery {
	consul, err := consul.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	return &ConsulDiscovery{
		consulConfig: consulConfig,
		kitConsul:    kitconsul.NewClient(consul),
	}
}

func (cd *ConsulDiscovery) Register(ctx context.Context, instance core.ServiceInstance) error {
	serviceRegistration := &consul.AgentServiceRegistration{
		ID:   instance.GetName(),
		Name: instance.GetName(),
		Tags: []string{"grpc"},
		Port: 8080,
		Meta: map[string]string{
			"version": "1.0.0",
		},
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", instance.GetName(), 8080),
			Interval: "10s",
			Timeout:  "5s",
			Status:   "passing",
		},
	}
	return cd.kitConsul.Register(serviceRegistration)
}

func (cd *ConsulDiscovery) Deregister(ctx context.Context, instance core.ServiceInstance) error {
	return cd.kitConsul.Deregister(&consul.AgentServiceRegistration{
		ID: instance.GetName(),
	})
}

func (cd *ConsulDiscovery) Service(ctx context.Context, serverName string) ([]core.ServiceInstance, error) {
	instances, ok := cd.instanceMap.Load(serverName)
	if ok {
		return instances.([]core.ServiceInstance), nil
	}
	cd.mutex.Lock()
	defer cd.mutex.Unlock()
	instances, ok = cd.instanceMap.Load(serverName)
	if ok {
		return instances.([]core.ServiceInstance), nil
	}

	// 监控实例变化
	go func() {
		params := make(map[string]interface{})
		params["type"] = "service"
		params["service"] = serverName
		plan, err := watch.Parse(params)
		if err != nil {
			return
		}
		plan.Handler = func(idx uint64, data interface{}) {
			if data == nil {
				return
			}
			v, ok := data.([]*consul.ServiceEntry)
			if !ok {
				return
			}
			if len(v) == 0 {
				cd.instanceMap.Store(serverName, nil)
				return
			}
			healthInstances := make([]core.ServiceInstance, len(v))
			for _, service := range v {
				if service.Checks.AggregatedStatus() == consul.HealthPassing {
					healthInstances = append(healthInstances, &ServiceInstance{
						ID:   service.Service.ID,
						Name: service.Service.Service,
						Addr: service.Service.Address,
						Port: service.Service.Port,
					})
				}
			}
			cd.instanceMap.Store(serverName, healthInstances)
		}
		defer plan.Stop()
		plan.Run(cd.consulConfig.Address)
	}()

	services, _, err := cd.kitConsul.Service(serverName, "", true, nil)
	if err != nil {
		cd.instanceMap.Store(serverName, nil)
		return nil, err
	}
	serviceInstances := make([]core.ServiceInstance, len(services))
	for i := 0; i < len(services); i++ {
		serviceInstances[i] = &ServiceInstance{
			ID:   services[i].Service.ID,
			Name: services[i].Service.Service,
			Addr: services[i].Service.Address,
			Port: services[i].Service.Port,
			Tags: services[i].Service.Tags,
			Meta: services[i].Service.Meta,
		}
	}
	cd.instanceMap.Store(serverName, serviceInstances)
	return serviceInstances, nil
}
