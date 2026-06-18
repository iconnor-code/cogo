package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc"
)

type NacosDiscovery struct {
	logger core.ILogger
	nacos  *client.Nacos
	mu     sync.Mutex
	next   map[string]int
}

func NewNacosDiscovery(logger core.ILogger, nacos *client.Nacos) *NacosDiscovery {
	return &NacosDiscovery{
		logger: logger,
		nacos:  nacos,
		next:   make(map[string]int),
	}
}

func (nd *NacosDiscovery) Discover(_ context.Context, serverName, serviceName, methodName string, tags []string, resp any, opts ...grpc.DialOption) (core.Endpoint, error) {
	fullMethodName := fmt.Sprintf("/%s.%s/%s", serverName, serviceName, methodName)

	return func(ctx context.Context, request any) (response any, err error) {
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			instance, err := nd.nextInstance(serverName, tags)
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

func (nd *NacosDiscovery) nextInstance(serverName string, clusters []string) (string, error) {
	instances, err := nd.nacos.NamingClient().SelectInstances(vo.SelectInstancesParam{
		ServiceName: serverName,
		GroupName:   nd.nacos.GroupName(),
		Clusters:    clusters,
		HealthyOnly: true,
	})
	if err != nil {
		return "", cerrs.Wrap(err, "nacos discover failed")
	}
	if len(instances) == 0 {
		return "", cerrs.New("no healthy nacos service instance found: " + serverName)
	}

	nd.mu.Lock()
	defer nd.mu.Unlock()

	index := nd.next[serverName] % len(instances)
	nd.next[serverName]++

	service := instances[index]
	return fmt.Sprintf("%s:%d", service.Ip, service.Port), nil
}
