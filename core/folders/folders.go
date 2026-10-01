// NetFiles folders package — folder layout management
// Created by Mohammed Salah
package folders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Standard subfolder names
const (
	PublicDir  = "Public"
	InboxDir   = "Inbox"
	SentDir    = "_Sent"
	OnlineFile = "_Online.txt"
)

// CreateLayout creates the root folder structure
func CreateLayout(rootFolder string, myName string) error {
	dirs := []string{
		rootFolder,
		filepath.Join(rootFolder, PublicDir),
		filepath.Join(rootFolder, InboxDir),
		filepath.Join(rootFolder, SentDir),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create folder %s: %w", dir, err)
		}
	}

	// Create _Online.txt
	onlinePath := filepath.Join(rootFolder, OnlineFile)
	err := os.WriteFile(onlinePath, []byte("0 peers online\n"), 0644)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", onlinePath, err)
	}

	_ = SetPermissions(rootFolder)

	return nil
}

// SetPermissions grants Modify access to Users and Everyone so any logged-in user can drop files
func SetPermissions(rootFolder string) error {
	cmd := exec.Command("icacls", rootFolder, "/grant", "Users:(OI)(CI)M", "/grant", "Everyone:(OI)(CI)M", "/T", "/C", "/Q")
	return cmd.Run()
}

// EnsurePeerFolder creates a folder for a peer if it doesn't exist
func EnsurePeerFolder(rootFolder string, peerName string) error {
	dir := filepath.Join(rootFolder, peerName)
	return os.MkdirAll(dir, 0755)
}

// RemovePeerFolder removes a peer's folder if empty
func RemovePeerFolder(rootFolder string, peerName string) error {
	dir := filepath.Join(rootFolder, peerName)

	// Only remove if empty
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // doesn't exist, that's fine
	}
	if len(entries) > 0 {
		return nil // not empty, keep it
	}

	return os.Remove(dir)
}

// RenamePeerFolder renames a peer's folder
func RenamePeerFolder(rootFolder string, oldName string, newName string) error {
	oldDir := filepath.Join(rootFolder, oldName)
	newDir := filepath.Join(rootFolder, newName)

	// Check if old folder exists
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return nil // nothing to rename
	}

	// Check if new name already exists
	if _, err := os.Stat(newDir); err == nil {
		return fmt.Errorf("folder %s already exists", newName)
	}

	return os.Rename(oldDir, newDir)
}

// MarkOffline renames a peer folder to "<name> (offline)"
func MarkOffline(rootFolder string, peerName string) error {
	offlineName := peerName + " (offline)"
	return RenamePeerFolder(rootFolder, peerName, offlineName)
}

// MarkOnline renames a peer folder back from "<name> (offline)" to "<name>"
func MarkOnline(rootFolder string, peerName string) error {
	offlineName := peerName + " (offline)"
	return RenamePeerFolder(rootFolder, offlineName, peerName)
}

// UpdateOnlineFile writes the current online peers list
func UpdateOnlineFile(rootFolder string, peers []PeerInfo) error {
	onlinePath := filepath.Join(rootFolder, OnlineFile)

	var content string
	content = fmt.Sprintf("%d peers online\n", len(peers))
	content += "---\n"
	for _, p := range peers {
		content += fmt.Sprintf("%-20s %s:%d\n", p.Name, p.IP, p.Port)
	}

	return os.WriteFile(onlinePath, []byte(content), 0644)
}

// PeerInfo holds basic peer information for display
type PeerInfo struct {
	Name string
	IP   string
	Port int
}

// IsReservedName checks if a folder name conflicts with system folders
func IsReservedName(name string) bool {
	reserved := map[string]bool{
		"public":  true,
		"inbox":   true,
		"_sent":   true,
		"_online": true,
	}
	return reserved[filepath.Clean(name)]
}

// GetPeerFolders lists all peer folders (excludes system folders)
func GetPeerFolders(rootFolder string) ([]string, error) {
	entries, err := os.ReadDir(rootFolder)
	if err != nil {
		return nil, err
	}

	systemDirs := map[string]bool{
		PublicDir: true,
		InboxDir:  true,
		SentDir:   true,
	}

	var peers []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if systemDirs[e.Name()] {
			continue
		}
		peers = append(peers, e.Name())
	}

	return peers, nil
}
