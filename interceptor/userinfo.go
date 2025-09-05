package interceptor

import (
	"context"
	"slices"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"github.com/iconnor-code/cogo/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UserInfoInterceptor(whiteList ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx := ctx.Value(core.SrvCtx).(core.ISrvCtx)

		// skip white list
		if slices.Contains(whiteList, info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata is required")
		}

		accessTokens := md.Get("access_token")
		if len(accessTokens) == 0 || len(accessTokens[0]) == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "access_token is required")
		}

		jwtToken := token.NewJwtToken(srvCtx.Config())
		userInfo, err := jwtToken.ParseToken(accessTokens[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "access_token is invalid")
		}

		srvCtx.SetUserInfo(&srvctx.UserInfo{
			UserID:    userInfo["user_id"].(uint32),
			UserEmail: userInfo["user_email"].(string),
		})

		return handler(ctx, req)
	}
}
