// NetFiles — offline LAN file-sharing tool
// Created by Mohammed Salah
//
// Main entry point: handles CLI commands and service startup.
package main

import (
	"fmt"
	"os"
	"strings"

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
		// No args: if running in a terminal, open TUI; otherwise start background service
		if isTerminal() {
			cmdStart(nil)
		} else {
			cmdStart(nil)
		}
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

func cmdStart(args []string) {
	flags := parseStartFlags(args)

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

	// Port check & auto-pick
	httpPort, err := ports.FindFreePort(cfg.HTTPPort)
	if err != nil {
		fmt.Printf("[FAIL] لا يمكن العثور على منفذ HTTP حر: %v\n", err)
		os.Exit(1)
	}
	if httpPort != cfg.HTTPPort {
		fmt.Printf("[INFO] المنفذ %d مشغول، يستخدم المنفذ %d بدلاً\n", cfg.HTTPPort, httpPort)
	}
	cfg.HTTPPort = httpPort

	udpPort, err := ports.FindFreePort(cfg.UDPPort)
	if err != nil {
		fmt.Printf("[FAIL] لا يمكن العثور على منفذ UDP حر: %v\n", err)
		os.Exit(1)
	}
	if udpPort != cfg.UDPPort {
		fmt.Printf("[INFO] المنفذ %d مشغول، يستخدم المنفذ %d بدلاً\n", cfg.UDPPort, udpPort)
	}
	cfg.UDPPort = udpPort

	// Identity: name prompt or CLI flag
	if flags.name != "" {
		cfg.DisplayName = identity.SanitizeName(flags.name)
	} else if flags.number > 0 {
		cfg.DisplayName = identity.NumberToName(flags.number, cfg.NamePrefix)
	}

	if cfg.DeviceID == "" {
		cfg.DeviceID = identity.GenerateDeviceID()
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
				fmt.Printf("[FAIL] خطأ في إدخال الاسم: %v\n", err)
				os.Exit(1)
			}
			cfg.DisplayName = name
		} else {
			// Fallback: use IP octets
			cfg.DisplayName = identity.FallbackName()
		}
	}

	// Create folder layout
	err = folders.CreateLayout(cfg.RootFolder, cfg.DisplayName)
	if err != nil {
		fmt.Printf("[FAIL] لا يمكن إنشاء المجلدات: %v\n", err)
		os.Exit(1)
	}

	// Add firewall rules (may need elevation)
	err = firewall.EnsureRules(cfg.HTTPPort, cfg.UDPPort)
	if err != nil {
		fmt.Printf("[WARN] لم يتم إضافة قواعد جدار الحماية: %v\n", err)
		fmt.Println("       قد تحتاج لتشغيل البرنامج كمسؤول")
	}

	// Register autostart if requested
	if flags.autostart {
		err = config.SetAutostart(true)
		if err != nil {
			fmt.Printf("[WARN] لم يتم تسجيل البدء التلقائي: %v\n", err)
		}
	}

	// Save config
	err = config.Save(cfg)
	if err != nil {
		fmt.Printf("[WARN] لم يتم حفظ الإعدادات: %v\n", err)
	}

	fmt.Println("[ OK ] NetFiles يعمل الآن")
	fmt.Printf("       الاسم: %s\n", cfg.DisplayName)
	fmt.Printf("       المعرف: %s\n", cfg.ShortID())
	fmt.Printf("       المنافذ: HTTP %d / UDP %d\n", cfg.HTTPPort, cfg.UDPPort)
	fmt.Printf("       المجلد: %s\n", cfg.RootFolder)

	// Start the HTTP server (blocks)
	srv := server.New(cfg)
	fmt.Println("[ OK ] الخادم جاهز للاستقبال")
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Printf("[FAIL] خطأ في الخادم: %v\n", err)
		os.Exit(1)
	}
}

func cmdStop() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] لم يتم العثور على إعدادات NetFiles")
		os.Exit(1)
	}

	// Remove firewall rules
	err = firewall.RemoveRules()
	if err != nil {
		fmt.Printf("[WARN] لم يتم إزالة قواعد جدار الحماية: %v\n", err)
	}

	// Remove autostart
	err = config.SetAutostart(false)
	if err != nil {
		fmt.Printf("[WARN] لم يتم إزالة البدء التلقائي: %v\n", err)
	}

	fmt.Println("[ OK ] تم إيقاف NetFiles")
	_ = cfg
}

func cmdUninstall(args []string) {
	force := false
	for _, a := range args {
		if strings.EqualFold(a, "-force") || strings.EqualFold(a, "--force") {
			force = true
		}
	}

	if !force && isTerminal() {
		fmt.Print("هل تريد حذف NetFiles بالكامل؟ سيتم حذف جميع الملفات. [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			fmt.Println("تم الإلغاء.")
			return
		}
	}

	// Stop first
	cmdStop()

	cfg, _ := config.Load()
	rootFolder := config.DefaultRootFolder
	if cfg != nil {
		rootFolder = cfg.RootFolder
	}

	// Remove folders
	err := os.RemoveAll(rootFolder)
	if err != nil {
		fmt.Printf("[WARN] لم يتم حذف المجلد %s: %v\n", rootFolder, err)
	} else {
		fmt.Printf("[ OK ] تم حذف %s\n", rootFolder)
	}

	// Remove config directory
	err = os.RemoveAll(config.ConfigDir)
	if err != nil {
		fmt.Printf("[WARN] لم يتم حذف الإعدادات: %v\n", err)
	} else {
		fmt.Println("[ OK ] تم حذف الإعدادات")
	}

	fmt.Println("[ OK ] تم إزالة NetFiles بالكامل")
}

func cmdStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] لم يتم العثور على إعدادات NetFiles. البرنامج غير مثبت.")
		os.Exit(1)
	}

	fmt.Println("=== حالة NetFiles ===")
	fmt.Printf("الاسم:    %s\n", cfg.DisplayName)
	fmt.Printf("المعرف:   %s\n", cfg.ShortID())
	fmt.Printf("المنافذ:  HTTP %d / UDP %d\n", cfg.HTTPPort, cfg.UDPPort)
	fmt.Printf("المجلد:   %s\n", cfg.RootFolder)

	ips := identity.GetLocalIPs()
	if len(ips) > 0 {
		fmt.Printf("العناوين: %s\n", strings.Join(ips, ", "))
	}

	// Check if running
	running := ports.IsPortInUse(cfg.HTTPPort)
	if running {
		fmt.Println("الحالة:   🟢 يعمل")
	} else {
		fmt.Println("الحالة:   🔴 متوقف")
	}
}

func cmdRename(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("[FAIL] لم يتم العثور على إعدادات NetFiles")
		os.Exit(1)
	}

	var newName string
	if len(args) > 0 {
		input := strings.Join(args, " ")
		newName = identity.ParseNameInput(input, cfg.NamePrefix)
	} else if isTerminal() {
		name, err := identity.PromptName(cfg.NamePrefix)
		if err != nil {
			fmt.Printf("[FAIL] خطأ: %v\n", err)
			os.Exit(1)
		}
		newName = name
	} else {
		fmt.Println("[FAIL] يجب تحديد الاسم: netfiles rename <name|number>")
		os.Exit(1)
	}

	oldName := cfg.DisplayName
	cfg.DisplayName = newName
	err = config.Save(cfg)
	if err != nil {
		fmt.Printf("[FAIL] لم يتم حفظ الاسم الجديد: %v\n", err)
		os.Exit(1)
	}

	// Rename the folder if it exists
	err = folders.RenamePeerFolder(cfg.RootFolder, oldName, newName)
	if err != nil {
		fmt.Printf("[WARN] لم يتم تغيير اسم المجلد: %v\n", err)
	}

	fmt.Printf("[ OK ] تم تغيير الاسم: %s -> %s\n", oldName, newName)
}

func cmdSetup(args []string) {
	// Stage (f) will add the pacman-style setup screen
	// For now, just run start
	cmdStart(args)
}

func cmdAbout() {
	fmt.Println()
	fmt.Println(`        /\            /\`)
	fmt.Println(`       /  \__      __/  \`)
	fmt.Println(`      / /\   \____/   /\ \`)
	fmt.Println(`     / /  \  (O)  (O)  /  \ \`)
	fmt.Println(`     \/    \____\/____/    \/`)
	fmt.Println(`            \  \/\/  /`)
	fmt.Println(`             \______/`)
	fmt.Println()
	fmt.Printf("  NetFiles v%s\n", Version)
	fmt.Printf("  %s\n", Credit)
	fmt.Println()
}

func printUsage() {
	fmt.Println()
	fmt.Println("NetFiles — أداة مشاركة الملفات عبر الشبكة المحلية")
	fmt.Printf("%s\n\n", Credit)
	fmt.Println("الاستخدام:")
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
