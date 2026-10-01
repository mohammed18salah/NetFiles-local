// NetFiles explorer package — Windows File Explorer integration
// Created by Mohammed Salah
//
// Implements:
// 1. Navigation Pane root entry via HKLM CLSID + Desktop\NameSpace + System.IsPinnedToNameSpaceTree=1
// 2. Network Shortcuts (%APPDATA%\Microsoft\Windows\Network Shortcuts) -> appears under This PC > Network locations
// 3. Public Desktop and User Desktop shortcuts
// 4. desktop.ini customization for the folder
// 5. SHChangeNotify for instant Explorer live refresh
package explorer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	CLSID_NetFiles = "{D45E2001-A482-4E90-951E-478310000001}"
	TagValue       = "NetFilesTool"
)

var (
	modshell32         = windows.NewLazySystemDLL("shell32.dll")
	procSHChangeNotify = modshell32.NewProc("SHChangeNotify")
)

// NotifyExplorer triggers a live refresh of all open File Explorer windows
func NotifyExplorer() {
	// SHCNE_ASSOCCHANGED = 0x08000000, SHCNF_IDLIST = 0x0000
	procSHChangeNotify.Call(0x08000000, 0, 0, 0)
}

// Install registers all Explorer integration points
func Install(folderPath, exePath string) error {
	var lastErr error

	// 1. Register Navigation-Pane CLSID in HKLM (visible to all users)
	if err := registerNavPaneCLSID(folderPath, exePath); err != nil {
		lastErr = fmt.Errorf("nav pane CLSID: %w", err)
	}

	// 2. Attach to Desktop NameSpace
	if err := registerDesktopNamespace(); err != nil {
		lastErr = fmt.Errorf("desktop namespace: %w", err)
	}

	// 3. Hide on Desktop surface (keeps it strictly in the left navigation pane)
	_ = hideOnDesktopSurface()

	// 4. Create desktop.ini with icon
	_ = setupDesktopIni(folderPath, exePath)

	// 5. Create Desktop shortcuts (Public Desktop & User Desktop)
	_ = createDesktopShortcuts(folderPath, exePath)

	// 6. Create Network Shortcut (shows under This PC > Network locations)
	_ = createNetworkShortcut(folderPath, exePath)

	// 7. Refresh Explorer
	NotifyExplorer()

	return lastErr
}

// Uninstall cleans up all Explorer integrations tagged NetFilesTool
func Uninstall(folderPath string) error {
	// 1. Remove Desktop NameSpace
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, fmt.Sprintf(`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Desktop\NameSpace\%s`, CLSID_NetFiles))

	// 2. Remove CLSID
	deleteKeyTree(registry.LOCAL_MACHINE, fmt.Sprintf(`SOFTWARE\Classes\CLSID\%s`, CLSID_NetFiles))

	// 3. Remove from HideDesktopIcons
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\HideDesktopIcons\NewStartPanel`, registry.SET_VALUE)
	if err == nil {
		_ = k.DeleteValue(CLSID_NetFiles)
		k.Close()
	}

	// 4. Delete shortcuts
	if pubDesktop := os.Getenv("PUBLIC"); pubDesktop != "" {
		_ = os.Remove(filepath.Join(pubDesktop, "Desktop", "NetFiles.lnk"))
	}
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		_ = os.Remove(filepath.Join(userProfile, "Desktop", "NetFiles.lnk"))
	}
	if appData := os.Getenv("APPDATA"); appData != "" {
		_ = os.Remove(filepath.Join(appData, `Microsoft\Windows\Network Shortcuts\NetFiles.lnk`))
	}

	// 5. Refresh Explorer
	NotifyExplorer()

	return nil
}

func registerNavPaneCLSID(folderPath, exePath string) error {
	basePath := fmt.Sprintf(`SOFTWARE\Classes\CLSID\%s`, CLSID_NetFiles)

	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, basePath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()

	_ = k.SetStringValue("", "NetFiles")
	_ = k.SetStringValue("InfoTip", "NetFiles Offline LAN Sharing")
	_ = k.SetStringValue("Tag", TagValue)
	_ = k.SetDWordValue("System.IsPinnedToNameSpaceTree", 1)
	_ = k.SetDWordValue("SortOrderIndex", 0x42)

	// DefaultIcon
	kIcon, _, err := registry.CreateKey(k, "DefaultIcon", registry.ALL_ACCESS)
	if err == nil {
		_ = kIcon.SetStringValue("", fmt.Sprintf(`"%s",0`, exePath))
		kIcon.Close()
	}

	// InProcServer32
	kInproc, _, err := registry.CreateKey(k, "InProcServer32", registry.ALL_ACCESS)
	if err == nil {
		_ = kInproc.SetStringValue("", `shell32.dll`)
		_ = kInproc.SetStringValue("ThreadingModel", "Apartment")
		kInproc.Close()
	}

	// Instance
	kInst, _, err := registry.CreateKey(k, "Instance", registry.ALL_ACCESS)
	if err == nil {
		_ = kInst.SetStringValue("CLSID", "{0E5AAE11-A675-4c65-8786-691C4DC797E0}")
		kProp, _, err := registry.CreateKey(kInst, "InitPropertyBag", registry.ALL_ACCESS)
		if err == nil {
			_ = kProp.SetDWordValue("Attributes", 0x11)
			_ = kProp.SetStringValue("TargetFolderPath", folderPath)
			kProp.Close()
		}
		kInst.Close()
	}

	// ShellFolder
	kShell, _, err := registry.CreateKey(k, "ShellFolder", registry.ALL_ACCESS)
	if err == nil {
		_ = kShell.SetDWordValue("Attributes", 0xF080004D)
		_ = kShell.SetDWordValue("FolderValueFlags", 0x28)
		kShell.Close()
	}

	return nil
}

func registerDesktopNamespace() error {
	keyPath := fmt.Sprintf(`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Desktop\NameSpace\%s`, CLSID_NetFiles)
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()
	_ = k.SetStringValue("", "NetFiles")
	_ = k.SetStringValue("Tag", TagValue)
	return nil
}

func hideOnDesktopSurface() error {
	keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\HideDesktopIcons\NewStartPanel`
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyPath, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer k.Close()
	return k.SetDWordValue(CLSID_NetFiles, 1)
}

func setupDesktopIni(folderPath, exePath string) error {
	iniPath := filepath.Join(folderPath, "desktop.ini")
	content := fmt.Sprintf("[.ShellClassInfo]\r\nIconResource=%s,0\r\nInfoTip=NetFiles Offline LAN Sharing\r\n[ViewState]\r\nMode=\r\nVid=\r\nFolderType=Generic\r\n", exePath)
	_ = os.WriteFile(iniPath, []byte(content), 0644)

	// Set system & hidden attributes on desktop.ini
	_ = runHidden("attrib", "+s", "+h", iniPath)
	// Set read-only attribute on folder (required by Windows to process desktop.ini)
	_ = runHidden("attrib", "+r", folderPath)
	return nil
}

func createDesktopShortcuts(folderPath, exePath string) error {
	psScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$targets = @(
    "$env:PUBLIC\Desktop\NetFiles.lnk",
    "$env:USERPROFILE\Desktop\NetFiles.lnk"
)
foreach ($target in $targets) {
    if ($target -and (Test-Path (Split-Path $target))) {
        $s = $WshShell.CreateShortcut($target)
        $s.TargetPath = "%s"
        $s.Description = "NetFiles Offline LAN Sharing"
        $s.IconLocation = "%s,0"
        $s.Save()
    }
}
`, folderPath, exePath)

	return runHidden("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
}

func createNetworkShortcut(folderPath, exePath string) error {
	psScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$netShortcuts = "$env:APPDATA\Microsoft\Windows\Network Shortcuts"
if (Test-Path $netShortcuts) {
    $s = $WshShell.CreateShortcut("$netShortcuts\NetFiles.lnk")
    $s.TargetPath = "%s"
    $s.Description = "NetFiles Offline LAN Sharing"
    $s.IconLocation = "%s,0"
    $s.Save()
}
`, folderPath, exePath)

	return runHidden("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
}

func deleteKeyTree(k registry.Key, path string) {
	subKey, err := registry.OpenKey(k, path, registry.ALL_ACCESS)
	if err != nil {
		return
	}
	subNames, _ := subKey.ReadSubKeyNames(-1)
	for _, name := range subNames {
		deleteKeyTree(subKey, name)
	}
	subKey.Close()
	_ = registry.DeleteKey(k, path)
}

func runHidden(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
