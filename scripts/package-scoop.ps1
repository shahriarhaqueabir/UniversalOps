param([string]$Version)
# Generates a Scoop manifest for Universal-Ops
$BaseUrl = "https://github.com/shahriarhaqueabir/AllOpsFull/releases/download"

$manifest = @{
    version = $Version
    description = "High-performance native operations platform for systems, network, and security auditing."
    homepage = "https://github.com/shahriarhaqueabir/AllOpsFull"
    license = "MIT"
    architecture = @{
        "64bit" = @{
            url = "$BaseUrl/v$Version/universal-ops-$Version-windows-amd64.exe#/universal-ops.exe"
            hash = "PASTE_HASH_HERE"
        }
    }
    bin = "universal-ops.exe"
    checkver = "github"
    autoupdate = @{
        architecture = @{
            "64bit" = @{
                url = "$BaseUrl/v$version/universal-ops-$version-windows-amd64.exe#/universal-ops.exe"
            }
        }
    }
}

$manifest | ConvertTo-Json -Depth 10 | Out-File "universal-ops.json" -Encoding utf8
