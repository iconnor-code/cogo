// Package discovery idiscovery implement
package discovery

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	kitlb "github.com/go-kit/kit/sd/lb"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
)

type KitConsulDiscovery struct {
	logger core.ILogger
	consul kitconsul.Client
}

func NewKitConsulDiscovery(logger core.ILogger, consul *client.Consul) *KitConsulDiscovery {
	return &KitConsulDiscovery{
		logger: logger,
		consul: consul.DefaultClient(),
	}
}

func (kcd *KitConsulDiscovery) Discover(_ context.Context, serverName, serviceName, methodName string, tags []string, resp any, opts ...grpc.DialOption) (endpoint.Endpoint, error) {
	// 创建服务发现实例
	instancer := kitconsul.NewInstancer(kcd.consul, kcd.logger, serverName, tags, true)

	// 创建工厂函数，用于为每个服务实例创建端点
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.NewClient(instance, opts...)
		if err != nil {
			return nil, nil, err
		}

		endpoint := func(ctx context.Context, request any) (response any, err error) {
			// 构建完整的 gRPC 方法路径
			fullMethodName := fmt.Sprintf("/%s.%s/%s", serverName, serviceName, methodName)

			err = conn.Invoke(ctx, fullMethodName, request, resp, grpc.StaticMethod())
			if err != nil {
				return nil, err
			}

			return resp, nil
		}
		return endpoint, conn, nil
	}

	// 创建端点集合器，用于管理服务实例的端点
	endpointer := sd.NewEndpointer(instancer, factory, kcd.logger)

	// 创建负载均衡器
	// 这里使用轮询策略，还可以选择随机、加权等策略
	balancer := kitlb.NewRoundRobin(endpointer)

	// 添加重试机制
	retry := kitlb.Retry(3, 500*time.Millisecond, balancer)

	return retry, nil
}
