package registry

import (
	"fmt"

	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/etcd"
	etcdnaming "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type GrpcDiscover struct {
	config   *config.Conf
	etcd     *etcd.EtcdClient
	resolver resolver.Builder
}

func NewGrpcDiscover(config *config.Conf, etcd *etcd.EtcdClient) (*GrpcDiscover, error) {
	name, err := etcdnaming.NewBuilder(etcd.Client)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return &GrpcDiscover{
		config:   config,
		etcd:     etcd,
		resolver: name,
	}, nil
}

func (g *GrpcDiscover) GetServiceConn(serverID uint8) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s/%d", "etcd://", serverID),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(g.resolver),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, roundrobin.Name)),
	)
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return conn, nil
}
