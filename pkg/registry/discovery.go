package registry

import (
	"errors"
	"fmt"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/etcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

type EtcdDiscover struct {
	config map[string]any
	etcd   *etcd.EtcdClient
}

func NewGrpcDiscover(opts ...core.DiscoveryOption) (*EtcdDiscover, error) {
	d := &EtcdDiscover{}
	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (g *EtcdDiscover) GetGrpcClientConn(serverName string) (any, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s/%d", "etcd://", serverName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithResolvers(g.resolver),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)),
	)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return conn, nil
}

func WithEtcdDiscoverConfig(config core.IConfig) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		confMap := config.Get("discovery").(map[string]any)
		if confMap == nil {
			return errors.New("discovery config is not found")
		}
		discovery := d.(*EtcdDiscover)
		discovery.config = confMap
		return nil
	}
}

func WithEtcdDiscoverEtcdClient(etcd *etcd.EtcdClient) core.DiscoveryOption {
	return func(d core.IDiscovery) error {
		discovery := d.(*EtcdDiscover)
		discovery.etcd = etcd
		return nil
	}
}
