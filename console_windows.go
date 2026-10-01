// NetFiles — console initialization for Windows
// Created by Mohammed Salah
package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func initConsole() {
	// Enable UTF-8 code page for full unicode support
	_ = windows.SetConsoleOutputCP(65001)
	_ = windows.SetConsoleCP(65001)

	// Enable Virtual Terminal Processing for ANSI colors and escape sequences
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err == nil {
		_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
