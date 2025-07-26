package registry

import (
	"errors"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/etcd"
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
