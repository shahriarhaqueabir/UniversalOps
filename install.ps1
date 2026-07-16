# OpsForAll Installation Script for Windows
# Usage: irm https://raw.githubusercontent.com/shahriarhaqueabir/AllOpsFull/main/install.ps1 | iex

$Version = "1.3.0"
$Url = "https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$Version/opsforall-v$Version-windows-amd64.exe"
$Dest = "$HOME\AppData\Local\OpsForAll"
$Exe = "$Dest\opsforall.exe"

Write-Host "Installing OpsForAll v$Version..." -ForegroundColor Cyan

if (!(Test-Path $Dest)) {
    New-Item -ItemType Directory -Path $Dest | Out-Null
}

Write-Host "Downloading binaries..."
Invoke-WebRequest -Uri $Url -OutFile $Exe

# Create shortcut if possible
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$HOME\Desktop\OpsForAll.lnk")
$Shortcut.TargetPath = $Exe
$Shortcut.Save()

Write-Host "Installation complete. Shortcut created on Desktop." -ForegroundColor Green
Write-Host "Starting OpsForAll..."
Start-Process $Exe
