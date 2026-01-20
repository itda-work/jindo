#Requires -Version 5.1
<#
.SYNOPSIS
    jd - Claude Code configuration manager installation script for Windows

.DESCRIPTION
    Downloads and installs jd to the specified directory.

.EXAMPLE
    # Install latest version
    irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex

.EXAMPLE
    # Install to custom directory
    $env:JD_INSTALL_DIR = "C:\tools"; irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex

.EXAMPLE
    # Install specific version
    $env:VERSION = "v0.1.0"; irm https://cdn.jsdelivr.net/gh/itda-work/itda-jindo@main/install.ps1 | iex

.NOTES
    Installs to: $env:USERPROFILE\.local\bin\jd.exe
#>

$ErrorActionPreference = "Stop"

$Repo = "itda-work/itda-jindo"
$BinaryName = "jd"
$DefaultInstallDir = Join-Path $env:USERPROFILE ".local\bin"

function Write-Info {
    param([string]$Message)
    Write-Host "INFO: " -ForegroundColor Blue -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "SUCCESS: " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warning {
    param([string]$Message)
    Write-Host "WARNING: " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error-Exit {
    param([string]$Message)
    Write-Host "ERROR: " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Error-Exit "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Error-Exit "Failed to get latest version: $_"
    }
}

function Install-Jd {
    Write-Host ""
    Write-Host "  +--------------------------------------+"
    Write-Host "  |                                      |"
    Write-Host "  |   jd - Claude Code Config Manager    |"
    Write-Host "  |                                      |"
    Write-Host "  +--------------------------------------+"
    Write-Host ""

    Write-Info "Installing jd..."

    # Detect architecture
    $arch = Get-Architecture
    Write-Info "Detected architecture: windows/$arch"

    # Get version
    $version = $env:VERSION
    if (-not $version) {
        Write-Info "Fetching latest version..."
        $version = Get-LatestVersion
    }
    Write-Info "Version: $version"

    # Determine install directory
    $installDir = if ($env:JD_INSTALL_DIR) { $env:JD_INSTALL_DIR } else { $DefaultInstallDir }
    $installPath = Join-Path $installDir "$BinaryName.exe"

    # Create install directory if it doesn't exist
    if (-not (Test-Path $installDir)) {
        Write-Info "Creating directory: $installDir"
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    # Download
    $filename = "$BinaryName-windows-$arch.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$filename"

    Write-Info "Downloading $filename..."

    $tempFile = Join-Path $env:TEMP "$BinaryName-$([guid]::NewGuid()).exe"

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
    }
    catch {
        Write-Error-Exit "Download failed: $_"
    }

    # Install
    Write-Info "Installing to $installPath..."
    Move-Item -Path $tempFile -Destination $installPath -Force

    # Verify
    if (Test-Path $installPath) {
        Write-Success "Successfully installed jd $version"
        Write-Host ""
        & $installPath version
        Write-Host ""

        # Check if install directory is in PATH
        $pathDirs = $env:PATH -split ";"
        if ($installDir -notin $pathDirs) {
            Write-Warning "Note: $installDir is not in your PATH"
            Write-Host ""
            Write-Host "  To add it permanently, run:" -ForegroundColor Cyan
            Write-Host "    [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$installDir', 'User')"
            Write-Host ""
            Write-Host "  Or add to current session:" -ForegroundColor Cyan
            Write-Host "    `$env:PATH += ';$installDir'"
        }
    }
    else {
        Write-Error-Exit "Installation failed"
    }
}

Install-Jd
