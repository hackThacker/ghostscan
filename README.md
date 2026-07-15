# GhostScan

<p align="center">
  <img src="https://img.shields.io/github/v/release/hackThacker/ghostscan" alt="GitHub Release">
  <img src="https://img.shields.io/github/go-mod/go-version/hackThacker/ghostscan" alt="Go Version">
  <img src="https://img.shields.io/github/license/hackThacker/ghostscan" alt="License">
  <img src="https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-blue" alt="Platforms">
</p>

**GhostScan** is a multi-threaded security scanning tool written in Go. It detects common web vulnerabilities including **LFI**, **Open Redirect (OR)**, **SQL Injection (SQLi)**, **Cross-Site Scripting (XSS)**, and **CRLF Injection** — all from a single, self-contained binary with **zero runtime dependencies** (except Chrome for XSS/OR).

---

## Features

| Scanner | Type | Detection Method | Threaded |
|---|---|---|---|
| **LFI** | Local File Inclusion | HTTP 200 + body regex match | ✅ 0–10 workers |
| **OR** | Open Redirect | Chrome headless (chromedp) URL check | ✅ 2 workers, 3-context pool |
| **SQLi** | SQL Injection | Time-based (response ≥ 10 s) | ✅ 0–10 workers |
| **XSS** | Cross-Site Scripting | Chrome headless alert capture | ✅ 2 workers, 3-context pool |
| **CRLF** | CRLF Injection | HTTP header + body regex match | ✅ 1–10 workers |

- **50+ rotating User-Agents** to evade simple filtering
- **Retry HTTP client** — 3 retries with exponential backoff on 500/502/504
- **HTML report generator** — animated SVG, glitch-text CSS, vulnerability timeline
- **Embedded payloads** — 80,000+ payloads compiled into the binary (no external files needed)
- **Custom payloads supported** — answer `n` at the default-payload prompt to load your own file
- **Self-update** — checks for new releases from within the tool (option 6)

---

## Supported Platforms

| OS | Architecture | Binary |
|---|---|---|
| Linux | amd64 | `ghostscan` |
| Linux | arm64 | `ghostscan` |
| macOS | amd64 | `ghostscan` |
| macOS | arm64 (Apple Silicon) | `ghostscan` |
| Windows | amd64 | `ghostscan.exe` |
| Windows | arm64 | `ghostscan.exe` |

---

## Installation

### Download a pre-built binary (recommended)

Download the latest release from the [GitHub Releases page](https://github.com/hackThacker/ghostscan/releases/latest).

Available archives:

| Platform | Archive |
|---|---|
| Linux amd64 | `ghostscan_v2.1.1_linux_amd64.zip` |
| Linux arm64 | `ghostscan_v2.1.1_linux_arm64.zip` |
| macOS amd64 | `ghostscan_v2.1.1_darwin_amd64.zip` |
| macOS arm64 | `ghostscan_v2.1.1_darwin_arm64.zip` |
| Windows amd64 | `ghostscan_v2.1.1_windows_amd64.zip` |
| Windows arm64 | `ghostscan_v2.1.1_windows_arm64.zip` |

Extract and run:

```bash
# Linux / macOS
unzip ghostscan_v2.1.1_linux_amd64.zip
chmod +x ghostscan
./ghostscan

# Windows (PowerShell)
Expand-Archive ghostscan_v2.1.1_windows_amd64.zip .
.\ghostscan.exe
```

> Verify the download integrity using the checksums file:
>
> ```bash
> # Linux / macOS
> sha256sum -c ghostscan_v2.1.1_checksums.txt
>
> # Windows (PowerShell)
> Get-Content ghostscan_v2.1.1_checksums.txt | ForEach-Object {
>     $hash, $file = $_ -split ' ', 2
>     $actual = (Get-FileHash $file.Trim() -Algorithm SHA256).Hash.ToLower()
>     if ($actual -eq $hash.Replace('sha256:','')) { "OK: $file" } else { "MISMATCH: $file" }
> }
> ```

### Install via `go install`

```bash
go install github.com/hackthacker/ghostscan@latest
```

> Requires Go 1.22 or later.

---

## Build from Source

```bash
# Clone the repository
git clone https://github.com/hackThacker/ghostscan.git
cd ghostscan

# Build for your local OS/arch
go build -ldflags="-s -w" -o ghostscan .

# Run
./ghostscan          # Linux / macOS
.\ghostscan.exe      # Windows
```

### Cross-compile

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ghostscan .

# Linux arm64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ghostscan .

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ghostscan .

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ghostscan .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ghostscan.exe .

# Windows arm64
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o ghostscan.exe .
```

### Build all release archives (PowerShell)

```powershell
# From the project root on Windows:
.\scripts\build_release.ps1
```

This produces all 6 platform archives under `dist/` plus a SHA256 checksums file.

---

## Requirements

| Requirement | Details |
|---|---|
| Go | 1.22 or later (build from source) |
| Chrome / Chromium | Required **at runtime** for the XSS and OR scanners only |

The LFI, SQLi, and CRLF scanners use the standard HTTP client — no browser needed.

---

## Usage

```
   ____ _               _   ____
  / ___| |__   ___  ___| |_/ ___|  ___ __ _ _ __
 | |  _| '_ \ / _ \/ __| __\___ \ / __/ _ \| '_ \
 | |_| | | | | (_) \__ \ |_ ___) | (_| (_| | | | |
  \____|_| |_|\___/|___/\__|____/ \___\__,_|_| |_|

────────────────────────────────────────────────────────────────────────
┌────────────────────────────────────────────────────────────────────────┐
│1] LFi Scanner                                                          │
│2] OR Scanner                                                           │
│3] SQLi Scanner                                                         │
│4] XSS Scanner                                                          │
│5] CRLF Scanner                                                         │
│6] tool Update                                                          │
│7] Exit                                                                 │
└────────────────────────────────────────────────────────────────────────┘
────────────────────────────────────────────────────────────────────────
           Created by hackthacker | github.com/hackthacker
────────────────────────────────────────────────────────────────────────
        Select an option by entering the corresponding number:
────────────────────────────────────────────────────────────────────────
```

Each scanner follows the same interactive flow:

1. **URL input** — provide a file path containing one URL per line, or type a single URL
2. **Payloads** — press Enter to use embedded defaults, or type `n` for a custom payload file
3. **Thread count** — set concurrency level (0 = sequential, 1–10 = concurrent)
4. **Results** — real-time per-payload output in the terminal
5. **Summary** — total found / scanned / time taken
6. **HTML report** — optionally generate an animated HTML security report

---

## Scanner Reference

### 1] LFI Scanner

Detects Local File Inclusion by sending payloads appended to the URL and matching the response body against regex patterns.

- **Default success criteria**: `root:x:0:` (matches `/etc/passwd` leak)
- **Detection**: HTTP 200 + body matches any supplied regex
- **Payloads**: 70,466+ embedded paths (common system files, configs, logs)
- **Custom criteria**: Enter comma-separated regex patterns (e.g. `root:,/bin/bash`)

**Example flow:**

```
[?] Enter the path to the input file containing the URLs (or press Enter to input a single URL):
[?] Enter a single URL to scan: http://example.com/page.php?file=
[?] Use default LFI payloads? (Y/n):
[?] Enter the success criteria patterns (comma-separated, press Enter for 'root:x:0:'):
[?] Enter the number of concurrent threads (0-10, press Enter for 5): 5

→ Scanning URL: http://example.com/page.php?file=

[→] Scanning with payload: /etc/passwd
[✓] Vulnerable: http://example.com/page.php?file=%2Fetc%2Fpasswd - Response Time: 1.23 seconds
[→] Scanning with payload: /etc/shadow
[✗] Not Vulnerable: http://example.com/page.php?file=%2Fetc%2Fshadow - Response Time: 0.45 seconds

┌──────────────────────────────────┐
│→ Scanning finished.              │
│• Total found: 1                  │
│• Total scanned: 2                │
│• Time taken: 2 seconds           │
└──────────────────────────────────┘
```

### 2] OR Scanner

Detects Open Redirect vulnerabilities using headless Chrome.

- **Detection**: Chrome navigates to the test URL; if `window.location.href` resolves to a domain containing `google.com` → vulnerable
- **Requires**: Chrome or Chromium installed and in `$PATH`

### 3] SQLi Scanner (Time-Based)

Detects blind SQL injection by measuring HTTP response latency.

- **Detection**: Response time ≥ 10 seconds → vulnerable
- **Payload types**: Generic, MySQL, MSSQL, Oracle, PostgreSQL, XOR, or custom file
- **Optional cookie**: Supports authenticated scanning via a Cookie header

**Payload type menu:**

```
[?] Select SQL injection payload type:
  1) Generic
  2) Mysql
  3) Mssql
  4) Oracle
  5) Postgresql
  6) Xor
  7) Custom file
[?] Enter choice (1-7):
```

### 4] XSS Scanner

Detects reflected XSS by injecting payloads via headless Chrome and catching JavaScript `alert()` calls.

- **Detection**: `window.alert()` is intercepted; if triggered → vulnerable
- **Timeout**: Configurable wait per page (default 0.5 s)
- **URL mutation**: Substitutes each query-string parameter value with the payload; falls back to `?test=<payload>` if no params exist
- **Requires**: Chrome or Chromium installed and in `$PATH`

### 5] CRLF Scanner

Detects CRLF injection by analysing response headers and body using regex patterns.

- **Payloads**: 30 hardcoded payloads covering URL-encoding, UTF-8 overlong encoding, and double-encoding variants
- **Detection**: Matches `Location:` / `Set-Cookie:` header injection patterns in the HTTP response

---

## Configuration

All configuration is interactive — there are no config files or environment variables.

| Prompt | Scanner | Default |
|---|---|---|
| URL list file or single URL | All | — |
| Default payloads? (Y/n) | LFI, OR, XSS | Y (embedded) |
| SQL injection type | SQLi | — |
| Cookie header | SQLi | (none) |
| Success criteria regex | LFI | `root:x:0:` |
| Timeout per page (seconds) | XSS | `0.5` |
| Thread count (0–10) | LFI, SQLi | `5` |
| Thread count (1–10) | CRLF | `5` |
| Generate HTML report? (y/n) | All | — |
| Report filename | All | varies |

---

## HTML Reports

After each scan you can optionally generate a self-contained HTML report:

- Animated SVG scanner logo
- Glitch-text CSS effects
- Scan timeline
- Statistics grid (found / scanned / time / rate)
- Colour-coded list of all vulnerable URLs

Reports are saved to your current working directory.

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run checks: `go vet ./...` and `go build ./...`
5. Commit: `git commit -am 'feat: add my feature'`
6. Push: `git push origin feat/my-feature`
7. Open a Pull Request

**Guidelines:**
- Keep scanner logic in its own `<scanner>.go` file
- Add new payloads to the `payloads/` directory (they are embedded at build time)
- Update this README for any new user-facing features

---

## License

[MIT](LICENSE) © 2025 hackthacker