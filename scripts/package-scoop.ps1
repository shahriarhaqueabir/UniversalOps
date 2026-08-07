param(
    [string]$Version,
    [string]$Sha256 = ""
)
# Generates a Scoop manifest for UniversalOps.
# Usage (from repo root, pwsh):
#   & scripts/package-scoop.ps1 -Version v1.5.0 -Sha256 <real-sha256-of-exe>
# Accepts "v1.5.0" (git tag) or "1.5.0". The release download URL uses the tag (with v),
# while the asset name and manifest version use the bare version (no v).
$BaseUrl = "https://github.com/shahriarhaqueabir/UniversalOps/releases/download"

$tag = if ($Version.StartsWith('v')) { $Version } else { "v$Version" }
$ver = $Version.TrimStart('v')

if ([string]::IsNullOrWhiteSpace($Sha256) -or $Sha256 -match 'PASTE|PLACEHOLDER|REPLACE') {
    throw "A real SHA256 hash is required (-Sha256). Refusing to emit a manifest with a placeholder hash."
}

$manifest = @{
    version     = $ver
    description = "High-performance native operations platform for systems, network, and security auditing."
    homepage    = "https://github.com/shahriarhaqueabir/UniversalOps"
    license     = "MIT"
    architecture = @{
        "64bit" = @{
            url  = "$BaseUrl/$tag/universal-ops-$ver-windows-amd64.exe#/universal-ops.exe"
            hash = $Sha256
        }
    }
    bin         = "universal-ops.exe"
    checkver    = "github"
    autoupdate  = @{
        architecture = @{
            "64bit" = @{
                url = "$BaseUrl/v`$version/universal-ops-`$version-windows-amd64.exe#/universal-ops.exe"
            }
        }
    }
}

$manifest | ConvertTo-Json -Depth 10 | Out-File "universal-ops.json" -Encoding utf8
Write-Host "Wrote universal-ops.json (version $ver)"
