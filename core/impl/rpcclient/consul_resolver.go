package rpcclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

type consulResolverBuilder struct {
	client          *api.Client
	logger          core.ILogger
	refreshInterval time.Duration
	queryTimeout    time.Duration
}

func newConsulResolverBuilder(client *api.Client, logger core.ILogger, refreshInterval, queryTimeout time.Duration) *consulResolverBuilder {
	return &consulResolverBuilder{client: client, logger: logger, refreshInterval: refreshInterval, queryTimeout: queryTimeout}
}

func (b *consulResolverBuilder) Scheme() string { return "consul" }

func (b *consulResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	service := target.Endpoint()
	if service == "" {
		return nil, fmt.Errorf("consul resolver service name is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &consulResolver{
		client:          b.client,
		logger:          b.logger,
		refreshInterval: b.refreshInterval,
		queryTimeout:    b.queryTimeout,
		service:         service,
		cc:              cc,
		ctx:             ctx,
		cancel:          cancel,
		resolveNow:      make(chan struct{}, 1),
	}
	if err := r.update(); err != nil {
		cc.ReportError(err)
		b.logger.Warn("initial consul resolve failed", zap.String("service", service), zap.Error(err))
	}
	r.wg.Add(1)
	go r.watch()
	return r, nil
}

type consulResolver struct {
	client          *api.Client
	logger          core.ILogger
	refreshInterval time.Duration
	queryTimeout    time.Duration
	service         string
	cc              resolver.ClientConn
	ctx             context.Context
	cancel          context.CancelFunc
	resolveNow      chan struct{}
	wg              sync.WaitGroup
}

func (r *consulResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case r.resolveNow <- struct{}{}:
	default:
	}
}

func (r *consulResolver) Close() {
	r.cancel()
	r.wg.Wait()
}

func (r *consulResolver) watch() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
		case <-r.resolveNow:
		}
		if err := r.update(); err != nil {
			r.cc.ReportError(err)
			r.logger.Warn("consul resolver refresh failed", zap.String("service", r.service), zap.Error(err))
		}
	}
}

func (r *consulResolver) update() error {
	ctx, cancel := context.WithTimeout(r.ctx, r.queryTimeout)
	defer cancel()
	entries, _, err := r.client.Health().ServiceMultipleTags(
		r.service,
		nil,
		true,
		(&api.QueryOptions{}).WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("resolve consul service %s: %w", r.service, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no healthy consul service instance found: %s", r.service)
	}

	addresses := make([]resolver.Address, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		address := entry.Service.Address
		if address == "" {
			address = entry.Node.Address
		}
		endpoint := fmt.Sprintf("%s:%d", address, entry.Service.Port)
		if _, ok := seen[endpoint]; ok {
			continue
		}
		seen[endpoint] = struct{}{}
		addresses = append(addresses, resolver.Address{Addr: endpoint})
	}
	return r.cc.UpdateState(resolver.State{Addresses: addresses})
}
