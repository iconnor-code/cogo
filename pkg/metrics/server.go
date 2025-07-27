// Package metrics
package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/iconnor-code/cogo/core"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	conf   core.IConfig
	logger core.ILogger
	server *http.Server
}

func WithMetricsConfig(conf core.IConfig) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*MetricsServer)
		server.conf = conf
		return nil
	}
}

func WithMetricsLogger(logger core.ILogger) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*MetricsServer)
		server.logger = logger
		return nil
	}
}

func NewMetricsServer(opts ...core.ServerOption) (*MetricsServer, error) {
	s := &MetricsServer{}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func (s *MetricsServer) Start() error {
	httpSrv := &http.Server{
		Addr:    s.conf.Get("metrics.listen").(string),
		Handler: promhttp.Handler(),
	}
	go func() {
		s.logger.Info("metrics server start", zap.String("listen", s.conf.Get("metrics.listen").(string)))
		if listenErr := httpSrv.ListenAndServe(); listenErr != nil {
			s.logger.Error("metrics server start failed", zap.Error(listenErr))
		}
	}()
	s.server = httpSrv
	return nil
}

func (s *MetricsServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
