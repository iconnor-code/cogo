package interceptor

import (
	"context"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RecoveryInterceptor(logger core.ILogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic error",
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", res),
					zap.Any("error", r),
					zap.StackSkip("stack", 1),
				)
				err = cerr.ErrInternalPanic
			}
		}()
		return handler(ctx, req)
	}
}
