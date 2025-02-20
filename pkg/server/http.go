package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/iconnor-code/cogo/pkg/config"
	"github.com/iconnor-code/cogo/pkg/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

type HttpServer struct {
	conf   *config.Conf
	logger *logger.Logger
	wg     *sync.WaitGroup
	server *http.Server
}

func NewHttpServer(conf *config.Conf, logger *logger.Logger, wg *sync.WaitGroup) *HttpServer {
	return &HttpServer{
		conf:   conf,
		logger: logger,
		wg:     wg,
	}
}

func (s *HttpServer) Start(handler *runtime.ServeMux) {
	startHttpServer := func() *http.Server {
		httpSrv := &http.Server{
			Handler: handler,
			Addr:    s.conf.Http.Listen,
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Log().Info("Http Server Starting", zap.String("address", httpSrv.Addr))
			if err := httpSrv.ListenAndServe(); err != nil {
				s.logger.Log().Error("Http Server Error", zap.Error(err))
			}
		}()
		return httpSrv
	}

	startHttpsServer := func() *http.Server {
		httpsSrv := &http.Server{
			Handler: handler,
			Addr:    s.conf.Http.Listen,
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.logger.Log().Info("Https Server Starting", zap.String("listen", httpsSrv.Addr))
			if err := httpsSrv.ListenAndServeTLS(s.conf.Http.CertFile, s.conf.Http.KeyFile); err != nil {
				s.logger.Log().Error("Https Server Error", zap.Error(err))
			}
		}()
		return httpsSrv
	}

	if s.conf.Http.SSL {
		if s.conf.Http.CertFile == "" || s.conf.Http.KeyFile == "" {
			s.logger.Log().Fatal("Https Server Error: cert_file or key_file is empty")
		}
		s.server = startHttpsServer()
	} else {
		s.server = startHttpServer()
	}
}

func (s *HttpServer) WaitStop(stopSingal chan struct{}) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		<-stopSingal

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Log().Fatal("Http(s) Server Shutdown Error", zap.Error(err))
		}
		s.logger.Log().Info("Http(s) Server Exited")
	}()
}
