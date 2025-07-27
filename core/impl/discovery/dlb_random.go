package discovery

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/iconnor-code/cogo/core"
)

type RandomLoadBalance struct {
	instances []core.IServerInstance
	rwlock    sync.RWMutex
}

func NewRandomLoadBalance() *RandomLoadBalance {
	return &RandomLoadBalance{}
}

func (rlb *RandomLoadBalance) GetInstance(ctx context.Context, serverName string) (core.IServerInstance, error) {
	rlb.rwlock.RLock()
	defer rlb.rwlock.Unlock()
	if len(rlb.instances) == 0 {
		return nil, fmt.Errorf("none of instances for servername:%s", serverName)
	}
	if len(rlb.instances) == 1 {
		return rlb.instances[0], nil
	}

	bigidx, err := rand.Int(rand.Reader, big.NewInt(int64(len(rlb.instances)-1)))
	if err != nil {
		return nil, err
	}
	return rlb.instances[bigidx.Int64()], nil
}

func (rlb *RandomLoadBalance) RefreshInstance(ctx context.Context, serverName string, instances []core.IServerInstance) error {
	rlb.rwlock.Lock()
	defer rlb.rwlock.Unlock()
	rlb.instances = instances
	return nil
}
