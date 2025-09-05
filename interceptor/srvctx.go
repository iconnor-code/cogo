package interceptor

import (
	"context"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"google.golang.org/grpc"
)

func SrvCtxInterceptor(config core.IConfig, logger core.ILogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx := srvctx.NewSrvCtx(logger, config)
		ctx = context.WithValue(ctx, core.SrvCtx, srvCtx)
		return handler(ctx, req)
	}
}
