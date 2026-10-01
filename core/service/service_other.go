//go:build !windows

// NetFiles service package — non-Windows stub
// Created by Mohammed Salah
package service

import "fmt"

func Install(exePath string) error {
	return fmt.Errorf("service installation only supported on Windows")
}

func Uninstall() error {
	return nil
}

func Start() error {
	return nil
}

func Stop() error {
	return nil
}

func QueryStatus() (string, error) {
	return "Not Installed", nil
}

func IsRunning() bool {
	return false
}

func IsInstalled() bool {
	return false
}

func Run() error {
	return fmt.Errorf("service runner only supported on Windows")
}
