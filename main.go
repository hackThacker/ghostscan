package main

import (
	"fmt"
	"os"
)

// version is set at build time via -ldflags. Defaults to the constant below.
var version = versionDefault

const versionDefault = "v1.0.0"

// main is the entry point of GhostScan. It mirrors the Python original: show
// the menu, dispatch the user's selection, and loop until they choose "7" (Exit)
// or send Ctrl-C.
func main() {
	enableWindowsVT()
	disableInsecureWarnings()
	clearScreen()

	for {
		tryRunMainPass()
	}
}

// tryRunMainPass performs a single menu iteration. After a scanner runs
// (options 1-6) control returns here for the next iteration. Option "7" (Exit)
// and any unrecognised input call printExitMenu + os.Exit(0).
func tryRunMainPass() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(colorRed+"[!] Recovered from panic:", r, colorReset)
			pausePrompt()
		}
	}()

	displayMenu()
	choice := promptString(fmt.Sprintf("\n%s[?] Select an option (1-7): %s", colorCyan, colorReset))
	handleSelection(choice)
}

// disableInsecureWarnings is a no-op kept for symmetry with the Python's
// `urllib3.disable_warnings`.
func disableInsecureWarnings() {}

// handleSelection dispatches the user's menu choice.
//
//	1 → LFI    2 → OR     3 → SQLi
//	4 → XSS    5 → CRLF   6 → tool Update
//	7 / anything else → Exit
func handleSelection(selection string) {
	switch selection {
	case "1":
		clearScreen()
		runLFIScanner()
	case "2":
		clearScreen()
		runORScanner()
	case "3":
		clearScreen()
		runSQLScanner()
	case "4":
		clearScreen()
		runXSSScanner()
	case "5":
		clearScreen()
		runCRLFScanner()
	case "6":
		clearScreen()
		runUpdate()
		clearScreen()
	default:
		// "7" or any other input → exit
		printExitMenu()
		os.Exit(0)
	}
}
