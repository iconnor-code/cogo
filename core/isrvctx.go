package core

import (
	"context"
)

type SrvCtxKey string

const SrvCtx SrvCtxKey = "srvctx"

func SrvCtxFromContext(ctx context.Context) (ISrvCtx, bool) {
	if ctx == nil {
		return nil, false
	}
	srvCtx, ok := ctx.Value(SrvCtx).(ISrvCtx)
	return srvCtx, ok
}

type IBizInfo interface {
	GetBizID() int32
	GetBizName() string
	GetCallerBizID() int32
	GetCallerBizName() string
}

type IUserInfo interface {
	GetUserID() uint32
	GetUserName() string
	GetIsAdmin() bool
}

type ISrvCtx interface {
	Logger() ILogger
	SetField(key SrvCtxKey, value any)
	GetField(key SrvCtxKey) (any, bool)
	SetBizInfo(bizInfo IBizInfo)
	GetBizInfo() IBizInfo
	SetUserInfo(userInfo IUserInfo)
	GetUserInfo() IUserInfo
}
