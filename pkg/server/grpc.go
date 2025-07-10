package server

import (
	"context"
	"errors"
	"net"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	conf       map[string]any
	logger     core.ILogger
	listener   net.Listener
	registry   core.IRegistry
	baseServer *grpc.Server
}

func WithGrpcConfig(conf core.IConfig) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*GrpcServer)
		confMap := conf.Get("grpc").(map[string]any)
		if confMap == nil {
			return errors.New("grpc config is not found")
		}
		server.conf = confMap
		return nil
	}
}

func WithGrpcLogger(logger core.ILogger) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*GrpcServer)
		server.logger = logger
		return nil
	}
}

func WithGrpcRegistry(registry core.IRegistry) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*GrpcServer)
		server.registry = registry
		return nil
	}
}

func WithGrpcBaseServer(baseServer *grpc.Server) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*GrpcServer)
		server.baseServer = baseServer
		return nil
	}
}

func NewGrpcServer(opts ...core.ServerOption) (*GrpcServer, error) {
	s := &GrpcServer{}
	for _, opt := range opts {
		opt(s)
	}
	listener, err := net.Listen("tcp", s.conf["listen"].(string))
	if err != nil {
		return nil, cerrs.Wrap("failed to listen", err)
	}
	s.listener = listener
	return s, nil
}

func (s *GrpcServer) Start() error {
	go func() {
		s.logger.Info("grpc server start", zap.String("listen", s.conf["listen"].(string)))
		err := s.baseServer.Serve(s.listener)
		if err != nil {
			s.logger.Error("grpc server start failed", zap.Error(err))
		}
	}()

	if s.registry != nil {
		s.logger.Info("grpc server register", zap.String("name", s.conf["name"].(string)), zap.String("addr", s.conf["addr"].(string)))
		if err := s.registry.Register(context.Background(), s.conf["name"].(string), s.conf["addr"].(string)); err != nil {
			return cerr.WithStack(err)
		}
	}

	return nil
}

func (s *GrpcServer) Stop() error {
	if s.registry != nil {
		if err := s.registry.Deregister(context.Background(), s.conf["name"].(string), s.conf["addr"].(string)); err != nil {
			return err
		}
	}
	if err := s.listener.Close(); err != nil {
		return err
	}

	return nil
}
