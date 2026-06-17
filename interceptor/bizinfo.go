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

func BizInfoInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx, ok := core.SrvCtxFromContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "srvctx is required")
		}
		config := srvCtx.Config()

		bizInfo := &srvctx.BizInfo{}
		bizID, err := core.GetInt(config, "biz_id")
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		bizName, err := core.GetString(config, "biz_name")
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		bizInfo.BizID = int32(bizID)
		bizInfo.BizName = bizName

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
