//go:build !windows

// NetFiles explorer package — non-Windows stub
// Created by Mohammed Salah
package explorer

func NotifyExplorer() {}

func Install(folderPath, exePath string) error {
	return nil
}

func Uninstall(folderPath string) error {
	return nil
}
