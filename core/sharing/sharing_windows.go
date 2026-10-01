// NetFiles sharing package — Windows SMB and Network Discovery automation
// Created by Mohammed Salah
package sharing

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const ShareName = "NetFiles"

// ShareInfo holds information about the network share
type ShareInfo struct {
	ShareName   string
	LocalPath   string
	UNCPath     string
	IPAddress   string
	ComputerName string
	IsShared    bool
	DiscoveryOK bool
}

// IsElevated returns true if running with Administrator privileges
func IsElevated() bool {
	token := windows.GetCurrentProcessToken()
	return token.IsElevated()
}

// Elevate restarts the current process with Administrator privileges
func Elevate(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cwd, _ := os.Getwd()

	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(exe)
	argPtr, _ := windows.UTF16PtrFromString(strings.Join(args, " "))
	dirPtr, _ := windows.UTF16PtrFromString(cwd)

	return windows.ShellExecute(0, verbPtr, exePtr, argPtr, dirPtr, windows.SW_NORMAL)
}

// SetupShare configures the folder, Windows SMB sharing, firewall, and Network Discovery
func SetupShare(folderPath string) (*ShareInfo, error) {
	// 1. Ensure directories exist
	dirs := []string{
		folderPath,
		filepath.Join(folderPath, "Public"),
		filepath.Join(folderPath, "Inbox"),
		filepath.Join(folderPath, "_Sent"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// 2. Grant full NTFS permissions to Everyone & Users
	_ = runHidden("icacls", folderPath, "/grant", "Everyone:(OI)(CI)F", "/T", "/C", "/Q")
	_ = runHidden("icacls", folderPath, "/grant", "Users:(OI)(CI)F", "/T", "/C", "/Q")

	// 3. Remove existing share if already present to avoid conflicts
	if IsShareActive() {
		_ = runHidden("net", "share", ShareName, "/DELETE", "/Y")
	}

	// 4. Create SMB Share with full access for Everyone
	err := runHidden("net", "share", fmt.Sprintf("%s=%s", ShareName, folderPath), "/GRANT:Everyone,FULL", "/REMARK:NetFiles LAN Offline Share")
	if err != nil {
		return nil, fmt.Errorf("failed to create Windows share: %w", err)
	}

	// 5. Configure Windows Network Profile to Private (so discovery & sharing are active)
	psScript := "Get-NetConnectionProfile | Set-NetConnectionProfile -NetworkCategory Private -ErrorAction SilentlyContinue"
	_ = runHidden("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)

	// 6. Enable Windows Firewall rules for File Sharing and Network Discovery
	_ = runHidden("netsh", "advfirewall", "firewall", "set", "rule", "group=File and Printer Sharing", "new", "enable=Yes")
	_ = runHidden("netsh", "advfirewall", "firewall", "set", "rule", "group=Network Discovery", "new", "enable=Yes")

	// 7. Enable and start Windows Network Discovery & SMB services
	// LanmanServer (Server service)
	_ = runHidden("sc", "config", "LanmanServer", "start=", "auto")
	_ = runHidden("net", "start", "LanmanServer")

	// FDResPub (Function Discovery Resource Publication - makes this PC visible in Explorer Network)
	_ = runHidden("sc", "config", "FDResPub", "start=", "auto")
	_ = runHidden("net", "start", "FDResPub")

	// FDHost (Function Discovery Provider Host)
	_ = runHidden("sc", "config", "FDHost", "start=", "auto")
	_ = runHidden("net", "start", "FDHost")

	// 8. Configure registry for guest/password-less network access
	configureGuestSharing()

	return GetShareInfo(folderPath)
}

// RemoveShare removes the SMB share from Windows Network
func RemoveShare() error {
	if !IsShareActive() {
		return nil
	}
	return runHidden("net", "share", ShareName, "/DELETE", "/Y")
}

// IsShareActive checks if the NetFiles share currently exists
func IsShareActive() bool {
	cmd := exec.Command("net", "share", ShareName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// GetShareInfo returns the current network share status
func GetShareInfo(folderPath string) (*ShareInfo, error) {
	hostname, _ := os.Hostname()
	ip := getLANIP()

	info := &ShareInfo{
		ShareName:    ShareName,
		LocalPath:    folderPath,
		ComputerName: hostname,
		IPAddress:    ip,
		UNCPath:      fmt.Sprintf(`\\%s\%s`, hostname, ShareName),
		IsShared:     IsShareActive(),
		DiscoveryOK:  true,
	}

	return info, nil
}

func configureGuestSharing() {
	// Enable NullSessionShares so machines on the LAN can access without domain accounts
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`, registry.SET_VALUE|registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		existing, _, err := k.GetStringsValue("NullSessionShares")
		hasNetFiles := false
		if err == nil {
			for _, s := range existing {
				if strings.EqualFold(s, ShareName) {
					hasNetFiles = true
					break
				}
			}
		}
		if !hasNetFiles {
			existing = append(existing, ShareName)
			_ = k.SetStringsValue("NullSessionShares", existing)
		}
		_ = k.SetDWordValue("AutoShareWks", 1)
	}

	// Allow blank passwords / everyone includes anonymous
	k2, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Lsa`, registry.SET_VALUE)
	if err == nil {
		defer k2.Close()
		_ = k2.SetDWordValue("everyoneincludesanonymous", 1)
		_ = k2.SetDWordValue("LimitBlankPasswordUse", 0)
	}
}

func runHidden(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func getLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	var fallback string
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				// Prefer standard private LAN IP (192.168.x.x or 10.x.x.x)
				if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
					return ipStr
				}
				fallback = ipStr
			}
		}
	}
	if fallback != "" {
		return fallback
	}
	return "127.0.0.1"
}
