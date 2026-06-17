// Package discovery idiscovery implement
package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
)

type ConsulDiscovery struct {
	logger core.ILogger
	consul *api.Client
	mu     sync.Mutex
	next   map[string]int
}

func NewConsulDiscovery(logger core.ILogger, consul *client.Consul) *ConsulDiscovery {
	return &ConsulDiscovery{
		logger: logger,
		consul: consul.DefaultClient(),
		next:   make(map[string]int),
	}
}

func (cd *ConsulDiscovery) Discover(_ context.Context, serverName, serviceName, methodName string, tags []string, resp any, opts ...grpc.DialOption) (core.Endpoint, error) {
	fullMethodName := fmt.Sprintf("/%s.%s/%s", serverName, serviceName, methodName)

	return func(ctx context.Context, request any) (response any, err error) {
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			instance, err := cd.nextInstance(serverName, tags)
			if err != nil {
				lastErr = err
			} else {
				lastErr = invoke(ctx, instance, fullMethodName, request, resp, opts...)
				if lastErr == nil {
					return resp, nil
				}
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}
		return nil, lastErr
	}, nil
}

func (cd *ConsulDiscovery) nextInstance(serverName string, tags []string) (string, error) {
	entries, _, err := cd.consul.Health().ServiceMultipleTags(serverName, tags, true, nil)
	if err != nil {
		return "", cerrs.Wrap(err, "consul discover failed")
	}
	if len(entries) == 0 {
		return "", cerrs.New("no healthy consul service instance found: " + serverName)
	}

	cd.mu.Lock()
	defer cd.mu.Unlock()

	index := cd.next[serverName] % len(entries)
	cd.next[serverName]++

	service := entries[index].Service
	address := service.Address
	if address == "" {
		address = entries[index].Node.Address
	}
	return fmt.Sprintf("%s:%d", address, service.Port), nil
}

func invoke(ctx context.Context, instance, method string, request any, resp any, opts ...grpc.DialOption) error {
	conn, err := grpc.NewClient(instance, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Invoke(ctx, method, request, resp, grpc.StaticMethod())
}
