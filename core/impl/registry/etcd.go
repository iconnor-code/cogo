package registry

import (
	"context"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type KitEtcdRetistry struct {
	config      core.IConfig
	registryKey string
}

func WithEtcdClient(etcd *client.EtcdClient) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).etcdClient = etcd
		return nil
	}
}

func WithEtcdRegisterLeaseTTL(ttl int64) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).leaseTTL = ttl
		return nil
	}
}

func (r *Registry) etcdRegister(ctx context.Context) error {
	lease, err := r.etcdClient.Grant(ctx, r.leaseTTL)
	if err != nil {
		return cerrs.Wrap(err)
	}
	r.leaseID = lease.ID

	_, err = r.etcdClient.Put(ctx, r.config.Get("registry.name").(string), r.instanceID, clientv3.WithLease(lease.ID))
	if err != nil {
		return cerrs.Wrap(err)
	}
	r.keepAlive(ctx)

	r.logger.Info("etcd register", zap.String("key", r.config.Get("registry.name").(string)), zap.String("value", r.instanceID), zap.Int64("lease_id", int64(r.leaseID)), zap.Int64("lease_ttl", r.leaseTTL))

	return nil
}

func (r *Registry) keepAlive(ctx context.Context) error {
	var err *error

	ctx, cancel := context.WithCancel(ctx)
	r.leaseCancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(r.leaseTTL-1) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_, doneerr := r.etcdClient.Revoke(context.Background(), r.leaseID)
				if doneerr != nil {
					*err = doneerr
					return
				}
				return
			case <-ticker.C:
				_, keepAliveErr := r.etcdClient.KeepAliveOnce(ctx, r.leaseID)
				if keepAliveErr != nil {
					*err = keepAliveErr
					return
				}
			}
		}
	}()

	return *err
}

func (r *Registry) etcdDeRegister(ctx context.Context) error {
	if r.leaseCancel != nil {
		r.leaseCancel()
	}
	_, err := r.etcdClient.Delete(ctx, r.config.Get("registry.name").(string))
	if err != nil {
		return err
	}
	return nil
}
