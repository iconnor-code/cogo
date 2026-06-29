package client

import (
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	*clientv3.Client
}

func NewEtcdClient(config core.IConfig) (*EtcdClient, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: config.GetEtcd().Endpoints,
	})
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	return &EtcdClient{
		Client: client,
	}, nil
}

func (e *EtcdClient) Close() error {
	return e.Client.Close()
}
