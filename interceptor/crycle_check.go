package interceptor

import (
	"context"
	"slices"

	"github.com/iconnor-code/cogo/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const CallerMethodsKey = "caller_methods"

func CycleCheckInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}
		logger := srvCtx.Logger()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata is required")
		}

		callerMethods := md.Get(CallerMethodsKey)
		if slices.Contains(callerMethods, info.FullMethod) {
			logger.Error("cycle call detected", zap.Strings("caller methods", callerMethods), zap.String("current method", info.FullMethod))
			return nil, status.Errorf(codes.Aborted, "cycle call detected!")
		}

		ctx = metadata.AppendToOutgoingContext(ctx, CallerMethodsKey, info.FullMethod)

		return handler(ctx, req)
	}
}
