package server

import (
	"context"
	"errors"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type testLogger struct{}

func (l *testLogger) Log(...any) error       { return nil }
func (l *testLogger) Debug(string, ...any)   {}
func (l *testLogger) Info(string, ...any)    {}
func (l *testLogger) Warn(string, ...any)    {}
func (l *testLogger) Error(string, ...any)   {}
func (l *testLogger) Fatal(string, ...any)   {}
func (l *testLogger) Panic(string, ...any)   {}
func (l *testLogger) AddGlobalFields(...any) {}

type testRegistry struct {
	registered    bool
	deregistered  bool
	err           error
	deregisterErr error
}

type testServer struct {
	started bool
	stopped bool
	err     error
}

func (s *testServer) Start() error {
	if s.err != nil {
		return s.err
	}
	s.started = true
	return nil
}

func (s *testServer) Stop() error {
	s.stopped = true
	return nil
}

func (r *testRegistry) Register(context.Context) error {
	r.registered = true
	return r.err
}

func TestGrpcServerStartClosesListenerWhenRegistrationFails(t *testing.T) {
	registry := &testRegistry{err: errors.New("register failed")}
	listener := bufconn.Listen(1024)
	s := &GrpcServer{
		conf:       &cogoconfig.Config{Config: core.Config{GRPC: core.GRPCConfig{Listen: "bufconn"}}},
		logger:     &testLogger{},
		listener:   listener,
		registry:   registry,
		baseServer: grpc.NewServer(),
	}

	if err := s.Start(); err == nil {
		t.Fatal("expected registration error")
	}
	if _, err := listener.Accept(); err == nil {
		t.Fatal("expected listener to be closed")
	}
}

func (r *testRegistry) DeRegister(context.Context) error {
	r.deregistered = true
	return r.deregisterErr
}

func TestGrpcServerStopStillStopsWhenDeregisterFails(t *testing.T) {
	listener := bufconn.Listen(1024)
	s := &GrpcServer{
		conf:       &cogoconfig.Config{Config: core.Config{GRPC: core.GRPCConfig{Listen: "bufconn"}}},
		logger:     &testLogger{},
		listener:   listener,
		registry:   &testRegistry{deregisterErr: errors.New("deregister failed")},
		baseServer: grpc.NewServer(),
	}
	go func() { _ = s.baseServer.Serve(listener) }()

	if err := s.Stop(); err == nil {
		t.Fatal("expected deregistration error")
	}
	if _, err := listener.Accept(); err == nil {
		t.Fatal("expected listener to be closed despite deregistration error")
	}
}

func TestHTTPServerStartReturnsListenError(t *testing.T) {
	s, err := NewHTTPServer(&cogoconfig.Config{
		Config: core.Config{HTTP: core.HTTPConfig{Listen: "invalid-address"}},
	}, &testLogger{}, runtime.NewServeMux())
	if err != nil {
		t.Fatalf("new http server: %v", err)
	}

	if err := s.Start(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestMetricsServerStartReturnsListenError(t *testing.T) {
	s, err := NewMetricsServer(&cogoconfig.Config{
		Config: core.Config{Metrics: core.MetricsConfig{Listen: "invalid-address"}},
	}, &testLogger{})
	if err != nil {
		t.Fatalf("new metrics server: %v", err)
	}

	if err := s.Start(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestMetricsServerStopBeforeStartIsNoop(t *testing.T) {
	s, err := NewMetricsServer(&cogoconfig.Config{
		Config: core.Config{Metrics: core.MetricsConfig{Listen: "127.0.0.1:0"}},
	}, &testLogger{})
	if err != nil {
		t.Fatalf("new metrics server: %v", err)
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
}

func TestGrpcServerStartStopRegistersAndDeregisters(t *testing.T) {
	registry := &testRegistry{}
	s := &GrpcServer{
		conf:       &cogoconfig.Config{Config: core.Config{GRPC: core.GRPCConfig{Listen: "bufconn"}}},
		logger:     &testLogger{},
		listener:   bufconn.Listen(1024),
		registry:   registry,
		baseServer: grpc.NewServer(),
	}

	if err := s.Start(); err != nil {
		t.Fatalf("start grpc server: %v", err)
	}
	if !registry.registered {
		t.Fatalf("expected registry.Register to be called")
	}

	if err := s.Stop(); err != nil {
		t.Fatalf("stop grpc server: %v", err)
	}
	if !registry.deregistered {
		t.Fatalf("expected registry.DeRegister to be called")
	}
}

func TestRegistryEnabledRequiresCompleteConfig(t *testing.T) {
	if registryEnabled(&cogoconfig.Config{}) {
		t.Fatalf("expected registry to be disabled when config is empty")
	}

	config := &cogoconfig.Config{
		Config: core.Config{
			Consul: core.ConsulConfig{Address: "127.0.0.1:8500"},
			Registry: core.RegistryConfig{
				Name:    "account",
				Address: "127.0.0.1",
				Port:    10000,
			},
		},
	}
	if !registryEnabled(config) {
		t.Fatalf("expected registry to be enabled when config is complete")
	}
}

func TestGRPCEndpointPrefersGatewayEndpoint(t *testing.T) {
	endpoint, err := GRPCEndpoint(&cogoconfig.Config{
		Config: core.Config{GRPC: core.GRPCConfig{GatewayEndpoint: "gateway:9090", Listen: ":9090"}},
	})
	if err != nil {
		t.Fatalf("grpc endpoint: %v", err)
	}
	if endpoint != "gateway:9090" {
		t.Fatalf("expected gateway endpoint, got %q", endpoint)
	}
}

func TestGRPCEndpointFallsBackToListen(t *testing.T) {
	endpoint, err := GRPCEndpoint(&cogoconfig.Config{
		Config: core.Config{GRPC: core.GRPCConfig{Listen: ":9090"}},
	})
	if err != nil {
		t.Fatalf("grpc endpoint: %v", err)
	}
	if endpoint != "127.0.0.1:9090" {
		t.Fatalf("expected listen endpoint, got %q", endpoint)
	}
}

func TestServerGroupStopsStartedServersOnStartError(t *testing.T) {
	first := &testServer{}
	second := &testServer{err: errors.New("boom")}
	group := NewServerGroup()
	group.AddServer("first", first)
	group.AddServer("second", second)

	if err := group.Start(); err == nil {
		t.Fatalf("expected start error")
	}
	if !first.started {
		t.Fatalf("expected first server to start")
	}
	if !first.stopped {
		t.Fatalf("expected first server to stop after later start error")
	}
	if second.started {
		t.Fatalf("second server should not be marked started")
	}
}

func TestNewGrpcServerGroupAddsMetricsOnlyWhenEnabled(t *testing.T) {
	config := &cogoconfig.Config{
		Config: core.Config{Metrics: core.MetricsConfig{Enable: false, Listen: "127.0.0.1:0"}},
	}
	group, err := NewGrpcServerGroup(config, &testLogger{}, func(core.IConfig, core.ILogger) (*testServer, error) {
		return &testServer{}, nil
	})
	if err != nil {
		t.Fatalf("new grpc server group: %v", err)
	}
	if len(group.servers) != 1 {
		t.Fatalf("expected one server when metrics disabled, got %d", len(group.servers))
	}

	config.Metrics.Enable = true
	group, err = NewGrpcServerGroup(config, &testLogger{}, func(core.IConfig, core.ILogger) (*testServer, error) {
		return &testServer{}, nil
	})
	if err != nil {
		t.Fatalf("new grpc server group with metrics: %v", err)
	}
	if len(group.servers) != 2 {
		t.Fatalf("expected two servers when metrics enabled, got %d", len(group.servers))
	}
}

var _ core.IRegistry = (*testRegistry)(nil)
