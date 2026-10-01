// NetFiles emblem package — Batman-style Bat emblem constants & renderer
// Created by Mohammed Salah
package emblem

import (
	"fmt"
	"os"
	"strings"
)

// Classic perfectly symmetric half-block Batman emblem
var HalfBlockEmblem = []string{
	`               ▄▄       ▄▄               `,
	`             ▄█▀▀█▄   ▄█▀▀█▄             `,
	`     ▄██▄  ▄████████▄████████▄  ▄██▄     `,
	`    ████████████████▀████████████████    `,
	`   █████████████████▄█████████████████   `,
	`  █████████████████████████████████████  `,
	` ▄█████████████████████████████████████▄ `,
	` ▀█████████████████████████████████████▀ `,
	`  ▀██████████████▀  ▀██████████████▀     `,
	`    ▀█████████▀       ▀█████████▀        `,
	`      ▀█████▀           ▀█████▀          `,
	`        ▀█▀               ▀█▀            `,
}

// Braille representation for modern consoles (Windows Terminal)
var BrailleEmblem = []string{
	`               ⢀⡀       ⢀⡀               `,
	`             ⢀⣼⣿⣷⣄   ⣠⣾⣿⣧⡀             `,
	`     ⢠⣴⣶⡄ ⣸⣿⣿⣿⣿⣿⣦⣴⣿⣿⣿⣿⣿⣇ ⢠⣶⣦⡄     `,
	`    ⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷    `,
	`   ⣸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣇   `,
	`  ⣰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣆  `,
	` ⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡄ `,
	` ⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃ `,
	`  ⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟  ⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋  `,
	`    ⠙⢿⣿⣿⣿⣿⣿⣿⣿⠟       ⠻⣿⣿⣿⣿⣿⣿⣿⡿⠋    `,
	`      ⠙⢿⣿⣿⣿⠟           ⠻⣿⣿⣿⡿⠋      `,
	`        ⠙⠟               ⠻⠋        `,
}

// Render prints the emblem centered with purple glow and light gray/white core
func Render(useBraille bool) {
	lines := HalfBlockEmblem
	if useBraille || (os.Getenv("WT_SESSION") != "" && !strings.Contains(strings.Join(os.Args, " "), "--blocks")) {
		lines = BrailleEmblem
	}

	fmt.Println()
	for _, line := range lines {
		// Purple glow + bright white/gray emblem
		fmt.Printf("       \033[1;35m%s\033[0m\n", line)
	}
	fmt.Println()
}
