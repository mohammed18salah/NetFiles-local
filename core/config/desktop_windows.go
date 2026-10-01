// NetFiles config package — desktop directory detection for Windows
// Created by Mohammed Salah
package config

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// GetDesktopDir resolves the active Desktop directory on Windows.
// It properly resolves OneDrive redirected Desktops, Shell Folders, and standard paths.
func GetDesktopDir() string {
	// 1. Try Windows Known Folder API (handles OneDrive / redirected folders automatically)
	if p, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, 0); err == nil && p != "" {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}

	// 2. Try User Shell Folders registry key (expand %USERPROFILE% etc.)
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		val, _, err := k.GetStringValue("Desktop")
		if err == nil && val != "" {
			expanded := os.ExpandEnv(val)
			if fi, err := os.Stat(expanded); err == nil && fi.IsDir() {
				return expanded
			}
		}
	}

	// 3. Try Shell Folders registry key
	k2, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders`, registry.QUERY_VALUE)
	if err == nil {
		defer k2.Close()
		val, _, err := k2.GetStringValue("Desktop")
		if err == nil && val != "" {
			expanded := os.ExpandEnv(val)
			if fi, err := os.Stat(expanded); err == nil && fi.IsDir() {
				return expanded
			}
		}
	}

	// 4. Fallback to USERPROFILE\Desktop
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		d := filepath.Join(userProfile, "Desktop")
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			return d
		}
	}

	// 5. Fallback to UserHomeDir
	if home, err := os.UserHomeDir(); err == nil {
		d := filepath.Join(home, "Desktop")
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			return d
		}
		return home
	}

	return `C:\`
}
