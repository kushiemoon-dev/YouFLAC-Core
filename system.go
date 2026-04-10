package core

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the directory that contains the config file.
func GetConfigDir() string {
	return filepath.Dir(GetConfigPathWithEnv())
}

// OpenConfigFolder opens the config directory in the system file manager.
func OpenConfigFolder() error {
	dir := GetConfigDir()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "windows":
		cmd = exec.Command("explorer", dir)
	default: // linux and others
		cmd = exec.Command("xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open config folder: %w", err)
	}
	return nil
}
