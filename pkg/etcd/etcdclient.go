package etcd

import (
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	*clientv3.Client
}

func NewEtcdClient(config core.IConfig) (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: config.Get("etcd.endpoints").([]string),
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
