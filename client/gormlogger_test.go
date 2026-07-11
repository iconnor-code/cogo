package client

import (
	"context"
	"testing"
)

func TestGormZapLoggerParamsFilter(t *testing.T) {
	t.Parallel()

	const query = "SELECT * FROM user_token_pairs WHERE access_token = ? AND refresh_token = ?"
	logger := NewGormZapLogger(nil)

	gotQuery, gotParams := logger.ParamsFilter(context.Background(), query, "access-secret", "refresh-secret")

	if gotQuery != query {
		t.Fatalf("ParamsFilter() query = %q, want %q", gotQuery, query)
	}
	if gotParams != nil {
		t.Fatalf("ParamsFilter() params = %#v, want nil", gotParams)
	}
}
