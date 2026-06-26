package interceptor

import (
	"context"
	"time"

	"github.com/iconnor-code/cogo/cerrs"
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
		defer func() {
			if info.FullMethod == "/grpc.health.v1.Health/Check" {
				return
			}
			logger.Info("request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("took", time.Since(start)),
			)
		}()

		resp, err := handler(ctx, req)

		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return resp, nil
		}
		if err == nil {
			logger.Debug("request detail",
				zap.Any("request", req),
				zap.Any("response", resp),
				zap.Any("context", ctx),
			)
			return resp, nil
		}

		if st, ok := status.FromError(err); ok {
			logger.Warn("request failed",
				zap.String("code", st.Code().String()),
				zap.String("message", st.Message()),
				zap.String("method", info.FullMethod),
				zap.Any("request", req),
			)
			return nil, err
		}

		if customErr, ok := err.(*cerrs.CError); ok {
			if customErr.GetCode() == cerrs.UnknownErrCode {
				logger.Error("internal custom error",
					zap.Any("code", customErr.GetCode()),
					zap.Any("msg", customErr.Error()),
					zap.String("method", info.FullMethod),
					zap.Any("request", req),
					zap.Any("response", resp),
					zap.Error(customErr),
				)
				return nil, cerrs.NewWithCode(cerrs.UnknownErrCode, "internal error occurred")
			}
			return nil, customErr
		}

		logger.Error("internal error",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
			zap.Any("response", resp),
			zap.Error(err),
		)
		return nil, cerrs.NewWithCode(cerrs.UnknownErrCode, "internal error occurred")
	}
}
