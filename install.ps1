# Universal-Ops Installation Script for Windows (Portable Mode)
# Usage: irm https://raw.githubusercontent.com/shahriarhaqueabir/AllOpsFull/main/install.ps1 | iex

$Version = "1.3.1"
$Url = "https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$Version/universal-ops-v$Version-windows-amd64.exe"
$Dest = "$HOME\Desktop\UniversalOps"
$Exe = "$Dest\universal-ops.exe"

Write-Host "Setting up Universal-Ops Portable v$Version..." -ForegroundColor Cyan

if (!(Test-Path $Dest)) {
    New-Item -ItemType Directory -Path $Dest | Out-Null
}

# Create folder structure for zero-manual work
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
Write-Host "Starting Universal-Ops..."
Start-Process $Exe -WorkingDirectory $Dest
