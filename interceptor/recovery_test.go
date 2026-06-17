package interceptor

import (
	"context"
	"testing"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testLogger struct{}

func (l *testLogger) Log(...any) error       { return nil }
func (l *testLogger) Debug(string, ...any)   {}
func (l *testLogger) Info(string, ...any)    {}
func (l *testLogger) Warn(string, ...any)    {}
func (l *testLogger) Error(string, ...any)   {}
func (l *testLogger) Fatal(string, ...any)   {}
func (l *testLogger) Panic(string, ...any)   {}
func (l *testLogger) AddGlobalFields(...any) {}

type testConfig struct{}

func (c *testConfig) Get(string) any { return nil }
func (c *testConfig) ReLoad() error  { return nil }

type testSrvCtx struct {
	logger core.ILogger
	config core.IConfig
}

func (s *testSrvCtx) Logger() core.ILogger                { return s.logger }
func (s *testSrvCtx) Config() core.IConfig                { return s.config }
func (s *testSrvCtx) SetField(core.SrvCtxKey, any)        {}
func (s *testSrvCtx) GetField(core.SrvCtxKey) (any, bool) { return nil, false }
func (s *testSrvCtx) SetBizInfo(core.IBizInfo)            {}
func (s *testSrvCtx) GetBizInfo() core.IBizInfo           { return nil }
func (s *testSrvCtx) SetUserInfo(core.IUserInfo)          {}
func (s *testSrvCtx) GetUserInfo() core.IUserInfo         { return nil }

func TestRecoveryInterceptorRecoverPanic(t *testing.T) {
	itc := RecoveryInterceptor()
	ctx := context.WithValue(context.Background(), core.SrvCtx, &testSrvCtx{
		logger: &testLogger{},
		config: &testConfig{},
	})

	info := &grpc.UnaryServerInfo{FullMethod: "/svc/method"}
	_, err := itc(ctx, "req", info, func(context.Context, any) (any, error) {
		panic("boom")
	})
	if err == nil {
		t.Fatalf("expected error after panic recovery")
	}

	var cerr *cerrs.CError
	if !cerrs.As(err, &cerr) {
		t.Fatalf("expected CError")
	}
	if cerr.GetCode() != cerrs.UnknownErrCode {
		t.Fatalf("expected unknown err code")
	}
}

func TestRecoveryInterceptorMissingSrvCtx(t *testing.T) {
	itc := RecoveryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/method"}
	_, err := itc(context.Background(), "req", info, func(context.Context, any) (any, error) {
		t.Fatalf("handler should not be called")
		return nil, nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
}
