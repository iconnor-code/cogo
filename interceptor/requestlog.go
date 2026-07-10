package interceptor

import (
	"context"
	"time"

	"github.com/iconnor-code/cogo/core"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RequestLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}
		logger := srvCtx.Logger()
		start := time.Now()

		resp, err := handler(ctx, req)

		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return resp, err
		}

		duration := time.Since(start)
		if err == nil {
			logger.Info("request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("took", duration),
			)
			return resp, nil
		}

		if st, ok := status.FromError(err); ok {
			fields := []any{
				zap.String("method", info.FullMethod),
				zap.Duration("took", duration),
				zap.String("code", st.Code().String()),
			}
			switch st.Code() {
			case codes.Canceled:
				logger.Info("request canceled", fields...)
			case codes.InvalidArgument, codes.Unauthenticated, codes.PermissionDenied,
				codes.NotFound, codes.AlreadyExists, codes.FailedPrecondition:
				logger.Warn("request failed", fields...)
			default:
				logger.Error("request failed", fields...)
			}
			return nil, err
		}

		logger.Error("request failed",
			zap.String("method", info.FullMethod),
			zap.Duration("took", duration),
			zap.Error(err),
		)
		return nil, err
	}
}
