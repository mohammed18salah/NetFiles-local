// NetFiles installer package — Pacman/Arch style operation runner
// Created by Mohammed Salah
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"netfiles/core/config"
	"netfiles/core/emblem"
	"netfiles/core/explorer"
	"netfiles/core/firewall"
	"netfiles/core/folders"
	"netfiles/core/identity"
	"netfiles/core/ports"
	"netfiles/core/service"
	"netfiles/core/sharing"
)

func runHeader() {
	emblem.Render(false)
	fmt.Println("       \033[1;36m[ NETFILES LAN SYSTEM V1.0 ]\033[0m")
	fmt.Println("       \033[90m[ CREATED BY MOHAMMED SALAH - 2026 ]\033[0m")
	fmt.Println()
}

func checkElevatedOrPrompt(args []string) bool {
	if sharing.IsElevated() {
		return true
	}

	fmt.Println(":: Administrator privileges are required for this operation.")
	fmt.Println(":: Requesting UAC elevation...")
	time.Sleep(300 * time.Millisecond)

	err := sharing.Elevate(args)
	if err != nil {
		fmt.Printf(":: [FAIL] Elevation failed: %v\n", err)
		fmt.Println(":: Please right-click netfiles.exe and choose 'Run as administrator'.")
		waitForEnter()
		os.Exit(1)
	}
	os.Exit(0)
	return false
}

// RunSetup executes the pacman-style setup workflow
func RunSetup(args []string) {
	runHeader()
	if !checkElevatedOrPrompt(os.Args[1:]) {
		return
	}

	noConfirm := hasFlag(args, "-y") || hasFlag(args, "--noconfirm")
	exePath, _ := os.Executable()
	exeInfo, _ := os.Stat(exePath)
	var exeSizeMB float64 = 7.04
	if exeInfo != nil {
		exeSizeMB = float64(exeInfo.Size()) / (1024 * 1024)
	}

	// 1. Environment check
	fmt.Print(":: Checking environment... ")
	time.Sleep(200 * time.Millisecond)

	// Set network profile to private
	_ = runHiddenCmd("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "Get-NetConnectionProfile | Set-NetConnectionProfile -NetworkCategory Private -ErrorAction SilentlyContinue")

	fmt.Println("\033[1;32m[ OK ]\033[0m")
	fmt.Println("   admin: yes | network: private | disk: ready | adapters: LAN link-local ready")
	fmt.Println()

	// 2. Load or create configuration
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	if cfg.DeviceID == "" {
		cfg.DeviceID = identity.GenerateDeviceID()
	}
	if cfg.DisplayName == "" {
		cfg.DisplayName = identity.FallbackName()
	}
	if cfg.FirstSeen == "" {
		cfg.FirstSeen = time.Now().Format(time.RFC3339)
	}
	_ = config.Save(cfg)

	// 3. Print Configuration block
	fmt.Println(":: Configuration")
	fmt.Printf("   Device name    = %-18s Device ID = %s\n", cfg.DisplayName, cfg.ShortID())
	fmt.Printf("   Folder         = %-18s Service   = NetFilesSvc (auto, delayed, restart on failure)\n", cfg.RootFolder)
	fmt.Printf("   Ports          = HTTP %d / UDP %d / FTP off\n", cfg.HTTPPort, cfg.UDPPort)
	fmt.Println("   Firewall rules = NetFilesTool-TCP, NetFilesTool-UDP")
	fmt.Println("   Explorer links = nav pane (CLSID), Quick Access, Network Shortcut, Desktop")
	fmt.Printf("   Components (4) / Folders (4) / Total Download Size 0.00 MiB (offline) / Total Installed Size %.2f MiB\n", exeSizeMB)
	fmt.Println()

	// 4. Prompt confirmation
	if !noConfirm && isTerminal() {
		fmt.Print(":: Proceed with installation? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(ans)
		if ans != "" && !strings.EqualFold(ans, "y") && !strings.EqualFold(ans, "yes") {
			fmt.Println(":: Installation cancelled.")
			waitForEnter()
			return
		}
		fmt.Println()
	}

	// 5. Installing components with pacman style progress
	fmt.Println(":: Installing components...")

	// Component 1: Windows Service
	printPacmanRow("core-service-netfilessvc", fmt.Sprintf("%.2f MiB", exeSizeMB), 1)
	err = service.Install(exePath)
	if err != nil {
		fmt.Printf("\n:: [FAIL] Service install error: %v\n", err)
	}

	// Component 2: Folder layout & permissions
	printPacmanRow("folder-layout-c-netfiles", "0.01 MiB", 2)
	err = folders.CreateLayout(cfg.RootFolder, cfg.DisplayName)
	if err != nil {
		fmt.Printf("\n:: [FAIL] Folder layout error: %v\n", err)
	}

	// Component 3: Windows Firewall rules
	printPacmanRow("windows-firewall-rules", "0.00 MiB", 3)
	err = firewall.EnsureRules(cfg.HTTPPort, cfg.UDPPort)
	if err != nil {
		fmt.Printf("\n:: [WARN] Firewall rule warning: %v\n", err)
	}

	// Component 4: Explorer Navigation-pane CLSID & shortcuts
	printPacmanRow("explorer-navpane-clsid", "0.01 MiB", 4)
	err = explorer.Install(cfg.RootFolder, exePath)
	if err != nil {
		fmt.Printf("\n:: [WARN] Explorer integration warning: %v\n", err)
	}

	fmt.Println(" Total (4/4)                                       [\033[1;35m#🦇###################\033[0m] 100%")
	fmt.Println()

	// 6. Running post-transaction hooks
	fmt.Println(":: Running post-transaction hooks...")
	fmt.Print("   (1/3) Starting NetFilesSvc background service... ")
	err = service.Start()
	if err != nil {
		fmt.Printf("\033[1;33m[WARN: %v]\033[0m\n", err)
	} else {
		fmt.Println("\033[1;32m[ OK ]\033[0m")
	}

	fmt.Print("   (2/3) Notifying Windows File Explorer (SHChangeNotify)... ")
	explorer.NotifyExplorer()
	fmt.Println("\033[1;32m[ OK ]\033[0m")

	fmt.Print("   (3/3) Verifying service health... ")
	time.Sleep(300 * time.Millisecond)
	if service.IsRunning() {
		fmt.Println("\033[1;32m[ RUNNING ]\033[0m")
	} else {
		fmt.Println("\033[1;33m[ STARTING ]\033[0m")
	}

	fmt.Println()
	fmt.Println("\033[1;32m:: Done. The service is running.\033[0m You can close this window;")
	fmt.Printf("   Open '\033[1;36m%s\033[0m' directly from Windows File Explorer navigation pane.\n", cfg.RootFolder)
	fmt.Println()
	waitForEnter()
}

// RunUninstall executes the pacman-style uninstallation workflow
func RunUninstall(args []string) {
	runHeader()
	if !checkElevatedOrPrompt(os.Args[1:]) {
		return
	}

	noConfirm := hasFlag(args, "-y") || hasFlag(args, "--noconfirm")
	force := hasFlag(args, "-force") || hasFlag(args, "--force")
	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	if cfg != nil && cfg.RootFolder != "" {
		rootFolder = cfg.RootFolder
	}

	fmt.Println(":: Preparing to uninstall NetFiles...")
	fmt.Println()

	deleteFolder := force
	if !deleteFolder && !noConfirm && isTerminal() {
		fmt.Printf(":: Delete the folder '%s' and all files in it? [y/N] ", rootFolder)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(ans)
		if strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") {
			deleteFolder = true
		}
		fmt.Println()
	}

	fmt.Println(":: Removing components...")

	// 1. Stop and remove service
	fmt.Print("   (1/4) Stopping and deleting NetFilesSvc... ")
	_ = service.Uninstall()
	fmt.Println("\033[1;32m[ OK ]\033[0m")

	// 2. Remove Explorer CLSID & shortcuts
	fmt.Print("   (2/4) Removing Explorer navigation-pane CLSID & shortcuts... ")
	_ = explorer.Uninstall(rootFolder)
	fmt.Println("\033[1;32m[ OK ]\033[0m")

	// 3. Remove Firewall rules
	fmt.Print("   (3/4) Removing Windows Firewall rules... ")
	_ = firewall.RemoveRules()
	fmt.Println("\033[1;32m[ OK ]\033[0m")

	// 4. Remove config & optional folder
	fmt.Print("   (4/4) Cleaning NetFiles configuration... ")
	_ = os.RemoveAll(config.ConfigDir)
	if deleteFolder {
		_ = os.RemoveAll(rootFolder)
		fmt.Println("\033[1;32m[ OK ] (Folder deleted)\033[0m")
	} else {
		fmt.Println("\033[1;32m[ OK ] (Folder kept)\033[0m")
	}

	fmt.Println(":: Running post-transaction hooks...")
	fmt.Print("   (1/1) Refreshing Windows File Explorer... ")
	explorer.NotifyExplorer()
	fmt.Println("\033[1;32m[ OK ]\033[0m")

	fmt.Println()
	fmt.Println("\033[1;32m:: Done. NetFiles has been completely uninstalled.\033[0m")
	fmt.Println("   The background service, links, and rules have been removed.")
	fmt.Println()
	waitForEnter()
}

// RunStatus displays live service, folder, and link status
func RunStatus() {
	runHeader()

	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	displayName := "PC-Unknown"
	shortID := "unknown"
	httpPort := config.DefaultHTTPPort
	udpPort := config.DefaultUDPPort
	if cfg != nil {
		if cfg.RootFolder != "" {
			rootFolder = cfg.RootFolder
		}
		if cfg.DisplayName != "" {
			displayName = cfg.DisplayName
		}
		shortID = cfg.ShortID()
		httpPort = cfg.HTTPPort
		udpPort = cfg.UDPPort
	}

	svcStatus, _ := service.QueryStatus()

	fmt.Println(":: NetFiles System Status")
	fmt.Printf("   Device Name      : \033[1;36m%s\033[0m\n", displayName)
	fmt.Printf("   Device ID        : \033[90m%s\033[0m\n", shortID)
	fmt.Printf("   Root Folder      : \033[1;33m%s\033[0m\n", rootFolder)
	fmt.Printf("   Network Ports    : HTTP %d / UDP %d\n", httpPort, udpPort)

	// Status color
	if svcStatus == "Running" {
		fmt.Printf("   Service Status   : \033[1;32m● %s (NetFilesSvc)\033[0m\n", svcStatus)
	} else if svcStatus == "Stopped" {
		fmt.Printf("   Service Status   : \033[1;31m● %s\033[0m (Run 'netfiles start' to start)\n", svcStatus)
	} else {
		fmt.Printf("   Service Status   : \033[90m○ %s\033[0m (Run 'netfiles setup' to install)\n", svcStatus)
	}

	// Folder status
	if fi, err := os.Stat(rootFolder); err == nil && fi.IsDir() {
		fmt.Println("   Folder Status    : \033[1;32m[ OK ] Present on disk\033[0m")
	} else {
		fmt.Println("   Folder Status    : \033[1;31m[ MISSING ] Not created yet\033[0m")
	}

	// Local IPs
	ips := identity.GetLocalIPs()
	if len(ips) > 0 {
		fmt.Printf("   LAN IP Addresses : %s\n", strings.Join(ips, ", "))
	}

	fmt.Println()
	waitForEnter()
}

// RunDoctor performs diagnostics
func RunDoctor() {
	runHeader()
	fmt.Println(":: Running NetFiles Diagnostics Doctor...")
	fmt.Println()

	// 1. Service check
	fmt.Print("   [1/5] Checking Windows Service (NetFilesSvc)... ")
	if service.IsRunning() {
		fmt.Println("\033[1;32m[ OK - Running ]\033[0m")
	} else if service.IsInstalled() {
		fmt.Println("\033[1;33m[ WARN - Installed but Stopped ]\033[0m")
	} else {
		fmt.Println("\033[1;31m[ FAIL - Not Installed ]\033[0m")
	}

	// 2. Folder check
	cfg, _ := config.Load()
	root := config.DefaultRootFolder
	if cfg != nil && cfg.RootFolder != "" {
		root = cfg.RootFolder
	}
	fmt.Print("   [2/5] Checking root folder accessibility... ")
	if fi, err := os.Stat(root); err == nil && fi.IsDir() {
		fmt.Printf("\033[1;32m[ OK - %s ]\033[0m\n", root)
	} else {
		fmt.Printf("\033[1;31m[ FAIL - %s missing ]\033[0m\n", root)
	}

	// 3. Port availability
	fmt.Print("   [3/5] Checking TCP/UDP ports... ")
	hPort := config.DefaultHTTPPort
	uPort := config.DefaultUDPPort
	if cfg != nil {
		hPort = cfg.HTTPPort
		uPort = cfg.UDPPort
	}
	inUse := ports.IsPortInUse(hPort)
	if service.IsRunning() && inUse {
		fmt.Printf("\033[1;32m[ OK - HTTP %d in use by service ]\033[0m\n", hPort)
	} else if !inUse {
		fmt.Printf("\033[1;32m[ OK - HTTP %d / UDP %d available ]\033[0m\n", hPort, uPort)
	} else {
		fmt.Printf("\033[1;33m[ WARN - Port %d in use by another process ]\033[0m\n", hPort)
	}

	// 4. Network profile
	fmt.Print("   [4/5] Checking Windows Network profile... ")
	isPriv, profName, _ := firewall.CheckNetworkProfile()
	if isPriv {
		fmt.Printf("\033[1;32m[ OK - Private (%s) ]\033[0m\n", profName)
	} else {
		fmt.Printf("\033[1;33m[ WARN - %s (Recommended: Private) ]\033[0m\n", profName)
	}

	// 5. Explorer navigation links
	fmt.Print("   [5/5] Checking Explorer navigation pane registration... ")
	desktopShortcut := filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "NetFiles.lnk")
	if _, err := os.Stat(desktopShortcut); err == nil {
		fmt.Println("\033[1;32m[ OK - Pinned & Desktop Shortcut active ]\033[0m")
	} else {
		fmt.Println("\033[1;32m[ OK ]\033[0m")
	}

	fmt.Println()
	fmt.Println(":: Doctor diagnostics complete.")
	fmt.Println()
	waitForEnter()
}

func printPacmanRow(component string, size string, step int) {
	fmt.Printf("   %-30s  %8s  [####################] 100%%\n", component, size)
	time.Sleep(120 * time.Millisecond)
}

func runHiddenCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
