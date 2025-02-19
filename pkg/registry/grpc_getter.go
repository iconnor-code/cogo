package registry

import (
	"fmt"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/cerr"

	clientv3 "go.etcd.io/etcd/client/v3"
	etcdnaming "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcGetter struct {
	config *config.Conf
	etcd   *clientv3.Client
}

func NewGrpcGetter(config *config.Conf) (*GrpcGetter, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: config.Etcd.Endpoints,
	})
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return &GrpcGetter{
		config: config,
		etcd:   etcd,
	}, nil
}

func (g *GrpcGetter) GetService(key string) (*grpc.ClientConn, error) {
	name, err := etcdnaming.NewBuilder(g.etcd)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	conn, err := grpc.NewClient(fmt.Sprintf("%s/%s", "etcd://", key),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(name),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)),
	)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return conn, nil
}
