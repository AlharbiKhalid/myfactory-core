<#
.SYNOPSIS
MyFactory installer for native Windows PowerShell.

.DESCRIPTION
Downloads a prebuilt myfactory.exe release asset from GitHub Releases,
verifies its SHA-256 checksum, and installs it into a user-owned directory.
Requires no administrator privileges and no Python/Go runtime.

.USAGE
  irm "https://raw.githubusercontent.com/AlharbiKhalid/myfactory-core/main/scripts/install.ps1" | iex

Configuration via environment variables:
  MYFACTORY_REPOSITORY   GitHub "owner/repo" (default: AlharbiKhalid/myfactory-core)
  MYFACTORY_VERSION      Release tag, e.g. v0.3.0 (default: latest)
  MYFACTORY_INSTALL_DIR  Install directory (default: %LOCALAPPDATA%\Programs\myfactory)
  MYFACTORY_UPDATE_PATH  Set to "1" to append the install dir to the user PATH.
#>

$ErrorActionPreference = "Stop"

function Write-Log([string]$Message) { Write-Host "[myfactory-install] $Message" }
function Fail([string]$Message) { Write-Error "[myfactory-install] ERROR: $Message"; exit 1 }

$Repo = if ($env:MYFACTORY_REPOSITORY) { $env:MYFACTORY_REPOSITORY } else { "AlharbiKhalid/myfactory-core" }
$InstallDir = if ($env:MYFACTORY_INSTALL_DIR) { $env:MYFACTORY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\myfactory" }

# --- Detect architecture ------------------------------------------------------

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { Fail "Unsupported CPU architecture: $env:PROCESSOR_ARCHITECTURE" }
}

# --- Resolve version ----------------------------------------------------------

if ($env:MYFACTORY_VERSION) {
    $Version = $env:MYFACTORY_VERSION
    if ($Version -notmatch '^v') { $Version = "v$Version" }
} else {
    Write-Log "Resolving latest release of $Repo"
    try {
        $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -Method Head -MaximumRedirection 5 -UseBasicParsing
        $finalUrl = $response.BaseResponse.ResponseUri.AbsoluteUri
        if (-not $finalUrl) { $finalUrl = $response.BaseResponse.RequestMessage.RequestUri.AbsoluteUri }
    } catch {
        Fail "Could not resolve the latest release. Set MYFACTORY_VERSION explicitly. ($_)"
    }
    $Version = ($finalUrl -split '/')[-1]
    if ($Version -notmatch '^v') { Fail "Could not parse latest release tag from: $finalUrl" }
}
Write-Log "Installing MyFactory $Version for windows/$arch"

# --- Download and verify ------------------------------------------------------

$asset = "myfactory_${Version}_windows_${arch}.zip"
$baseUrl = "https://github.com/$Repo/releases/download/$Version"
$tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("myfactory-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmpDir | Out-Null

try {
    Write-Log "Downloading $asset"
    try {
        Invoke-WebRequest -Uri "$baseUrl/$asset" -OutFile (Join-Path $tmpDir $asset) -UseBasicParsing
        Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile (Join-Path $tmpDir "checksums.txt") -UseBasicParsing
    } catch {
        Fail "Download failed (does release $Version ship windows/$arch?): $_"
    }

    Write-Log "Verifying SHA-256 checksum"
    $expectedLine = Get-Content (Join-Path $tmpDir "checksums.txt") | Where-Object { $_ -match [regex]::Escape($asset) } | Select-Object -First 1
    if (-not $expectedLine) { Fail "checksums.txt has no entry for $asset" }
    $expected = ($expectedLine -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -Path (Join-Path $tmpDir $asset)).Hash.ToLowerInvariant()
    if ($expected -ne $actual) { Fail "Checksum mismatch for $asset (expected $expected, got $actual). Aborting." }
    Write-Log "Checksum OK"

    # --- Extract and install --------------------------------------------------

    Expand-Archive -Path (Join-Path $tmpDir $asset) -DestinationPath $tmpDir -Force
    $exe = Join-Path $tmpDir "myfactory.exe"
    if (-not (Test-Path $exe)) { Fail "Archive did not contain myfactory.exe." }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $installPath = Join-Path $InstallDir "myfactory.exe"
    Move-Item -Force -Path $exe -Destination $installPath
    Write-Log "Installed: $installPath"

    # --- Verify and report ----------------------------------------------------

    & $installPath version
    if ($LASTEXITCODE -ne 0) { Fail "Installed binary failed to run." }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $onPath = ($userPath -split ';' | Where-Object { $_ -eq $InstallDir }).Count -gt 0
    if ($onPath) {
        Write-Log "Done. Open a new terminal and try: myfactory --help"
    } elseif ($env:MYFACTORY_UPDATE_PATH -eq "1") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
        Write-Log "Added $InstallDir to your user PATH. Open a new terminal and try: myfactory --help"
    } else {
        Write-Host ""
        Write-Host "$InstallDir is not on your PATH. Add it (user scope, no admin needed):"
        Write-Host ""
        Write-Host "    [Environment]::SetEnvironmentVariable('Path', `"$userPath;$InstallDir`", 'User')"
        Write-Host ""
        Write-Host "Or re-run this installer with MYFACTORY_UPDATE_PATH=1 to do it automatically."
        Write-Host "Then open a new terminal and run: myfactory --help"
    }
} finally {
    Remove-Item -Recurse -Force -Path $tmpDir -ErrorAction SilentlyContinue
}
