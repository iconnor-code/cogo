package server

import (
	"context"
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

type HTTPServer struct {
	config  core.IConfig
	logger  core.ILogger
	handler http.Handler
	server  *http.Server
}

func NewHTTPServer(config core.IConfig, logger core.ILogger, handler *runtime.ServeMux) (*HTTPServer, error) {
	return NewHTTPServerWithHandler(config, logger, handler)
}

func NewHTTPServerWithHandler(config core.IConfig, logger core.ILogger, handler http.Handler) (*HTTPServer, error) {
	s := &HTTPServer{
		config:  config,
		logger:  logger,
		handler: handler,
	}
	return s, nil
}

func NewHTTPGatewayServerGroup[T core.IServer](
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

	mux, err := newGatewayMux(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("init grpc gateway: %w", err)
	}
	handler := NewSwaggerHandler(mux, swaggerOption)
	httpServer, err := NewHTTPServerWithHandler(config, logger, handler)
	if err != nil {
		return nil, fmt.Errorf("init http server: %w", err)
	}

	group := NewServerGroup()
	group.AddServer("grpc", grpcServer)
	group.AddServer("http", httpServer)
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

func (s *HTTPServer) Start() error {
	startHTTPServer := func() (*http.Server, error) {
		listen := s.config.GetHTTP().Listen
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
		listen := s.config.GetHTTP().Listen
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
	httpConf := s.config.GetHTTP()
	if httpConf.SSL.CertFile != "" || httpConf.SSL.KeyFile != "" {
		if httpConf.SSL.CertFile == "" {
			return cerrs.New("https ssl cert_file is required")
		}
		if httpConf.SSL.KeyFile == "" {
			return cerrs.New("https ssl key_file is required")
		}
		s.server, err = startHTTPSServer(map[string]string{
			"cert_file": httpConf.SSL.CertFile,
			"key_file":  httpConf.SSL.KeyFile,
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
