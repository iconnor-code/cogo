package interceptor

import (
	"testing"

	"github.com/iconnor-code/cogo/cerrs"
	"google.golang.org/grpc/codes"
)

func TestCustomErrorMapping(t *testing.T) {
	tests := []struct {
		name string
		code cerrs.CerrCode
		want codes.Code
	}{
		{name: "invalid argument", code: 4000, want: codes.InvalidArgument},
		{name: "permission denied", code: 4030, want: codes.PermissionDenied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grpcCodeForCustomError(tt.code); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCustomErrorMessageDoesNotExposeCaller(t *testing.T) {
	err := cerrs.NewWithCode(4000, "email and password are required").(*cerrs.CError)
	if got := customErrorMessage(err); got != "email and password are required" {
		t.Fatalf("unexpected public message %q", got)
	}
}
