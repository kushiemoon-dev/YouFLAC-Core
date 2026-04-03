package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveConfig_RespectsConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONFIG_DIR", dir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.Theme = "dark"

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		t.Errorf("config.json not found in CONFIG_DIR: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after save: %v", err)
	}
	if loaded.Theme != "dark" {
		t.Errorf("got theme %q, want %q", loaded.Theme, "dark")
	}
}
