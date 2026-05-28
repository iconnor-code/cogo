package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HTTPServer struct {
	conf   map[string]any
	logger core.ILogger
	// wg      *sync.WaitGroup
	handler *runtime.ServeMux
	server  *http.Server
}

func NewHTTPServer(config core.IConfig, logger core.ILogger, handler *runtime.ServeMux) (*HTTPServer, error) {
	conf := config.Get("http").(map[string]any)
	if conf == nil {
		return nil, cerrs.New("http config is not found")
	}
	s := &HTTPServer{
		conf:    conf,
		logger:  logger,
		handler: handler,
	}
	return s, nil
}

func (s *HTTPServer) Start() error {
	startHTTPServer := func() (*http.Server, error) {
		httpSrv := &http.Server{
			Handler: s.handler,
			Addr:    s.conf["listen"].(string),
		}
		listener, err := net.Listen("tcp", httpSrv.Addr)
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		go func() {
			s.logger.Info("http server start", zap.String("listen", s.conf["listen"].(string)))
			err := httpSrv.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				s.logger.Error("http server start failed", zap.Error(err))
			}
		}()
		return httpSrv, nil
	}

	startHTTPSServer := func(sslConfMap map[string]string) (*http.Server, error) {
		httpsSrv := &http.Server{
			Handler: s.handler,
			Addr:    s.conf["listen"].(string),
		}
		listener, err := net.Listen("tcp", httpsSrv.Addr)
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		go func() {
			s.logger.Info("https server start", zap.String("listen", s.conf["listen"].(string)))
			err := httpsSrv.ServeTLS(listener, sslConfMap["cert_file"], sslConfMap["key_file"])
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
		s.server, err = startHTTPSServer(sslConfMap)
		if err != nil {
			return err
		}
	} else {
		s.server, err = startHTTPServer()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HTTPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
