package etcd

import (
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	*clientv3.Client
}

func NewEtcdClient(config *config.Conf) (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: config.Etcd.Endpoints,
	})
	if err != nil {
		return nil, cerr.WithStack(err)
	}
	return &EtcdClient{
		Client: client,
	}, nil
}

func (e *EtcdClient) Close() error {
	return e.Client.Close()
}
