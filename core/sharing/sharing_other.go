//go:build !windows

// NetFiles sharing package — non-Windows stub
// Created by Mohammed Salah
package sharing

import (
	"fmt"
	"os"
)

const ShareName = "NetFiles"

type ShareInfo struct {
	ShareName    string
	LocalPath    string
	UNCPath      string
	IPAddress    string
	ComputerName string
	IsShared     bool
	DiscoveryOK  bool
}

func IsElevated() bool {
	return os.Geteuid() == 0
}

func Elevate(args []string) error {
	return fmt.Errorf("elevation not supported on this platform")
}

func SetupShare(folderPath string) (*ShareInfo, error) {
	hostname, _ := os.Hostname()
	return &ShareInfo{
		ShareName:    ShareName,
		LocalPath:    folderPath,
		ComputerName: hostname,
		IPAddress:    "127.0.0.1",
		UNCPath:      folderPath,
		IsShared:     true,
		DiscoveryOK:  true,
	}, nil
}

func RemoveShare() error {
	return nil
}

func IsShareActive() bool {
	return false
}

func GetShareInfo(folderPath string) (*ShareInfo, error) {
	hostname, _ := os.Hostname()
	return &ShareInfo{
		ShareName:    ShareName,
		LocalPath:    folderPath,
		ComputerName: hostname,
		IPAddress:    "127.0.0.1",
		UNCPath:      folderPath,
		IsShared:     false,
		DiscoveryOK:  false,
	}, nil
}
