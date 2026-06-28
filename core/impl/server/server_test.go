package server

import (
	"context"
	"errors"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type testConfig struct {
	data map[string]any
}

func (c *testConfig) Get(key string) any { return c.data[key] }
func (c *testConfig) ReLoad() error      { return nil }

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
	registered   bool
	deregistered bool
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
	return nil
}

func (r *testRegistry) DeRegister(context.Context) error {
	r.deregistered = true
	return nil
}

func TestHTTPServerStartReturnsListenError(t *testing.T) {
	s, err := NewHTTPServer(&testConfig{data: map[string]any{
		"http": map[string]any{"listen": "invalid-address"},
	}}, &testLogger{}, runtime.NewServeMux())
	if err != nil {
		t.Fatalf("new http server: %v", err)
	}

	if err := s.Start(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestMetricsServerStartReturnsListenError(t *testing.T) {
	s, err := NewMetricsServer(&testConfig{data: map[string]any{
		"metrics": map[string]any{"listen": "invalid-address"},
	}}, &testLogger{})
	if err != nil {
		t.Fatalf("new metrics server: %v", err)
	}

	if err := s.Start(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestMetricsServerStopBeforeStartIsNoop(t *testing.T) {
	s, err := NewMetricsServer(&testConfig{data: map[string]any{
		"metrics": map[string]any{"listen": "127.0.0.1:0"},
	}}, &testLogger{})
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
		conf:       &testConfig{data: map[string]any{"grpc.listen": "bufconn"}},
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

func TestGRPCEndpointPrefersGatewayEndpoint(t *testing.T) {
	endpoint, err := GRPCEndpoint(&testConfig{data: map[string]any{
		"grpc": map[string]any{
			"gateway_endpoint": "gateway:9090",
			"listen":           ":9090",
		},
	}})
	if err != nil {
		t.Fatalf("grpc endpoint: %v", err)
	}
	if endpoint != "gateway:9090" {
		t.Fatalf("expected gateway endpoint, got %q", endpoint)
	}
}

func TestGRPCEndpointFallsBackToListen(t *testing.T) {
	endpoint, err := GRPCEndpoint(&testConfig{data: map[string]any{
		"grpc": map[string]any{"listen": ":9090"},
	}})
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

var _ core.IRegistry = (*testRegistry)(nil)
