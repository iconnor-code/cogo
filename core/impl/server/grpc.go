// Package server
package server

import (
	"context"
	"net"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	conf       core.IConfig
	logger     core.ILogger
	listener   net.Listener
	registry   core.IRegistry
	baseServer *grpc.Server
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
	listen, err := core.GetString(s.conf, "grpc.listen")
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, cerrs.Wrap(err)
	}
	s.listener = listener
	return s, nil
}

func (s *GrpcServer) Start() error {
	go func() {
		listen, _ := core.GetString(s.conf, "grpc.listen")
		s.logger.Info("grpc server start", zap.String("listen", listen))
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
