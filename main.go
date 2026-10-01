// NetFiles — offline LAN file-sharing & Windows Network setup tool
// Created by Mohammed Salah
//
// Portable single-binary tool for Windows 10/11.
// Configures native Windows Network sharing (SMB) on Desktop with zero background overhead.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"netfiles/core/config"
	"netfiles/core/folders"
	"netfiles/core/identity"
	"netfiles/core/ports"
	"netfiles/core/server"
	"netfiles/core/sharing"
)

const (
	Version = "1.0.0"
	Credit  = "Created by Mohammed Salah"
)

func main() {
	initConsole()

	args := os.Args[1:]
	cmd := ""
	if len(args) > 0 {
		cmd = strings.ToLower(args[0])
	}

	switch cmd {
	case "setup", "install", "share":
		cmdSetup(args[1:])
	case "start":
		// If -serve or -daemon flag is passed, run the HTTP daemon
		if hasFlag(args, "-serve") || hasFlag(args, "--serve") || hasFlag(args, "-daemon") {
			cmdServe(args[1:])
		} else {
			cmdSetup(args[1:])
		}
	case "serve":
		cmdServe(args[1:])
	case "stop":
		cmdStop()
	case "uninstall", "remove", "unshare":
		cmdUninstall(args[1:])
	case "status":
		cmdStatus()
	case "rename":
		cmdRename(args[1:])
	case "about", "--version", "-v":
		cmdAbout()
	case "":
		// Double-clicked or launched with no args: run one-time setup
		cmdSetup(nil)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func waitForEnter() {
	if isTerminal() {
		fmt.Println("  Press Enter to exit...")
		var buf [1]byte
		_, _ = os.Stdin.Read(buf[:])
	}
}

func printSplash() {
	fmt.Println()
	fmt.Println("        \033[1;35m/\\            /\\\033[0m")
	fmt.Println("       \033[1;35m/  \\__      __/  \\\033[0m")
	fmt.Println("      \033[1;35m/ /\\   \\____/   /\\ \\\033[0m")
	fmt.Println("     \033[1;35m/ /  \\  \033[1;33m(O)\033[1;35m  \033[1;33m(O)\033[1;35m  /  \\ \\\033[0m")
	fmt.Println("     \033[1;35m\\/    \\____\\/____/    \\/\033[0m")
	fmt.Println("            \033[1;35m\\  \\/\\/  /\033[0m")
	fmt.Println("             \033[1;35m\\______/\033[0m")
	fmt.Println()
	fmt.Printf("       \033[1;36mNetFiles\033[0m \033[90mv%s\033[0m\n", Version)
	fmt.Printf("       \033[90m%s\033[0m\n", Credit)
	fmt.Println()
}

// cmdSetup performs the one-time Windows Network SMB sharing configuration
func cmdSetup(args []string) {
	printSplash()

	// Check if running as administrator; if not, request elevation
	if !sharing.IsElevated() {
		fmt.Println("  [!] Administrator privileges are required to configure Windows Network share.")
		fmt.Println("      Requesting elevation...")
		time.Sleep(300 * time.Millisecond)

		err := sharing.Elevate(os.Args[1:])
		if err != nil {
			fmt.Printf("  [FAIL] Could not elevate privileges: %v\n", err)
			fmt.Println("         Please right-click netfiles.exe and choose 'Run as administrator'.")
			waitForEnter()
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	targetFolder := cfg.RootFolder
	if targetFolder == "" {
		targetFolder = config.GetDefaultRootFolder()
		cfg.RootFolder = targetFolder
	}

	fmt.Println("  === One-Time Windows Network Setup ===")
	fmt.Println("  [1/4] Preparing Desktop folder structure...")
	err = folders.CreateLayout(targetFolder, cfg.DisplayName)
	if err != nil {
		fmt.Printf("  [FAIL] Failed to create folder layout: %v\n", err)
		waitForEnter()
		os.Exit(1)
	}

	fmt.Println("  [2/4] Setting folder permissions for LAN access...")
	fmt.Println("  [3/4] Enabling Network Discovery and File Sharing services...")
	fmt.Println("  [4/4] Creating Windows Network SMB share 'NetFiles'...")

	info, err := sharing.SetupShare(targetFolder)
	if err != nil {
		fmt.Printf("  [FAIL] Error configuring Windows share: %v\n", err)
		waitForEnter()
		os.Exit(1)
	}

	// Save configuration
	if cfg.DeviceID == "" {
		cfg.DeviceID = identity.GenerateDeviceID()
	}
	if cfg.FirstSeen == "" {
		cfg.FirstSeen = time.Now().Format(time.RFC3339)
	}
	_ = config.Save(cfg)

	// Display completion card
	fmt.Println()
	fmt.Println("  \033[1;32m============================================================\033[0m")
	fmt.Println("  \033[1;32m   [ OK ] NetFiles Shared Successfully on Windows Network   \033[0m")
	fmt.Println("  \033[1;32m============================================================\033[0m")
	fmt.Printf("   Share Name    : \033[1;36m%s\033[0m\n", info.ShareName)
	fmt.Printf("   Network Path  : \033[1;33m%s\033[0m\n", info.UNCPath)
	fmt.Printf("   Local Folder  : %s\n", info.LocalPath)
	fmt.Printf("   Computer Name : %s\n", info.ComputerName)
	fmt.Printf("   LAN IP        : %s\n", info.IPAddress)
	fmt.Println("  ------------------------------------------------------------")
	fmt.Println("   \033[1mHow to access from other computers on your LAN:\033[0m")
	fmt.Println("   1. Open Windows Explorer -> Click 'Network' on the left.")
	fmt.Printf("   2. Double-click '\033[1;36m%s\033[0m' -> Open '\033[1;36mNetFiles\033[0m'.\n", info.ComputerName)
	fmt.Printf("   OR type in address bar: \033[1;33m%s\033[0m\n", info.UNCPath)
	fmt.Println("  ------------------------------------------------------------")
	fmt.Println("   \033[90mNOTE: This was a one-time setup. You do NOT need to keep\033[0m")
	fmt.Println("   \033[90m      this tool running. Windows manages the share natively.\033[0m")
	fmt.Println("  \033[1;32m============================================================\033[0m")
	fmt.Println()

	waitForEnter()
}

func cmdStatus() {
	printSplash()

	cfg, _ := config.Load()
	targetFolder := config.GetDefaultRootFolder()
	if cfg != nil && cfg.RootFolder != "" {
		targetFolder = cfg.RootFolder
	}

	info, _ := sharing.GetShareInfo(targetFolder)

	fmt.Println("  === NetFiles Status ===")
	fmt.Printf("  Local Folder  : %s\n", info.LocalPath)
	fmt.Printf("  Computer Name : %s\n", info.ComputerName)
	fmt.Printf("  LAN IP        : %s\n", info.IPAddress)
	fmt.Printf("  Network Path  : %s\n", info.UNCPath)

	if info.IsShared {
		fmt.Println("  Status        : \033[1;32m● Active on Windows Network (Shared)\033[0m")
	} else {
		fmt.Println("  Status        : \033[1;31m● Not Shared\033[0m (Run 'netfiles setup' to enable)")
	}
	fmt.Println()
	waitForEnter()
}

func cmdUninstall(args []string) {
	printSplash()

	if !sharing.IsElevated() {
		fmt.Println("  [!] Administrator privileges are required to remove network share.")
		fmt.Println("      Requesting elevation...")
		_ = sharing.Elevate(os.Args[1:])
		os.Exit(0)
	}

	force := false
	for _, a := range args {
		if strings.EqualFold(a, "-force") || strings.EqualFold(a, "--force") {
			force = true
		}
	}

	if !force && isTerminal() {
		fmt.Print("  Remove NetFiles share from Windows Network? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(line)
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			fmt.Println("  Cancelled.")
			waitForEnter()
			return
		}
	}

	// Remove SMB share
	err := sharing.RemoveShare()
	if err != nil {
		fmt.Printf("  [WARN] Failed to remove share: %v\n", err)
	} else {
		fmt.Println("  [ OK ] Windows Network share removed")
	}

	// Remove config
	_ = os.RemoveAll(config.ConfigDir)
	fmt.Println("  [ OK ] NetFiles configuration removed")

	if force {
		cfg, _ := config.Load()
		targetFolder := config.GetDefaultRootFolder()
		if cfg != nil && cfg.RootFolder != "" {
			targetFolder = cfg.RootFolder
		}
		_ = os.RemoveAll(targetFolder)
		fmt.Printf("  [ OK ] Folder %s deleted\n", targetFolder)
	}

	fmt.Println()
	fmt.Println("  [ OK ] NetFiles has been completely uninstalled.")
	fmt.Println()
	waitForEnter()
}

func cmdStop() {
	err := sharing.RemoveShare()
	if err != nil {
		fmt.Printf("[WARN] Failed to unshare: %v\n", err)
	} else {
		fmt.Println("[ OK ] NetFiles network share stopped")
	}
}

func cmdRename(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] NetFiles config not found. Run setup first.")
		os.Exit(1)
	}

	var newName string
	if len(args) > 0 {
		input := strings.Join(args, " ")
		newName = identity.ParseNameInput(input, cfg.NamePrefix)
	} else if isTerminal() {
		name, err := identity.PromptName(cfg.NamePrefix)
		if err != nil {
			fmt.Printf("[FAIL] Error: %v\n", err)
			os.Exit(1)
		}
		newName = name
	} else {
		fmt.Println("[FAIL] Name required: netfiles rename <name|number>")
		os.Exit(1)
	}

	oldName := cfg.DisplayName
	cfg.DisplayName = newName
	_ = config.Save(cfg)

	fmt.Printf("[ OK ] Renamed: %s -> %s\n", oldName, newName)
}

func cmdAbout() {
	printSplash()
	waitForEnter()
}

func cmdServe(args []string) {
	printSplash()

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	httpPort, _ := ports.FindFreePort(cfg.HTTPPort)
	cfg.HTTPPort = httpPort

	udpPort, _ := ports.FindFreePort(cfg.UDPPort)
	cfg.UDPPort = udpPort

	srv := server.New(cfg)
	fmt.Printf("  [ OK ] NetFiles HTTP transfer server listening on port %d\n", cfg.HTTPPort)
	fmt.Println("         Press Ctrl+C to stop")
	fmt.Println()
	_ = srv.ListenAndServe()
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if strings.EqualFold(a, flag) {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Println()
	fmt.Println("NetFiles — Offline LAN File Sharing Tool")
	fmt.Printf("%s\n\n", Credit)
	fmt.Println("Usage:")
	fmt.Println("  netfiles                   # One-time setup (creates Desktop share on Windows Network)")
	fmt.Println("  netfiles setup             # One-time setup & network sharing")
	fmt.Println("  netfiles status            # Check network share status")
	fmt.Println("  netfiles uninstall [-Force]# Remove share and configuration")
	fmt.Println("  netfiles rename <name>     # Update device display name")
	fmt.Println("  netfiles serve             # (Optional) Start standalone HTTP server")
	fmt.Println("  netfiles about             # Show app information")
	fmt.Println()
}
