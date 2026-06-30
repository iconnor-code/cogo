// Package server
package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/client"
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/registry"
	cogointerceptor "github.com/iconnor-code/cogo/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type managedServer struct {
	name   string
	server core.IServer
}

type ServerGroup struct {
	servers []managedServer
}

type GrpcServer struct {
	conf       core.IConfig
	logger     core.ILogger
	listener   net.Listener
	registry   core.IRegistry
	baseServer *grpc.Server
}

type GrpcServiceOption struct {
	PublicMethods          []string
	TokenRevocationChecker cogointerceptor.TokenRevocationChecker
	UnaryInterceptors      []grpc.UnaryServerInterceptor
	RegisterServices       func(*grpc.Server) error
}

func WithGrpcRegistry(registry core.IRegistry) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*GrpcServer)
		server.registry = registry
		return nil
	}
}

func NewGrpcServer(config core.IConfig, logger core.ILogger, bs *grpc.Server, opts ...core.ServerOption) (*GrpcServer, error) {
	s := &GrpcServer{
		conf:       config,
		logger:     logger,
		baseServer: bs,
	}
	for _, opt := range opts {
		opt(s)
	}
	listen := s.conf.GetGRPC().Listen
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	s.listener = listener
	return s, nil
}

func NewGrpcServiceServer(config core.IConfig, logger core.ILogger, opt GrpcServiceOption) (*GrpcServer, error) {
	baseServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		unaryInterceptors(config, logger, opt)...,
	))
	if err := registerUnaryServices(baseServer, opt); err != nil {
		return nil, err
	}
	if err := registerHealthServer(baseServer, config); err != nil {
		return nil, err
	}

	opts := make([]core.ServerOption, 0, 1)
	if registryEnabled(config) {
		reg, err := newConsulRegistry(config, logger)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithGrpcRegistry(reg))
	}
	return NewGrpcServer(config, logger, baseServer, opts...)
}

func (s *GrpcServer) Start() error {
	go func() {
		s.logger.Info("grpc server start", zap.String("listen", s.conf.GetGRPC().Listen))
		err := s.baseServer.Serve(s.listener)
		if err != nil {
			s.logger.Error("grpc server failed", zap.Error(err))
		}
	}()

	if s.registry != nil {
		if err := s.registry.Register(context.Background()); err != nil {
			return cerrs.Wrap(err)
		}
	}

	return nil
}

func (s *GrpcServer) Stop() error {
	if s.registry != nil {
		if err := s.registry.DeRegister(context.Background()); err != nil {
			return err
		}
	}
	s.baseServer.GracefulStop()

	return nil
}

func NewServerGroup(servers ...core.IServer) *ServerGroup {
	group := &ServerGroup{}
	for _, srv := range servers {
		group.AddServer("", srv)
	}
	return group
}

func (s *ServerGroup) AddServer(name string, server core.IServer) {
	s.servers = append(s.servers, managedServer{name: name, server: server})
}

func NewGrpcServerGroup[T core.IServer](
	config core.IConfig,
	logger core.ILogger,
	newGrpcServer func(core.IConfig, core.ILogger) (T, error),
) (*ServerGroup, error) {
	grpcServer, err := newGrpcServer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("init grpc server: %w", err)
	}

	group := NewServerGroup()
	group.AddServer("grpc", grpcServer)
	if err := addMetricsServer(config, logger, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *ServerGroup) Start() error {
	for i, srv := range s.servers {
		if err := srv.server.Start(); err != nil {
			_ = s.stopStarted(i - 1)
			return fmt.Errorf("start %s server: %w", srv.name, err)
		}
	}
	return nil
}

func (s *ServerGroup) Stop() error {
	return s.stopStarted(len(s.servers) - 1)
}

func (s *ServerGroup) stopStarted(last int) error {
	var errs error
	for i := last; i >= 0; i-- {
		if err := s.servers[i].server.Stop(); err != nil {
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
		cogointerceptor.SrvCtxInterceptor(config, logger),
		cogointerceptor.RecoveryInterceptor(),
		cogointerceptor.CycleCheckInterceptor(),
		cogointerceptor.RequestLogInterceptor(),
		cogointerceptor.BizInfoInterceptor(),
		cogointerceptor.UserInfoInterceptorWithOptions(
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

func registerHealthServer(baseServer *grpc.Server, config core.IConfig) error {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(baseServer, healthServer)

	healthServer.SetServingStatus(config.GetRegistry().Name, grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	return nil
}

func newConsulRegistry(config core.IConfig, logger core.ILogger) (core.IRegistry, error) {
	consul, err := client.NewConsul(config)
	if err != nil {
		return nil, err
	}
	return registry.NewRegistry(config, logger, registry.WithConsulClient(consul))
}

func registryEnabled(config core.IConfig) bool {
	registryConf := config.GetRegistry()
	return config.GetConsul().Address != "" &&
		registryConf.Name != "" &&
		registryConf.Address != "" &&
		registryConf.Port != 0
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
	group.AddServer("metrics", metricsServer)
	return nil
}
