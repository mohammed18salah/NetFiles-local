// NetFiles autostart — Windows implementation
// Created by Mohammed Salah
//go:build windows

package config

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// SetAutostart adds or removes the autostart registry entry
func SetAutostart(enable bool) error {
	keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("لم يتم فتح سجل البدء التلقائي: %w", err)
	}
	defer key.Close()

	valueName := AppName

	if enable {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("لم يتم العثور على مسار البرنامج: %w", err)
		}
		return key.SetStringValue(valueName, fmt.Sprintf(`"%s" start -y`, exe))
	}

	// Disable: delete the value (ignore if not found)
	_ = key.DeleteValue(valueName)
	return nil
}
