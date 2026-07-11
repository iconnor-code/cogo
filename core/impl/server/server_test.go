package server

import (
	"context"
	"errors"
	"strings"
	"sync"
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
	started     bool
	stopped     bool
	startErr    error
	waitErr     error
	shutdownErr error
	done        chan struct{}
	once        sync.Once
}

func (s *testServer) Start(context.Context) error {
	if s.startErr != nil {
		return s.startErr
	}
	if s.done == nil {
		s.done = make(chan struct{})
	}
	s.started = true
	if s.waitErr != nil {
		s.once.Do(func() { close(s.done) })
	}
	return nil
}

func (s *testServer) Wait() error {
	<-s.done
	return s.waitErr
}

func (s *testServer) Shutdown(context.Context) error {
	s.stopped = true
	s.once.Do(func() {
		if s.done == nil {
			s.done = make(chan struct{})
		}
		close(s.done)
	})
	return s.shutdownErr
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
		serveErr:   make(chan error, 1),
	}

	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected registration error")
	}
	if _, err := s.listener.Accept(); err == nil {
		t.Fatal("expected listener to be closed")
	}
}

func TestNewGrpcServerReturnsOptionError(t *testing.T) {
	wantErr := errors.New("invalid option")
	_, err := NewGrpcServer(&cogoconfig.Config{}, &testLogger{}, grpc.NewServer(), func(*GrpcServer) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected option error, got %v", err)
	}
}

func (r *testRegistry) DeRegister(context.Context) error {
	r.deregistered = true
	return r.deregisterErr
}

func TestGrpcServerStopStillStopsWhenDeregisterFails(t *testing.T) {
	listener := bufconn.Listen(1024)
	registry := &testRegistry{deregisterErr: errors.New("deregister failed")}
	s := &GrpcServer{
		conf:       &cogoconfig.Config{Config: core.Config{GRPC: core.GRPCConfig{Listen: "bufconn"}}},
		logger:     &testLogger{},
		listener:   listener,
		registry:   registry,
		baseServer: grpc.NewServer(),
		serveErr:   make(chan error, 1),
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start grpc server: %v", err)
	}

	if err := s.Shutdown(context.Background()); err == nil {
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

	if err := s.Start(context.Background()); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestHTTPServerStartValidatesTLSCertificate(t *testing.T) {
	s, err := NewHTTPServer(&cogoconfig.Config{Config: core.Config{HTTP: core.HTTPConfig{
		Listen: "127.0.0.1:0",
		SSL: core.HTTPSSLConfig{
			CertFile: "missing-cert.pem",
			KeyFile:  "missing-key.pem",
		},
	}}}, &testLogger{}, runtime.NewServeMux())
	if err != nil {
		t.Fatalf("new http server: %v", err)
	}

	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected TLS certificate error during startup")
	}
}

func TestMetricsServerStartReturnsListenError(t *testing.T) {
	s, err := NewMetricsServer(&cogoconfig.Config{
		Config: core.Config{Metrics: core.MetricsConfig{Listen: "invalid-address"}},
	}, &testLogger{})
	if err != nil {
		t.Fatalf("new metrics server: %v", err)
	}

	if err := s.Start(context.Background()); err == nil {
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

	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
}

func TestGrpcServerStartStopRegistersAndDeregisters(t *testing.T) {
	registry := &testRegistry{}
	listener := bufconn.Listen(1024)
	s := &GrpcServer{
		conf:       &cogoconfig.Config{Config: core.Config{GRPC: core.GRPCConfig{Listen: "bufconn"}}},
		logger:     &testLogger{},
		listener:   listener,
		registry:   registry,
		baseServer: grpc.NewServer(),
		serveErr:   make(chan error, 1),
	}

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("start grpc server: %v", err)
	}
	if !registry.registered {
		t.Fatalf("expected registry.Register to be called")
	}

	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("stop grpc server: %v", err)
	}
	if !registry.deregistered {
		t.Fatalf("expected registry.DeRegister to be called")
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
	second := &testServer{startErr: errors.New("boom")}
	group := NewServerGroup()
	group.AddServer("first", first)
	group.AddServer("second", second)

	if err := group.start(context.Background()); err == nil {
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

func TestServerGroupRunReturnsRuntimeFailureAndStopsPeers(t *testing.T) {
	failed := &testServer{waitErr: errors.New("serve failed")}
	peer := &testServer{}
	group := NewServerGroup(failed, peer)

	err := group.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "serve failed") {
		t.Fatalf("expected runtime failure, got %v", err)
	}
	if !peer.stopped {
		t.Fatal("expected peer server to be stopped after runtime failure")
	}
}

func TestServerGroupRunPreservesFailureWhenContextIsAlreadyCanceled(t *testing.T) {
	failed := &testServer{waitErr: errors.New("serve failed")}
	group := NewServerGroup(failed)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := group.Run(ctx)
	if err == nil || !strings.Contains(err.Error(), "serve failed") {
		t.Fatalf("expected runtime failure alongside cancellation, got %v", err)
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

func TestHTTPGatewayGroupCreatesResourcesLazilyAndCleansUpStartFailure(t *testing.T) {
	var gatewayCtx context.Context
	grpcServer := &testServer{}
	group, err := NewHTTPGatewayServerGroup(
		&cogoconfig.Config{Config: core.Config{HTTP: core.HTTPConfig{Listen: "invalid-address"}}},
		&testLogger{},
		func(core.IConfig, core.ILogger) (*testServer, error) { return grpcServer, nil },
		func(ctx context.Context, _ core.IConfig) (*runtime.ServeMux, error) {
			gatewayCtx = ctx
			return runtime.NewServeMux(), nil
		},
		SwaggerOption{},
	)
	if err != nil {
		t.Fatalf("new gateway group: %v", err)
	}
	if gatewayCtx != nil {
		t.Fatal("gateway resources must not be created by the group constructor")
	}
	if err := group.start(context.Background()); err == nil {
		t.Fatal("expected HTTP listen error")
	}
	if gatewayCtx == nil {
		t.Fatal("expected gateway resources to be created during start")
	}
	select {
	case <-gatewayCtx.Done():
	default:
		t.Fatal("expected gateway context to be canceled")
	}
	if !grpcServer.stopped {
		t.Fatal("expected grpc server rollback after gateway start failure")
	}
}

func TestServerLifecycleRejectsInvalidCallOrder(t *testing.T) {
	s, err := NewMetricsServer(&cogoconfig.Config{}, &testLogger{})
	if err != nil {
		t.Fatalf("new metrics server: %v", err)
	}
	if err := s.Wait(); !errors.Is(err, ErrServerNotStarted) {
		t.Fatalf("wait before start: got %v, want %v", err, ErrServerNotStarted)
	}
	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected missing listen address error")
	}
	if err := s.Start(context.Background()); !errors.Is(err, ErrServerAlreadyStarted) {
		t.Fatalf("second start: got %v, want %v", err, ErrServerAlreadyStarted)
	}
}

var _ core.IRegistry = (*testRegistry)(nil)
