//go:build !windows

// NetFiles config package — desktop directory detection for non-Windows
// Created by Mohammed Salah
package config

import (
	"os"
	"path/filepath"
)

// GetDesktopDir resolves the Desktop directory on Linux/macOS.
func GetDesktopDir() string {
	home, err := os.UserHomeDir()
	if err == nil {
		desktop := filepath.Join(home, "Desktop")
		if fi, err := os.Stat(desktop); err == nil && fi.IsDir() {
			return desktop
		}
		return home
	}
	return "/tmp"
}
