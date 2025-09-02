package interceptor

import (
	"context"
	"strconv"

	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func BizInfoInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		srvCtx := ctx.Value(core.SrvCtx).(core.ISrvCtx)
		bizInfo := &impl.BizInfo{}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata is required")
		}

		var bizID uint32
		var bizName string
		for _, v := range md.Get("biz_id") {
			bizID64, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "biz_id is invalid")
			}
			bizID = uint32(bizID64)
		}
		for _, v := range md.Get("biz_name") {
			bizName = v
		}

		if bizID == 0 {
			bizIDint := srvCtx.Config().Get("biz_id").(int)
			bizID = uint32(bizIDint)
			bizName = srvCtx.Config().Get("biz_name").(string)
		}
		bizInfo.BizID = bizID
		bizInfo.BizName = bizName

		srvCtx.SetBizInfo(bizInfo)
		return handler(ctx, req)
	}
}
