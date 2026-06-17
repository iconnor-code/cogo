package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigReturnsIndependentInstances(t *testing.T) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "first.yaml")
	secondPath := filepath.Join(dir, "second.yaml")
	if err := os.WriteFile(firstPath, []byte("name: first\n"), 0o600); err != nil {
		t.Fatalf("write first config: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("name: second\n"), 0o600); err != nil {
		t.Fatalf("write second config: %v", err)
	}

	first, err := NewConfig(WithFilePath(firstPath))
	if err != nil {
		t.Fatalf("new first config: %v", err)
	}
	second, err := NewConfig(WithFilePath(secondPath))
	if err != nil {
		t.Fatalf("new second config: %v", err)
	}

	if got := first.Get("name"); got != "first" {
		t.Fatalf("first name = %v, want first", got)
	}
	if got := second.Get("name"); got != "second" {
		t.Fatalf("second name = %v, want second", got)
	}
}
