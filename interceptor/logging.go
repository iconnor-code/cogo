package interceptor

import (
	"context"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
	"github.com/iconnor-code/cogo/core"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingInterceptor(logger core.ILogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		defer func() {
			logger.Info("request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("took", time.Since(start)),
			)
		}()

		resp, err := handler(ctx, req)
		if err == nil {
			logger.Debug("request detail",
				zap.Any("request", req),
				zap.Any("response", resp),
				zap.Any("context", ctx),
			)
			return resp, nil
		}

		if customErr, ok := err.(*cerrs.CError); ok {
			if customErr.GetCode() == cerrs.InternalErrCode {
				logger.Error("internal custom error",
					zap.Any("code", customErr.GetCode()),
					zap.Any("msg", customErr.Error()),
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", resp),
					zap.Error(customErr),
				)
				return nil, cerrs.InternalError
			}
			return nil, customErr
		}

		logger.Error("internal error",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
			zap.Any("response", resp),
			zap.Error(err),
		)
		return nil, cerrs.InternalError
	}
}
