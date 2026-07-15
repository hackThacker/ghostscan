package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// runXSSScanner implements the Python XSS scanner using chromedp.
//
//   - pool of 3 headless Chrome contexts from a single allocator
//   - up to 2 concurrent workers
//   - generatePayloadUrls generates one or more URL variants per (url, payload):
//     substitute each query param value, or fragment params, or fall back to
//     `?test=payload` / `#payload` if neither exist
//   - navigate, wait up to `timeout` for an alert (we hook window.alert via JS
//     to capture the message into a global flag before navigation)
//   - if the alert global is set → vulnerable
//   - scan summary, HTML report
func runXSSScanner() {
	clearScreen()
	printXSSBanner()
	fmt.Println(colorGreen + "Welcome to the XSS Testing Tool!" + colorReset)

	urls := promptForURLsMSG("Welcome to the XSS Scanner!\n")
	payloads := promptPayloadsOrDefault("XSS", xssDefault)

	timeout := 0.5
	if raw := promptString("[?] Enter the timeout duration for each request (Press Enter for 0.5): "); raw != "" {
		if t, err := parseFloatSafe(raw); err == nil {
			timeout = t
		}
	}

	clearScreen()
	fmt.Printf("%s[i] Starting scan...\n\n%s", colorCyan, colorReset)

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-browser-side-navigation", true),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-notifications", true),
	)
	allocator, cancelAllocator := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAllocator()

	const poolSize = 3
	const workerCount = 2
	ctxSlots := make(chan context.Context, poolSize)
	cancelSlots := make(chan context.CancelFunc, poolSize)
	for i := 0; i < poolSize; i++ {
		ctx, cancel := chromedp.NewContext(allocator)
		ctxSlots <- ctx
		cancelSlots <- cancel
	}

	state := NewScanState()
	startTime := now()

	type job struct {
		url     string
		payload string
	}
	var wg sync.WaitGroup
	jobs := make(chan job)

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ctx := <-ctxSlots
				runXSSTest(ctx, j.url, j.payload, timeout, state)
				ctxSlots <- ctx
			}
		}()
	}

	for _, u := range urls {
		printURLBox(u)
		for _, p := range payloads {
			jobs <- job{url: u, payload: p}
		}
	}
	close(jobs)
	wg.Wait()

	_, vulnURLs, found, scanned := state.Snapshot()
	printScanSummaryBox(found, scanned, elapsedSeconds(startTime))
	offerHTMLReportDefault("Cross-Site Scripting (XSS)", state, startTime)
	_ = vulnURLs

	// Tear down contexts safely.
	close(ctxSlots)
	close(cancelSlots)
	go func() {
		for c := range cancelSlots {
			c()
		}
	}()
	time.Sleep(50 * time.Millisecond)
}

// runXSSTest injects an alert-capturing JS shim before navigation, then checks
// the captured flag after page load + a short wait. Mirrors the Python's
// WebDriverWait(EC.alert_is_present()) approach.
func runXSSTest(ctx context.Context, url, payload string, timeout float64, state *ScanState) {
	payloadURLs := generatePayloadUrls(url, payload)
	if len(payloadURLs) == 0 {
		return
	}
	for _, target := range payloadURLs {
		var alertText string
		var triggered bool
		// Inject a capture script before navigation.
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(`(window.__loxsAlerts=[]); (window.__loxsCaptureAlert=function(m){window.__loxsAlerts.push(String(m||''));}); window.alert = function(m){window.__loxsCaptureAlert(m);}; true;`, &triggered).Do(ctx)
			}),
			chromedp.Navigate(target),
			chromedp.Sleep(time.Duration(timeout*float64(time.Second))),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var arr []string
				if err := chromedp.Evaluate(`window.__loxsAlerts || []`, &arr).Do(ctx); err == nil {
					if len(arr) > 0 {
						alertText = arr[0]
						triggered = true
					}
				}
				return nil
			}),
		)
		if err != nil {
			// Unexpected alert presentation can throw; treat as a note and continue.
			errLower := strings.ToLower(err.Error())
			if strings.Contains(errLower, "alert") {
				fmt.Printf("%s[!] Unexpected alert: %v%s\n", colorYellow, err, colorReset)
				continue
			}
			fmt.Printf("%s[!] Error scanning %s: %v%s\n", colorRed, target, err, colorReset)
			state.AddScanned()
			continue
		}

		state.AddScanned()
		if triggered && alertText != "" {
			fmt.Printf("%s[✓]%s Vulnerable:%s %s%s%s - Alert Text: %s\n",
				colorGreen, colorCyan, colorGreen, target, colorCyan, colorReset, alertText)
			state.AddVulnerable(target)
		} else if triggered {
			fmt.Printf("%s[✓]%s Vulnerable:%s %s%s%s\n",
				colorGreen, colorCyan, colorGreen, target, colorCyan, colorReset)
			state.AddVulnerable(target)
		} else {
			fmt.Printf("%s[✗]%s Not Vulnerable:%s %s%s\n",
				colorRed, colorCyan, colorRed, target, colorReset)
		}
	}
}

func parseFloatSafe(s string) (float64, error) {
	// Minimal float parser to avoid strconv dependency proliferation.
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	} else if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	intPart := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			break
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("parse float")
		}
		intPart = intPart*10 + int(c-'0')
	}
	out := float64(intPart)
	if neg {
		out = -out
	}
	return out, nil
}
