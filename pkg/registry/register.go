package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/etcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type KitEtcdDiscovery struct {
	config      map[string]any
	etcd        *etcd.EtcdClient
	leaseTTL    int64
	leaseID     clientv3.LeaseID
	cancel      context.CancelFunc
	registryKey string
}

func WithEtcdRegisterConfig(config core.IConfig) core.RegistryOption {
	return func(r core.IRegistry) error {
		confMap := config.Get("registry").(map[string]any)
		if confMap == nil {
			return errors.New("registry config is not found")
		}
		registry := r.(*KitEtcdDiscovery)
		registry.config = confMap
		return nil
	}
}

func WithEtcdRegisterEtcdClient(etcd *etcd.EtcdClient) core.RegistryOption {
	return func(r core.IRegistry) error {
		registry := r.(*KitEtcdDiscovery)
		registry.etcd = etcd
		return nil
	}
}

func WithEtcdRegisterLeaseTTL(ttl int64) core.RegistryOption {
	return func(r core.IRegistry) error {
		registry := r.(*KitEtcdDiscovery)
		registry.leaseTTL = ttl
		return nil
	}
}

func NewEtcdRegister(opts ...core.RegistryOption) (*KitEtcdDiscovery, error) {
	registry := &KitEtcdDiscovery{}
	for _, opt := range opts {
		opt(registry)
	}
	return registry, nil
}

func (r *KitEtcdDiscovery) Register(ctx context.Context, serverName string, serverAddr string) error {
	r.registryKey = serverName
	lease, err := r.etcd.Grant(ctx, r.leaseTTL)
	if err != nil {
		return cerr.WithStack(err)
	}
	r.leaseID = lease.ID

	_, err = r.etcd.Put(ctx, r.registryKey, serverAddr, clientv3.WithLease(lease.ID))
	if err != nil {
		return cerr.WithStack(err)
	}
	r.keepAlive(ctx)

	return nil
}

func (r *KitEtcdDiscovery) keepAlive(ctx context.Context) error {
	var err *error

	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(r.leaseTTL-1) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_, doneerr := r.etcd.Revoke(context.Background(), r.leaseID)
				if doneerr != nil {
					*err = doneerr
					return
				}
				return
			case <-ticker.C:
				_, keepAliveErr := r.etcd.KeepAliveOnce(ctx, r.leaseID)
				if keepAliveErr != nil {
					*err = keepAliveErr
					return
				}
			}
		}
	}()

	return *err
}

func (r *KitEtcdDiscovery) DeRegister(ctx context.Context, serverName string, serverAddr string) error {
	if r.cancel != nil {
		r.cancel()
	}
	_, err := r.etcd.Delete(ctx, r.registryKey)
	if err != nil {
		return err
	}
	return nil
}

func (r *KitEtcdDiscovery) GetServer(serverName string) (any, error) {
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
