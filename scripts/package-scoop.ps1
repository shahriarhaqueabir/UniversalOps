param(
    [Parameter(Mandatory = $true)][string]$Version,
    [Parameter(Mandatory = $true)][string]$WindowsAmd64Sha,
    [string]$BaseUrl = "https://example.com/hawkward/releases/download"
)

@{
    version = $Version
    description = "Operations platform for the terminal"
    homepage = "https://example.com/hawkward"
    license = "MIT"
    architecture = @{
        "64bit" = @{
            url = "$BaseUrl/v$Version/hawkward-$Version-windows-amd64.exe"
            hash = $WindowsAmd64Sha
        }
    }
    bin = @(@("hawkward-$Version-windows-amd64.exe", "hawkward"))
} | ConvertTo-Json -Depth 5
