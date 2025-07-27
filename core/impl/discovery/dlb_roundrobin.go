package discovery

import (
	"context"

	"github.com/iconnor-code/cogo/core"
)

type RoundRobinLoadBalance struct{}

func NewRoundRobinLoadBalance() *RoundRobinLoadBalance {
	return &RoundRobinLoadBalance{}
}

func (rrlb *RoundRobinLoadBalance) GetInstance(ctx context.Context, serverName string) (core.IServerInstance, error) {
	return nil, nil
}

func (rrlb *RoundRobinLoadBalance) RefreshInstance(ctx context.Context, serverName string, instances []core.IServerInstance) error {
	return nil
}
