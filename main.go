// NetFiles — offline LAN file-sharing & Windows Service control panel
// Created by Mohammed Salah
//
// Portable single-binary tool for Windows 10/11.
// Architecture:
// 1. SERVICE: Headless Windows Service "NetFilesSvc" (golang.org/x/sys/windows/svc)
// 2. CONTROL PANEL: Setup/Uninstall/Status/Doctor control interface
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"netfiles/core/config"
	"netfiles/core/emblem"
	"netfiles/core/identity"
	"netfiles/core/service"
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

	// 1. Windows Service entry point (invoked by Windows SCM)
	if cmd == "service" {
		err := service.Run()
		if err != nil {
			os.Exit(1)
		}
		return
	}

	// 2. CLI commands
	switch cmd {
	case "setup", "install":
		RunSetup(args[1:])
	case "uninstall", "remove":
		RunUninstall(args[1:])
	case "status":
		RunStatus()
	case "doctor":
		RunDoctor()
	case "start":
		if err := service.Start(); err != nil {
			fmt.Printf("[FAIL] Could not start NetFilesSvc: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[ OK ] NetFilesSvc started")
	case "stop":
		if err := service.Stop(); err != nil {
			fmt.Printf("[FAIL] Could not stop NetFilesSvc: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[ OK ] NetFilesSvc stopped")
	case "rename":
		cmdRename(args[1:])
	case "about", "--version", "-v":
		cmdAbout()
	case "menu", "panel", "control", "":
		RunControlPanel()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
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
		fmt.Println("Press Enter to continue...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
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

	fmt.Printf("[ OK ] Device renamed: %s -> %s\n", oldName, newName)
	waitForEnter()
}

func cmdAbout() {
	emblem.Render(false)
	fmt.Printf("       \033[1;36mNetFiles\033[0m \033[90mv%s\033[0m\n", Version)
	fmt.Printf("       \033[90m%s\033[0m\n\n", Credit)
	waitForEnter()
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
	fmt.Println("NetFiles — Offline LAN File Sharing System")
	fmt.Printf("%s\n\n", Credit)
	fmt.Println("Usage:")
	fmt.Println("  netfiles                   # Control panel / auto setup")
	fmt.Println("  netfiles setup [-y]        # Install service, folder & Explorer links")
	fmt.Println("  netfiles uninstall [-Force]# Remove service, rules & links")
	fmt.Println("  netfiles status            # View system and service status")
	fmt.Println("  netfiles doctor            # Run network & service diagnostics")
	fmt.Println("  netfiles start / stop      # Control background service")
	fmt.Println("  netfiles rename <name>     # Change this PC's network name")
	fmt.Println("  netfiles about             # Show version and info")
	fmt.Println("  netfiles service           # Windows SCM entry point (internal)")
	fmt.Println()
}
