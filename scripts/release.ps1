param(
    [string]$Version = "dev",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Ext = "" }
)

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

foreach ($target in $targets) {
    $name = "hawkward-$Version-$($target.GOOS)-$($target.GOARCH)$($target.Ext)"
    $path = Join-Path $OutputDir $name
    Write-Host "Building $name"
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    go build -trimpath -ldflags="-s -w" -o $path .\cmd\hawkward\
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

$checksumPath = Join-Path $OutputDir "checksums.txt"
Get-ChildItem $OutputDir -File |
    Where-Object { $_.Name -ne "checksums.txt" } |
    ForEach-Object {
        $hash = Get-FileHash -Algorithm SHA256 $_.FullName
        "$($hash.Hash.ToLowerInvariant())  $($_.Name)"
    } | Set-Content -Encoding ascii $checksumPath

Write-Host "Release artifacts written to $OutputDir"
Write-Host "Checksums written to $checksumPath"
