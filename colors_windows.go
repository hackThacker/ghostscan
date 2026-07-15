//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// enableWindowsVT enables virtual terminal processing on Windows consoles so
// ANSI escape sequences render correctly.
func enableWindowsVT() {
	const enableVirtualTerminal = 0x0004
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
	mode |= enableVirtualTerminal
	setConsoleMode.Call(handle, uintptr(mode))
}
