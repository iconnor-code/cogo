package srvctx

import (
	"testing"

	"github.com/iconnor-code/cogo/core"
)

type testLogger struct{}

func (l *testLogger) Log(...any) error       { return nil }
func (l *testLogger) Debug(string, ...any)   {}
func (l *testLogger) Info(string, ...any)    {}
func (l *testLogger) Warn(string, ...any)    {}
func (l *testLogger) Error(string, ...any)   {}
func (l *testLogger) Fatal(string, ...any)   {}
func (l *testLogger) Panic(string, ...any)   {}
func (l *testLogger) AddGlobalFields(...any) {}

type testConfig struct{}

func (c *testConfig) Get(string) any { return nil }
func (c *testConfig) ReLoad() error  { return nil }

func TestSrvCtxSetGetField(t *testing.T) {
	s := NewSrvCtx(&testLogger{}, &testConfig{})
	s.SetField(core.SrvCtxKey("k"), "v")
	got, ok := s.GetField(core.SrvCtxKey("k"))
	if !ok {
		t.Fatalf("expected field exists")
	}
	if got != "v" {
		t.Fatalf("expected value v, got %v", got)
	}
}

func TestSrvCtxBizAndUserInfo(t *testing.T) {
	s := NewSrvCtx(&testLogger{}, &testConfig{})
	b := &BizInfo{BizID: 1, BizName: "biz"}
	u := &UserInfo{UserID: 2, UserEmail: "u@test", IsAdmin: true}

	s.SetBizInfo(b)
	s.SetUserInfo(u)

	if s.GetBizInfo().GetBizID() != 1 {
		t.Fatalf("unexpected biz id")
	}
	if s.GetUserInfo().GetUserID() != 2 {
		t.Fatalf("unexpected user id")
	}
	if !s.GetUserInfo().GetIsAdmin() {
		t.Fatalf("unexpected admin flag")
	}
}
