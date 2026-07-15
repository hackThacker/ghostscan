package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
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

// enableWindowsVT enables virtual terminal processing on Windows consoles so
// ANSI escape sequences render correctly. On non-Windows it is a no-op.
func enableWindowsVT() {
	if runtime.GOOS != "windows" {
		return
	}
	const enableVirtualTerminla = 0x0004
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleHandle := kernel32.NewProc("GetStdHandle")
	const stdOutputHandle = ^uintptr(0) - 11 + 1 // STD_OUTPUT_HANDLE = -11
	handle, _, _ := getConsoleHandle.Call(uintptr(stdOutputHandle))
	if handle == uintptr(0) {
		return
	}
	var mode uint32
	res, _, _ := getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if res == 0 {
		return
	}
	mode |= enableVirtualTerminla
	setConsoleMode.Call(handle, uintptr(mode))
}

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
