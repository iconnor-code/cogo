package interceptor

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iconnor-code/cogo/core"
	cogoconfig "github.com/iconnor-code/cogo/core/impl/config"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	tokenpkg "github.com/iconnor-code/cogo/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func makeAccessToken(secret string, claims jwt.MapClaims) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return s
}

func TestToUint32(t *testing.T) {
	tests := []struct {
		name    string
		in      any
		want    uint32
		wantErr bool
	}{
		{name: "uint32", in: uint32(1), want: 1},
		{name: "int", in: int(2), want: 2},
		{name: "float64", in: float64(3), want: 3},
		{name: "negative", in: int(-1), wantErr: true},
		{name: "overflow", in: int64(math.MaxUint32 + 1), wantErr: true},
		{name: "uint64 overflow", in: uint64(math.MaxUint32 + 1), wantErr: true},
		{name: "float64 fraction", in: float64(1.5), wantErr: true},
		{name: "unsupported", in: "x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toUint32(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("want %d got %d", tt.want, got)
			}
		})
	}
}

func TestUserInfoInterceptorSuccess(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	token := makeAccessToken("secret", jwt.MapClaims{
		"user_id":    float64(123),
		"user_email": "u@test.com",
		"is_admin":   true,
		"exp":        time.Now().Add(time.Hour).Unix(),
		"jti":        "token-1",
	})

	md := metadata.Pairs(tokenpkg.AuthorizationHeader, tokenpkg.BearerScheme+" "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptor(conf)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		user := sctx.GetUserInfo()
		if user == nil {
			return nil, errors.New("user info not set")
		}
		if user.GetUserID() != 123 {
			return nil, errors.New("unexpected user id")
		}
		if user.GetUserName() != "u@test.com" {
			return nil, errors.New("unexpected user email")
		}
		if !user.GetIsAdmin() {
			return nil, errors.New("unexpected admin flag")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserInfoInterceptorInvalidToken(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	md := metadata.Pairs(tokenpkg.AuthorizationHeader, tokenpkg.BearerScheme+" bad-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptor(conf)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUserInfoInterceptorMissingToken(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	ctx := context.WithValue(context.Background(), core.SrvCtx, sctx)

	itc := UserInfoInterceptor(conf)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestUserInfoInterceptorRejectsMalformedAuthorization(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(tokenpkg.AuthorizationHeader, "Basic credentials"))
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptor(conf)
	_, err := itc(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/svc/m"}, func(context.Context, any) (any, error) {
		return "ok", nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestUserInfoInterceptorExpiredToken(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	token := makeAccessToken("secret", jwt.MapClaims{
		"user_id":    float64(123),
		"user_email": "u@test.com",
		"exp":        time.Now().Add(-time.Hour).Unix(),
		"jti":        "token-1",
	})
	md := metadata.Pairs(tokenpkg.AuthorizationHeader, tokenpkg.BearerScheme+" "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptor(conf)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUserInfoInterceptorRevokedToken(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	token := makeAccessToken("secret", jwt.MapClaims{
		"user_id":    float64(123),
		"user_email": "u@test.com",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"jti":        "token-1",
	})
	md := metadata.Pairs(tokenpkg.AuthorizationHeader, tokenpkg.BearerScheme+" "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptorWithOptions(conf, nil, WithTokenRevocationChecker(fakeRevocationChecker{revoked: true}))
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUserInfoInterceptorRevocationCheckFailure(t *testing.T) {
	conf := &cogoconfig.Config{Config: core.Config{JWT: core.JWTConfig{AccessSecret: "secret"}}}
	sctx := srvctx.NewSrvCtx(&testLogger{})
	token := makeAccessToken("secret", jwt.MapClaims{
		"user_id":    float64(123),
		"user_email": "u@test.com",
		"exp":        time.Now().Add(time.Hour).Unix(),
		"jti":        "token-1",
	})
	md := metadata.Pairs(tokenpkg.AuthorizationHeader, tokenpkg.BearerScheme+" "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, core.SrvCtx, sctx)

	itc := UserInfoInterceptorWithOptions(conf, nil, WithTokenRevocationChecker(fakeRevocationChecker{err: errors.New("redis unavailable")}))
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/m"}
	_, err := itc(ctx, nil, info, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
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

type fakeRevocationChecker struct {
	revoked bool
	err     error
}

func (f fakeRevocationChecker) IsTokenRevoked(context.Context, string) (bool, error) {
	return f.revoked, f.err
}
