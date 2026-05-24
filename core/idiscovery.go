package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc"
)

type IDiscovery interface {
	Discover(ctx context.Context, serverName, serviceName, methodName string, tags []string, resp any, opts ...grpc.DialOption) (endpoint.Endpoint, error)
}

type DiscoveryOption func(d IDiscovery) error
