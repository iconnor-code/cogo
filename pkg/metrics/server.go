package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type PrometheusServer struct {
	conf   *config.Conf
	logger *logger.Logger
	wg     *sync.WaitGroup
	server *http.Server
}

func NewPrometheusServer(conf *config.Conf, logger *logger.Logger, wg *sync.WaitGroup) *PrometheusServer {
	return &PrometheusServer{
		conf:   conf,
		logger: logger,
		wg:     wg,
	}
}

func (s *PrometheusServer) Start() {
	httpSrv := &http.Server{
		Addr:    s.conf.Metrics.Listen,
		Handler: promhttp.Handler(),
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.logger.Log().Info("Prometheus Server Starting", zap.String("address", s.conf.Metrics.Listen))
		if err := httpSrv.ListenAndServe(); err != nil {
			s.logger.Log().Error("Prometheus Server Error", zap.Error(err))
		}
	}()
	s.server = httpSrv
}

func (s *PrometheusServer) WaitStop(stopSingal chan struct{}) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		<-stopSingal

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Log().Fatal("Prometheus Server Shutdown Error", zap.Error(err))
		}
		s.logger.Log().Info("Prometheus Server Exited")
	}()
}