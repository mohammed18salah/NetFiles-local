// NetFiles service package — Windows Service management & execution
// Created by Mohammed Salah
package service

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"netfiles/core/config"
	"netfiles/core/folders"
	"netfiles/core/ports"
	"netfiles/core/server"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	ServiceName        = "NetFilesSvc"
	ServiceDisplayName = "NetFiles LAN Service"
	ServiceDesc        = "NetFiles offline LAN file sharing background service. Created by Mohammed Salah."
)

// Install registers NetFilesSvc in Windows SCM with delayed auto-start & restart on failure
func Install(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to Service Control Manager: %w", err)
	}
	defer m.Disconnect()

	// Check if service already exists
	s, err := m.OpenService(ServiceName)
	if err == nil {
		// Existing service: update config
		defer s.Close()
		err = s.UpdateConfig(mgr.Config{
			ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS,
			StartType:        mgr.StartAutomatic,
			DelayedAutoStart: true,
			DisplayName:      ServiceDisplayName,
			Description:      ServiceDesc,
			BinaryPathName:   fmt.Sprintf(`"%s" service`, exePath),
		})
		if err != nil {
			return fmt.Errorf("failed to update service config: %w", err)
		}
		setRecovery(s)
		return nil
	}

	// Create new service
	s, err = m.CreateService(ServiceName, exePath, mgr.Config{
		ServiceType:      windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
		DisplayName:      ServiceDisplayName,
		Description:      ServiceDesc,
		BinaryPathName:   fmt.Sprintf(`"%s" service`, exePath),
	}, "service")
	if err != nil {
		return fmt.Errorf("failed to create service %s: %w", ServiceName, err)
	}
	defer s.Close()

	setRecovery(s)
	return nil
}

func setRecovery(s *mgr.Service) {
	recovery := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 10 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
	}
	_ = s.SetRecoveryActions(recovery, 86400)
}

// Uninstall stops and deletes NetFilesSvc from Windows SCM
func Uninstall() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return nil // Not installed, nothing to do
	}
	defer s.Close()

	// Stop service if running
	status, err := s.Query()
	if err == nil && status.State != svc.Stopped {
		_, _ = s.Control(svc.Stop)
		// Wait up to 5 seconds for service to stop
		for i := 0; i < 10; i++ {
			time.Sleep(500 * time.Millisecond)
			status, err = s.Query()
			if err != nil || status.State == svc.Stopped {
				break
			}
		}
	}

	return s.Delete()
}

// Start launches NetFilesSvc
func Start() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", ServiceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err == nil && status.State == svc.Running {
		return nil // Already running
	}

	return s.Start()
}

// Stop stops NetFilesSvc
func Stop() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", ServiceName, err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	return err
}

// QueryStatus returns the current service status ("Running", "Stopped", "Not Installed", etc.)
func QueryStatus() (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return queryStatusViaSC()
	}
	defer m.Disconnect()

	namePtr, err := windows.UTF16PtrFromString(ServiceName)
	if err != nil {
		return "Unknown", err
	}

	h, err := windows.OpenService(m.Handle, namePtr, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return queryStatusViaSC()
	}
	s := &mgr.Service{Name: ServiceName, Handle: h}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return queryStatusViaSC()
	}

	switch status.State {
	case svc.Running:
		return "Running", nil
	case svc.Stopped:
		return "Stopped", nil
	case svc.StartPending:
		return "Starting", nil
	case svc.StopPending:
		return "Stopping", nil
	default:
		return "Paused", nil
	}
}

func queryStatusViaSC() (string, error) {
	cmd := exec.Command("sc", "query", ServiceName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "Not Installed", nil
	}
	outStr := string(out)
	if strings.Contains(outStr, "RUNNING") {
		return "Running", nil
	} else if strings.Contains(outStr, "STOPPED") {
		return "Stopped", nil
	} else if strings.Contains(outStr, "START_PENDING") {
		return "Starting", nil
	}
	return "Not Installed", nil
}

// IsRunning returns true if the service is currently running
func IsRunning() bool {
	st, err := QueryStatus()
	return err == nil && st == "Running"
}

// IsInstalled returns true if the service is registered in SCM
func IsInstalled() bool {
	st, _ := QueryStatus()
	return st != "Not Installed" && st != "Unknown"
}

// Run is the entrypoint invoked by the Windows Service Control Manager
func Run() error {
	return svc.Run(ServiceName, &serviceHandler{})
}

type serviceHandler struct{}

func (h *serviceHandler) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Auto-pick ports if needed
	if httpPort, err := ports.FindFreePort(cfg.HTTPPort); err == nil {
		cfg.HTTPPort = httpPort
	}
	if udpPort, err := ports.FindFreePort(cfg.UDPPort); err == nil {
		cfg.UDPPort = udpPort
	}

	// Ensure folder layout
	_ = folders.CreateLayout(cfg.RootFolder, cfg.DisplayName)

	srv := server.New(cfg)
	serverDone := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverDone <- err
		}
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				// Stop the HTTP server
				_ = srv.Close()
				changes <- svc.Status{State: svc.Stopped}
				return
			default:
			}
		case <-serverDone:
			changes <- svc.Status{State: svc.Stopped}
			return
		}
	}
}
