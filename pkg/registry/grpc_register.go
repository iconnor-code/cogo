package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/logger"
	"github.com/iconnor-code/cogo/pkg/utils"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.uber.org/zap"
)

type GrpcRegister struct {
	config   *config.Conf
	logger   *logger.Logger
	etcd     *clientv3.Client
	manager  endpoints.Manager
	leaseTTL int64

	leaseID     clientv3.LeaseID
	cancel      context.CancelFunc
	registryKey string
}

func NewGrpcRegister(config *config.Conf, logger *logger.Logger) (*GrpcRegister, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: config.Etcd.Endpoints,
	})
	if err != nil {
		return nil, cerr.WithStack(err)
	}

	manager, err := endpoints.NewManager(etcd, config.Registry.Key)
	if err != nil {
		return nil, cerr.WithStack(err)
	}

	registry := &GrpcRegister{
		config:      config,
		logger:      logger,
		etcd:        etcd,
		manager:     manager,
		leaseTTL:    3,
		registryKey: config.Registry.Key,
	}
	return registry, nil
}

func (r *GrpcRegister) Register(ctx context.Context) error {
	var err error
	host := strings.Split(r.config.Grpc.Listen, ":")
	hostName := host[0]
	hostPort := host[1]
	if r.config.Registry.Hostname != "" && r.config.Registry.Hostname != "0.0.0.0" {
		hostName = r.config.Registry.Hostname
	}
	if hostName == "" {
		hostName, err = utils.GetLocalIP()
		if err != nil {
			return cerr.WithStack(err)
		}
	}

	r.registryKey = fmt.Sprintf("%s/%s:%s", r.config.Registry.Key, hostName, hostPort)
	value := fmt.Sprintf("%s:%s", hostName, hostPort)

	lease, err := r.etcd.Grant(ctx, r.leaseTTL)
	if err != nil {
		return cerr.WithStack(err)
	}
	hexLeaseID := fmt.Sprintf("0x%x", lease.ID)
	r.logger.Log().Info("GrpcRegistry leaseID", zap.Any("leaseID", lease.ID), zap.String("hexLeaseID", hexLeaseID))

	r.leaseID = lease.ID

	err = r.manager.AddEndpoint(ctx, r.registryKey,
		endpoints.Endpoint{
			Addr: value,
		},
		clientv3.WithLease(lease.ID),
	)
	if err != nil {
		return cerr.WithStack(err)
	}

	r.keepAlive(ctx)

	return nil
}

func (r *GrpcRegister) keepAlive(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(r.leaseTTL-1) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_, err := r.etcd.Revoke(context.Background(), r.leaseID)
				if err != nil {
					r.logger.Log().Error("GrpcRegistry revoke error", zap.Error(err))
				}
				return
			case <-ticker.C:
				_, err := r.etcd.KeepAliveOnce(ctx, r.leaseID)
				if err != nil {
					r.logger.Log().Error("GrpcRegistry keepAlive error", zap.Error(err))
					return
				}
			}
		}
	}()
}

func (r *GrpcRegister) Deregister(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}

	err := r.manager.DeleteEndpoint(ctx, r.registryKey)
	if err != nil {
		return cerr.WithStack(err)
	}
	return nil
}