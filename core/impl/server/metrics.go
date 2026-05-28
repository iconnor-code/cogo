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
	conf   map[string]any
	logger core.ILogger
	server *http.Server
}

func NewMetricsServer(config core.IConfig, logger core.ILogger) (*MetricsServer, error) {
	conf, ok := config.Get("metrics").(map[string]any)
	if !ok {
		return nil, cerrs.New("metrics config error")
	}
	s := &MetricsServer{
		conf:   conf,
		logger: logger,
	}
	return s, nil
}

func (s *MetricsServer) Start() error {
	httpSrv := &http.Server{
		Addr:    s.conf["listen"].(string),
		Handler: promhttp.Handler(),
	}
	listener, err := net.Listen("tcp", httpSrv.Addr)
	if err != nil {
		return cerrs.Wrap(err)
	}
	go func() {
		s.logger.Info("metrics server start", zap.String("listen", s.conf["listen"].(string)))
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
