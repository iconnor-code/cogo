package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

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

type HTTPServer struct {
	config   core.IConfig
	logger   core.ILogger
	handler  http.Handler
	server   *http.Server
	serveErr chan error
}

type contextLifecycle struct {
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
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

	gatewayCtx, cancelGateway := context.WithCancel(context.Background())
	mux, err := newGatewayMux(gatewayCtx, config)
	if err != nil {
		cancelGateway()
		return nil, fmt.Errorf("init grpc gateway: %w", err)
	}
	handler := NewSwaggerHandler(mux, swaggerOption)
	httpServer, err := NewHTTPServerWithHandler(config, logger, handler)
	if err != nil {
		cancelGateway()
		return nil, fmt.Errorf("init http server: %w", err)
	}

	group := NewServerGroup()
	group.AddServer("grpc", grpcServer)
	group.AddServer("gateway", newContextLifecycle(cancelGateway))
	group.AddServer("http", httpServer)
	if err := addMetricsServer(config, logger, group); err != nil {
		cancelGateway()
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

func (s *HTTPServer) Start(context.Context) error {
	startHTTPServer := func() (*http.Server, net.Listener, error) {
		listen := s.config.GetHTTP().Listen
		httpSrv := &http.Server{
			Handler: s.handler,
			Addr:    listen,
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
		listen := s.config.GetHTTP().Listen
		httpsSrv := &http.Server{
			Handler: s.handler,
			Addr:    listen,
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
		err      error
		listener net.Listener
		protocol = "http"
	)
	httpConf := s.config.GetHTTP()
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
	return nil
}

func (s *HTTPServer) Wait() error {
	return <-s.serveErr
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Join(err, s.server.Close())
	}
	return nil
}

func newContextLifecycle(cancel context.CancelFunc) *contextLifecycle {
	return &contextLifecycle{cancel: cancel, done: make(chan struct{})}
}

func (s *contextLifecycle) Start(context.Context) error { return nil }

func (s *contextLifecycle) Wait() error {
	<-s.done
	return nil
}

func (s *contextLifecycle) Shutdown(context.Context) error {
	s.once.Do(func() {
		s.cancel()
		close(s.done)
	})
	return nil
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
