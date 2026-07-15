package main

import "fmt"

// displayMenu prints the main menu banner, options box, authors, and prompt.
func displayMenu() {
	title := "\n" +
		"   ____ _               _   ____                 \n" +
		"  / ___| |__   ___  ___| |_/ ___|  ___ __ _ _ __  \n" +
		" | |  _| '_ \\ / _ \\/ __| __\\___ \\ / __/ _ \\| '_ \\ \n" +
		" | |_| | | | | (_) \\__ \\ |_ ___) | (_| (_| | | | |\n" +
		"  \\____|_| |_|\\___/|___/\\__|____/ \\___\\__,_|_| |_|\n"
	colorPrintln(colorOrange+colorBold, centerStr(title, 72))
	fmt.Println(colorWhite + colorBold + repeatStr("─", 72))
	borderColor := colorCyan + colorBold
	optionColor := colorWhite + colorBold

	fmt.Println(borderColor + "┌" + repeatStr("─", 72) + "┐")

	options := []string{
		"1] LFi Scanner",
		"2] OR Scanner",
		"3] SQLi Scanner",
		"4] XSS Scanner",
		"5] CRLF Scanner",
		"6] tool Update",
		"7] Exit",
	}

	for _, option := range options {
		fmt.Println(borderColor + "│" + optionColor + leftPadTo(option, 72) + borderColor + "│")
	}

	fmt.Println(borderColor + "└" + repeatStr("─", 72) + "┘")
	authors := fmt.Sprintf("Created by hackthacker | github.com/hackthacker | Version: %s", version)
	instructions := "Select an option by entering the corresponding number:"

	fmt.Println(colorWhite + colorBold + repeatStr("─", 72))
	fmt.Println(colorWhite + colorBold + centerStr(authors, 72))
	fmt.Println(colorWhite + colorBold + repeatStr("─", 72))
	fmt.Println(colorWhite + colorBold + centerStr(instructions, 72))
	fmt.Println(colorWhite + colorBold + repeatStr("─", 72))
}

// leftPadTo right-justifies s within width by trailing spaces (equivalent to ljust).
func leftPadTo(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + repeatStr(" ", width-len(s))
}

// printExitMenu prints the green EXIT ASCII panel and credit line.
func printExitMenu() {
	clearScreen()
	panel := "\n" +
		"         ______               ______              \n" +
		"        |   __ \\.--.--.-----.|   __ \\.--.--.-----.\n" +
		"        |   __ <|  |  |  -__||   __ <|  |  |  -__|\n" +
		"        |______/|___  |_____||______/|___  |_____|\n" +
		"                |_____|              |_____|      \n" +
		"\n" +
		"  Created by hackthacker | github.com/hackthacker\n"
	colorPrintln(colorGreen+colorBold, panel)
	fmt.Print(colorRed + "\n\nSession Off..\n")
}

// printLFIBanner prints the LFI scanner ASCII panel.
func printLFIBanner() {
	panel := "\n" +
		"    __    __________   _____                                 \n" +
		"   / /   / ____/  _/  / ___/_________ _____  ____  ___  _____\n" +
		"  / /   / /_   / /    \\__ \\/ ___/ __ `/ __ \\/ __ \\/ _ \\/ ___/\n" +
		" / /___/ __/ _/ /    ___/ / /__/ /_/ / / / / / / /  __/ /    \n" +
		"/_____/_/   /___/   /____/\\___/\\__,_/_/ /_/_/ /_/\\___/_/     \n" +
		"                                                        \n"
	printGreenPanel(panel)
	colorPrintln(colorGreen, "Welcome to the LFI Testing Tool!\n")
}

// printSQLBanner prints the SQLi scanner ASCII panel.
func printSQLBanner() {
	panel := "\n" +
		"                                                       \n" +
		"                  ___                                         \n" +
		"      _________ _/ (_)  ______________ _____  ____  ___  _____\n" +
		"    / ___/ __ `/ / /  / ___/ ___/ __ `/ __ \\/ __ \\/ _ \\/ ___/\n" +
		"   (__  ) /_/ / / /  (__  ) /__/ /_/ / / / / / / /  __/ /    \n" +
		"  /____/\\__, /_/_/  /____/\\___/\\__,_/_/ /_/_/ /_/\\___/_/     \n" +
		"            /_/                                                \n"
	printGreenPanel(panel)
	colorPrintln(colorGreen, "Welcome to the SQL Testing Tool!\n")
}

// printXSSBanner prints the XSS scanner ASCII panel.
func printXSSBanner() {
	panel := "\n" +
		"    _  __________  ____________   _  ___  __________\n" +
		"   | |/_/ __/ __/ / __/ ___/ _ | / |/ / |/ / __/ _  |\n" +
		"   >  <_\\ \\_\\ \\  _\\ \\/ /__/ __ |/    /    / _// , _/\n" +
		"  /_/|_/___/___/ /___/\\___/_/ |_/_/|_/_/|_/___/_/|_|  \n"
	printGreenPanel(panel)
	colorPrintln(colorGreen, "Welcome to the XSS Testing Tool!\n")
}

// printORBanner prints the Open Redirect scanner ASCII panel.
func printORBanner() {
	panel := "\n" +
		"   ____  ___    ____________   _  ___  __________\n" +
		"  / __ \\/ _ \\  / __/ ___/ _ | / |/ / |/ / __/ _  |\n" +
		" / /_/ / , _/ _\\ \\/ /__/ __ |/    /    / _// , _/\n" +
		"/____//_/|_| /___/\\___/_/ |_/_/|_/_/|_/___/_/|_| \n" +
		"            \n"
	printGreenPanel(panel)
	colorPrintln(colorGreen, "Welcome to the Open Redirect Testing Tool!\n")
}

// printCRLFBanner prints the CRLF scanner ASCII panel.
func printCRLFBanner() {
	panel := "\n" +
		"   __________  __    ______\n" +
		"  / ____/ __ \\/ /   / ____/  ______________ _____  ____  ___  _____\n" +
		" / /   / /_/ / /   / /_     / ___/ ___/ __ `/ __ \\/ __ \\/ _ \\/ ___/\n" +
		"/ /___/ _, _/ /___/ __/    (__  ) /__/ /_/ / / / / / / /  __/ /\n" +
		"\\____/_/ |_/_____/_/      /____/\\___/\\__,_/_/ /_/_/ /_/\\___/_/\n" +
		"        \n"
	printGreenPanel(panel)
	colorPrintln(colorGreen, "Welcome to the CRLF Injection Testing Tool!\n")
}

// printUpdateBanner prints the Update ASCII panel.
func printUpdateBanner() {
	panel := "\n" +
		"██    ██ ███████ ███████  ███████ ████████ ███████ \n" +
		"██    ██ ██   ██ ██    ██ ██   ██    ██    ██      \n" +
		"██    ██ ███████ ██    ██ ███████    ██    █████   \n" +
		"██    ██ ██      ██    ██ ██   ██    ██    ██      \n" +
		"████████ ██      ███████  ██   ██    ██    ███████ \n"
	printGreenPanel(panel)
	fmt.Print(colorCyan + "Welcome to the GhostScan updater!\n")
}

// printGreenPanel prints a bordered green ASCII panel.
func printGreenPanel(panel string) {
	fmt.Println(colorGreen + colorBold + panel + colorReset)
}

// printScanSummaryBox prints a yellow-bordered summary box like the Python's
// print_scan_summary used across the scanners.
func printScanSummaryBox(totalFound, totalScanned int, timeTaken int64) {
	summary := []string{
		"→ Scanning finished.",
		fmt.Sprintf("• Total found: %s%d%s", colorGreen, totalFound, colorYellow),
		fmt.Sprintf("• Total scanned: %d", totalScanned),
		fmt.Sprintf("• Time taken: %d seconds", timeTaken),
	}
	maxLen := 0
	for _, line := range summary {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	borderTop := "┌" + repeatStr("─", maxLen+2) + "┐"
	borderBot := "└" + repeatStr("─", maxLen+2) + "┘"

	fmt.Println(colorYellow + "\n" + borderTop)
	for _, line := range summary {
		padding := maxLen - len(line)
		fmt.Printf("%s│ %s%s │%s", colorYellow, line, repeatStr(" ", padding), colorYellow)
		fmt.Print(colorReset + "\n")
	}
	fmt.Print(colorYellow + borderBot + colorReset + "\n")
}

// printURLBox prints the per-URL scanning box, like the Python "━ Scanning URL:" box.
func printURLBox(url string) {
	boxContent := " → Scanning URL: " + url + " "
	boxWidth := max(len(boxContent)+2, 40)
	fmt.Printf("\n%s┌%s┐\n", colorYellow, repeatStr("─", boxWidth-2))
	fmt.Printf("%s│%s│\n", colorYellow, centerStr(boxContent, boxWidth-2))
	fmt.Printf("%s└%s┘\n\n", colorYellow, repeatStr("─", boxWidth-2))
}
