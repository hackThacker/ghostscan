package main

import (
	"embed"
	"fmt"
	"strings"
)

// Embedded default payload files — compiled into the binary so no payloads/
// folder is needed at runtime. Users can still override by providing a custom
// file path at the prompt.

//go:embed payloads/lfi.txt
var lfiDefault string

//go:embed payloads/or.txt
var orDefault string

//go:embed payloads/xss.txt
var xssDefault string

//go:embed payloads/sqli
var sqliFS embed.FS

// sqliFileMap maps the displayed SQL type names to their embedded filenames.
var sqliFileMap = map[string]string{
	"generic":    "payloads/sqli/generic.txt",
	"mysql":      "payloads/sqli/mysql.txt",
	"mssql":      "payloads/sqli/mssql",
	"oracle":     "payloads/sqli/oracle.txt",
	"postgresql": "payloads/sqli/postgresql.txt",
	"xor":        "payloads/sqli/xor.txt",
}

// sqliTypeNames is the ordered list for the picker menu.
var sqliTypeNames = []string{"generic", "mysql", "mssql", "oracle", "postgresql", "xor"}

// splitPayloadText splits a multiline string into non-empty trimmed lines.
// Handles both \r\n and \n line endings.
func splitPayloadText(text string) []string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	var out []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// loadEmbeddedPayload parses an embedded payload string into a slice of lines.
func loadEmbeddedPayload(embedded string) []string {
	return splitPayloadText(embedded)
}

// loadEmbeddedSQLi reads one SQLi payload file from the embedded FS by type name.
func loadEmbeddedSQLi(sqlType string) ([]string, error) {
	path, ok := sqliFileMap[sqlType]
	if !ok {
		return nil, fmt.Errorf("unknown SQL type: %s", sqlType)
	}
	data, err := sqliFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading embedded %s: %w", path, err)
	}
	return splitPayloadText(string(data)), nil
}

// listEmbeddedSQLiTypes returns the available embedded SQLi payload types.
func listEmbeddedSQLiTypes() []string {
	// Return a copy so the caller can't mutate the package-level slice.
	out := make([]string, len(sqliTypeNames))
	copy(out, sqliTypeNames)
	return out
}

// promptPayloadsOrDefault asks the user if they want default (embedded) payloads.
// If they press Enter or 'y', it returns the embedded payloads.
// If they press 'n', it prompts for a custom file path (like the original flow).
func promptPayloadsOrDefault(scanName string, embeddedData string) []string {
	yn := promptString(fmt.Sprintf("[?] Use default %s payloads? (Y/n): ", scanName))
	if strings.ToLower(yn) != "n" && embeddedData != "" {
		fmt.Println(colorGreen + "[i] Using embedded defaults." + colorReset)
		return loadEmbeddedPayload(embeddedData)
	}
	// Fall through to file prompt.
	welcome := fmt.Sprintf("Welcome to the %s Scanner! - hackthacker\n", scanName)
	return promptForPayloads(welcome)
}

// promptSQLiPayloads shows a menu of embedded SQLi types and optionally falls
// back to a custom file.
func promptSQLiPayloads() []string {
	fmt.Println(colorCyan + "[?] Select SQL injection payload type:" + colorReset)
	types := listEmbeddedSQLiTypes()
	for i, t := range types {
		first := strings.ToUpper(t[:1]) + t[1:]
		fmt.Printf("  %d) %s%s\n", i+1, first, colorReset)
	}
	fmt.Printf("  %d) Custom file\n", len(types)+1)
	choice := promptInt(fmt.Sprintf("%s[?] Enter choice (1-%d): %s", colorCyan, len(types)+1, colorReset), 1)

	if choice >= 1 && choice <= len(types) {
		selected := types[choice-1]
		fmt.Printf("%s[i] Loading embedded %s payloads...%s\n", colorGreen, selected, colorReset)
		payloads, err := loadEmbeddedSQLi(selected)
		if err != nil {
			fmt.Printf("%s[!] Error loading embedded payloads: %v. Using custom file prompt.%s\n", colorRed, err, colorReset)
			return promptForPayloads("Welcome to GhostScan SQL-Injector! - hackthacker\n")
		}
		if len(payloads) == 0 {
			fmt.Printf("%s[!] Empty embedded payloads. Using custom file.%s\n", colorYellow, colorReset)
			return promptForPayloads("Welcome to GhostScan SQL-Injector! - hackthacker\n")
		}
		fmt.Printf("%s[i] Loaded %d %s payloads.%s\n", colorGreen, len(payloads), selected, colorReset)
		return payloads
	}

	// Choice 7 or invalid → custom file
	fmt.Printf("%s[i] Custom file selected.%s\n", colorYellow, colorReset)
	return promptForPayloads("Welcome to GhostScan SQL-Injector! - hackthacker\n")
}

// loadCRLFPayloads returns either default hardcoded CRLF payloads (if user
// presses Enter / Y) or loads them from a custom file.
func loadCRLFPayloads(domain string) []string {
	yn := promptString("[?] Use default CRLF payloads? (Y/n): ")
	if strings.ToLower(yn) != "n" {
		fmt.Println(colorGreen + "[i] Using hardcoded CRLF defaults." + colorReset)
		return generateCRLFPayloads(domain)
	}
	// Custom file
	welcome := "Welcome to the CRLF Injection Testing Tool!\n"
	custom := promptForPayloads(welcome)
	// Replace {{Hostname}} placeholder in custom payloads too
	for i, p := range custom {
		custom[i] = strings.ReplaceAll(p, "{{Hostname}}", domain)
	}
	return custom
}
