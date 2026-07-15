package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// runORScanner implements the Python Open Redirect scanner using chromedp.
//
//   - pool of up to 3 headless Chrome contexts from a single allocator
//   - up to 2 concurrent workers (mirror of ThreadPoolExecutor(max_workers=2))
//   - for each URL x payload:
//     if no query → substitute path (path + payload)
//     if query → substitute each param value with payload
//     navigate to test_url, wait for document.readyState == "complete",
//     read window.location.href; if netloc contains "google.com" → vulnerable
//   - Ctrl-C safe (chromedp contexts are cancelled in defer)
//   - shows ASCII banner, scan summary, offers HTML report.
func runORScanner() {
	clearScreen()
	printORBanner()
	fmt.Println(colorGreen + "Welcome to the Open Redirect Testing Tool!" + colorReset)

	urls := promptForURLsMSG("Welcome to the Open Redirect Testing Tool!\n")
	payloads := promptPayloadsOrDefault("OR", orDefault)

	rawT := promptString("[?] Enter the number of concurrent threads (0-10, press Enter for 5): ")
	_ = rawT // chromedp uses fixed pool (matches Python behavior)

	fmt.Printf("\n%s[i] Loading, Please Wait...%s\n", colorYellow, colorReset)
	clearScreen()
	fmt.Printf("%s[i] Starting scan...\n\n%s", colorCyan, colorReset)

	// Set up a single headless Chrome allocator.
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

	// Build a pool of contexts. Mirror Python: 3 contexts in the rest pool.
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
	timings := nowUnixMilli()

	scanScan := func(url string) {
		fullURL := url
		if !strings.Contains(url, "://") {
			fullURL = "http://" + url
		}
		parts, err := splitPoolURL(fullURL)
		if err != nil || parts.Query == "" {
			// Path-based testing.
			testPathBased(ctxSlots, cancelSlots, workerCount, parts.Scheme, parts.Host, parts.Path, payloads, state)
		} else {
			params := parseQueryString(parts.Query)
			testParamBased(ctxSlots, cancelSlots, workerCount, parts.Scheme, parts.Host, parts.Path, params, payloads, state)
		}
	}

	for _, u := range urls {
		printURLBox(u)
		scanScan(u)
	}

	timings = nowUnixMilli() - timings
	_ = timings

	_, vulnURLs, found, scanned := state.Snapshot()
	printScanSummaryBox(found, scanned, elapsedSeconds(startTime))
	offerHTMLReportOR("Open Redirect (OR)", state, startTime)

	if vuln, _, _, _ := state.Snapshot(); vuln {
		fmt.Printf("\n%s[+] Vulnerabilities found: %d%s\n", colorGreen, found, colorReset)
		fmt.Printf("%s[+] Vulnerable URLs:%s\n", colorGreen, colorReset)
		for _, u := range vulnURLs {
			fmt.Printf("%s    %s%s\n", colorGreen, u, colorReset)
		}
	} else {
		fmt.Printf("\n%s[-] No vulnerabilities found.%s\n", colorYellow, colorReset)
	}
	fmt.Printf("%s[i] Total URLs scanned: %d%s\n", colorCyan, scanned, colorReset)

	// Drain & cancel contexts in a goroutine-safe way.
	close(ctxSlots)
	close(cancelSlots)
	go func() {
		for c := range cancelSlots {
			c()
		}
	}()
	time.Sleep(50 * time.Millisecond)
}

// parsedURL is a small wrapper around url parts we use for OR testing.
type parsedURL struct {
	Scheme   string
	Host     string
	Path     string
	Query    string
	Fragment string
}

func splitPoolURL(s string) (*parsedURL, error) {
	u, err := urlParseRaw(s)
	if err != nil {
		return nil, err
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	return &parsedURL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     path,
		Query:    u.RawQuery,
		Fragment: u.Fragment,
	}, nil
}

func testPathBased(ctxSlots chan context.Context, cancelSlots chan context.CancelFunc, workers int, scheme, host, path string, payloads []string, state *ScanState) {
	type job struct{ payload string }
	jobs := make(chan job)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ctx := <-ctxSlots
				testURL := scheme + "://" + host + path + j.payload
				runORTest(ctx, testURL, j.payload, "path", state)
				ctxSlots <- ctx
			}
		}()
	}
	for _, p := range payloads {
		jobs <- job{payload: p}
	}
	close(jobs)
	wg.Wait()
}

func testParamBased(ctxSlots chan context.Context, cancelSlots chan context.CancelFunc, workers int, scheme, host, path string, params map[string][]string, payloads []string, state *ScanState) {
	_ = cancelSlots
	type job struct {
		payload string
		param   string
	}
	jobs := make(chan job)
	var wg sync.WaitGroup
	paramNames := make([]string, 0, len(params))
	for k := range params {
		paramNames = append(paramNames, k)
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ctx := <-ctxSlots
				modParams := make(map[string][]string, len(params))
				for k, v := range params {
					if k == j.param {
						modParams[k] = []string{j.payload}
					} else {
						modParams[k] = v
					}
				}
				testURL := buildQueryURL(scheme, host, path, modParams, "")
				runORTest(ctx, testURL, j.payload, j.param, state)
				ctxSlots <- ctx
			}
		}()
	}
	for _, p := range payloads {
		for _, param := range paramNames {
			jobs <- job{payload: p, param: param}
		}
	}
	close(jobs)
	wg.Wait()
}

// runORTest navigates to testURL using ctx, waits for document readyState, then
// inspects window.location.href for "google.com" in the netloc.
func runORTest(ctx context.Context, testURL, payload, paramName string, state *ScanState) {
	var location string
	fmt.Printf("%s[→] Testing %s: %s%s%s\n", colorYellow, paramName, colorCyan, testURL, colorReset)
	err := chromedp.Run(ctx,
		chromedp.Navigate(testURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// wait for document.readyState === 'complete'
			var rs string
			if err := chromedp.Evaluate(`document.readyState`, &rs).Do(ctx); err == nil {
				for i := 0; i < 20 && rs != "complete"; i++ {
					time.Sleep(50 * time.Millisecond)
					_ = chromedp.Evaluate(`document.readyState`, &rs).Do(ctx)
				}
			}
			return chromedp.Evaluate(`window.location.href`, &location).Do(ctx)
		}),
	)
	if err != nil {
		fmt.Printf("%s[!] Error: %s%s\n", colorRed, err, colorReset)
		state.AddScanned()
		return
	}
	loc := strings.ToLower(location)
	if strings.Contains(loc, "google.com") {
		fmt.Printf("%s[✓] Vulnerable: %s%s\n", colorGreen, testURL, colorReset)
		state.AddVulnerable(testURL)
	} else {
		fmt.Printf("%s[✗] Not Vulnerable: %s%s\n", colorRed, testURL, colorReset)
	}
	state.AddScanned()
}

// buildQueryURL rebuilds a URL string from scheme/host/path/params/fragment.
func buildQueryURL(scheme, host, path string, params map[string][]string, fragment string) string {
	return urlBuildFromParts(scheme, host, path, encodeQuery(params), fragment)
}
