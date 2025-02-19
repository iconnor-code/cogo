package server

import (
	"context"
	"net"
	"sync"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/cerr"
	"github.com/iconnor-code/cogo/pkg/logger"
	"github.com/iconnor-code/cogo/pkg/registry"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	conf     *config.Conf
	logger   *logger.Logger
	wg       *sync.WaitGroup
	listener net.Listener
	registry *registry.GrpcRegister
}

func NewGrpcServer(conf *config.Conf, logger *logger.Logger, wg *sync.WaitGroup, registry *registry.GrpcRegister) *GrpcServer {
	grpcListener, err := net.Listen("tcp", conf.Grpc.Listen)
	if err != nil {
		logger.Log().Fatal("Grpc Listen Error", zap.Error(err))
	}
	return &GrpcServer{
		conf:     conf,
		logger:   logger,
		wg:       wg,
		listener: grpcListener,
		registry: registry,
	}
}

func (s *GrpcServer) Start(baseServer *grpc.Server) error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Log().Info("Grpc Server Starting", zap.String("listen", s.conf.Grpc.Listen))
		if err := baseServer.Serve(s.listener); err != nil {
			s.logger.Log().Error("Grpc Server Error", zap.Error(err))
		}
	}()

	if err := s.registry.Register(context.Background()); err != nil {
		return cerr.WithStack(err)
	}

	return nil
}

func (s *GrpcServer) WaitStop(stopSingal chan struct{}) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		<-stopSingal

		if err := s.registry.Deregister(context.Background()); err != nil {
			s.logger.Log().Error("Grpc Server Deregister Error", zap.Error(err))
		}

		if err := s.listener.Close(); err != nil {
			s.logger.Log().Error("Grpc Server Shutdown Error", zap.Error(err))
		}

		s.logger.Log().Info("Grpc Server Exited")
	}()
}
