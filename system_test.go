package core

import (
	"strings"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()
	if dir == "" {
		t.Fatal("GetConfigDir returned empty string")
	}
	// Should not end with config.json
	if strings.HasSuffix(dir, "config.json") {
		t.Errorf("GetConfigDir returned file path, want directory: %s", dir)
	}
}

func TestOpenConfigFolder(t *testing.T) {
	// OpenConfigFolder launches xdg-open/open/explorer; in CI those binaries
	// may not exist. We verify it either succeeds or returns a wrapped error —
	// it must never panic.
	err := OpenConfigFolder()
	if err != nil {
		if !strings.Contains(err.Error(), "failed to open config folder") {
			t.Errorf("unexpected error format: %v", err)
		}
	}
}
