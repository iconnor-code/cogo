package server

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	config   core.IConfig
	logger   core.ILogger
	server   *http.Server
	serveErr chan error
}

func NewMetricsServer(config core.IConfig, logger core.ILogger) (*MetricsServer, error) {
	s := &MetricsServer{
		config:   config,
		logger:   logger,
		serveErr: make(chan error, 1),
	}
	return s, nil
}

func (s *MetricsServer) Start(context.Context) error {
	listen := s.config.GetMetrics().Listen
	httpSrv := &http.Server{
		Addr:    listen,
		Handler: promhttp.Handler(),
	}
	listener, err := net.Listen("tcp", httpSrv.Addr)
	if err != nil {
		return cerrs.Wrap(err)
	}
	go func() {
		s.logger.Info("metrics server start", zap.String("listen", listen))
		serveErr := httpSrv.Serve(listener)
		if errors.Is(serveErr, http.ErrServerClosed) || errors.Is(serveErr, net.ErrClosed) {
			serveErr = nil
		}
		s.serveErr <- serveErr
	}()
	s.server = httpSrv
	return nil
}

func (s *MetricsServer) Wait() error {
	return <-s.serveErr
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Join(err, s.server.Close())
	}
	return nil
}
