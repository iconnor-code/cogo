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
	config core.IConfig
	logger core.ILogger
	// wg      *sync.WaitGroup
	handler *runtime.ServeMux
	server  *http.Server
}

func NewHTTPServer(config core.IConfig, logger core.ILogger, handler *runtime.ServeMux) (*HTTPServer, error) {
	s := &HTTPServer{
		config:  config,
		logger:  logger,
		handler: handler,
	}
	return s, nil
}

func (s *HTTPServer) Start() error {
	startHTTPServer := func() (*http.Server, error) {
		listen, err := core.GetString(s.config, "http.listen")
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		httpSrv := &http.Server{
			Handler: s.handler,
			Addr:    listen,
		}
		listener, err := net.Listen("tcp", httpSrv.Addr)
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		go func() {
			s.logger.Info("http server start", zap.String("listen", listen))
			err := httpSrv.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				s.logger.Error("http server start failed", zap.Error(err))
			}
		}()
		return httpSrv, nil
	}

	startHTTPSServer := func(sslConfMap map[string]string) (*http.Server, error) {
		listen, err := core.GetString(s.config, "http.listen")
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		httpsSrv := &http.Server{
			Handler: s.handler,
			Addr:    listen,
		}
		listener, err := net.Listen("tcp", httpsSrv.Addr)
		if err != nil {
			return nil, cerrs.Wrap(err)
		}
		go func() {
			s.logger.Info("https server start", zap.String("listen", listen))
			err := httpsSrv.ServeTLS(listener, sslConfMap["cert_file"], sslConfMap["key_file"])
			if err != nil && err != http.ErrServerClosed {
				s.logger.Error("https server start failed", zap.Error(err))
			}
		}()
		return httpsSrv, nil
	}

	var err error
	sslConf := s.config.Get("http.ssl")
	if sslConf != nil {
		sslConfMap, ok := sslConf.(map[string]any)
		if !ok {
			return cerrs.New(fmt.Sprintf("https ssl config is error: %+v", sslConf))
		}
		certFile, ok := sslConfMap["cert_file"].(string)
		if !ok || certFile == "" {
			return cerrs.New("https ssl cert_file is required")
		}
		keyFile, ok := sslConfMap["key_file"].(string)
		if !ok || keyFile == "" {
			return cerrs.New("https ssl key_file is required")
		}
		s.server, err = startHTTPSServer(map[string]string{
			"cert_file": certFile,
			"key_file":  keyFile,
		})
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
