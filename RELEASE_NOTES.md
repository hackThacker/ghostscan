## GhostScan v2.1.2

Multi-threaded web vulnerability scanner for LFI, Open Redirect, SQL Injection, XSS, and CRLF Injection.

### What's included

- **LFI Scanner** — 70,000+ embedded payloads, configurable success-criteria regex
- **OR Scanner** — headless Chrome redirect detection via chromedp
- **SQLi Scanner** — time-based detection with 6 payload databases (Generic, MySQL, MSSQL, Oracle, PostgreSQL, XOR)
- **XSS Scanner** — headless Chrome `alert()` capture via chromedp
- **CRLF Scanner** — 30 payload variants covering URL-encoded, UTF-8 overlong, and double-encoded injections
- **HTML Reports** — animated scan reports saved locally
- **Self-contained binary** — all payloads embedded; no external files required at runtime

### Download

| Platform | Archive |
|---|---|
| Linux amd64 | `ghostscan_v2.1.2_linux_amd64.zip` |
| Linux arm64 | `ghostscan_v2.1.2_linux_arm64.zip` |
| macOS amd64 | `ghostscan_v2.1.2_darwin_amd64.zip` |
| macOS arm64 | `ghostscan_v2.1.2_darwin_arm64.zip` |
| Windows amd64 | `ghostscan_v2.1.2_windows_amd64.zip` |
| Windows arm64 | `ghostscan_v2.1.2_windows_arm64.zip` |

Verify your download using `ghostscan_v2.1.2_checksums.txt`.

### Requirements

- Chrome or Chromium must be installed **at runtime** for the XSS and OR scanners.
- All other scanners (LFI, SQLi, CRLF) require no browser.

### Install via go install

```bash
go install -v github.com/hackthacker/ghostscan@latest
```
