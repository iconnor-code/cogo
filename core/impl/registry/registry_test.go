package registry

import (
	"strings"
	"testing"

	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
)

type testLogger struct{}

func (*testLogger) Log(...any) error       { return nil }
func (*testLogger) Debug(string, ...any)   {}
func (*testLogger) Info(string, ...any)    {}
func (*testLogger) Warn(string, ...any)    {}
func (*testLogger) Error(string, ...any)   {}
func (*testLogger) Fatal(string, ...any)   {}
func (*testLogger) Panic(string, ...any)   {}
func (*testLogger) AddGlobalFields(...any) {}

func TestNewDefaultDisablesRegistryWithoutConsul(t *testing.T) {
	got, err := NewDefault(&cogoconfig.Config{}, &testLogger{})
	if err != nil {
		t.Fatalf("new default registry: %v", err)
	}
	if got != nil {
		t.Fatal("expected registry to be disabled")
	}
}

func TestNewRegistryRejectsInvalidClientConfiguration(t *testing.T) {
	config := &cogoconfig.Config{}
	tests := []struct {
		name    string
		opts    []Option
		wantErr string
	}{
		{name: "missing client", wantErr: "exactly one"},
		{name: "multiple clients", opts: []Option{WithConsulClient(&client.Consul{}), WithEtcdClient(&client.EtcdClient{}), WithEtcdRegisterLeaseTTL(5)}, wantErr: "cannot be configured together"},
		{name: "missing etcd ttl", opts: []Option{WithEtcdClient(&client.EtcdClient{})}, wantErr: "lease ttl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRegistry(config, &testLogger{}, tt.opts...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q error, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRegistryConstructorsRejectMissingDependencies(t *testing.T) {
	if _, err := NewRegistry(nil, &testLogger{}, WithConsulClient(&client.Consul{})); err == nil {
		t.Fatal("expected missing registry config error")
	}
	if _, err := NewRegistry(&cogoconfig.Config{}, nil, WithConsulClient(&client.Consul{})); err == nil {
		t.Fatal("expected missing registry logger error")
	}
	if _, err := NewDefault(nil, &testLogger{}); err == nil {
		t.Fatal("expected missing default registry config error")
	}
}

func TestNewDefaultRejectsIncompleteRegistryConfig(t *testing.T) {
	tests := []struct {
		name        string
		registry    core.RegistryConfig
		wantErrPart string
	}{
		{name: "missing name", registry: core.RegistryConfig{Provider: "consul", Address: "127.0.0.1", Port: 10000}, wantErrPart: "name"},
		{name: "missing address", registry: core.RegistryConfig{Provider: "consul", Name: "account", Port: 10000}, wantErrPart: "address"},
		{name: "invalid port", registry: core.RegistryConfig{Provider: "consul", Name: "account", Address: "127.0.0.1", Port: 70000}, wantErrPart: "port"},
		{name: "invalid interval", registry: core.RegistryConfig{Provider: "consul", Name: "account", Address: "127.0.0.1", Port: 10000, HealthCheck: core.RegistryHealthCheckConfig{Interval: "bad", Timeout: "5s"}}, wantErrPart: "interval"},
		{name: "invalid timeout", registry: core.RegistryConfig{Provider: "consul", Name: "account", Address: "127.0.0.1", Port: 10000, HealthCheck: core.RegistryHealthCheckConfig{Interval: "3s", Timeout: "0s"}}, wantErrPart: "timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &cogoconfig.Config{Config: core.Config{
				Consul:   core.ConsulConfig{Address: "127.0.0.1:8500"},
				Registry: tt.registry,
			}}
			_, err := NewDefault(config, &testLogger{})
			if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("expected %q error, got %v", tt.wantErrPart, err)
			}
		})
	}
}

func TestNewDefaultBuildsCompleteConsulRegistry(t *testing.T) {
	config := &cogoconfig.Config{Config: core.Config{
		Consul: core.ConsulConfig{Address: "127.0.0.1:8500"},
		Registry: core.RegistryConfig{
			Provider: "consul",
			Name:     "account",
			Address:  "127.0.0.1",
			Port:     10000,
			HealthCheck: core.RegistryHealthCheckConfig{
				Interval: "3s",
				Timeout:  "5s",
			},
		},
	}}

	got, err := NewDefault(config, &testLogger{})
	if err != nil {
		t.Fatalf("new default registry: %v", err)
	}
	if got == nil {
		t.Fatal("expected consul registry")
	}
}
