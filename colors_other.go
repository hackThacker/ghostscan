//go:build !windows

package main

// enableWindowsVT is a no-op on Linux and macOS — ANSI escape sequences
// work natively in those terminal emulators without any setup.
func enableWindowsVT() {}
