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

# ── Fetch latest release tag ──────────────────────────────────────────────────
Write-Host 'Detecting latest release...'

try {
    # -UseBasicParsing is not a valid parameter on Invoke-RestMethod (PS5 or PS7);
    # it only exists on Invoke-WebRequest. JSON parsing here is always basic/native.
    $Release = Invoke-RestMethod `
        -Uri "https://api.github.com/repos/$Repo/releases/latest"
} catch {
    Write-Error "Failed to fetch release info: $_"
    exit 1
}

$LatestTag   = $Release.tag_name
$DownloadUrl = "https://github.com/$Repo/releases/download/$LatestTag/$AssetName"

Write-Host "Installing zeus $LatestTag (windows/$Arch)..."

# ── Create install directory ──────────────────────────────────────────────────
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
    Write-Host "Created directory: $InstallDir"
}

# ── Download to temp file ─────────────────────────────────────────────────────
$TempFile = Join-Path $env:TEMP "zeus_install_$(Get-Random).exe"

$ChecksumFile = Join-Path $env:TEMP "zeus_checksum_$(Get-Random).sha256"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing

    # Verify download succeeded and file is non-empty
    if (-not (Test-Path $TempFile) -or (Get-Item $TempFile).Length -eq 0) {
        Write-Error "Downloaded file is empty. Check that the release asset '$AssetName' exists."
        exit 1
    }

    # Verify SHA256 checksum against the sidecar published in the release
    Invoke-WebRequest -Uri "$DownloadUrl.sha256" -OutFile $ChecksumFile -UseBasicParsing
    $ExpectedHash = (Get-Content $ChecksumFile -Raw).Trim().ToUpper()
    $ActualHash   = (Get-FileHash $TempFile -Algorithm SHA256).Hash.ToUpper()

    if ($ActualHash -ne $ExpectedHash) {
        Write-Error "Checksum mismatch.`n  expected: $ExpectedHash`n  got:      $ActualHash"
        exit 1
    }
    Write-Host 'Checksum OK.'

    $Dest = Join-Path $InstallDir $ExeName
    Move-Item -Path $TempFile -Destination $Dest -Force
    Write-Host "Installed zeus to: $Dest"
} catch {
    if (Test-Path $TempFile)      { Remove-Item $TempFile      -Force -ErrorAction SilentlyContinue }
    if (Test-Path $ChecksumFile)  { Remove-Item $ChecksumFile  -Force -ErrorAction SilentlyContinue }
    Write-Error "Installation failed: $_"
    exit 1
} finally {
    if (Test-Path $ChecksumFile)  { Remove-Item $ChecksumFile  -Force -ErrorAction SilentlyContinue }
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
