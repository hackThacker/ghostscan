package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// stdinReader is the shared buffered reader for interactive prompts.
var stdinReader = bufio.NewReader(os.Stdin)

// promptString reads a line from stdin after printing the given prompt.
// Mirrors the Python prompt_toolkit usage (without tab completion).
func promptString(promptText string) string {
	fmt.Print(promptText)
	line, _ := stdinReader.ReadString('\n')
	return strings.TrimSpace(line)
}

// promptFilePath reads a file path from stdin after printing the prompt.
// Equivalent to get_file_path in the Python source.
func promptFilePath(promptText string) string {
	return promptString(promptText)
}

// promptInt reads an integer from stdin; returns the default value when the
// input is empty or cannot be parsed.
func promptInt(promptText string, def int) int {
	raw := promptString(promptText)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

// promptThreadCount reads a thread count 0-10 with default 5 (used by LFI/SQLi).
func promptThreadCount() int {
	raw := promptString("[?] Enter the number of concurrent threads (0-10, press Enter for 5): ")
	if raw == "" {
		return 5
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > 10 {
		return 5
	}
	return n
}

// promptThreadCount1To10 reads a thread count 1-10 with default 5 (used by CRLF).
func promptThreadCount1To10() int {
	raw := promptString("[?] Enter the number of concurrent threads (1-10, press Enter for 5): ")
	if raw == "" {
		return 5
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 10 {
		return 5
	}
	return n
}

// promptYesNo prompts for a y/n answer.
func promptYesNo(promptText string) bool {
	raw := promptString(promptText)
	return strings.ToLower(raw) == "y"
}

// pausePrompt prints a "Press Enter to continue" prompt and waits for input.
func pausePrompt() {
	promptString("\n[i] Press Enter to continue...")
}
