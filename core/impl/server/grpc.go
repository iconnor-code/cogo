// Package server
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	cogointerceptor "github.com/iconnor-code/cogo/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type managedServer struct {
	name   string
	server core.Server
}

type ServerGroup struct {
	servers         []managedServer
	started         int
	results         chan serverResult
	shutdownTimeout time.Duration
}

type serverResult struct {
	name string
	err  error
}

type GrpcServer struct {
	conf       core.IConfig
	logger     core.ILogger
	listener   net.Listener
	registry   core.IRegistry
	closers    []io.Closer
	closeOnce  sync.Once
	closeErr   error
	baseServer *grpc.Server
	health     *health.Server
	serveErr   chan error
	lifecycle  componentLifecycle
}

type GrpcServerOption func(*GrpcServer) error

type GrpcServiceOption struct {
	PublicMethods          []string
	TokenRevocationChecker cogointerceptor.TokenRevocationChecker
	UnaryInterceptors      []grpc.UnaryServerInterceptor
	RegisterServices       func(*grpc.Server) error
	Registry               core.IRegistry
	Closers                []io.Closer
}

func WithGrpcRegistry(registry core.IRegistry) GrpcServerOption {
	return func(server *GrpcServer) error {
		server.registry = registry
		return nil
	}
}

func NewGrpcServer(config core.IConfig, logger core.ILogger, bs *grpc.Server, opts ...GrpcServerOption) (*GrpcServer, error) {
	if err := validateServerDependencies(config, logger); err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, errors.New("grpc base server is required")
	}
	s := &GrpcServer{
		conf:       config,
		logger:     logger,
		baseServer: bs,
		serveErr:   make(chan error, 1),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func NewGrpcServiceServer(config core.IConfig, logger core.ILogger, opt GrpcServiceOption) (*GrpcServer, error) {
	if err := validateServerDependencies(config, logger); err != nil {
		return nil, err
	}
	baseServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		unaryInterceptors(config, logger, opt)...,
	))
	if err := registerUnaryServices(baseServer, opt); err != nil {
		_ = closeResources(opt.Closers)
		return nil, err
	}
	healthServer := registerHealthServer(baseServer, config)

	server, err := NewGrpcServer(config, logger, baseServer, WithGrpcRegistry(opt.Registry))
	if err != nil {
		_ = closeResources(opt.Closers)
		return nil, err
	}
	server.health = healthServer
	server.closers = append(server.closers, opt.Closers...)
	return server, nil
}

func (s *GrpcServer) Start(ctx context.Context) (err error) {
	if err := s.lifecycle.beginStart(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.lifecycle.markStartFailed()
			err = errors.Join(err, s.Close())
		}
	}()

	listen := s.conf.GetGRPC().Listen
	if strings.TrimSpace(listen) == "" {
		return errors.New("grpc listen address is required")
	}
	listener := s.listener
	if listener == nil {
		var err error
		listener, err = net.Listen("tcp", listen)
		if err != nil {
			return cerrs.Wrap(err)
		}
		s.listener = listener
	}

	go func() {
		s.logger.Info("grpc server start", zap.String("listen", listen))
		serveErr := s.baseServer.Serve(listener)
		if errors.Is(serveErr, grpc.ErrServerStopped) || errors.Is(serveErr, net.ErrClosed) {
			serveErr = nil
		}
		s.serveErr <- serveErr
	}()

	if s.registry != nil {
		if err := s.registry.Register(ctx); err != nil {
			s.baseServer.Stop()
			_ = listener.Close()
			<-s.serveErr
			return cerrs.Wrap(err)
		}
	}

	select {
	case serveErr := <-s.serveErr:
		if s.registry != nil {
			_ = s.registry.DeRegister(ctx)
		}
		if serveErr == nil {
			return errors.New("grpc server stopped during startup")
		}
		return fmt.Errorf("serve grpc: %w", serveErr)
	default:
		s.lifecycle.markStarted()
		return nil
	}
}

func (s *GrpcServer) Wait() error {
	if err := s.lifecycle.claimWait(); err != nil {
		return err
	}
	return <-s.serveErr
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	shutdown, err := s.lifecycle.beginShutdown()
	if err != nil || !shutdown {
		return err
	}
	defer s.lifecycle.markStopped()

	var errs error
	if s.health != nil {
		s.health.Shutdown()
	}
	if s.registry != nil {
		if err := s.registry.DeRegister(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	stopped := make(chan struct{})
	go func() {
		s.baseServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-ctx.Done():
		s.baseServer.Stop()
		<-stopped
		errs = errors.Join(errs, ctx.Err())
	}
	errs = errors.Join(errs, s.Close())
	return errs
}

// Close releases resources owned by the server even if Start never succeeded.
// It is idempotent and is also called by Shutdown.
func (s *GrpcServer) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = closeResources(s.closers)
		s.closers = nil
	})
	return s.closeErr
}

func closeResources(closers []io.Closer) error {
	var errs error
	for i := len(closers) - 1; i >= 0; i-- {
		if closers[i] != nil {
			errs = errors.Join(errs, closers[i].Close())
		}
	}
	return errs
}

func newServerGroup(servers ...core.Server) *ServerGroup {
	group := &ServerGroup{shutdownTimeout: 10 * time.Second}
	for _, srv := range servers {
		group.addServer("", srv)
	}
	return group
}

func (s *ServerGroup) addServer(name string, server core.Server) {
	s.servers = append(s.servers, managedServer{name: name, server: server})
}

func NewGrpcServerGroup[T core.Server](
	config core.IConfig,
	logger core.ILogger,
	newGrpcServer func(core.IConfig, core.ILogger) (T, error),
) (*ServerGroup, error) {
	if err := validateServerDependencies(config, logger); err != nil {
		return nil, err
	}
	if newGrpcServer == nil {
		return nil, errors.New("grpc server factory is required")
	}
	grpcServer, err := newGrpcServer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("init grpc server: %w", err)
	}

	group := newServerGroup()
	group.addServer("grpc", grpcServer)
	if err := addMetricsServer(config, logger, group); err != nil {
		return nil, errors.Join(err, closeOwned(grpcServer))
	}
	return group, nil
}

func (s *ServerGroup) start(ctx context.Context) error {
	s.results = make(chan serverResult, len(s.servers))
	s.started = 0
	for i, srv := range s.servers {
		if err := srv.server.Start(ctx); err != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
			defer cancel()
			currentShutdownErr := srv.server.Shutdown(shutdownCtx)
			shutdownErr := s.shutdownStarted(shutdownCtx, i-1)
			return errors.Join(fmt.Errorf("start %s server: %w", srv.name, err), currentShutdownErr, shutdownErr)
		}
		s.started++
		go func(srv managedServer) {
			s.results <- serverResult{name: srv.name, err: srv.server.Wait()}
		}(srv)
	}
	return nil
}

func (s *ServerGroup) Run(ctx context.Context) error {
	if err := s.start(ctx); err != nil {
		return err
	}

	var (
		runErr   error
		received int
	)
	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			runErr = ctx.Err()
		}
	case result := <-s.results:
		received++
		runErr = serverResultError(result)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	shutdownErr := s.shutdown(shutdownCtx)

	var waitErr error
	for received < s.started {
		select {
		case result := <-s.results:
			received++
			if result.err != nil {
				waitErr = errors.Join(waitErr, fmt.Errorf("%s server failed: %w", result.name, result.err))
			}
		case <-shutdownCtx.Done():
			waitErr = errors.Join(waitErr, fmt.Errorf("wait for server shutdown: %w", shutdownCtx.Err()))
			return errors.Join(runErr, shutdownErr, waitErr)
		}
	}
	return errors.Join(runErr, shutdownErr, waitErr)
}

func serverResultError(result serverResult) error {
	if result.err == nil {
		return fmt.Errorf("%s server stopped unexpectedly", result.name)
	}
	return fmt.Errorf("%s server failed: %w", result.name, result.err)
}

func (s *ServerGroup) shutdown(ctx context.Context) error {
	return s.shutdownStarted(ctx, s.started-1)
}

func (s *ServerGroup) shutdownStarted(ctx context.Context, last int) error {
	var errs error
	for i := last; i >= 0; i-- {
		if err := s.servers[i].server.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("stop %s server: %w", s.servers[i].name, err))
		}
	}
	return errs
}

func publicMethodsWithHealth(methods []string) []string {
	publicMethods := append([]string{}, methods...)
	return append(publicMethods,
		grpc_health_v1.Health_Check_FullMethodName,
		grpc_health_v1.Health_Watch_FullMethodName,
	)
}

func unaryInterceptors(config core.IConfig, logger core.ILogger, opt GrpcServiceOption) []grpc.UnaryServerInterceptor {
	interceptors := []grpc.UnaryServerInterceptor{
		cogointerceptor.SrvCtxInterceptor(logger),
		cogointerceptor.RequestLogInterceptor(),
		cogointerceptor.ErrorInterceptor(),
		cogointerceptor.RecoveryInterceptor(),
		cogointerceptor.CycleCheckInterceptor(),
		cogointerceptor.BizInfoInterceptor(config),
		cogointerceptor.UserInfoInterceptorWithOptions(
			config,
			publicMethodsWithHealth(opt.PublicMethods),
			cogointerceptor.WithTokenRevocationChecker(opt.TokenRevocationChecker),
		),
	}
	return append(interceptors, opt.UnaryInterceptors...)
}

func registerUnaryServices(baseServer *grpc.Server, opt GrpcServiceOption) error {
	if opt.RegisterServices == nil {
		return nil
	}
	return opt.RegisterServices(baseServer)
}

func registerHealthServer(baseServer *grpc.Server, config core.IConfig) *health.Server {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(baseServer, healthServer)

	healthServer.SetServingStatus(config.GetRegistry().Name, grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	return healthServer
}

func metricsEnabled(config core.IConfig) bool {
	return config.GetMetrics().Enable
}

func addMetricsServer(config core.IConfig, logger core.ILogger, group *ServerGroup) error {
	if !metricsEnabled(config) {
		return nil
	}
	metricsServer, err := NewMetricsServer(config, logger)
	if err != nil {
		return fmt.Errorf("init metrics server: %w", err)
	}
	group.addServer("metrics", metricsServer)
	return nil
}

func closeOwned(server any) error {
	closer, ok := server.(io.Closer)
	if !ok {
		return nil
	}
	return closer.Close()
}

func validateServerDependencies(config core.IConfig, logger core.ILogger) error {
	if config == nil {
		return errors.New("server config is required")
	}
	if logger == nil {
		return errors.New("server logger is required")
	}
	return nil
}
