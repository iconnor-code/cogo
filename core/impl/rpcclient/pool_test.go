package rpcclient

import (
	"strings"
	"testing"

	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
)

func TestPoolReusesDNSConnection(t *testing.T) {
	config := &cogoconfig.Config{Config: core.Config{Discovery: core.DiscoveryConfig{
		Provider: "dns",
		Services: map[string]string{"account": "dns:///account:10000"},
	}}}
	pool, err := NewPool(config, nil)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })

	first, err := pool.Conn("account")
	if err != nil {
		t.Fatalf("first connection: %v", err)
	}
	second, err := pool.Conn("account")
	if err != nil {
		t.Fatalf("second connection: %v", err)
	}
	if first != second {
		t.Fatal("expected connection to be reused")
	}
}

func TestPoolRejectsMissingTargetAndUseAfterClose(t *testing.T) {
	config := &cogoconfig.Config{Config: core.Config{Discovery: core.DiscoveryConfig{Provider: "dns"}}}
	pool, err := NewPool(config, nil)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	if _, err := pool.Conn("account"); err == nil || !strings.Contains(err.Error(), "target") {
		t.Fatalf("expected missing target error, got %v", err)
	}
	if err := pool.Close(); err != nil {
		t.Fatalf("close pool: %v", err)
	}
	if _, err := pool.Conn("account"); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed pool error, got %v", err)
	}
}

func TestPoolAllowsDisabledDiscoveryUntilConnectionRequested(t *testing.T) {
	pool, err := NewPool(&cogoconfig.Config{}, nil)
	if err != nil {
		t.Fatalf("new disabled pool: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if _, err := pool.Conn("account"); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled discovery error, got %v", err)
	}
}

func TestPoolRejectsInvalidConsulTimeout(t *testing.T) {
	config := &cogoconfig.Config{Config: core.Config{
		Consul: core.ConsulConfig{Address: "127.0.0.1:8500"},
		Discovery: core.DiscoveryConfig{
			Provider: "consul",
			Timeout:  "0s",
		},
	}}
	if _, err := NewPool(config, &testLogger{}); err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected invalid timeout error, got %v", err)
	}
}

type testLogger struct{}

func (*testLogger) Log(...any) error       { return nil }
func (*testLogger) Debug(string, ...any)   {}
func (*testLogger) Info(string, ...any)    {}
func (*testLogger) Warn(string, ...any)    {}
func (*testLogger) Error(string, ...any)   {}
func (*testLogger) Fatal(string, ...any)   {}
func (*testLogger) Panic(string, ...any)   {}
func (*testLogger) AddGlobalFields(...any) {}
