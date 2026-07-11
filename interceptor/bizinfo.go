// Package interceptor GRPC服务端拦截器.
package interceptor

import (
	"context"
	"strconv"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BizInfoConfig interface {
	GetBizID() int
	GetBizName() string
}

func BizInfoInterceptor(config BizInfoConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}
		bizInfo := &srvctx.BizInfo{}
		bizInfo.BizID = int32(config.GetBizID())
		bizInfo.BizName = config.GetBizName()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata is required")
		}

		for _, bizID := range md.Get("biz_id") {
			originalBizID, err := strconv.ParseInt(bizID, 10, 32)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "biz_id is invalid")
			}
			bizInfo.OriginalBizID = append(bizInfo.OriginalBizID, int32(originalBizID))
		}

		bizInfo.OriginalBizName = md.Get("biz_name")

		srvCtx.SetBizInfo(bizInfo)
		return handler(ctx, req)
	}
}
