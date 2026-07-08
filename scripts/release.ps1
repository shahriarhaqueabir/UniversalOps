param(
    [string]$Version = "dev",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$targets = @(
    @{ Platform = "windows/amd64"; Ext = ".exe" },
    @{ Platform = "linux/amd64"; Ext = "" },
    @{ Platform = "darwin/amd64"; Ext = "" },
    @{ Platform = "darwin/arm64"; Ext = "" }
)

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

foreach ($target in $targets) {
    $name = "hawkward-$Version-$($target.Platform.Replace('/','-'))$($target.Ext)"
    $path = Join-Path $OutputDir $name
    Write-Host "Building $name"
    wails build -platform $target.Platform -trimpath -ldflags="-s -w" -o $path
}

$checksumPath = Join-Path $OutputDir "checksums.txt"
Get-ChildItem $OutputDir -File |
    Where-Object { $_.Name -ne "checksums.txt" } |
    ForEach-Object {
        $hash = Get-FileHash -Algorithm SHA256 $_.FullName
        "$($hash.Hash.ToLowerInvariant())  $($_.Name)"
    } | Set-Content -Encoding ascii $checksumPath

Write-Host "Release artifacts written to $OutputDir"
Write-Host "Checksums written to $checksumPath"
