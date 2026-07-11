package srvctx

import (
	"sync"
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

func TestSrvCtxSetGetField(t *testing.T) {
	s := NewSrvCtx(&testLogger{})
	s.SetField(core.SrvCtxKey("k"), "v")
	got, ok := s.GetField(core.SrvCtxKey("k"))
	if !ok {
		t.Fatalf("expected field exists")
	}
	if got != "v" {
		t.Fatalf("expected value v, got %v", got)
	}
}

func TestSrvCtxSupportsConcurrentAccess(t *testing.T) {
	s := NewSrvCtx(&testLogger{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(value int) {
			defer wg.Done()
			s.SetField(core.SrvCtxKey("shared"), value)
			s.SetBizInfo(&BizInfo{BizID: int32(value)})
		}(i)
		go func() {
			defer wg.Done()
			_, _ = s.GetField(core.SrvCtxKey("shared"))
			_ = s.GetBizInfo()
		}()
	}
	wg.Wait()
}

func TestSrvCtxBizAndUserInfo(t *testing.T) {
	s := NewSrvCtx(&testLogger{})
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
