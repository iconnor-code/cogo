// Package rpcclient owns reusable gRPC client connections and service-name
// resolution. Business packages should depend on generated protobuf clients,
// not on this infrastructure package directly.
package rpcclient

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultConsulRefreshInterval = 10 * time.Second
const defaultConsulQueryTimeout = 3 * time.Second

var roundRobinServiceConfig = `{"loadBalancingConfig":[{"round_robin":{}}]}`

// Pool lazily creates and reuses one gRPC ClientConn per logical service.
// Close must be called by the process lifecycle owner.
type Pool struct {
	config   core.DiscoveryConfig
	resolver *consulResolverBuilder

	mu     sync.Mutex
	conns  map[string]*grpc.ClientConn
	closed bool
}

// NewPool builds the configured discovery strategy. An empty provider is
// allowed so services without downstream dependencies do not need discovery.
func NewPool(config core.IConfig, logger core.ILogger) (*Pool, error) {
	if config == nil {
		return nil, errors.New("rpc client config is required")
	}

	discoveryConfig := config.GetDiscovery()
	provider := strings.ToLower(strings.TrimSpace(discoveryConfig.Provider))
	pool := &Pool{config: discoveryConfig, conns: make(map[string]*grpc.ClientConn)}

	switch provider {
	case "", "none", "dns":
		return pool, nil
	case "consul":
		if logger == nil {
			return nil, errors.New("rpc client logger is required for consul discovery")
		}
		if strings.TrimSpace(config.GetConsul().Address) == "" {
			return nil, errors.New("consul address is required for consul discovery")
		}
		refreshInterval := defaultConsulRefreshInterval
		if value := strings.TrimSpace(discoveryConfig.RefreshInterval); value != "" {
			parsed, err := time.ParseDuration(value)
			if err != nil || parsed <= 0 {
				return nil, fmt.Errorf("discovery refresh interval must be a positive duration: %q", value)
			}
			refreshInterval = parsed
		}
		queryTimeout := defaultConsulQueryTimeout
		if value := strings.TrimSpace(discoveryConfig.Timeout); value != "" {
			parsed, err := time.ParseDuration(value)
			if err != nil || parsed <= 0 {
				return nil, fmt.Errorf("discovery timeout must be a positive duration: %q", value)
			}
			queryTimeout = parsed
		}
		consul, err := client.NewConsul(config)
		if err != nil {
			return nil, err
		}
		pool.resolver = newConsulResolverBuilder(consul.DefaultClient(), logger, refreshInterval, queryTimeout)
		return pool, nil
	default:
		return nil, fmt.Errorf("unsupported discovery provider %q", discoveryConfig.Provider)
	}
}

// Conn returns a shared connection for service. DNS targets can be ordinary
// Kubernetes Service names or dns:/// targets for headless Services.
func (p *Pool) Conn(service string) (*grpc.ClientConn, error) {
	service = strings.TrimSpace(service)
	if service == "" {
		return nil, errors.New("service name is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, errors.New("rpc client pool is closed")
	}
	if conn := p.conns[service]; conn != nil {
		return conn, nil
	}

	target, opts, err := p.targetAndOptions(service)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("create grpc client for %s: %w", service, err)
	}
	p.conns[service] = conn
	return conn, nil
}

func (p *Pool) targetAndOptions(service string) (string, []grpc.DialOption, error) {
	provider := strings.ToLower(strings.TrimSpace(p.config.Provider))
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(roundRobinServiceConfig),
	}
	if provider == "consul" {
		return "consul:///" + service, append(opts, grpc.WithResolvers(p.resolver)), nil
	}
	if provider == "" || provider == "none" {
		return "", nil, errors.New("service discovery is disabled")
	}
	target := strings.TrimSpace(p.config.Services[service])
	if target == "" {
		return "", nil, fmt.Errorf("discovery target for service %q is required", service)
	}
	return target, opts, nil
}

// Close closes all connections. It is safe to call more than once.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	var errs error
	for service, conn := range p.conns {
		if err := conn.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("close grpc client for %s: %w", service, err))
		}
	}
	p.conns = nil
	return errs
}
