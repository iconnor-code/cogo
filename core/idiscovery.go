package core

import (
	"context"
)

type DiscoveryLoadBalance string

const (
	RoundRobin DiscoveryLoadBalance = "roundrobin"
	Random     DiscoveryLoadBalance = "random"
)

type IServerInstance interface {
	GetName() string
	GetAddr() string
	GetID() string
}

type IDiscovery interface {
	GetServer(ctx context.Context, serverName string) (IServerInstance, error)
}

type DiscoveryOption func(d IDiscovery) error

type IDiscoveryLoadBalance interface {
	GetInstance(ctx context.Context, serverName string) (IServerInstance, error)
	RefreshInstance(ctx context.Context, serverName string, instances []IServerInstance) error
}
