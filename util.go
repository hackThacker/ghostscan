package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// now returns the current Unix timestamp in seconds.
func now() int64 {
	return time.Now().Unix()
}

// elapsedSeconds computes int(time.time() - start_time) from a Unix ts.
func elapsedSeconds(startUnix int64) int64 {
	return time.Now().Unix() - startUnix
}

// readLinesFile reads a text file and returns non-empty trimmed lines.
func readLinesFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

// fileExists checks if a path is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// promptForURLs reads the URL input (file or single URL) with retry.
func promptForURLs(welcomeMsg string) []string {
	for {
		urlInput := promptFilePath("[?] Enter the path to the input file containing the URLs (or press Enter to input a single URL): ")
		if urlInput != "" {
			if !fileExists(urlInput) {
				fmt.Printf("%s[!] Error reading input file: %s. Exception: File not found: %s%s\n", colorRed, urlInput, urlInput, colorReset)
				pausePrompt()
				clearScreen()
				colorPrintln(colorGreen, welcomeMsg)
				continue
			}
			urls, err := readLinesFile(urlInput)
			if err != nil {
				fmt.Printf("%s[!] Error reading input file: %s. Exception: %s%s\n", colorRed, urlInput, err, colorReset)
				pausePrompt()
				clearScreen()
				colorPrintln(colorGreen, welcomeMsg)
				continue
			}
			return urls
		}
		singleURL := promptString(colorCyan + "[?] Enter a single URL to scan: ")
		if singleURL != "" {
			return []string{singleURL}
		}
		fmt.Println(colorRed + "[!] You must provide either a file with URLs or a single URL.")
		pausePrompt()
		clearScreen()
		colorPrintln(colorGreen, welcomeMsg)
	}
}

// promptForPayloads reads the payload file path with retry.
func promptForPayloads(welcomeMsg string) []string {
	for {
		path := promptFilePath("[?] Enter the path to the payloads file: ")
		if !fileExists(path) {
			fmt.Printf("%s[!] Error reading payload file: %s. Exception: File not found: %s%s\n", colorRed, path, path, colorReset)
			pausePrompt()
			clearScreen()
			colorPrintln(colorGreen, welcomeMsg)
			continue
		}
		payloads, err := readLinesFile(path)
		if err != nil {
			fmt.Printf("%s[!] Error reading payload file: %s. Exception: %s%s\n", colorRed, path, err, colorReset)
			pausePrompt()
			clearScreen()
			colorPrintln(colorGreen, welcomeMsg)
			continue
		}
		return payloads
	}
}

// promptForValidFilePath keeps prompting for a valid file path (XSS scanner style).
func promptForValidFilePath(welcomeMsg, promptText string) string {
	for {
		path := strings.TrimSpace(promptFilePath(promptText))
		if path == "" {
			fmt.Println(colorRed + "[!] You must provide a file containing the payloads.")
			pausePrompt()
			clearScreen()
			colorPrintln(colorGreen, welcomeMsg)
			continue
		}
		if fileExists(path) {
			return path
		}
		fmt.Println(colorRed + "[!] Error reading the input file.")
		pausePrompt()
		clearScreen()
		colorPrintln(colorGreen, welcomeMsg)
	}
}
