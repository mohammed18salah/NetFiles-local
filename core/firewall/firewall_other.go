// NetFiles firewall — Linux stub
// Created by Mohammed Salah
//go:build !windows

package firewall

// EnsureRules is a no-op on non-Windows platforms
// TODO: implement via iptables/ufw
func EnsureRules(httpPort, udpPort int) error {
	return nil
}

// RemoveRules is a no-op on non-Windows platforms
func RemoveRules() error {
	return nil
}

// CheckNetworkProfile always returns true on non-Windows
func CheckNetworkProfile() (isPrivate bool, profileName string, err error) {
	return true, "Linux", nil
}

// SwitchToPrivate is a no-op on non-Windows
func SwitchToPrivate() error {
	return nil
}
