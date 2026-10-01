// NetFiles autostart — Linux stub
// Created by Mohammed Salah
//go:build !windows

package config

// SetAutostart is a no-op on non-Windows platforms
// TODO: implement via systemd user service or ~/.config/autostart
func SetAutostart(enable bool) error {
	return nil
}
