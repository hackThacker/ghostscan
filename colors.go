package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ANSI color constants mirroring the Python Color class and colorama Fore/Style.
const (
	colorBlue       = "\033[94m"
	colorGreen      = "\033[1;92m"
	colorYellow     = "\033[93m"
	colorRed        = "\033[91m"
	colorPurple     = "\033[95m"
	colorCyan       = "\033[96m"
	colorReset      = "\033[0m"
	colorOrange     = "\033[38;5;208m"
	colorBold       = "\033[1m"
	colorUnbold     = "\033[22m"
	colorItalic     = "\033[3m"
	colorUnitalic   = "\033[23m"
	colorWhite      = "\033[97m"
	colorLightBlack = "\033[90m"
	colorMagenta    = "\033[95m"
)

// clearScreen clears the terminal screen.
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// centerStr centers a string within the given width using space padding.
func centerStr(s string, width int) string {
	if len(s) >= width {
		return s
	}
	total := width - len(s)
	left := total / 2
	right := total - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// repeatStr returns n copies of s.
func repeatStr(s string, n int) string {
	return strings.Repeat(s, n)
}

// colorPrint prints a message with the given ANSI color prefix, then resets.
func colorPrint(color, msg string) {
	fmt.Print(color + msg + colorReset)
}

// colorPrintln prints a message with color and a trailing newline.
func colorPrintln(color, msg string) {
	fmt.Println(color + msg + colorReset)
}
