package core

import (
	"context"

	"google.golang.org/grpc"
)

type Endpoint func(ctx context.Context, request any) (response any, err error)

type IDiscovery interface {
	Discover(ctx context.Context, serverName, serviceName, methodName string, tags []string, resp any, opts ...grpc.DialOption) (Endpoint, error)
}

type DiscoveryOption func(d IDiscovery) error
