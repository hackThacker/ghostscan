package main

import "sync"

// ScanState holds the shared mutable scan results, protected by a mutex.
// This mirrors the Python scan_state dict used across scanners.
type ScanState struct {
	mu                 sync.Mutex
	VulnerabilityFound bool
	VulnerableURLs     []string
	TotalFound         int
	TotalScanned       int
}

// NewScanState creates an initialised ScanState.
func NewScanState() *ScanState {
	return &ScanState{
		VulnerableURLs: []string{},
	}
}

// Lock acquires the mutex.
func (s *ScanState) Lock() { s.mu.Lock() }

// Unlock releases the mutex.
func (s *ScanState) Unlock() { s.mu.Unlock() }

// AddVulnerable marks a URL as vulnerable and updates counters. The caller
// must NOT hold the lock.
func (s *ScanState) AddVulnerable(url string) {
	s.mu.Lock()
	s.VulnerabilityFound = true
	s.VulnerableURLs = append(s.VulnerableURLs, url)
	s.TotalFound++
	s.mu.Unlock()
}

// AddScanned records one more scanned item. The caller must NOT hold the lock.
func (s *ScanState) AddScanned() {
	s.mu.Lock()
	s.TotalScanned++
	s.mu.Unlock()
}

// AddScannedN records n more scanned items.
func (s *ScanState) AddScannedN(n int) {
	s.mu.Lock()
	s.TotalScanned += n
	s.mu.Unlock()
}

// Snapshot returns a consistent copy of the scan results.
func (s *ScanState) Snapshot() (vuln bool, urls []string, found, scanned int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	urlsCopy := make([]string, len(s.VulnerableURLs))
	copy(urlsCopy, s.VulnerableURLs)
	return s.VulnerabilityFound, urlsCopy, s.TotalFound, s.TotalScanned
}
