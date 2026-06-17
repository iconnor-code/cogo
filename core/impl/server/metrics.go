package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	config core.IConfig
	logger core.ILogger
	server *http.Server
}

func NewMetricsServer(config core.IConfig, logger core.ILogger) (*MetricsServer, error) {
	s := &MetricsServer{
		config: config,
		logger: logger,
	}
	return s, nil
}

func (s *MetricsServer) Start() error {
	listen, err := core.GetString(s.config, "metrics.listen")
	if err != nil {
		return cerrs.Wrap(err)
	}
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
		if listenErr := httpSrv.Serve(listener); listenErr != nil && listenErr != http.ErrServerClosed {
			s.logger.Error("metrics server start failed", zap.Error(listenErr))
		}
	}()
	s.server = httpSrv
	return nil
}

func (s *MetricsServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if s.server == nil {
		return nil
	}
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
