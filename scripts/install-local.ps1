<#
.SYNOPSIS
Build MyFactory from the current local checkout and install myfactory.exe
(for development/testing on Windows). End users should use scripts/install.ps1,
which downloads prebuilt release binaries and needs no toolchain.

Requires: Go 1.23+.

.USAGE
  powershell -ExecutionPolicy Bypass -File scripts\install-local.ps1

Configuration:
  MYFACTORY_INSTALL_DIR  Install directory (default: %LOCALAPPDATA%\Programs\myfactory)
#>

$ErrorActionPreference = "Stop"

function Write-Log([string]$Message) { Write-Host "[myfactory-install-local] $Message" }
function Fail([string]$Message) { Write-Error "[myfactory-install-local] ERROR: $Message"; exit 1 }

$coreDir = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path (Join-Path $coreDir "cmd\myfactory\main.go"))) { Fail "run this from a myfactory-core checkout." }
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { Fail "Go 1.23+ is required to build from source. End users: use scripts/install.ps1 instead." }

$InstallDir = if ($env:MYFACTORY_INSTALL_DIR) { $env:MYFACTORY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\myfactory" }
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$commit = "unknown"
try { $commit = (git -C $coreDir rev-parse --short HEAD 2>$null).Trim() } catch {}
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$mod = "github.com/AlharbiKhalid/myfactory-core/internal/version"
$ldflags = "-X $mod.Version=dev-local -X $mod.Commit=$commit -X $mod.Date=$date"

$installPath = Join-Path $InstallDir "myfactory.exe"
Write-Log "Building myfactory.exe from $coreDir"
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags $ldflags -o $installPath (Join-Path $coreDir "cmd\myfactory")
if ($LASTEXITCODE -ne 0) { Fail "go build failed." }
Write-Log "Installed: $installPath"

& $installPath version

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$onPath = ($userPath -split ';' | Where-Object { $_ -eq $InstallDir }).Count -gt 0
if (-not $onPath) {
    Write-Host ""
    Write-Host "$InstallDir is not on your PATH. Add it (user scope, no admin needed):"
    Write-Host "    [Environment]::SetEnvironmentVariable('Path', `"$userPath;$InstallDir`", 'User')"
    Write-Host "Then open a new terminal and run: myfactory --help"
}
