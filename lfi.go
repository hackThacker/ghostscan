package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// runLFIScanner implements the LFI (Local File Inclusion) scan mode.
// It mirrors the Python run_lfi_scanner function line-for-line:
//   - prompt for URL list (file or single URL)
//   - prompt for payload file
//   - prompt for success-criteria regex patterns (comma-separated, default "root:x:0:")
//   - prompt for thread count 0-10 (0 = sequential, default 5)
//   - for each URL x payload: HTTP GET, if status==200 and body matches
//     any success-criteria regex, mark vulnerable
//   - threaded worker pool bounded by `threads` (0 means sequential loop)
//   - prints per-payload scan lines, summary box, offers HTML report.
func runLFIScanner() {
	clearScreen()
	printLFIBanner()

	welcome := "Welcome to the LFI Testing Tool! - hackthacker\n"
	colorPrintln(colorGreen, welcome)

	urls := promptForURLs(welcome)
	payloads := promptPayloadsOrDefault("LFI", lfiDefault)

	scInput := promptString("[?] Enter the success criteria patterns (comma-separated, e.g: 'root:,admin:', press Enter for 'root:x:0:'): ")
	var successCriteria []*regexp.Regexp
	if scInput == "" {
		successCriteria = []*regexp.Regexp{regexp.MustCompile(`root:x:0:`)}
	} else {
		// comma-separated list of regex patterns
		parts := splitComma(scInput)
		for _, p := range parts {
			re, err := regexp.Compile(p)
			if err != nil {
				fmt.Printf("%s[!] Invalid regex %q: %v%s\n", colorRed, p, err, colorReset)
				return
			}
			successCriteria = append(successCriteria, re)
		}
	}

	threads := promptThreadCount()
	fmt.Printf("\n%s[i] Loading, Please Wait...%s\n", colorYellow, colorReset)
	clearScreen()
	fmt.Printf("%s[i] Starting scan...\n\n%s", colorCyan, colorReset)

	state := NewScanState()
	startTime := now()

	client := newHTTPClient()

	runOne := func(url, payload string) {
		encoded := urlEncodeAll(payload)
		target := url + encoded
		scanStart := time.Now()
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", randomUserAgent())
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "close")

		resp, err := client.Do(req)
		rt := roundTime(scanStart)
		if err != nil {
			fmt.Printf("%s[!] Error accessing %s: %s%s\n", colorRed, target, err, colorReset)
			return
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		vulnerable := false
		if resp.StatusCode == 200 {
			for _, re := range successCriteria {
				if re.Match(bodyBytes) {
					vulnerable = true
					break
				}
			}
		}
		fmt.Printf("%s[→] Scanning with payload: %s%s\n", colorYellow, payload, colorReset)
		if vulnerable {
			fmt.Printf("%s[✓]%s Vulnerable: %s%s%s - Response Time: %.2f seconds\n",
				colorGreen, colorCyan, colorGreen, target, colorCyan, rt.Seconds())
			state.AddVulnerable(target + " " + payload) // mirror Python's append + add-found
			_ = url + urlEncodeAll(payload)
		} else {
			fmt.Printf("%s[✗]%s Not Vulnerable: %s%s%s - Response Time: %.2f seconds\n",
				colorRed, colorCyan, colorRed, target, colorCyan, rt.Seconds())
		}
		state.AddScanned()
	}

	// 0 = sequential (Python: for url in urls -> for payload in payloads -> perform serially)
	if threads == 0 {
		for _, url := range urls {
			printURLBox(url)
			for _, payload := range payloads {
				runOne(url, payload)
			}
		}
	} else {
		// Worker pool bounded by `threads`.
		for _, url := range urls {
			printURLBox(url)
			runURLPool(url, payloads, threads, runOne)
		}
	}

	// Summary
	_, vulnURLs, found, scanned := state.Snapshot()
	printScanSummaryBox(found, scanned, elapsedSeconds(startTime))
	offerHTMLReport("Local File Inclusion (LFI)", state, startTime)

	if vuln, _, _, _ := state.Snapshot(); vuln {
		fmt.Printf("\n%s[+] Vulnerabilities found: %d%s\n", colorGreen, found, colorReset)
		fmt.Println(colorGreen + "[+] " + "Vulnerable URLs:")
		for _, u := range vulnURLs {
			fmt.Printf("%s    %s%s\n", colorGreen, u, colorReset)
		}
	} else {
		fmt.Printf("\n%s[-] No vulnerabilities found.%s\n", colorYellow, colorReset)
	}
	fmt.Printf("%s[i] Total URLs scanned: %d%s\n", colorCyan, scanned, colorReset)
}

// runURLPool fans payloads out across `workers` goroutines for one URL,
// calling fn(url, payload) for each. Used by LFI for the threaded case.
func runURLPool(url string, payloads []string, workers int, fn func(string, string)) {
	type job struct{ payload string }
	jobs := make(chan job)
	done := make(chan struct{})
	for w := 0; w < workers; w++ {
		go func() {
			for j := range jobs {
				fn(url, j.payload)
			}
			done <- struct{}{}
		}()
	}
	for _, p := range payloads {
		jobs <- job{payload: p}
	}
	close(jobs)
	for i := 0; i < workers; i++ {
		<-done
	}
}

// splitComma splits a comma-separated string and trims whitespace.
func splitComma(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == ',' {
			t := trimSpaces(cur)
			if t != "" {
				out = append(out, t)
			}
			cur = ""
			continue
		}
		cur += string(r)
	}
	t := trimSpaces(cur)
	if t != "" {
		out = append(out, t)
	}
	return out
}

func trimSpaces(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// urlEncodeAll URL-encodes the string the same way Python's quote(s, safe="")
// does: encode every byte except letters/digits/unreserved.
func urlEncodeAll(s string) string {
	out := make([]byte, 0, len(s)*3)
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		unreserved := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~'
		if unreserved {
			out = append(out, c)
		} else {
			out = append(out, '%', hex[c>>4], hex[c&0xF])
		}
	}
	return string(out)
}

// roundTime computes a duration between two timestamps with millisecond precision.
func roundTime(start time.Time) time.Duration {
	return time.Since(start)
}


