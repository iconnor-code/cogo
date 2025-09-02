package core

type SrvCtxKey string

const SrvCtx SrvCtxKey = "srvctx"

type IBizInfo interface {
	GetBizID() uint32
	GetBizName() string
}

type IUserInfo interface {
	GetUserID() uint32
	GetUserName() string
}

type ISrvCtx interface {
	Logger() ILogger
	Config() IConfig
	SetField(key SrvCtxKey, value any)
	GetField(key SrvCtxKey) (any, bool)
	SetBizInfo(bizInfo IBizInfo)
	GetBizInfo() IBizInfo
	SetUserInfo(userInfo IUserInfo)
	GetUserInfo() IUserInfo
}
