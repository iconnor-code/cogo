package interceptor

import (
	"context"
	"errors"

	"github.com/iconnor-code/cogo/cerrs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor translates service errors into stable transport errors.
func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if _, ok := status.FromError(err); ok {
			return nil, err
		}

		var customErr *cerrs.CError
		if errors.As(err, &customErr) {
			if customErr.Kind() == cerrs.KindInternal {
				return nil, status.Error(codes.Internal, "internal error occurred")
			}
			return nil, status.Error(grpcCodeForKind(customErr.Kind()), customErr.PublicMessage())
		}
		return nil, status.Error(codes.Internal, "internal error occurred")
	}
}

func grpcCodeForKind(kind cerrs.Kind) codes.Code {
	switch kind {
	case cerrs.KindInvalidArgument:
		return codes.InvalidArgument
	case cerrs.KindUnauthenticated:
		return codes.Unauthenticated
	case cerrs.KindPermissionDenied:
		return codes.PermissionDenied
	case cerrs.KindNotFound:
		return codes.NotFound
	case cerrs.KindAlreadyExists:
		return codes.AlreadyExists
	case cerrs.KindFailedPrecondition:
		return codes.FailedPrecondition
	case cerrs.KindUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}
