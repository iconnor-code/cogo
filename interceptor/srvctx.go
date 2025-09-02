package interceptor

import (
	"context"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl"
	"google.golang.org/grpc"
)

func SrvCtxInterceptor(logger core.ILogger, config core.IConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx := impl.NewSrvCtx(ctx, logger, config)
		ctx = context.WithValue(ctx, core.SrvCtx, srvCtx)
		return handler(ctx, req)
	}
}
