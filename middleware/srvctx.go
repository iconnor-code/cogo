package middleware

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/iconnor-code/cogo/core"
	"github.com/iconnor-code/cogo/core/impl"
)

func SrvCtxMiddleware(logger core.ILogger, config core.IConfig) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			srvCtx := impl.NewSrvCtx(ctx, logger, config)
			return next(srvCtx, request)
		}
	}
}
