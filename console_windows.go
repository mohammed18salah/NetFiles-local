// NetFiles — console initialization for Windows
// Created by Mohammed Salah
package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func initConsole() {
	// Enable Virtual Terminal Processing for ANSI colors without altering console font or codepage
	handle := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err == nil {
		_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
