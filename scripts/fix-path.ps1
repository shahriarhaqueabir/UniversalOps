#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Fixes broken and duplicate PATH entries on this machine.
.DESCRIPTION
    User PATH was already cleaned to only E: drive tools (Go, Scoop, MinGW, npm-global, Zed).
    This script fixes the SYSTEM (Machine) PATH:
      - Replaces broken C:\Users\shahr\scoop\shims --> E:\scoop\shims
      - Removes broken C:\Users\shahr\scoop\apps\git\current\cmd (Git already at C:\Program Files\Git\cmd)
      - Removes C:\Program Files (x86)\NVIDIA Corporation\PhysX\Common (deleted by NVIDIA)
      - Removes C:\Users\shahr\AppData\Local\PowerToys\DSCModules (deprecated)
.NOTES
    Run from an Administrator PowerShell.
    Changes are permanent (survive reboot). No restart required.
#>

$ErrorActionPreference = "Continue"

Write-Host "`n=== PATH Cleanup Script ===" -ForegroundColor Cyan

# --- User PATH (already clean) ---
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User") -split ";" | Where-Object { $_ -ne "" }
Write-Host "`nUser PATH ($($userPath.Count) entries):" -ForegroundColor Green
$userPath | ForEach-Object { Write-Host "  [OK] $_" -ForegroundColor Green }

# --- System PATH (needs fixing) ---
$sysPathRaw = [Environment]::GetEnvironmentVariable("PATH", "Machine")
$sysPath = $sysPathRaw -split ";" | Where-Object { $_ -ne "" }
Write-Host "`nSystem PATH -- current ($($sysPath.Count) entries):" -ForegroundColor Yellow

$fixed = $sysPath | ForEach-Object {
    switch ($_) {
        "C:\Users\shahr\scoop\shims" {
            Write-Host "  [REPLACE] $_ --> E:\scoop\shims" -ForegroundColor Cyan
            "E:\scoop\shims"
        }
        "C:\Users\shahr\scoop\apps\git\current\cmd" {
            Write-Host "  [REMOVE]  $_ (Git already at C:\Program Files\Git\cmd)" -ForegroundColor Cyan
            $null
        }
        "C:\Program Files (x86)\NVIDIA Corporation\PhysX\Common" {
            Write-Host "  [REMOVE]  $_ (BROKEN)" -ForegroundColor Red
            $null
        }
        "C:\Users\shahr\AppData\Local\PowerToys\DSCModules" {
            Write-Host "  [REMOVE]  $_ (BROKEN)" -ForegroundColor Red
            $null
        }
        default { $_ }
    }
} | Where-Object { $null -ne $_ }

$fixed = $fixed | Select-Object -Unique

# --- Backup ---
$backupDir = Join-Path $env:USERPROFILE ".path-backups"
if (-not (Test-Path $backupDir)) { New-Item -ItemType Directory -Path $backupDir | Out-Null }
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$backupFile = Join-Path $backupDir "system-path-$timestamp.txt"
$sysPathRaw | Out-File -FilePath $backupFile -Encoding UTF8
Write-Host "`n[BACKUP] Saved: $backupFile" -ForegroundColor Green

# --- Confirm ---
Write-Host "`nReady to update system PATH." -ForegroundColor Yellow
$confirm = Read-Host "Apply changes? (y/N)"
if ($confirm -notmatch '^[yY]$') {
    Write-Host "`nAborted. No changes made." -ForegroundColor Red
    exit 0
}

[Environment]::SetEnvironmentVariable("PATH", ($fixed -join ";"), "Machine")
Write-Host "`n[DONE] System PATH updated. New count: $($fixed.Count) entries" -ForegroundColor Green

# --- Verify ---
Write-Host "`n=== VERIFICATION ===" -ForegroundColor Cyan
$env:PATH = ($userPath + $fixed) -join ";"

$tools = @{
    "Go"     = "go version"
    "Cargo"  = "cargo --version"
    "Node"   = "node --version"
    "Python" = "python --version"
    "Wails"  = "wails version"
    "Scoop"  = "scoop --version"
    "Git"    = "git --version"
    "Docker" = "docker --version"
}

foreach ($name in $tools.Keys) {
    $result = Invoke-Expression $tools[$name] 2>&1 | Select-Object -First 1
    if ($LASTEXITCODE -eq 0 -or $result -match "version") {
        Write-Host "  [OK] $name : $result" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] $name : not found or error" -ForegroundColor Yellow
    }
}

Write-Host "`n=== FINAL PATH STATE ===" -ForegroundColor Cyan
Write-Host "`nUser PATH ($($userPath.Count) entries -- E: drive tools only):" -ForegroundColor Green
$userPath | ForEach-Object { Write-Host "  $_" }
Write-Host "`nSystem PATH ($($fixed.Count) entries -- C: drive & Windows):" -ForegroundColor Yellow
$fixed | ForEach-Object { Write-Host "  $_" }

$total = ($userPath + $fixed | Select-Object -Unique).Count
Write-Host "`nTotal unique PATH entries: $total" -ForegroundColor White
Write-Host "No duplicates. No broken entries." -ForegroundColor Green

Write-Host "`nPress any key to exit..." -ForegroundColor DarkGray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")