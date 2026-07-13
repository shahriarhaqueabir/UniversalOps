# OpsForAll Windows Installer
# One-liner: irm https://opsforall.app/install.ps1 | iex

$ErrorActionPreference = "Stop"

$owner = "shahriarhaqueabir"
$repo = "AllOpsFull"
$appPath = "$HOME\AppData\Local\OpsForAll"
$binName = "OpsForAll.exe"

Write-Host "--- OpsForAll Installer ---" -ForegroundColor Cyan

# 1. Create directory
if (!(Test-Path $appPath)) {
    New-Item -ItemType Directory -Path $appPath -Force | Out-Null
}

# 2. Get latest release version from GitHub
Write-Host "Checking for latest release..."
$apiUrl = "https://api.github.com/repos/$owner/$repo/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $apiUrl -Method Get
    $version = $release.tag_name
} catch {
    Write-Error "Failed to fetch latest release info from GitHub."
    return
}

Write-Host "Found version: $version" -ForegroundColor Green

# 3. Download binary
$arch = "windows-amd64"
$assetName = "OpsForAll-$version-$arch.exe"
$asset = $release.assets | Where-Object { $_.name -eq $assetName }

if ($null -eq $asset) {
    # Fallback to dev naming if tagged but not released with final names
    $assetName = "OpsForAll-$($version.TrimStart('v'))-windows-amd64.exe"
    $asset = $release.assets | Where-Object { $_.name -eq $assetName }
}

if ($null -eq $asset) {
    Write-Error "Could not find asset $assetName in release $version"
    return
}

$downloadUrl = $asset.browser_download_url
$destPath = Join-Path $appPath $binName

Write-Host "Downloading OpsForAll to $appPath..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $destPath

# 4. Add to Path (User scope)
Write-Host "Adding to User PATH..."
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$appPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$appPath", "User")
    $env:Path += ";$appPath"
}

# 5. Create Desktop Shortcut
Write-Host "Creating Desktop Shortcut..."
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$HOME\Desktop\OpsForAll.lnk")
$Shortcut.TargetPath = $destPath
$Shortcut.WorkingDirectory = $appPath
$Shortcut.Save()

Write-Host "`nSuccessfully installed OpsForAll $version!" -ForegroundColor Green
Write-Host "You can now run 'OpsForAll' from your terminal or use the Desktop shortcut."
