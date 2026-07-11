package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSwaggerHandlerProvidesHealthEndpointWithoutSwagger(t *testing.T) {
	handler := NewSwaggerHandler(http.NotFoundHandler(), SwaggerOption{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("health response = %d %q", recorder.Code, recorder.Body.String())
	}
}
