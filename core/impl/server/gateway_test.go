package server

import "testing"

func TestIncomingHeaderMatcher(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{header: "x-biz-id", want: "biz_id"},
		{header: "X-Biz-ID", want: "biz_id"},
		{header: "x-biz-name", want: "biz_name"},
		{header: "access_token", want: "access_token"},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got, ok := incomingHeaderMatcher(tt.header)
			if !ok || got != tt.want {
				t.Fatalf("incomingHeaderMatcher(%q) = (%q, %v), want (%q, true)", tt.header, got, ok, tt.want)
			}
		})
	}
}
