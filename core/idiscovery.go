package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type IDiscovery interface {
	Discover(ctx context.Context, serverName string, tags []string) (endpoint.Endpoint, error)
}

type DiscoveryOption func(d IDiscovery) error
