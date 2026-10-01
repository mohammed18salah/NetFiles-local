// NetFiles server package — HTTP server for file transfers
// Created by Mohammed Salah
package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"netfiles/core/config"
)

// Server handles incoming file transfers
type Server struct {
	cfg  *config.Config
	mux  *http.ServeMux
	addr string
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:  cfg,
		mux:  http.NewServeMux(),
		addr: fmt.Sprintf(":%d", cfg.HTTPPort),
	}

	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/info", s.handleInfo)

	return s
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","name":"%s","device_id":"%s","version":"1.0.0"}`,
		s.cfg.DisplayName, s.cfg.DeviceID)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"name":"%s","device_id":"%s","port":%d}`,
		s.cfg.DisplayName, s.cfg.DeviceID, s.cfg.HTTPPort)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get headers
	senderName := r.Header.Get("X-From")
	scope := r.Header.Get("X-Scope") // "public" or "private"
	deviceID := r.Header.Get("X-Device")
	fileName := r.Header.Get("X-Filename")

	if fileName == "" {
		http.Error(w, "X-Filename header required", http.StatusBadRequest)
		return
	}

	if senderName == "" {
		senderName = "Unknown"
	}

	// Path traversal protection: clean the filename
	fileName = filepath.Base(fileName)
	if fileName == "." || fileName == ".." || fileName == "/" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Determine destination folder
	var destDir string
	if scope == "public" {
		destDir = filepath.Join(s.cfg.RootFolder, "Public")
	} else {
		// Private: put in Inbox/<sender>
		destDir = filepath.Join(s.cfg.RootFolder, "Inbox", senderName)
	}

	// Create dest directory
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		http.Error(w, "Cannot create directory", http.StatusInternalServerError)
		return
	}

	destPath := filepath.Join(destDir, fileName)

	// Security: verify resolved path stays inside the root
	absRoot, _ := filepath.Abs(s.cfg.RootFolder)
	absDest, _ := filepath.Abs(destPath)
	if !strings.HasPrefix(absDest, absRoot) {
		http.Error(w, "Path traversal blocked", http.StatusForbidden)
		return
	}

	// Write to .part file first, then rename
	partPath := destPath + ".part"
	f, err := os.Create(partPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot create file: %v", err), http.StatusInternalServerError)
		return
	}

	// Stream with 1MB buffer
	buf := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(f, r.Body, buf)
	f.Close()

	if err != nil {
		os.Remove(partPath)
		http.Error(w, fmt.Sprintf("Upload failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Rename .part to final name
	err = os.Rename(partPath, destPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("File rename failed: %v", err), http.StatusInternalServerError)
		return
	}

	_ = deviceID // Will be used for peer tracking in later stages
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"received","file":"%s"}`, fileName)
}
