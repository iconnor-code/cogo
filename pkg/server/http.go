package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HttpServer struct {
	conf   map[string]any
	logger core.ILogger
	// wg      *sync.WaitGroup
	handler *runtime.ServeMux
	server  *http.Server
}

func WithHttpHandler(handler *runtime.ServeMux) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*HttpServer)
		server.handler = handler
		return nil
	}
}

func WithHttpConfig(conf core.IConfig) core.ServerOption {
	return func(s core.IServer) error {
		server := s.(*HttpServer)
		confMap := conf.Get("http").(map[string]any)
		if confMap == nil {
			return cerrs.New("http config is not found")
		}
		server.conf = confMap
		return nil
	}
}

func WithHttpLogger(logger core.ILogger) core.ServerOption {
	return func(s core.IServer) error {
		s.(*HttpServer).logger = logger
		return nil
	}
}

func NewHttpServer(opts ...core.ServerOption) *HttpServer {
	s := &HttpServer{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *HttpServer) Start() error {
	startHttpServer := func() (*http.Server, error) {
		httpSrv := &http.Server{
			Handler: s.handler,
			Addr:    s.conf["listen"].(string),
		}
		go func() {
			s.logger.Info("http server start", zap.String("listen", s.conf["listen"].(string)))
			err := httpSrv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				s.logger.Error("http server start failed", zap.Error(err))
			}
		}()
		return httpSrv, nil
	}

	startHttpsServer := func(sslConfMap map[string]string) (*http.Server, error) {
		httpsSrv := &http.Server{
			Handler: s.handler,
			Addr:    s.conf["listen"].(string),
		}
		go func() {
			s.logger.Info("https server start", zap.String("listen", s.conf["listen"].(string)))
			err := httpsSrv.ListenAndServeTLS(sslConfMap["cert_file"], sslConfMap["key_file"])
			if err != nil && err != http.ErrServerClosed {
				s.logger.Error("https server start failed", zap.Error(err))
			}
		}()
		return httpsSrv, nil
	}

	var err error
	sslConf, ok := s.conf["ssl"]
	if ok && sslConf != nil {
		sslConfMap, ok := sslConf.(map[string]string)
		if !ok {
			return cerrs.New(fmt.Sprintf("https ssl config is error: %+v", sslConf))
		}
		s.server, err = startHttpsServer(sslConfMap)
		if err != nil {
			return err
		}
	} else {
		s.server, err = startHttpServer()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HttpServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
