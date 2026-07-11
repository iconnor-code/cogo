package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	config    core.IConfig
	logger    core.ILogger
	server    *http.Server
	serveErr  chan error
	lifecycle componentLifecycle
}

func NewMetricsServer(config core.IConfig, logger core.ILogger) (*MetricsServer, error) {
	s := &MetricsServer{
		config:   config,
		logger:   logger,
		serveErr: make(chan error, 1),
	}
	return s, nil
}

func (s *MetricsServer) Start(context.Context) (err error) {
	if err := s.lifecycle.beginStart(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.lifecycle.markStartFailed()
		}
	}()

	listen := s.config.GetMetrics().Listen
	if strings.TrimSpace(listen) == "" {
		return errors.New("metrics listen address is required")
	}
	httpSrv := &http.Server{
		Addr:              listen,
		Handler:           promhttp.Handler(),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		IdleTimeout:       defaultIdleTimeout,
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
	s.lifecycle.markStarted()
	return nil
}

func (s *MetricsServer) Wait() error {
	if err := s.lifecycle.claimWait(); err != nil {
		return err
	}
	return <-s.serveErr
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	shutdown, err := s.lifecycle.beginShutdown()
	if err != nil || !shutdown {
		return err
	}
	defer s.lifecycle.markStopped()
	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Join(err, s.server.Close())
	}
	return nil
}
