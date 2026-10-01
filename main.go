// NetFiles — offline LAN file-sharing tool
// Created by Mohammed Salah
//
// Main entry point: handles CLI commands and service startup.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"netfiles/core/config"
	"netfiles/core/firewall"
	"netfiles/core/folders"
	"netfiles/core/identity"
	"netfiles/core/ports"
	"netfiles/core/server"
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
	case "start":
		cmdStart(args[1:])
	case "stop":
		cmdStop()
	case "uninstall":
		cmdUninstall(args[1:])
	case "status":
		cmdStatus()
	case "rename":
		cmdRename(args[1:])
	case "setup":
		cmdSetup(args[1:])
	case "about", "--version":
		cmdAbout()
	case "":
		// No args: start service
		cmdStart(nil)
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

func cmdStart(args []string) {
	flags := parseStartFlags(args)

	// Always show bat splash on startup unless explicitly disabled with --plain
	if !flags.plain {
		printSplash()
	}

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Apply CLI flags
	if flags.port > 0 {
		cfg.HTTPPort = flags.port
	}
	if flags.folder != "" {
		cfg.RootFolder = flags.folder
	}
	if flags.password != "" {
		cfg.SetPassword(flags.password)
	}

	// Identity: name prompt or CLI flag
	if flags.name != "" {
		cfg.DisplayName = identity.SanitizeName(flags.name)
	} else if flags.number > 0 {
		cfg.DisplayName = identity.NumberToName(flags.number, cfg.NamePrefix)
	}

	if cfg.DeviceID == "" {
		cfg.DeviceID = identity.GenerateDeviceID()
	}

	if cfg.FirstSeen == "" {
		cfg.FirstSeen = time.Now().Format(time.RFC3339)
	}

	if cfg.DisplayName == "" {
		// Check name.txt next to exe
		exeName := identity.ReadNameFile()
		if exeName != "" {
			cfg.DisplayName = identity.SanitizeName(exeName)
		} else if isTerminal() {
			// Interactive prompt
			name, err := identity.PromptName(cfg.NamePrefix)
			if err != nil {
				fmt.Printf("[FAIL] Name input error: %v\n", err)
				os.Exit(1)
			}
			cfg.DisplayName = name
		} else {
			// Fallback: use IP octets
			cfg.DisplayName = identity.FallbackName()
		}
	}

	// Port check & auto-pick
	httpPort, err := ports.FindFreePort(cfg.HTTPPort)
	if err != nil {
		fmt.Printf("[FAIL] Cannot find free HTTP port: %v\n", err)
		os.Exit(1)
	}
	if httpPort != cfg.HTTPPort {
		fmt.Printf("[INFO] Port %d busy, using %d instead\n", cfg.HTTPPort, httpPort)
	}
	cfg.HTTPPort = httpPort

	udpPort, err := ports.FindFreePort(cfg.UDPPort)
	if err != nil {
		fmt.Printf("[FAIL] Cannot find free UDP port: %v\n", err)
		os.Exit(1)
	}
	if udpPort != cfg.UDPPort {
		fmt.Printf("[INFO] Port %d busy, using %d instead\n", cfg.UDPPort, udpPort)
	}
	cfg.UDPPort = udpPort

	// Create folder layout
	err = folders.CreateLayout(cfg.RootFolder, cfg.DisplayName)
	if err != nil {
		fmt.Printf("[FAIL] Cannot create folders: %v\n", err)
		os.Exit(1)
	}

	// Save config BEFORE firewall (firewall may UAC-prompt and block)
	err = config.Save(cfg)
	if err != nil {
		fmt.Printf("[WARN] Failed to save config: %v\n", err)
	}

	// Register autostart if requested
	if flags.autostart {
		err = config.SetAutostart(true)
		if err != nil {
			fmt.Printf("[WARN] Failed to register autostart: %v\n", err)
		}
	}

	// Add firewall rules (may need elevation — non-fatal)
	go func() {
		err := firewall.EnsureRules(cfg.HTTPPort, cfg.UDPPort)
		if err != nil {
			fmt.Printf("[WARN] Firewall rules not added: %v\n", err)
			fmt.Println("       You may need to run as Administrator")
		} else {
			fmt.Println("[ OK ] Firewall rules configured")
		}
	}()

	// Print startup info
	ips := identity.GetLocalIPs()
	ipStr := "unknown"
	if len(ips) > 0 {
		ipStr = ips[len(ips)-1] // prefer last (usually the LAN IP)
	}

	fmt.Println("  -----------------------------------------")
	fmt.Printf("  Name      : \033[1;32m%s\033[0m\n", cfg.DisplayName)
	fmt.Printf("  Device ID : \033[90m%s\033[0m\n", cfg.ShortID())
	fmt.Printf("  IP        : \033[36m%s\033[0m\n", ipStr)
	fmt.Printf("  HTTP Port : \033[36m%d\033[0m\n", cfg.HTTPPort)
	fmt.Printf("  UDP Port  : \033[36m%d\033[0m\n", cfg.UDPPort)
	fmt.Printf("  Folder    : \033[1;33m%s\033[0m\n", cfg.RootFolder)
	fmt.Println("  -----------------------------------------")
	fmt.Println()

	// Start the HTTP server (blocks)
	srv := server.New(cfg)
	fmt.Println("[ \033[32mOK\033[0m ] Server is listening")
	fmt.Println("       Press Ctrl+C to stop")
	fmt.Println()
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Printf("[FAIL] Server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdStop() {
	_, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] NetFiles config not found")
		os.Exit(1)
	}

	// Remove firewall rules
	err = firewall.RemoveRules()
	if err != nil {
		fmt.Printf("[WARN] Failed to remove firewall rules: %v\n", err)
	}

	// Remove autostart
	err = config.SetAutostart(false)
	if err != nil {
		fmt.Printf("[WARN] Failed to remove autostart: %v\n", err)
	}

	fmt.Println("[ OK ] NetFiles stopped")
}

func cmdUninstall(args []string) {
	force := false
	for _, a := range args {
		if strings.EqualFold(a, "-force") || strings.EqualFold(a, "--force") {
			force = true
		}
	}

	if !force && isTerminal() {
		fmt.Print("Remove NetFiles completely? All files will be deleted. [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Stop first
	cmdStop()

	cfg, _ := config.Load()
	rootFolder := config.GetDefaultRootFolder()
	if cfg != nil {
		rootFolder = cfg.RootFolder
	}

	// Remove folders
	err := os.RemoveAll(rootFolder)
	if err != nil {
		fmt.Printf("[WARN] Failed to delete %s: %v\n", rootFolder, err)
	} else {
		fmt.Printf("[ OK ] Deleted %s\n", rootFolder)
	}

	// Remove config directory
	err = os.RemoveAll(config.ConfigDir)
	if err != nil {
		fmt.Printf("[WARN] Failed to delete config: %v\n", err)
	} else {
		fmt.Println("[ OK ] Config deleted")
	}

	fmt.Println("[ OK ] NetFiles fully removed")
}

func cmdStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] NetFiles config not found. Not installed.")
		os.Exit(1)
	}

	printSplash()

	fmt.Println("  === NetFiles Status ===")
	fmt.Printf("  Name      : %s\n", cfg.DisplayName)
	fmt.Printf("  Device ID : %s\n", cfg.ShortID())
	fmt.Printf("  Ports     : HTTP %d / UDP %d\n", cfg.HTTPPort, cfg.UDPPort)
	fmt.Printf("  Folder    : %s\n", cfg.RootFolder)

	ips := identity.GetLocalIPs()
	if len(ips) > 0 {
		fmt.Printf("  IPs       : %s\n", strings.Join(ips, ", "))
	}

	// Check if running
	running := ports.IsPortInUse(cfg.HTTPPort)
	if running {
		fmt.Println("  Status    : \033[32m● Running\033[0m")
	} else {
		fmt.Println("  Status    : \033[31m● Stopped\033[0m")
	}
	fmt.Println()
}

func cmdRename(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] NetFiles config not found")
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
	err = config.Save(cfg)
	if err != nil {
		fmt.Printf("[FAIL] Failed to save new name: %v\n", err)
		os.Exit(1)
	}

	// Rename the folder if it exists
	err = folders.RenamePeerFolder(cfg.RootFolder, oldName, newName)
	if err != nil {
		fmt.Printf("[WARN] Failed to rename folder: %v\n", err)
	}

	fmt.Printf("[ OK ] Renamed: %s -> %s\n", oldName, newName)
}

func cmdSetup(args []string) {
	// Stage (f) will add the pacman-style setup screen
	// For now, just run start
	cmdStart(args)
}

func cmdAbout() {
	printSplash()
}

func printUsage() {
	fmt.Println()
	fmt.Println("NetFiles — Offline LAN File Sharing Tool")
	fmt.Printf("%s\n\n", Credit)
	fmt.Println("Usage:")
	fmt.Println("  netfiles start [-Port N] [-Name X | -Number N] [-Password X] [-Folder path] [-Autostart]")
	fmt.Println("  netfiles stop")
	fmt.Println("  netfiles uninstall [-Force]")
	fmt.Println("  netfiles status")
	fmt.Println("  netfiles rename [name|number]")
	fmt.Println("  netfiles setup")
	fmt.Println("  netfiles about")
	fmt.Println("  netfiles --version")
}

type startFlags struct {
	port      int
	name      string
	number    int
	password  string
	folder    string
	autostart bool
	noConfirm bool
	noAnim    bool
	plain     bool
}

func parseStartFlags(args []string) startFlags {
	f := startFlags{}
	for i := 0; i < len(args); i++ {
		a := strings.ToLower(args[i])
		switch a {
		case "-port":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &f.port)
			}
		case "-name":
			if i+1 < len(args) {
				i++
				f.name = args[i]
			}
		case "-number":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &f.number)
			}
		case "-password":
			if i+1 < len(args) {
				i++
				f.password = args[i]
			}
		case "-folder":
			if i+1 < len(args) {
				i++
				f.folder = args[i]
			}
		case "-autostart":
			f.autostart = true
		case "-y", "--noconfirm":
			f.noConfirm = true
		case "--no-anim":
			f.noAnim = true
		case "--plain":
			f.plain = true
		}
	}
	return f
}
