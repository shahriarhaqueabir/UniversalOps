param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

$Targets = @(
    @{ Platform = "windows/amd64"; Ext = ".exe" }
    @{ Platform = "linux/amd64"; Ext = "" }
    @{ Platform = "darwin/universal"; Ext = "" }
)

Write-Host "Building Universal-Ops v$Version..." -ForegroundColor Cyan

foreach ($target in $Targets) {
    Write-Host "Building for $($target.Platform)..."
    $name = "universal-ops-$Version-$($target.Platform.Replace('/','-'))$($target.Ext)"
    wails build -platform $($target.Platform) -o $name
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build for $($target.Platform)"
        exit $LASTEXITCODE
    }
}

Write-Host "Builds completed successfully." -ForegroundColor Green
