package cerrs

import (
	"errors"
	"testing"
)

func TestNewWithCode(t *testing.T) {
	err := NewWithCode(UnknownErrCode, "boom")
	cerr, ok := err.(*CError)
	if !ok {
		t.Fatalf("expected *CError, got %T", err)
	}
	if cerr.GetCode() != UnknownErrCode {
		t.Fatalf("expected code %d, got %d", UnknownErrCode, cerr.GetCode())
	}
}

func TestKindConstructorsSeparatePublicMessageAndCause(t *testing.T) {
	err := WrapKind(errors.New("database password"), KindNotFound, "user not found")
	var cerr *CError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected CError, got %T", err)
	}
	if cerr.Kind() != KindNotFound || cerr.PublicMessage() != "user not found" {
		t.Fatalf("unexpected public contract: kind=%v message=%q", cerr.Kind(), cerr.PublicMessage())
	}
	if !errors.Is(err, cerr.Unwrap()) {
		t.Fatal("expected internal cause to remain available")
	}
}

func TestWrapUnwrapIsAs(t *testing.T) {
	base := errors.New("base")
	wrapped := WrapWithCode(base, UnknownErrCode, "wrapped")

	if got := Unwrap(wrapped); got != base {
		t.Fatalf("expected unwrap base error, got %v", got)
	}
	if !Is(wrapped, base) {
		t.Fatalf("expected Is(wrapped, base) == true")
	}
	if !errors.Is(wrapped, base) {
		t.Fatalf("expected standard errors.Is(wrapped, base) == true")
	}

	var cerr *CError
	if !As(wrapped, &cerr) {
		t.Fatalf("expected As to extract *CError")
	}
	if !errors.As(wrapped, &cerr) {
		t.Fatalf("expected standard errors.As to extract *CError")
	}
	if cerr.GetCode() != UnknownErrCode {
		t.Fatalf("expected code %d, got %d", UnknownErrCode, cerr.GetCode())
	}
}
