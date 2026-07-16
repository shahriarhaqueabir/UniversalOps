param([string]$Version)
# Generates a Scoop manifest for OpsForAll
$BaseUrl = "https://github.com/shahriarhaqueabir/AllOpsFull/releases/download"

$manifest = @{
    version = $Version
    description = "High-performance native operations platform for systems, network, and security auditing."
    homepage = "https://github.com/shahriarhaqueabir/AllOpsFull"
    license = "MIT"
    architecture = @{
        "64bit" = @{
            url = "$BaseUrl/v$Version/opsforall-$Version-windows-amd64.exe#/opsforall.exe"
            hash = "PASTE_HASH_HERE"
        }
    }
    bin = "opsforall.exe"
    checkver = "github"
    autoupdate = @{
        architecture = @{
            "64bit" = @{
                url = "$BaseUrl/v$version/opsforall-$version-windows-amd64.exe#/opsforall.exe"
            }
        }
    }
}

$manifest | ConvertTo-Json -Depth 10 | Out-File "opsforall.json" -Encoding utf8
