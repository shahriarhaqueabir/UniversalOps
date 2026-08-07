# UniversalOps Installation Script for Windows (Portable Mode)
# Usage: irm https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.ps1 | iex

# Fetch latest version from GitHub releases
$apiUrl = "https://api.github.com/repos/shahriarhaqueabir/UniversalOps/releases/latest"
try {
    $latest = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
    $Version = $latest.tag_name -replace '^v', ''
} catch {
    Write-Host "Could not fetch latest version. Falling back to default." -ForegroundColor Yellow
    $Version = "1.6.2"
}

# Asset name matches release.yml CI output (lowercase, no v prefix)
$Url = "https://github.com/shahriarhaqueabir/UniversalOps/releases/download/v$Version/universal-ops-$Version-windows-amd64.exe"
$Dest = "$HOME\Desktop\UniversalOps"
$Exe = "$Dest\UniversalOps.exe"

Write-Host "Setting up UniversalOps Portable v$Version..." -ForegroundColor Cyan

# --- Prerequisites check ---
$webView2 = $null
$webView2 = Get-ItemProperty -Path "HKLM:\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" -ErrorAction SilentlyContinue
if (-not $webView2) {
    $webView2 = Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" -ErrorAction SilentlyContinue
}
if (-not $webView2) {
    Write-Host ""
    Write-Host "WebView2 Runtime is required but not found." -ForegroundColor Yellow
    Write-Host "Opening download page..." -ForegroundColor Yellow
    Start-Process "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
    Write-Host "Install WebView2 Runtime, then run this script again." -ForegroundColor Yellow
    exit 1
}

# --- Create folders ---
if (!(Test-Path $Dest)) {
    New-Item -ItemType Directory -Path $Dest | Out-Null
}

New-Item -ItemType Directory -Path "$Dest\data" -Force | Out-Null
New-Item -ItemType Directory -Path "$Dest\logs" -Force | Out-Null
New-Item -ItemType Directory -Path "$Dest\bin" -Force | Out-Null

Write-Host "Downloading binary..."
Invoke-WebRequest -Uri $Url -OutFile $Exe

# Create shortcut on Desktop
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$HOME\Desktop\UniversalOps.lnk")
$Shortcut.TargetPath = $Exe
$Shortcut.WorkingDirectory = $Dest
$Shortcut.Save()

Write-Host "Setup complete. The app is located at $Dest" -ForegroundColor Green
Write-Host "All data and logs will remain inside this folder."
Write-Host "Starting UniversalOps..."
Start-Process $Exe -WorkingDirectory $Dest
