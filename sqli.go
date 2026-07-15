package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// runSQLScanner implements the time-based SQL injection scanner from the Python.
// For each URL x payload: GET request, measure response time, if >= 10s -> vulnerable.
// Optional cookie header. Threaded with 1-10 workers (0 = sequential).
//
// Mirrors run_sql_scanner exactly:
//   - prompts URL list, payload list, optional cookie, thread count 0-10 (default 5)
//   - for single URL the stripped payload is shown; for multi-URL list, the
//     stripped payload is also computed by removing every URL from each entry
//   - encoded-url-with-payload mirrors Python's `quote(stripped_payload, safe="")`
//   - response time >= 10 seconds == vulnerable
func runSQLScanner() {
	clearScreen()
	printSQLBanner()
	fmt.Println(colorGreen + "Welcome to the SQL Testing Tool!" + colorReset)

	urls := promptForURLsMSG("Welcome to the GhostScan SQL-Injector! - hackthacker\n")
	payloads := promptSQLiPayloads()

	rawCookie := promptString("[?] Enter the cookie to include in the GET request (press Enter if none): ")
	var cookie string
	if rawCookie != "" {
		cookie = rawCookie
	}

	rawT := promptString("[?] Enter the number of concurrent threads (0-10, press Enter for 5): ")
	threads := 5
	if rawT != "" {
		// python int() on non-empty string → error, but the Python source does int(... strip() or 5)
		// so the empty default is 5; non-empty that fails to parse just falls through to default in Python's flow.
		// Replicate the "0-10 inclusive" guard.
		n, err := parseIntSafe(rawT)
		if err == nil && n >= 0 && n <= 10 {
			threads = n
		}
	}

	fmt.Printf("\n%s[i] Loading, Please Wait...%s\n", colorYellow, colorReset)
	// NOTE: Python's main() prints "Loading, Please Wait...", then immediately clear_screen().
	clearScreen()
	fmt.Printf("%s[i] Starting scan...\n\n%s", colorCyan, colorReset)

	client := newHTTPClient()
	state := NewScanState()
	startTime := now()
	singleURLScan := len(urls) == 1

	type job struct {
		url, payload string
	}
	process := func(url, payload string) {
		// Mirror Python: f"{url}{payload}"
		urlWithPayload := url + payload
		// Set headers + cookie
		req, err := http.NewRequest("GET", urlWithPayload, nil)
		if err != nil {
			state.AddScanned()
			return
		}
		req.Header.Set("User-Agent", randomUserAgent())
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Connection", "close")
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		start := nowUnixMilli()
		resp, err := client.Do(req)
		elapsedMs := nowUnixMilli() - start
		var responseTime float64 = float64(elapsedMs) / 1000.0
		success := err == nil
		_ = success
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		_ = errMsg
		if resp != nil {
			_ = resp.Body.Close()
		}

		vulnerable := responseTime >= 10.0

		// Python's print differs for single-url vs multi-url; replicate properly.
		stripped := strings.Replace(urlWithPayload, url, "", 1)
		encodedStripped := urlEncodeAll(stripped)
		encodedURL := url + encodedStripped
		encodedURLWithPayload := encodedURL
		scanningPayload := stripped
		if !singleURLScan {
			listStripped := urlWithPayload
			for _, u := range urls {
				listStripped = strings.Replace(listStripped, u, "", 1)
			}
			encodedListStripped := urlEncodeAll(listStripped)
			// Python: encoded_url_with_payload = url_with_payload.replace(list_stripped, encoded_stripped)
			encodedURLWithPayload = strings.Replace(urlWithPayload, listStripped, encodedListStripped, 1)
			scanningPayload = listStripped
		}

		fmt.Printf("%s[→] Scanning with payload: %s%s\n", colorYellow, scanningPayload, colorReset)
		if vulnerable {
			fmt.Printf("%s[✓]%s Vulnerable: %s%s%s - Response Time: %.2f seconds\n",
				colorGreen, colorCyan, colorGreen, encodedURLWithPayload, colorCyan, responseTime)
			state.AddVulnerable(urlWithPayload)
		} else {
			fmt.Printf("%s[✗]%s Not Vulnerable: %s%s%s - Response Time: %.2f seconds\n",
				colorRed, colorCyan, colorRed, encodedURLWithPayload, colorCyan, responseTime)
		}
		state.AddScanned()
	}

	if threads == 0 {
		for _, url := range urls {
			printURLBox(url)
			for _, payload := range payloads {
				process(url, payload)
			}
		}
	} else {
		// Worker pool.
		jobs := make(chan job)
		done := make(chan struct{})
		for w := 0; w < threads; w++ {
			go func() {
				for j := range jobs {
					process(j.url, j.payload)
				}
				done <- struct{}{}
			}()
		}
		for _, url := range urls {
			printURLBox(url)
			for _, payload := range payloads {
				jobs <- job{url: url, payload: payload}
			}
		}
		close(jobs)
		for i := 0; i < threads; i++ {
			<-done
		}
	}

	_, vulnURLs, found, scanned := state.Snapshot()
	printScanSummaryBox(found, scanned, elapsedSeconds(startTime))
	// SQL scanner's save_results skips report unless user says yes; still uses offerHTMLReport.
	offerHTMLReport("Structured Query Language Injection (SQLi)", state, startTime)

	fmt.Printf("\n%sScanning finished. Findings saved above.%s\n", colorYellow, colorReset)
	_ = vulnURLs
}

// nowUnixMilli returns Unix milliseconds.
func nowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// parseIntSafe is a forgiving integer parser.
func parseIntSafe(s string) (int, error) {
	n := 0
	sign := 1
	i := 0
	if len(s) > 0 && s[0] == '-' {
		sign = -1
		i = 1
	} else if len(s) > 0 && s[0] == '+' {
		i = 1
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not an int")
		}
		n = n*10 + int(c-'0')
	}
	return sign * n, nil
}

// promptForURLsMSG wraps promptForURLs with a custom welcome message.
func promptForURLsMSG(welcomeMsg string) []string {
	return promptForURLs(welcomeMsg)
}
