package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCore(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCore(tmpDir)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	defer c.Shutdown()

	if c.config == nil {
		t.Fatal("config should not be nil")
	}
	if c.queue == nil {
		t.Fatal("queue should not be nil")
	}
	if c.history == nil {
		t.Fatal("history should not be nil")
	}
}

func TestNewCoreSetsDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewCore(tmpDir)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	got := GetDataDir()
	if got != tmpDir {
		t.Errorf("GetDataDir() = %q, want %q", got, tmpDir)
	}
}

func TestShutdownPersistsQueue(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCore(tmpDir)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	c.Shutdown()
	queuePath := filepath.Join(tmpDir, "queue.json")
	if _, err := os.Stat(queuePath); os.IsNotExist(err) {
		t.Error("queue.json should be created on shutdown")
	}
}

func TestSetEventCallback(t *testing.T) {
	tmpDir := t.TempDir()
	c, err := NewCore(tmpDir)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}
	defer c.Shutdown()

	called := false
	c.SetEventCallback(func(event Event) {
		called = true
	})
	c.emitEvent("test", nil)
	if !called {
		t.Error("event callback should have been called")
	}
}
