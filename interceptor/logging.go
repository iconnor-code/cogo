package interceptor

import (
	"context"
	"time"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/pkg/cerr"

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

		if customErr, ok := err.(*cerr.CustomError); ok {
			code := customErr.Code / 1000
			if code == 4 {
				logger.Debug("client custom error",
					zap.Any("code", customErr.Code),
					zap.Any("msg", customErr.Msg),
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", resp),
					zap.Error(customErr.Err),
				)
			} else if code == 6 {
				logger.Error("external custom error",
					zap.Any("code", customErr.Code),
					zap.Any("msg", customErr.Msg),
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", resp),
					zap.Error(customErr.Err),
				)
			} else {
				logger.Error("internal custom error",
					zap.Any("code", customErr.Code),
					zap.Any("msg", customErr.Msg),
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", resp),
					zap.Error(customErr.Err),
				)
			}
			return nil, customErr
		}

		logger.Error("internal error",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
			zap.Any("response", resp),
			zap.Error(err),
		)
		return nil, cerr.ErrInternalError
	}
}
