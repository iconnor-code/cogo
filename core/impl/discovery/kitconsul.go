// Package discovery idiscovery implement
package discovery

import (
	"context"
	"io"
	"time"

	"net/rpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	kitlb "github.com/go-kit/kit/sd/lb"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
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

func (kcd *KitConsulDiscovery) Discovery(serverName, methodName string, tags []string) endpoint.Endpoint {
	// 创建服务发现实例
	instancer := kitconsul.NewInstancer(kcd.consul, kcd.logger, serverName, tags, true)

	// 创建工厂函数，用于为每个服务实例创建端点
	factory := func(instance string) (endpoint.Endpoint, io.Closer, error) {
		// 创建 RPC 客户端
		rpcClient, err := rpc.Dial("tcp", instance)
		if err != nil {
			return nil, nil, err
		}

		// 创建 go-kit 端点
		endpoint := func(ctx context.Context, request any) (response any, err error) {
			// 调用 RPC 方法
			// 注意：这里需要根据你的实际 RPC 服务方法名来调整
			methodName := "/grpc." + serverName + "." + methodName
			err = rpcClient.Call(methodName, request, response)
			if err != nil {
				return nil, err
			}

			return response, nil
		}

		return endpoint, rpcClient, nil
	}

	// 创建端点集合器，用于管理服务实例的端点
	endpointer := sd.NewEndpointer(instancer, factory, kcd.logger)

	// 创建负载均衡器
	// 这里使用轮询策略，还可以选择随机、加权等策略
	balancer := kitlb.NewRoundRobin(endpointer)

	// 添加重试机制
	retry := kitlb.Retry(3, 500*time.Millisecond, balancer)

	return retry
}
