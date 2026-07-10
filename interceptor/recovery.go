// Package interceptor
package interceptor

import (
	"context"
	"fmt"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}
		logger := srvCtx.Logger()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic error",
					zap.String("method", info.FullMethod),
					zap.Any("error", r),
					zap.StackSkip("stack", 1),
				)
				err = cerrs.Wrap(fmt.Errorf("%v", r), "panic recovered")
				res = nil
			}
		}()
		return handler(ctx, req)
	}
}
