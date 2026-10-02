// NetFiles control panel — Interactive management interface
// Created by Mohammed Salah
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"netfiles/core/config"
	"netfiles/core/explorer"
	"netfiles/core/folders"
	"netfiles/core/identity"
	"netfiles/core/service"
	"netfiles/core/sharing"
)

// clearScreen resets the terminal viewport
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// pause waits for the user to press Enter using the shared reader
func pause(reader *bufio.Reader) {
	fmt.Print("Press Enter to continue...")
	_, _ = reader.ReadString('\n')
}

// ensureElevated checks if process is admin, or asks to elevate
func ensureElevated(reader *bufio.Reader, actionDesc string) bool {
	if sharing.IsElevated() {
		return true
	}

	fmt.Println()
	fmt.Printf(":: Administrator privileges are required to %s.\n", actionDesc)
	fmt.Print(":: Request UAC elevation now? [Y/n] ")
	ans, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	ans = strings.TrimSpace(ans)
	if ans != "" && !strings.EqualFold(ans, "y") && !strings.EqualFold(ans, "yes") {
		fmt.Println(":: Action cancelled.")
		pause(reader)
		return false
	}

	fmt.Println(":: Requesting UAC elevation...")
	time.Sleep(300 * time.Millisecond)
	err = sharing.Elevate(nil)
	if err != nil {
		fmt.Printf(":: [FAIL] Elevation failed: %v\n", err)
		fmt.Println(":: Please right-click netfiles.exe and choose 'Run as administrator'.")
		pause(reader)
		return false
	}
	os.Exit(0)
	return false
}

// detectInstalled checks whether NetFiles is already installed or configured
func detectInstalled() (installed bool, rootFolder string, displayName string, svcStatus string) {
	cfg, _ := config.Load()
	rootFolder = config.DefaultRootFolder
	displayName = "PC-Unknown"
	if cfg != nil {
		if cfg.RootFolder != "" {
			rootFolder = cfg.RootFolder
		}
		if cfg.DisplayName != "" {
			displayName = cfg.DisplayName
		}
	}

	svcStatus, _ = service.QueryStatus()

	folderExists := false
	if fi, err := os.Stat(rootFolder); err == nil && fi.IsDir() {
		folderExists = true
	}

	configExists := false
	if _, err := os.Stat(config.ConfigPath()); err == nil {
		configExists = true
	}

	installed = service.IsInstalled() || folderExists || configExists
	return installed, rootFolder, displayName, svcStatus
}

// RunControlPanel is the interactive loop when netfiles.exe is run with no CLI arguments
func RunControlPanel() {
	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()
		installed, rootFolder, displayName, svcStatus := detectInstalled()
		var keepRunning bool
		if installed {
			keepRunning = showInstalledMenu(reader, rootFolder, displayName, svcStatus)
		} else {
			keepRunning = showNotInstalledMenu(reader)
		}
		if !keepRunning {
			break
		}
	}
}

// showInstalledMenu renders the control panel for an existing installation
func showInstalledMenu(reader *bufio.Reader, rootFolder, displayName, svcStatus string) bool {
	runHeader()

	folderStatus := "\033[1;32m[ OK - Present ]\033[0m"
	if fi, err := os.Stat(rootFolder); err != nil || !fi.IsDir() {
		folderStatus = "\033[1;31m[ MISSING ]\033[0m"
	}

	svcColor := "\033[1;32m"
	if svcStatus != "Running" {
		svcColor = "\033[1;31m"
	}

	modeStr := "\033[1;32m[ Administrator ]\033[0m"
	if !sharing.IsElevated() {
		modeStr = "\033[1;33m[ Standard User ]\033[0m"
	}

	fmt.Println(":: Existing NetFiles installation detected on this computer:")
	fmt.Printf("   Computer Name : \033[1;36m%s\033[0m\n", displayName)
	fmt.Printf("   Root Folder   : \033[1;33m%s\033[0m %s\n", rootFolder, folderStatus)
	fmt.Printf("   Service       : %s[ %s ]\033[0m (NetFilesSvc)\n", svcColor, svcStatus)
	fmt.Printf("   Access Mode   : %s\n", modeStr)
	fmt.Println()
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("|                       NETFILES CONTROL PANEL                         |")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("  [1] Rename this PC         - Change computer name on LAN")
	fmt.Println("  [2] Background Service     - Start / Stop / Restart NetFilesSvc")
	fmt.Println("  [3] Delete Folder & Files  - Permanently wipe NetFiles folder")
	fmt.Println("  [4] Uninstall NetFiles     - Remove service, firewall & links")
	fmt.Println("  [5] Repair Explorer Links  - Refresh navigation pane & Desktop links")
	fmt.Println("  [6] System Diagnostics     - Run Doctor health checks")
	fmt.Println("  [7] Re-run Full Setup      - Reinstall components & reset settings")
	fmt.Println("  [8] Exit")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Print("Select option [1-8]: ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		menuRenamePC(reader)
		return true
	case "2":
		menuServiceControl(reader)
		return true
	case "3":
		menuDeleteFolder(reader)
		return true
	case "4":
		menuUninstall(reader)
		return true
	case "5":
		menuRepairExplorerLinks(reader)
		return true
	case "6":
		RunDoctor()
		return true
	case "7":
		RunSetup(nil)
		return true
	case "8", "q", "exit":
		fmt.Println()
		fmt.Println(":: Exiting NetFiles Control Panel.")
		fmt.Println("   The NetFilesSvc background service remains active on your LAN.")
		time.Sleep(400 * time.Millisecond)
		return false
	default:
		fmt.Println(":: Invalid choice. Please select 1-8.")
		time.Sleep(800 * time.Millisecond)
		return true
	}
}

// showNotInstalledMenu renders setup options when no installation is detected
func showNotInstalledMenu(reader *bufio.Reader) bool {
	runHeader()
	fmt.Println(":: NetFiles is NOT installed on this computer yet.")
	fmt.Println()
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("|                          NETFILES SETUP MENU                         |")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("  [1] Install NetFiles now (Automatic full setup)")
	fmt.Println("  [2] Custom Setup (Configure computer name and root folder)")
	fmt.Println("  [3] System Diagnostics (Doctor)")
	fmt.Println("  [4] Exit")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Print("Select option [1-4]: ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1", "":
		RunSetup(nil)
		return true
	case "2":
		menuCustomSetup(reader)
		return true
	case "3":
		RunDoctor()
		return true
	case "4", "q", "exit":
		fmt.Println()
		fmt.Println(":: Exiting NetFiles.")
		time.Sleep(400 * time.Millisecond)
		return false
	default:
		fmt.Println(":: Invalid choice. Please select 1-4.")
		time.Sleep(800 * time.Millisecond)
		return true
	}
}

// menuRenamePC allows changing the computer name on LAN
func menuRenamePC(reader *bufio.Reader) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                         RENAME THIS COMPUTER                           ")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Printf("Current Computer Name: %s\n\n", cfg.DisplayName)
	fmt.Println("You can enter a full name (e.g. Lab-PC-05) or just a number (e.g. 5).")
	fmt.Print("Enter new name (or press Enter to cancel): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println(":: Rename cancelled. Computer name unchanged.")
		pause(reader)
		return
	}

	newName := identity.ParseNameInput(input, cfg.NamePrefix)
	oldName := cfg.DisplayName
	cfg.DisplayName = newName
	err = config.Save(cfg)
	if err != nil {
		fmt.Printf(":: [FAIL] Could not save config: %v\n", err)
		pause(reader)
		return
	}

	// Update folder layout if root exists
	if fi, err := os.Stat(cfg.RootFolder); err == nil && fi.IsDir() {
		_ = folders.CreateLayout(cfg.RootFolder, newName)
	}

	fmt.Println()
	fmt.Printf(":: [ OK ] Computer name updated: %s -> %s\n", oldName, newName)

	// If background service is running, restart it to broadcast the new name immediately
	if service.IsRunning() {
		if sharing.IsElevated() {
			fmt.Print(":: Restarting background service to broadcast new name... ")
			_ = service.Stop()
			time.Sleep(1 * time.Second)
			_ = service.Start()
			fmt.Println("[ OK ]")
		} else {
			fmt.Println(":: Notice: Background service will broadcast the new name on its next restart.")
		}
	}

	fmt.Println()
	pause(reader)
}

// menuServiceControl toggles or restarts the background service
func menuServiceControl(reader *bufio.Reader) {
	if !ensureElevated(reader, "control the background service") {
		return
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                     BACKGROUND SERVICE CONTROL                         ")
	fmt.Println("------------------------------------------------------------------------")

	status, _ := service.QueryStatus()
	fmt.Printf("Service Name   : NetFilesSvc\n")
	fmt.Printf("Current Status : %s\n\n", status)

	if status == "Running" {
		fmt.Println("  [1] Stop service")
		fmt.Println("  [2] Restart service")
		fmt.Println("  [3] Back to main menu")
		fmt.Print("\nSelect option [1-3]: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print(":: Stopping NetFilesSvc... ")
			if err := service.Stop(); err != nil {
				fmt.Printf("[FAIL: %v]\n", err)
			} else {
				fmt.Println("[ OK ]")
			}
		case "2":
			fmt.Print(":: Restarting NetFilesSvc... ")
			_ = service.Stop()
			time.Sleep(1 * time.Second)
			if err := service.Start(); err != nil {
				fmt.Printf("[FAIL: %v]\n", err)
			} else {
				fmt.Println("[ OK ]")
			}
		default:
			return
		}
	} else if status == "Stopped" {
		fmt.Println("  [1] Start service")
		fmt.Println("  [2] Back to main menu")
		fmt.Print("\nSelect option [1-2]: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choice = strings.TrimSpace(choice)

		if choice == "1" {
			fmt.Print(":: Starting NetFilesSvc... ")
			if err := service.Start(); err != nil {
				fmt.Printf("[FAIL: %v]\n", err)
			} else {
				fmt.Println("[ OK ]")
			}
		} else {
			return
		}
	} else {
		fmt.Println(":: NetFilesSvc is not installed.")
		fmt.Print("Install and start service now? [Y/n]: ")
		ans, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		ans = strings.TrimSpace(ans)
		if ans == "" || strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") {
			exePath, _ := os.Executable()
			if err := service.Install(exePath); err != nil {
				fmt.Printf(":: [FAIL] Could not install service: %v\n", err)
			} else {
				_ = service.Start()
				fmt.Println(":: [ OK ] NetFilesSvc installed and started.")
			}
		}
	}

	fmt.Println()
	pause(reader)
}

// menuDeleteFolder removes the NetFiles directory and its files
func menuDeleteFolder(reader *bufio.Reader) {
	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	if cfg != nil && cfg.RootFolder != "" {
		rootFolder = cfg.RootFolder
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                       DELETE FOLDER & FILES                            ")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Printf("Target Folder: %s\n\n", rootFolder)

	if fi, err := os.Stat(rootFolder); err != nil || !fi.IsDir() {
		fmt.Println(":: Notice: Folder does not exist on disk.")
		fmt.Print("Do you want to create a clean empty NetFiles folder? [y/N]: ")
		ans, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		ans = strings.TrimSpace(ans)
		if strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes") {
			name := "PC-001"
			if cfg != nil && cfg.DisplayName != "" {
				name = cfg.DisplayName
			}
			_ = folders.CreateLayout(rootFolder, name)
			explorer.NotifyExplorer()
			fmt.Println(":: [ OK ] Clean folder created.")
		}
		pause(reader)
		return
	}

	fmt.Println(":: WARNING: This will permanently delete this folder and ALL files inside it:")
	fmt.Println("   - Public shared files")
	fmt.Println("   - Inbox received files")
	fmt.Println("   - Sent files and history")
	fmt.Println("   - Peer offline/online folders")
	fmt.Println()
	fmt.Print("Type 'DELETE' or 'y' to confirm permanent deletion (or Enter to cancel): ")

	confirm, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	confirm = strings.TrimSpace(confirm)
	if !strings.EqualFold(confirm, "delete") && !strings.EqualFold(confirm, "yes") && !strings.EqualFold(confirm, "y") {
		fmt.Println(":: Deletion cancelled. Folder was kept untouched.")
		pause(reader)
		return
	}

	// Stop service temporarily if running to release file locks
	wasRunning := service.IsRunning()
	if wasRunning {
		fmt.Print(":: Stopping background service to release file locks... ")
		_ = service.Stop()
		time.Sleep(500 * time.Millisecond)
		fmt.Println("[ OK ]")
	}

	fmt.Print(":: Deleting folder and all files... ")
	err = os.RemoveAll(rootFolder)
	if err != nil {
		fmt.Printf("[FAIL: %v]\n", err)
	} else {
		fmt.Println("[ OK ]")
	}

	explorer.NotifyExplorer()

	fmt.Println()
	fmt.Print("Do you want to re-create a clean empty NetFiles folder now? [y/N]: ")
	recreate, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	recreate = strings.TrimSpace(recreate)
	if strings.EqualFold(recreate, "y") || strings.EqualFold(recreate, "yes") {
		name := "PC-001"
		if cfg != nil && cfg.DisplayName != "" {
			name = cfg.DisplayName
		}
		_ = folders.CreateLayout(rootFolder, name)
		fmt.Println(":: [ OK ] Clean empty folder created.")
		if wasRunning {
			_ = service.Start()
			fmt.Println(":: [ OK ] NetFilesSvc restarted.")
		}
	} else if wasRunning {
		fmt.Println(":: Notice: Background service remains stopped since root folder was removed.")
	}

	fmt.Println()
	pause(reader)
}

// menuUninstall removes the service, firewall rules, and Explorer shortcuts
func menuUninstall(reader *bufio.Reader) {
	if !ensureElevated(reader, "uninstall NetFiles components") {
		return
	}

	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	if cfg != nil && cfg.RootFolder != "" {
		rootFolder = cfg.RootFolder
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                          UNINSTALL NETFILES                            ")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("This will remove:")
	fmt.Println("  1. NetFilesSvc Windows background service")
	fmt.Println("  2. NetFiles Windows Firewall rules (TCP/UDP)")
	fmt.Println("  3. Windows File Explorer navigation pane link & shortcuts")
	fmt.Println("  4. NetFiles system configuration")
	fmt.Println()
	fmt.Printf("Do you also want to delete the folder '%s' and all files? [y/N]: ", rootFolder)

	deleteFolderAns, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	deleteFolderAns = strings.TrimSpace(deleteFolderAns)
	deleteFolder := strings.EqualFold(deleteFolderAns, "y") || strings.EqualFold(deleteFolderAns, "yes")

	fmt.Print("Are you sure you want to proceed with uninstallation? [y/N]: ")
	confirmAns, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	confirmAns = strings.TrimSpace(confirmAns)
	if !strings.EqualFold(confirmAns, "y") && !strings.EqualFold(confirmAns, "yes") {
		fmt.Println(":: Uninstallation cancelled.")
		pause(reader)
		return
	}

	fmt.Println()
	ExecuteUninstall(deleteFolder)
	fmt.Println()
	pause(reader)
}

// menuRepairExplorerLinks re-registers the CLSID and desktop shortcuts
func menuRepairExplorerLinks(reader *bufio.Reader) {
	if !ensureElevated(reader, "repair Explorer registry links") {
		return
	}

	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	if cfg != nil && cfg.RootFolder != "" {
		rootFolder = cfg.RootFolder
	}
	exePath, _ := os.Executable()

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                      REPAIR EXPLORER LINKS                             ")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Print(":: Registering Navigation-Pane CLSID & Desktop shortcuts... ")
	err := explorer.Install(rootFolder, exePath)
	if err != nil {
		fmt.Printf("[WARN: %v]\n", err)
	} else {
		fmt.Println("[ OK ]")
	}

	fmt.Print(":: Notifying Windows File Explorer (SHChangeNotify)... ")
	explorer.NotifyExplorer()
	fmt.Println("[ OK ]")

	fmt.Println()
	fmt.Println(":: [ OK ] File Explorer links and Desktop shortcuts repaired.")
	fmt.Println()
	pause(reader)
}

// menuCustomSetup configures custom parameters before running setup
func menuCustomSetup(reader *bufio.Reader) {
	if !ensureElevated(reader, "configure and install NetFiles") {
		return
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("                           CUSTOM SETUP                                 ")
	fmt.Println("------------------------------------------------------------------------")

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Computer name
	fmt.Printf("Current or suggested name: %s\n", identity.FallbackName())
	fmt.Print("Enter computer name on LAN (or Enter to keep default): ")
	nameInput, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	nameInput = strings.TrimSpace(nameInput)
	if nameInput != "" {
		cfg.DisplayName = identity.ParseNameInput(nameInput, cfg.NamePrefix)
	}

	// Root folder
	fmt.Printf("Default folder: %s\n", config.DefaultRootFolder)
	fmt.Print("Enter custom root folder (or Enter to keep default): ")
	folderInput, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	folderInput = strings.TrimSpace(folderInput)
	if folderInput != "" {
		abs, err := filepath.Abs(folderInput)
		if err == nil {
			cfg.RootFolder = abs
		} else {
			cfg.RootFolder = folderInput
		}
	}

	_ = config.Save(cfg)
	fmt.Println()
	fmt.Println(":: Custom configuration saved. Starting setup...")
	time.Sleep(500 * time.Millisecond)

	RunSetup(nil)
}
