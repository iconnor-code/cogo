// Package interceptor
package interceptor

import (
	"context"
	"fmt"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
		srvCtx := ctx.Value(core.SrvCtx).(core.ISrvCtx)
		logger := srvCtx.Logger()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic error",
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", res),
					zap.Any("error", r),
					zap.StackSkip("stack", 1),
				)
				err = cerrs.WrapWithCode(fmt.Errorf("%v", r), cerrs.UnknownErrCode, "internal error occurred")
				res = nil
			}
		}()
		return handler(ctx, req)
	}
}
