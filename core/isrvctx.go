package core

import (
	"context"
	"time"
)

type SrvCtxKey string

type IBizInfo interface {
	GetBizID() uint32
	GetBizName() string
}

type IUserInfo interface {
	GetUserID() uint32
	GetUserName() string
}

type ISrvCtx interface {
	Ctx() context.Context
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any

	Logger() ILogger
	Config() IConfig
	SetField(key SrvCtxKey, value any)
	GetField(key SrvCtxKey) (any, bool)
	SetBizInfo(bizInfo IBizInfo)
	GetBizInfo() IBizInfo
	SetUserInfo(userInfo IUserInfo)
	GetUserInfo() IUserInfo
}
