package clientopt

import (
	"context"
	"testing"

	"github.com/iconnor-code/cogo/core/impl/srvctx"
	"google.golang.org/grpc/metadata"
)

func TestContextWithBizInfoAddsOutgoingMetadata(t *testing.T) {
	ctx := ContextWithBizInfo(context.Background(), &srvctx.BizInfo{BizID: 102100, BizName: "blog"})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := md.Get("biz_id"); len(got) != 1 || got[0] != "102100" {
		t.Fatalf("biz_id metadata = %v", got)
	}
	if got := md.Get("biz_name"); len(got) != 1 || got[0] != "blog" {
		t.Fatalf("biz_name metadata = %v", got)
	}
}

func TestContextWithBizInfoAllowsNil(t *testing.T) {
	ctx := context.Background()
	if got := ContextWithBizInfo(ctx, nil); got != ctx {
		t.Fatal("expected nil business info to leave context unchanged")
	}
}
