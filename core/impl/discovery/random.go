package discovery

import (
	"context"

	"github.com/iconnor-code/cogo/core"
)

type RandomLoadBalance struct {
	instances map[string]core.ServerInstance
}

func (rlb *RandomLoadBalance) GetInstance(ctx context.Context, serverName string) (core.ServerInstance, error) {
	return nil, nil
}

func (rlb *RandomLoadBalance) PutInstance(ctx context.Context, serverName string, instance core.ServerInstance) error {
	return nil
}

func (rlb *RandomLoadBalance) RefreshAll(ctx context.Context, serverName string) error {
	return nil
}
