package registry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"go.uber.org/zap"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func (r *Registry) etcdRegistryKey() (string, error) {
	instanceID, err := r.getInstanceID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/services/%s/%s", r.config.GetRegistry().Name, instanceID), nil
}

func WithEtcdClient(etcd *client.EtcdClient) Option {
	return func(r *Registry) error {
		r.etcdClient = etcd
		return nil
	}
}

func WithEtcdRegisterLeaseTTL(ttl int64) Option {
	return func(r *Registry) error {
		r.leaseTTL = ttl
		return nil
	}
}

func (r *Registry) etcdRegister(ctx context.Context) error {
	lease, err := r.etcdClient.Grant(ctx, r.leaseTTL)
	if err != nil {
		return cerrs.Wrap(err)
	}
	r.leaseID = lease.ID
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}
	key, err := r.etcdRegistryKey()
	if err != nil {
		return err
	}

	_, err = r.etcdClient.Put(ctx, key, instanceID, clientv3.WithLease(lease.ID))
	if err != nil {
		return cerrs.Wrap(err)
	}
	r.keepAlive(ctx)

	r.logger.Info("etcd register", zap.String("key", key), zap.String("value", instanceID), zap.Int64("lease_id", int64(r.leaseID)), zap.Int64("lease_ttl", r.leaseTTL))

	return nil
}

func (r *Registry) keepAlive(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.leaseCancel = cancel

	go func() {
		interval := r.leaseTTL - 1
		if interval <= 0 {
			interval = 1
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, keepAliveErr := r.etcdClient.KeepAliveOnce(ctx, r.leaseID)
				if keepAliveErr != nil {
					r.logger.Error("etcd keepalive failed", zap.Error(keepAliveErr), zap.Int64("lease_id", int64(r.leaseID)))
					return
				}
			}
		}
	}()
}

func (r *Registry) etcdDeRegister(ctx context.Context) error {
	if r.leaseCancel != nil {
		r.leaseCancel()
	}
	key, err := r.etcdRegistryKey()
	if err != nil {
		return err
	}
	_, deleteErr := r.etcdClient.Delete(ctx, key)
	_, revokeErr := r.etcdClient.Revoke(ctx, r.leaseID)
	return errors.Join(deleteErr, revokeErr)
}
