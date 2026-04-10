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

func TestQualityFallbackOrder_Default(t *testing.T) {
	cfg := GetDefaultConfig()
	want := []string{"highest", "24bit", "16bit"}
	if len(cfg.QualityFallbackOrder) != len(want) {
		t.Fatalf("len = %d", len(cfg.QualityFallbackOrder))
	}
	for i := range want {
		if cfg.QualityFallbackOrder[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, cfg.QualityFallbackOrder[i], want[i])
		}
	}
}

func TestResolveFallbackOrder_UsesConfig(t *testing.T) {
	got := ResolveFallbackOrder([]string{"16bit", "24bit"}, "highest")
	want := []string{"16bit", "24bit"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolveFallbackOrder_FallsBackToDefault(t *testing.T) {
	got := ResolveFallbackOrder(nil, "24bit")
	if len(got) == 0 || got[0] != "24bit" {
		t.Errorf("expected preferred first, got %v", got)
	}
}
