// NetFiles config package — configuration management
// Created by Mohammed Salah
package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName           = "NetFilesTool"
	DefaultHTTPPort   = 47831
	DefaultUDPPort    = 47832
	DefaultRootFolder = `C:\NetFiles`
	DefaultNamePrefix = "PC-"
)

var ConfigDir string

func init() {
	if runtime.GOOS == "windows" {
		ConfigDir = filepath.Join(os.Getenv("ProgramData"), AppName)
	} else {
		ConfigDir = filepath.Join("/etc", "netfiles")
	}
}

type Config struct {
	DeviceID     string            `json:"device_id"`
	DisplayName  string            `json:"display_name"`
	NamePrefix   string            `json:"name_prefix"`
	HTTPPort     int               `json:"http_port"`
	UDPPort      int               `json:"udp_port"`
	RootFolder   string            `json:"root_folder"`
	PasswordHash string            `json:"password_hash,omitempty"`
	FirstSeen    string            `json:"first_seen"`
	Aliases      map[string]string `json:"aliases,omitempty"` // device_id -> local nickname
}

func Default() *Config {
	return &Config{
		NamePrefix: DefaultNamePrefix,
		HTTPPort:   DefaultHTTPPort,
		UDPPort:    DefaultUDPPort,
		RootFolder: DefaultRootFolder,
		Aliases:    make(map[string]string),
	}
}

func (c *Config) SetPassword(pw string) {
	if pw == "" {
		c.PasswordHash = ""
		return
	}
	h := sha256.Sum256([]byte(pw))
	c.PasswordHash = hex.EncodeToString(h[:])
}

func (c *Config) ShortID() string {
	if len(c.DeviceID) >= 8 {
		return c.DeviceID[:8]
	}
	return c.DeviceID
}

func configPath() string {
	return filepath.Join(ConfigDir, "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return nil, fmt.Errorf("لم يتم العثور على ملف الإعدادات: %w", err)
	}

	cfg := Default()
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("خطأ في قراءة الإعدادات: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	err := os.MkdirAll(ConfigDir, 0755)
	if err != nil {
		return fmt.Errorf("لم يتم إنشاء مجلد الإعدادات: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("خطأ في تحويل الإعدادات: %w", err)
	}

	return os.WriteFile(configPath(), data, 0644)
}
