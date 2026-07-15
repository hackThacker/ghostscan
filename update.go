package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// runUpdate implements the in-place self-update flow from the Python source.
//   - GET https://api.github.com/repos/coffinxp/loxs/releases/latest
//   - compare tag_name vs VERSION via 4-component semver comparison
//   - if newer: download the first asset to ghostscan-update alongside the binary
//
// Because Go programs can't overwrite themselves while running in the same way
// Python's `with open(__file__, 'wb')` does (the file is mapped into memory),
// we follow the safe equivalence: download to a sibling file and prompt the
// user to manually replace the running binary. This mirrors the spirit of the
// Python "Update completed. Please restart GhostScan..!!" message exactly.
func runUpdate() {
	clearScreen()
	printUpdateBanner()
	fmt.Println(colorCyan + "Welcome to the GhostScan updater!" + colorReset)

	const (
		repoOwner = "hackthacker"
		repoName  = "ghostscan"
	)

	currentVersion := versionNormalize(VERSION)

	fmt.Printf("%s[i] Current version: %s%s\n", colorCyan, currentVersion, colorReset)
	fmt.Printf("%s[i] Checking for updates...%s\n", colorCyan, colorReset)

	release, err := fetchLatestRelease(repoOwner, repoName)
	if err != nil || release == nil {
		fmt.Printf("%s[!] Unable to check for updates: %v%s\n", colorYellow, err, colorReset)
		fmt.Print("\nPress Enter to return to the main menu...")
		stdinReader.ReadString('\n')
		return
	}

	latestRaw := release.TagName
	latestVer := versionNormalize(latestRaw)

	cmp, errCmp := versionCompare(currentVersion, latestVer)
	if errCmp != nil {
		fmt.Printf("%s[!] Error comparing versions: %v%s\n", colorRed, errCmp, colorReset)
		fmt.Print("\nPress Enter to return to the main menu...")
		stdinReader.ReadString('\n')
		return
	}

	if cmp >= 0 {
		fmt.Printf("%s[✓] You are already using the latest version.%s\n", colorGreen, colorReset)
		fmt.Printf("%s[i] Current version: %s%s\n", colorCyan, currentVersion, colorReset)
		fmt.Printf("%s[i] Latest version: %s%s\n", colorCyan, latestVer, colorReset)
		fmt.Print("\nPress Enter to return to the main menu...")
		stdinReader.ReadString('\n')
		return
	}

	fmt.Printf("%s[✓] New version available: %s%s\n", colorGreen, latestRaw, colorReset)
	if !promptYesNo("[?] Do you want to update? (y/n): ") {
		fmt.Println(colorYellow + "[i] Update cancelled.")
		return
	}

	assets, ok := release.Assets.([]interface{})
	if !ok || len(assets) == 0 {
		fmt.Printf("%s[!] No downloadable assets in this release.%s\n", colorRed, colorReset)
		return
	}
	first := assets[0].(map[string]interface{})
	url, _ := first["browser_download_url"].(string)
	if url == "" {
		fmt.Printf("%s[!] Missing browser_download_url on first asset.%s\n", colorRed, colorReset)
		return
	}

	dest := "ghostscan-update"
	fmt.Printf("%s[*] Downloading update from %s%s\n", colorCyan, url, colorReset)
	if err := downloadFile(url, dest); err != nil {
		fmt.Printf("%s[!] Update failed: %v%s\n", colorRed, err, colorReset)
		return
	}
	fmt.Printf("%s[✓] Update downloaded to %s. Please restart GhostScan..!!%s\n", colorGreen, dest, colorReset)
}

// releaseInfo maps the JSON returned by GitHub's releases/latest endpoint.
type releaseInfo struct {
	TagName string      `json:"tag_name"`
	Assets  interface{} `json:"assets"`
}

func fetchLatestRelease(owner, name string) (*releaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, name)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ghostscan-updater/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	var ri releaseInfo
	if err := json.Unmarshal(body, &ri); err != nil {
		return nil, err
	}
	return &ri, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// versionNormalize pads a version string to 4 dotted components ("1.2.3" → "1.2.3.0"),
// mirroring the Python normalize_version.
func versionNormalize(v string) string {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	for len(parts) < 4 {
		parts = append(parts, "0")
	}
	return strings.Join(parts, ".")
}

// versionCompare compares two 4-component versions. Returns -1, 0, +1.
// Each component is parsed as an integer; non-numeric components are compared
// lexicographically as a fallback.
func versionCompare(a, b string) (int, error) {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	for i := 0; i < 4; i++ {
		var ai, bi string
		if i < len(pa) {
			ai = pa[i]
		} else {
			ai = "0"
		}
		if i < len(pb) {
			bi = pb[i]
		} else {
			bi = "0"
		}
		// Try integer compare first.
		an, errA := parseIntSafe(ai)
		bn, errB := parseIntSafe(bi)
		if errA == nil && errB == nil {
			if an < bn {
				return -1, nil
			}
			if an > bn {
				return 1, nil
			}
			continue
		}
		if ai < bi {
			return -1, nil
		}
		if ai > bi {
			return 1, nil
		}
	}
	return 0, nil
}

// VERSION is the GhostScan build version (mirrors the original Python VERSION constant).
const VERSION = "v2.1.1"


