<#
.SYNOPSIS
    Build GhostScan release archives for all supported platforms.

.DESCRIPTION
    Produces six ZIP archives under dist/ and a SHA256 checksums file.
    Runs go fmt and go vet before building.

    Archives follow the naming convention:
        ghostscan_<version>_<os>_<arch>.zip

    Windows archives contain: ghostscan.exe
    Linux/macOS archives contain: ghostscan

.EXAMPLE
    .\scripts\build_release.ps1

.EXAMPLE
    .\scripts\build_release.ps1 -Version v2.2.0
#>
[CmdletBinding()]
param(
    [string]$Version = "v2.1.0",
    [string]$DistDir = "dist"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Project  = "ghostscan"
$LdFlags  = "-s -w -X main.version=$Version"

$Platforms = @(
    @{ GOOS = "linux";   GOARCH = "amd64" },
    @{ GOOS = "linux";   GOARCH = "arm64" },
    @{ GOOS = "darwin";  GOARCH = "amd64" },
    @{ GOOS = "darwin";  GOARCH = "arm64" },
    @{ GOOS = "windows"; GOARCH = "amd64" },
    @{ GOOS = "windows"; GOARCH = "arm64" }
)

# ── Preflight ──────────────────────────────────────────────────────────────────

Write-Host "`n==> Running go fmt" -ForegroundColor Cyan
& go fmt ./...
if ($LASTEXITCODE -ne 0) { throw "go fmt failed" }

Write-Host "==> Running go vet" -ForegroundColor Cyan
& go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed" }

# ── Prepare dist/ ─────────────────────────────────────────────────────────────

if (Test-Path $DistDir) {
    Remove-Item -Recurse -Force $DistDir
}
New-Item -ItemType Directory -Path $DistDir | Out-Null

# ── Build each platform ───────────────────────────────────────────────────────

foreach ($plat in $Platforms) {
    $os   = $plat.GOOS
    $arch = $plat.GOARCH

    $ext      = if ($os -eq "windows") { ".exe" } else { "" }
    $binName  = "$Project$ext"
    $binPath  = Join-Path $DistDir $binName
    $archive  = "${Project}_${Version}_${os}_${arch}.zip"
    $archPath = Join-Path $DistDir $archive

    Write-Host "`n==> Building $archive" -ForegroundColor Green

    $env:GOOS   = $os
    $env:GOARCH = $arch
    & go build -ldflags $LdFlags -o $binPath .
    if ($LASTEXITCODE -ne 0) { throw "Build failed for $os/$arch" }

    # Compress: archive name never contains .exe; binary inside uses correct name
    Compress-Archive -Path $binPath -DestinationPath $archPath -Force
    Write-Host "    Created $archive" -ForegroundColor Gray

    Remove-Item $binPath
}

# ── Restore environment ────────────────────────────────────────────────────────

Remove-Item Env:\GOOS   -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

# ── SHA256 checksums ──────────────────────────────────────────────────────────

$ChecksumFile = Join-Path $DistDir "${Project}_${Version}_checksums.txt"
$lines = @()

Get-ChildItem -Path $DistDir -Filter "*.zip" | Sort-Object Name | ForEach-Object {
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower()
    $lines += "sha256:$hash  $($_.Name)"
    Write-Host "    sha256:$hash  $($_.Name)" -ForegroundColor Gray
}

$lines | Set-Content -Encoding UTF8 $ChecksumFile
Write-Host "`n==> Checksums written to $ChecksumFile" -ForegroundColor Cyan

# ── Summary ───────────────────────────────────────────────────────────────────

Write-Host "`n==> Release artifacts in $DistDir/:" -ForegroundColor Yellow
Get-ChildItem -Path $DistDir | ForEach-Object {
    $size = "{0,8:N0} KB" -f ($_.Length / 1KB)
    Write-Host "    $size  $($_.Name)"
}

Write-Host "`n==> Done. Ready to tag and publish:" -ForegroundColor Green
Write-Host "    git tag $Version && git push origin $Version"
Write-Host "    gh release create $Version dist/*.zip dist/*_checksums.txt --title `"GhostScan $Version`""
