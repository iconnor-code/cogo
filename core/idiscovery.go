package core

import "context"

type IDiscovery interface {
	GetGrpcClientConn(serverName string) (any, error)
}

type DiscoveryOption func(d IDiscovery) error

type IDiscoveryClient interface {
	Register(ctx context.Context, instance ServiceInstance) error
	Deregister(ctx context.Context, instance ServiceInstance) error
	Service(ctx context.Context, serverName string) ([]ServiceInstance, error)
}

type ServiceInstance interface {
	GetName() string
}
