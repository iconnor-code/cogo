package srvctx

import (
	"github.com/iconnor-code/cogo/core"
)

type BizInfo struct {
	OriginalBizID   []int32  `json:"original_biz_id"`
	OriginalBizName []string `json:"original_biz_name"`
	BizID           int32    `json:"biz_id"`
	BizName         string   `json:"biz_name"`
}

func (b *BizInfo) GetCallerBizID() int32 {
	if len(b.OriginalBizID) == 0 {
		return 0
	}
	return b.OriginalBizID[len(b.OriginalBizID)-1]
}
func (b *BizInfo) GetCallerBizName() string {
	if len(b.OriginalBizName) == 0 {
		return ""
	}
	return b.OriginalBizName[len(b.OriginalBizName)-1]
}
func (b *BizInfo) GetBizID() int32 {
	return b.BizID
}
func (b *BizInfo) GetBizName() string {
	return b.BizName
}

type UserInfo struct {
	UserID    uint32 `json:"user_id"`
	UserEmail string `json:"user_email"`
}

func (u *UserInfo) GetUserID() uint32 {
	return u.UserID
}
func (u *UserInfo) GetUserName() string {
	return u.UserEmail
}

type SrvCtx struct {
	logger   core.ILogger
	config   core.IConfig
	bizInfo  core.IBizInfo
	userInfo core.IUserInfo
	ext      map[core.SrvCtxKey]any
}

func NewSrvCtx(logger core.ILogger, config core.IConfig) *SrvCtx {
	return &SrvCtx{
		logger: logger,
		config: config,
		ext:    make(map[core.SrvCtxKey]any),
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
