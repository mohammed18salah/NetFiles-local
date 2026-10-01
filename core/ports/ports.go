// NetFiles ports package — port availability checking
// Created by Mohammed Salah
package ports

import (
	"fmt"
	"net"
	"time"
)

// Ports that should never be used
var reservedPorts = map[int]bool{
	80: true, 443: true, 445: true,
	3000: true, 8080: true, 8443: true,
	3389: true, 5985: true, 5986: true,
}

// FindFreePort starts from the preferred port and finds the next available one.
// It checks both TCP and UDP availability.
func FindFreePort(preferred int) (int, error) {
	maxAttempts := 100

	for i := 0; i < maxAttempts; i++ {
		port := preferred + i

		if port > 65535 {
			break
		}

		if reservedPorts[port] {
			continue
		}

		if isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free port found starting from %d", preferred)
}

// isPortAvailable checks if a port is free for both TCP and UDP
func isPortAvailable(port int) bool {
	// Check TCP
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()

	// Check UDP
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return false
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return false
	}
	conn.Close()

	return true
}

// IsPortInUse checks if a specific TCP port is currently in use
func IsPortInUse(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
