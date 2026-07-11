package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestSwaggerHandlerWithoutSpecFallsBackSafely(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	NewSwaggerHandler(api, SwaggerOption{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected API handler response, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	NewSwaggerHandler(nil, SwaggerOption{}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected not found fallback, got %d", recorder.Code)
	}
}
