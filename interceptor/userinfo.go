package interceptor

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"github.com/iconnor-code/cogo/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenRevocationChecker interface {
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
}

type UserInfoConfig interface {
	GetJWT() core.JWTConfig
}

type UserInfoOption func(*userInfoOptions)

type userInfoOptions struct {
	revocationChecker TokenRevocationChecker
}

func WithTokenRevocationChecker(checker TokenRevocationChecker) UserInfoOption {
	return func(opts *userInfoOptions) {
		opts.revocationChecker = checker
	}
}

func UserInfoInterceptor(config UserInfoConfig, whiteList ...string) grpc.UnaryServerInterceptor {
	return UserInfoInterceptorWithOptions(config, whiteList)
}

func UserInfoInterceptorWithOptions(config UserInfoConfig, whiteList []string, options ...UserInfoOption) grpc.UnaryServerInterceptor {
	opts := userInfoOptions{}
	for _, option := range options {
		option(&opts)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}

		// skip white list
		if slices.Contains(whiteList, info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "access_token is required")
		}

		accessTokens := md.Get("access_token")
		if len(accessTokens) == 0 || len(accessTokens[0]) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "access_token is required")
		}

		jwtToken := token.NewJwtToken(config)
		userInfo, err := jwtToken.ParseToken(accessTokens[0])
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access_token is invalid or expired")
		}
		tokenID, err := token.ClaimsString(userInfo, token.ClaimID)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access_token id is invalid")
		}
		if opts.revocationChecker != nil {
			revoked, err := opts.revocationChecker.IsTokenRevoked(ctx, tokenID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "check access_token revocation failed")
			}
			if revoked {
				return nil, status.Errorf(codes.Unauthenticated, "access_token is revoked")
			}
		}

		userID, err := toUint32(userInfo["user_id"])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "access_token user_id is invalid")
		}
		userEmail, ok := userInfo["user_email"].(string)
		if !ok || userEmail == "" {
			return nil, status.Errorf(codes.InvalidArgument, "access_token user_email is invalid")
		}
		isAdmin, _ := userInfo["is_admin"].(bool)

		srvCtx.SetUserInfo(&srvctx.UserInfo{
			UserID:    userID,
			UserEmail: userEmail,
			IsAdmin:   isAdmin,
		})

		return handler(ctx, req)
	}
}

func toUint32(v any) (uint32, error) {
	switch id := v.(type) {
	case uint32:
		return id, nil
	case uint64:
		if id > math.MaxUint32 {
			return 0, fmt.Errorf("user_id out of range: %d", id)
		}
		return uint32(id), nil
	case int:
		if id < 0 {
			return 0, fmt.Errorf("user_id is negative: %d", id)
		}
		return uint32(id), nil
	case int32:
		if id < 0 {
			return 0, fmt.Errorf("user_id is negative: %d", id)
		}
		return uint32(id), nil
	case int64:
		if id < 0 || id > math.MaxUint32 {
			return 0, fmt.Errorf("user_id out of range: %d", id)
		}
		return uint32(id), nil
	case float64:
		if id < 0 || id > math.MaxUint32 {
			return 0, fmt.Errorf("user_id out of range: %f", id)
		}
		if math.Trunc(id) != id {
			return 0, fmt.Errorf("user_id is not an integer: %f", id)
		}
		return uint32(id), nil
	default:
		return 0, fmt.Errorf("unsupported user_id type: %T", v)
	}
}
