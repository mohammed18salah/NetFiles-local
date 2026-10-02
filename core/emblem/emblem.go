// NetFiles emblem package — Original clean terminal ASCII bat emblem
// Created by Mohammed Salah
package emblem

import (
	"fmt"
)

// Original clean terminal ASCII bat that renders perfectly on every terminal font
var OriginalEmblem = []string{
	`        /\            /\`,
	`       /  \__      __/  \`,
	`      / /\   \____/   /\ \`,
	`     / /  \  (O)  (O)  /  \ \`,
	`     \/    \____\/____/    \/`,
	`            \  \/\/  /`,
	`             \______/`,
}

// Batman style ASCII emblem
var BatmanEmblem = []string{
	`       _,    _   _    ,_`,
	`  .o888P     Y8o8Y     Y888o.`,
	` d88888      88888      88888b`,
	`d888888b_  _d88888b_  _d888888b`,
	`8888888888888888888888888888888`,
	`8888888888888888888888888888888`,
	`YJGS8P"Y888P"Y888P"Y888P"Y8888P`,
	` Y888   '8'   Y8P   '8'   888Y`,
	`  '8o          V          o8'`,
	`    '                     '`,
}

// Render prints the original terminal ASCII bat emblem
func Render(useBraille bool) {
	fmt.Println()
	for _, line := range OriginalEmblem {
		fmt.Printf("       \033[1;35m%s\033[0m\n", line)
	}
	fmt.Println()
}
