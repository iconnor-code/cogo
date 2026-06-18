package registry

import (
	"context"

	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go.uber.org/zap"
)

func WithNacosClient(nacos *client.Nacos) core.RegistryOption {
	return func(r core.IRegistry) error {
		r.(*Registry).nacosClient = nacos
		return nil
	}
}

func (r *Registry) nacosRegister(ctx context.Context) error {
	_ = ctx
	name, address, port, err := r.serviceConfig()
	if err != nil {
		return err
	}
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}

	_, err = r.nacosClient.NamingClient().RegisterInstance(vo.RegisterInstanceParam{
		Ip:          address,
		Port:        uint64(port),
		ServiceName: name,
		GroupName:   r.nacosClient.GroupName(),
		ClusterName: r.nacosClient.ClusterName(),
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"instance_id": instanceID,
		},
	})
	if err != nil {
		return err
	}
	r.logger.Info("nacos register", zap.String("id", instanceID), zap.String("name", name), zap.String("address", address), zap.Int("port", port), zap.String("group", r.nacosClient.GroupName()), zap.String("cluster", r.nacosClient.ClusterName()))
	return nil
}

func (r *Registry) nacosDeRegister(ctx context.Context) error {
	_ = ctx
	name, address, port, err := r.serviceConfig()
	if err != nil {
		return err
	}
	instanceID, err := r.getInstanceID()
	if err != nil {
		return err
	}

	_, err = r.nacosClient.NamingClient().DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          address,
		Port:        uint64(port),
		ServiceName: name,
		GroupName:   r.nacosClient.GroupName(),
		Cluster:     r.nacosClient.ClusterName(),
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}
	r.logger.Info("nacos deregister", zap.String("id", instanceID), zap.String("name", name), zap.String("address", address), zap.Int("port", port), zap.String("group", r.nacosClient.GroupName()), zap.String("cluster", r.nacosClient.ClusterName()))
	return nil
}
