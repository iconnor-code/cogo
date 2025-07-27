package discovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/etcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

type KitEtcdDiscovery struct {
	config map[string]any
	etcd   *etcd.EtcdClient
}

func NewKitEtcdDiscovery(opts ...core.DiscoveryOption) (*KitEtcdDiscovery, error) {
	d := &KitEtcdDiscovery{}
	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func WithEtcdDiscoverConfig(config core.IConfig) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		confMap := config.Get("discovery").(map[string]any)
		if confMap == nil {
			return errors.New("discovery config is not found")
		}
		discovery := d.(*KitEtcdDiscovery)
		discovery.config = confMap
		return nil
	}
}

func WithEtcdDiscoverEtcdClient(etcd *etcd.EtcdClient) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		discovery := d.(*KitEtcdDiscovery)
		discovery.etcd = etcd
		return nil
	}
}

func (ked *KitEtcdDiscovery) GetServer(ctx context.Context, serverName string) (core.IServerInstance, error) {
	_, err := grpc.NewClient(fmt.Sprintf("%s/%s", "etcd://", serverName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithResolvers(g.resolver),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)),
	)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return &ServerInstance{}, nil
}
