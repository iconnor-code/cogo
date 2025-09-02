package impl

import (
	"context"

	"github.com/iconnor-code/cogo/core"
)

type BizInfo struct {
	BizID   uint32
	BizName string
}

func (b *BizInfo) GetBizID() uint32 {
	return b.BizID
}
func (b *BizInfo) GetBizName() string {
	return b.BizName
}

type UserInfo struct {
	UserID   uint32
	UserName string
}

func (u *UserInfo) GetUserID() uint32 {
	return u.UserID
}
func (u *UserInfo) GetUserName() string {
	return u.UserName
}

type SrvCtx struct {
	logger   core.ILogger
	config   core.IConfig
	bizInfo  core.IBizInfo
	userInfo core.IUserInfo
	ext      map[core.SrvCtxKey]any
}

func NewSrvCtx(ctx context.Context, logger core.ILogger, config core.IConfig) *SrvCtx {
	return &SrvCtx{
		logger: logger,
		config: config,
	}
}

func (s *SrvCtx) Logger() core.ILogger {
	return s.logger
}

func (s *SrvCtx) Config() core.IConfig {
	return s.config
}

func (s *SrvCtx) SetField(key core.SrvCtxKey, value any) {
	s.ext[key] = value
}

func (s *SrvCtx) GetField(key core.SrvCtxKey) (any, bool) {
	res, ok := s.ext[key]
	return res, ok
}

func (s *SrvCtx) SetBizInfo(bizInfo core.IBizInfo) {
	s.bizInfo = bizInfo
}

func (s *SrvCtx) GetBizInfo() core.IBizInfo {
	return s.bizInfo
}

func (s *SrvCtx) SetUserInfo(userInfo core.IUserInfo) {
	s.userInfo = userInfo
}

func (s *SrvCtx) GetUserInfo() core.IUserInfo {
	return s.userInfo
}
