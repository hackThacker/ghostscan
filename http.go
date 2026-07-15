package main

import (
	"crypto/tls"
	"net/http"
	"time"
)

// retryTransport wraps an http.RoundTripper and retries requests on
// 500/502/504 with exponential backoff, mirroring urllib3's Retry adapter.
type retryTransport struct {
	transport http.RoundTripper
	retries   int
	backoff   time.Duration
}

// newHTTPClient builds an *http.Client that mirrors the Python requests
// session with Retry adapter (3 retries, 0.3 backoff_factor, 500/502/504).
// TLS verification is disabled to mirror verify=False.
func newHTTPClient() *http.Client {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	base := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	rt := &retryTransport{
		transport: base,
		retries:   3,
		backoff:   300 * time.Millisecond,
	}
	return &http.Client{
		Transport: rt,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // equivalent to allow_redirects=False (overridden per-scanner)
		},
	}
}

// newHTTPClientNoRedirect builds a client that does NOT follow redirects,
// mirroring requests.get(..., allow_redirects=False) used by the CRLF scanner.
func newHTTPClientNoRedirect() *http.Client {
	c := newHTTPClient()
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	c.Timeout = 10 * time.Second
	return c
}

// RoundTrip implements http.RoundTripper with retry logic.
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error
	backoff := rt.backoff
	for attempt := 0; attempt <= rt.retries; attempt++ {
		resp, err := rt.transport.RoundTrip(req)
		if err == nil {
			if resp.StatusCode != 500 && resp.StatusCode != 502 && resp.StatusCode != 504 {
				return resp, nil
			}
			_ = resp.Body.Close()
			lastResp = resp
			lastErr = nil
		} else {
			lastErr = err
		}
		if attempt < rt.retries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return lastResp, lastErr
}

// newRequestWithUA creates a GET request with a random User-Agent header.
func newRequestWithUA(targetURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", randomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "close")
	return req, nil
}
