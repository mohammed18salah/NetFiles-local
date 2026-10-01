// NetFiles firewall package — Windows firewall rule management
// Created by Mohammed Salah
//go:build windows

package firewall

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

const (
	ruleTCP = "NetFilesTool-TCP"
	ruleUDP = "NetFilesTool-UDP"
)

// EnsureRules adds firewall rules for the given ports.
// Uses netsh advfirewall. Will self-elevate via UAC if needed.
func EnsureRules(httpPort, udpPort int) error {
	// Check if rules already exist
	if ruleExists(ruleTCP) && ruleExists(ruleUDP) {
		return nil
	}

	// Try to add rules (may fail without admin)
	err := addTCPRule(httpPort)
	if err != nil {
		// Try self-elevation
		elevErr := elevateAndAddRules(httpPort, udpPort)
		if elevErr != nil {
			return fmt.Errorf("failed to add firewall rules: %w", err)
		}
		return nil
	}

	err = addUDPRule(udpPort)
	if err != nil {
		return fmt.Errorf("failed to add UDP rule: %w", err)
	}

	return nil
}

// RemoveRules removes all NetFilesTool firewall rules
func RemoveRules() error {
	var lastErr error

	if ruleExists(ruleTCP) {
		err := runNetsh("advfirewall", "firewall", "delete", "rule", fmt.Sprintf("name=%s", ruleTCP))
		if err != nil {
			lastErr = err
		}
	}

	if ruleExists(ruleUDP) {
		err := runNetsh("advfirewall", "firewall", "delete", "rule", fmt.Sprintf("name=%s", ruleUDP))
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func addTCPRule(port int) error {
	return runNetsh("advfirewall", "firewall", "add", "rule",
		fmt.Sprintf("name=%s", ruleTCP),
		"dir=in",
		"action=allow",
		"protocol=TCP",
		fmt.Sprintf("localport=%d", port),
		"profile=private",
		"enable=yes",
	)
}

func addUDPRule(port int) error {
	return runNetsh("advfirewall", "firewall", "add", "rule",
		fmt.Sprintf("name=%s", ruleUDP),
		"dir=in",
		"action=allow",
		"protocol=UDP",
		fmt.Sprintf("localport=%d", port),
		"profile=private",
		"enable=yes",
	)
}

func ruleExists(name string) bool {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", fmt.Sprintf("name=%s", name))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return !strings.Contains(string(output), "No rules match")
}

func runNetsh(args ...string) error {
	cmd := exec.Command("netsh", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh error: %s: %w", string(output), err)
	}
	return nil
}

func elevateAndAddRules(httpPort, udpPort int) error {
	script := fmt.Sprintf(
		`netsh advfirewall firewall add rule name=%s dir=in action=allow protocol=TCP localport=%d profile=private enable=yes & `+
			`netsh advfirewall firewall add rule name=%s dir=in action=allow protocol=UDP localport=%d profile=private enable=yes`,
		ruleTCP, httpPort, ruleUDP, udpPort,
	)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Start-Process -FilePath "cmd.exe" -ArgumentList '/c %s' -Verb RunAs -Wait -WindowStyle Hidden`, script),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// CheckNetworkProfile checks if the current network profile is Private
func CheckNetworkProfile() (isPrivate bool, profileName string, err error) {
	cmd := exec.Command("powershell", "-Command",
		"(Get-NetConnectionProfile | Select-Object -First 1).NetworkCategory")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", fmt.Errorf("failed to check network profile: %w", err)
	}

	profile := strings.TrimSpace(string(output))
	return strings.EqualFold(profile, "Private"), profile, nil
}

// SwitchToPrivate attempts to set the network profile to Private
func SwitchToPrivate() error {
	cmd := exec.Command("powershell", "-Command",
		`Start-Process -FilePath "powershell" -ArgumentList '-Command Set-NetConnectionProfile -NetworkCategory Private' -Verb RunAs -Wait -WindowStyle Hidden`)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
