# install.ps1 — installer for zeus on Windows
# Usage: irm https://raw.githubusercontent.com/smitt14ua/zeus/main/install.ps1 | iex
#Requires -Version 5
$ErrorActionPreference = 'Stop'

$Repo       = 'smitt14ua/zeus'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\Zeus'
$ExeName    = 'zeus.exe'

# ── Architecture detection ────────────────────────────────────────────────────
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

$AssetName = "zeus-windows-$Arch.exe"

# ── Fetch release metadata ────────────────────────────────────────────────────
Write-Host 'Detecting latest release...'

try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
} catch {
    Write-Error "Failed to fetch release info: $_"
    exit 1
}

# Locate the specific asset object — gives us the download URL and GitHub-computed digest
$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
if (-not $Asset) {
    Write-Error "Asset '$AssetName' not found in release $($Release.tag_name). Check that the release has finished building."
    exit 1
}

$DownloadUrl = $Asset.browser_download_url

# GitHub automatically computes SHA256 digests for all release assets (format: "sha256:<hex>")
$ExpectedHash = $null
if ($Asset.digest -and $Asset.digest -like 'sha256:*') {
    $ExpectedHash = ($Asset.digest -replace '^sha256:', '').ToUpper()
}

Write-Host "Installing zeus $($Release.tag_name) (windows/$Arch)..."

# ── Create install directory ──────────────────────────────────────────────────
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
    Write-Host "Created directory: $InstallDir"
}

# ── Download to temp file ─────────────────────────────────────────────────────
$TempFile = Join-Path $env:TEMP "zeus_install_$(Get-Random).exe"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing

    if (-not (Test-Path $TempFile) -or (Get-Item $TempFile).Length -eq 0) {
        Write-Error "Downloaded file is empty."
        exit 1
    }

    # ── Verify SHA256 digest ──────────────────────────────────────────────────
    if ($ExpectedHash) {
        $ActualHash = (Get-FileHash $TempFile -Algorithm SHA256).Hash.ToUpper()
        if ($ActualHash -ne $ExpectedHash) {
            Write-Error "Checksum mismatch.`n  expected: $ExpectedHash`n  got:      $ActualHash"
            exit 1
        }
        Write-Host 'Checksum OK.'
    } else {
        Write-Warning 'Digest not present in API response — skipping verification.'
    }

    $Dest = Join-Path $InstallDir $ExeName
    Move-Item -Path $TempFile -Destination $Dest -Force
    Write-Host "Installed zeus to: $Dest"
} catch {
    if (Test-Path $TempFile) { Remove-Item $TempFile -Force -ErrorAction SilentlyContinue }
    Write-Error "Installation failed: $_"
    exit 1
}

# ── Add install directory to user PATH if missing ─────────────────────────────
$UserPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
if ($null -eq $UserPath) { $UserPath = '' }

# Split on ';' for exact entry comparison — avoids wildcard mismatch if path
# contains '[' or ']' (e.g. a username with brackets in LOCALAPPDATA).
$PathEntries = $UserPath -split ';' | Where-Object { $_ -ne '' }
if ($PathEntries -notcontains $InstallDir) {
    $NewPath = ($PathEntries + $InstallDir) -join ';'
    [Environment]::SetEnvironmentVariable('PATH', $NewPath, 'User')
    Write-Host "Added $InstallDir to your user PATH."
    Write-Host 'Restart your terminal for the PATH change to take effect.'
} else {
    Write-Host "$InstallDir is already in your PATH."
}

Write-Host ''
Write-Host 'zeus installed successfully.'
Write-Host 'Run "zeus --version" to verify (restart terminal first if PATH was just updated).'
