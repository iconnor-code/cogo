package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type captureLogger struct{ entries []string }

func (l *captureLogger) add(message string, fields ...any) {
	l.entries = append(l.entries, fmt.Sprint(message, fields))
}
func (l *captureLogger) Log(...any) error                    { return nil }
func (l *captureLogger) Debug(message string, fields ...any) { l.add("debug:"+message, fields...) }
func (l *captureLogger) Info(message string, fields ...any)  { l.add("info:"+message, fields...) }
func (l *captureLogger) Warn(message string, fields ...any)  { l.add("warn:"+message, fields...) }
func (l *captureLogger) Error(message string, fields ...any) { l.add("error:"+message, fields...) }
func (l *captureLogger) Fatal(string, ...any)                {}
func (l *captureLogger) Panic(string, ...any)                {}
func (l *captureLogger) AddGlobalFields(...any)              {}

func TestCustomErrorMapping(t *testing.T) {
	tests := []struct {
		name string
		kind cerrs.Kind
		want codes.Code
	}{
		{name: "invalid argument", kind: cerrs.KindInvalidArgument, want: codes.InvalidArgument},
		{name: "permission denied", kind: cerrs.KindPermissionDenied, want: codes.PermissionDenied},
		{name: "not found", kind: cerrs.KindNotFound, want: codes.NotFound},
		{name: "already exists", kind: cerrs.KindAlreadyExists, want: codes.AlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grpcCodeForKind(tt.kind); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestErrorInterceptorMapsWrappedCustomError(t *testing.T) {
	interceptor := ErrorInterceptor()
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		return nil, fmt.Errorf("use case: %w", cerrs.NewWithCode(4030, "permission denied"))
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("want %v, got %v", codes.PermissionDenied, status.Code(err))
	}
}

func TestErrorInterceptorHidesUnknownError(t *testing.T) {
	interceptor := ErrorInterceptor()
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		return nil, errors.New("database password leaked")
	})
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error occurred" {
		t.Fatalf("unexpected transport error: %v", err)
	}
}

func TestErrorInterceptorHidesRecoveredPanic(t *testing.T) {
	logger := &captureLogger{}
	serviceContext := srvctx.NewSrvCtx(logger)
	ctx := context.WithValue(context.Background(), core.SrvCtx, serviceContext)
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Panic"}

	_, err := ErrorInterceptor()(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return RecoveryInterceptor()(ctx, req, info, func(context.Context, any) (any, error) {
			panic("database password leaked")
		})
	})
	if status.Code(err) != codes.Internal || status.Convert(err).Message() != "internal error occurred" {
		t.Fatalf("unexpected transport error: %v", err)
	}
}

func TestRequestLogDoesNotLogPayload(t *testing.T) {
	logger := &captureLogger{}
	serviceContext := srvctx.NewSrvCtx(logger)
	ctx := context.WithValue(context.Background(), core.SrvCtx, serviceContext)
	secret := "super-secret-password"

	_, err := RequestLogInterceptor()(ctx, secret, &grpc.UnaryServerInfo{FullMethod: "/account.AuthService/Login"}, func(context.Context, any) (any, error) {
		return "super-secret-token", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	logged := strings.Join(logger.entries, " ")
	if strings.Contains(logged, "super-secret") {
		t.Fatalf("request log leaked payload: %s", logged)
	}
}

func TestRequestLogUsesFinalStatusSeverity(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantPrefix string
	}{
		{name: "business failure", err: status.Error(codes.InvalidArgument, "invalid"), wantPrefix: "warn:"},
		{name: "server failure", err: status.Error(codes.Internal, "internal"), wantPrefix: "error:"},
		{name: "canceled", err: status.Error(codes.Canceled, "canceled"), wantPrefix: "info:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &captureLogger{}
			serviceContext := srvctx.NewSrvCtx(logger)
			ctx := context.WithValue(context.Background(), core.SrvCtx, serviceContext)
			_, _ = RequestLogInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Service/Call"}, func(context.Context, any) (any, error) {
				return nil, tt.err
			})
			if len(logger.entries) != 1 || !strings.HasPrefix(logger.entries[0], tt.wantPrefix) {
				t.Fatalf("unexpected log entries: %v", logger.entries)
			}
		})
	}
}

var _ core.ILogger = (*captureLogger)(nil)

func TestErrorInterceptorDoesNotExposeWrappedCause(t *testing.T) {
	interceptor := ErrorInterceptor()
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		return nil, cerrs.WrapKind(errors.New("database password leaked"), cerrs.KindInvalidArgument, "invalid request")
	})
	if got := status.Convert(err).Message(); got != "invalid request" {
		t.Fatalf("unexpected public message %q", got)
	}
}
