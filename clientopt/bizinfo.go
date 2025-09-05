// Package clientopt GRPC客户端拦截器.
package clientopt

import (
	"context"
	"strconv"

	"github.com/iconnor-code/cogo/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func BizInfoOption(bizInfo core.IBizInfo) grpc.DialOption {
	return grpc.WithUnaryInterceptor(
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx,
				"biz_id", strconv.FormatInt(int64(bizInfo.GetBizID()), 10),
				"biz_name", bizInfo.GetBizName(),
			)
			return invoker(ctx, method, req, reply, cc, opts...)
		},
	)
}
