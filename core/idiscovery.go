package core

import (
	"context"
)

type DiscoveryLoadBalance string

const (
	RoundRobin DiscoveryLoadBalance = "roundrobin"
	Random     DiscoveryLoadBalance = "random"
)

type ServerInstance interface {
	GetName() string
	GetAddr() string
}

type IDiscovery interface {
	GetServer(ctx context.Context, serverName string) (ServerInstance, error)
}

type DiscoveryOption func(d IDiscovery) error

type IDiscoveryLoadBalance interface {
	GetInstance(ctx context.Context, serverName string) (ServerInstance, error)
	PutInstance(ctx context.Context, serverName string, instance ServerInstance) error
	RefreshAll(ctx context.Context, serverName string, instances []ServerInstance) error
}
