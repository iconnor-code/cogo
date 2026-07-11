package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/token"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GatewayRegister func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
type GatewayMuxFactory func(context.Context, core.IConfig) (*runtime.ServeMux, error)

const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultIdleTimeout       = 60 * time.Second
)

type HTTPServer struct {
	config    core.IConfig
	logger    core.ILogger
	handler   http.Handler
	server    *http.Server
	serveErr  chan error
	lifecycle componentLifecycle
}

type gatewayHTTPServer struct {
	config        core.IConfig
	logger        core.ILogger
	newGatewayMux GatewayMuxFactory
	swagger       SwaggerOption
	server        *HTTPServer
	cancel        context.CancelFunc
	lifecycle     componentLifecycle
}

func NewHTTPServer(config core.IConfig, logger core.ILogger, handler *runtime.ServeMux) (*HTTPServer, error) {
	return NewHTTPServerWithHandler(config, logger, handler)
}

func NewHTTPServerWithHandler(config core.IConfig, logger core.ILogger, handler http.Handler) (*HTTPServer, error) {
	s := &HTTPServer{
		config:   config,
		logger:   logger,
		handler:  handler,
		serveErr: make(chan error, 1),
	}
	return s, nil
}

func NewHTTPGatewayServerGroup[T core.Server](
	config core.IConfig,
	logger core.ILogger,
	newGrpcServer func(core.IConfig, core.ILogger) (T, error),
	newGatewayMux GatewayMuxFactory,
	swaggerOption SwaggerOption,
) (*ServerGroup, error) {
	grpcServer, err := newGrpcServer(config, logger)
	if err != nil {
		return nil, fmt.Errorf("init grpc server: %w", err)
	}

	group := NewServerGroup()
	group.AddServer("grpc", grpcServer)
	group.AddServer("http", &gatewayHTTPServer{
		config:        config,
		logger:        logger,
		newGatewayMux: newGatewayMux,
		swagger:       swaggerOption,
	})
	if err := addMetricsServer(config, logger, group); err != nil {
		return nil, err
	}
	return group, nil
}

func NewGatewayMux(ctx context.Context, config core.IConfig, registers ...GatewayRegister) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
	)

	endpoint, err := GRPCEndpoint(config)
	if err != nil {
		return nil, err
	}
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	for _, register := range registers {
		if err := register(ctx, mux, endpoint, opts); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func incomingHeaderMatcher(header string) (string, bool) {
	if strings.EqualFold(header, token.JwtTokenKey) {
		return token.JwtTokenKey, true
	}
	if strings.EqualFold(header, "x-biz-id") {
		return "biz_id", true
	}
	if strings.EqualFold(header, "x-biz-name") {
		return "biz_name", true
	}
	return runtime.DefaultHeaderMatcher(header)
}

func (s *HTTPServer) Start(context.Context) (err error) {
	if err := s.lifecycle.beginStart(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.lifecycle.markStartFailed()
		}
	}()

	httpConf := s.config.GetHTTP()
	if strings.TrimSpace(httpConf.Listen) == "" {
		return errors.New("http listen address is required")
	}

	startHTTPServer := func() (*http.Server, net.Listener, error) {
		listen := httpConf.Listen
		httpSrv := &http.Server{
			Handler:           s.handler,
			Addr:              listen,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			IdleTimeout:       defaultIdleTimeout,
		}
		listener, err := net.Listen("tcp", httpSrv.Addr)
		if err != nil {
			return nil, nil, cerrs.Wrap(err)
		}
		return httpSrv, listener, nil
	}

	startHTTPSServer := func(certFile, keyFile string) (*http.Server, net.Listener, error) {
		certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, nil, cerrs.Wrap(err)
		}
		listen := httpConf.Listen
		httpsSrv := &http.Server{
			Handler:           s.handler,
			Addr:              listen,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			IdleTimeout:       defaultIdleTimeout,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{certificate},
				MinVersion:   tls.VersionTLS12,
			},
		}
		listener, err := net.Listen("tcp", httpsSrv.Addr)
		if err != nil {
			return nil, nil, cerrs.Wrap(err)
		}
		return httpsSrv, listener, nil
	}

	var (
		listener net.Listener
		protocol = "http"
	)
	if httpConf.SSL.CertFile != "" || httpConf.SSL.KeyFile != "" {
		if httpConf.SSL.CertFile == "" {
			return cerrs.New("https ssl cert_file is required")
		}
		if httpConf.SSL.KeyFile == "" {
			return cerrs.New("https ssl key_file is required")
		}
		s.server, listener, err = startHTTPSServer(httpConf.SSL.CertFile, httpConf.SSL.KeyFile)
		protocol = "https"
		if err != nil {
			return err
		}
	} else {
		s.server, listener, err = startHTTPServer()
		if err != nil {
			return err
		}
	}
	go func() {
		s.logger.Info(protocol+" server start", zap.String("listen", httpConf.Listen))
		var serveErr error
		if protocol == "https" {
			serveErr = s.server.ServeTLS(listener, "", "")
		} else {
			serveErr = s.server.Serve(listener)
		}
		if errors.Is(serveErr, http.ErrServerClosed) || errors.Is(serveErr, net.ErrClosed) {
			serveErr = nil
		}
		s.serveErr <- serveErr
	}()
	s.lifecycle.markStarted()
	return nil
}

func (s *HTTPServer) Wait() error {
	if err := s.lifecycle.claimWait(); err != nil {
		return err
	}
	return <-s.serveErr
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
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

func (s *gatewayHTTPServer) Start(ctx context.Context) (err error) {
	if err := s.lifecycle.beginStart(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			s.lifecycle.markStartFailed()
		}
	}()

	gatewayCtx, cancel := context.WithCancel(context.Background())
	mux, err := s.newGatewayMux(gatewayCtx, s.config)
	if err != nil {
		cancel()
		return fmt.Errorf("init grpc gateway: %w", err)
	}

	httpServer, err := NewHTTPServerWithHandler(s.config, s.logger, NewSwaggerHandler(mux, s.swagger))
	if err != nil {
		cancel()
		return fmt.Errorf("init http server: %w", err)
	}
	if err := httpServer.Start(ctx); err != nil {
		cancel()
		return err
	}

	s.server = httpServer
	s.cancel = cancel
	s.lifecycle.markStarted()
	return nil
}

func (s *gatewayHTTPServer) Wait() error {
	if err := s.lifecycle.claimWait(); err != nil {
		return err
	}
	return s.server.Wait()
}

func (s *gatewayHTTPServer) Shutdown(ctx context.Context) error {
	shutdown, err := s.lifecycle.beginShutdown()
	if err != nil || !shutdown {
		return err
	}
	defer s.lifecycle.markStopped()

	err = s.server.Shutdown(ctx)
	s.cancel()
	return err
}

func GRPCEndpoint(config core.IConfig) (string, error) {
	grpcConf := config.GetGRPC()
	if grpcConf.GatewayEndpoint != "" {
		return grpcConf.GatewayEndpoint, nil
	}

	registryConf := config.GetRegistry()
	address := registryConf.Address
	if address == "" || address == "0.0.0.0" {
		address = "127.0.0.1"
	}

	if registryConf.Port != 0 {
		return fmt.Sprintf("%s:%d", address, registryConf.Port), nil
	}

	listen := grpcConf.Listen
	if strings.HasPrefix(listen, ":") {
		return "127.0.0.1" + listen, nil
	}
	if listen == "" {
		return "", cerrs.New("grpc endpoint is required")
	}
	return listen, nil
}
