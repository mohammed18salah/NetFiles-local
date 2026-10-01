// NetFiles identity package — device naming, validation, prompting
// Created by Mohammed Salah
package identity

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Characters not allowed in Windows folder names
var invalidChars = regexp.MustCompile(`[\\/:*?"<>|]`)

const MaxNameLength = 32

// GenerateDeviceID creates a new random UUID
func GenerateDeviceID() string {
	return uuid.New().String()
}

// SanitizeName removes invalid characters and trims to max length
func SanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = invalidChars.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	// Trim to max length (rune-safe)
	if utf8.RuneCountInString(name) > MaxNameLength {
		runes := []rune(name)
		name = string(runes[:MaxNameLength])
	}

	return name
}

// NumberToName converts a number to a zero-padded name with prefix
// e.g., 7 with prefix "PC-" -> "PC-007"
func NumberToName(num int, prefix string) string {
	if num > 999 {
		return fmt.Sprintf("%s%d", prefix, num)
	}
	return fmt.Sprintf("%s%03d", prefix, num)
}

// ParseNameInput determines if input is a number or a name
func ParseNameInput(input string, prefix string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Check if all digits
	allDigits := true
	for _, r := range input {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}

	if allDigits {
		var num int
		fmt.Sscanf(input, "%d", &num)
		if num > 0 {
			return NumberToName(num, prefix)
		}
	}

	return SanitizeName(input)
}

// PromptName shows a prompt and reads the name interactively
func PromptName(prefix string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter device number or name: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("Name cannot be empty")
			continue
		}

		name := ParseNameInput(input, prefix)
		if name == "" {
			fmt.Println("Invalid name")
			continue
		}

		fmt.Printf("Name set: %s\n", name)
		return name, nil
	}
}

// ReadNameFile checks for a name.txt file next to the executable
func ReadNameFile() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	nameFile := filepath.Join(filepath.Dir(exe), "name.txt")
	data, err := os.ReadFile(nameFile)
	if err != nil {
		return ""
	}

	// Read first non-empty line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return ""
}

// FallbackName generates a name from the last two IP octets
func FallbackName() string {
	ips := GetLocalIPs()
	for _, ip := range ips {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			return fmt.Sprintf("PC-%s%s", parts[2], parts[3])
		}
	}
	return "PC-Unknown"
}

// GetLocalIPs returns all non-loopback IPv4 addresses
func GetLocalIPs() []string {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP
		if ip.IsLoopback() || ip.To4() == nil {
			continue
		}

		ips = append(ips, ip.String())
	}

	return ips
}
