param(
    [string]$Version,
    [string]$Arm64Sha256 = "",
    [string]$Amd64Sha256 = ""
)
# Generates a Homebrew formula for Universal-Ops.
# NOTE: macOS binaries are built manually on a macOS host (scripts/release-gh.sh),
# so this script is NOT invoked by CI (release.yml publishes Windows only).
# Usage (from repo root, pwsh):
#   & scripts/package-homebrew.ps1 -Version v1.5.0 -Arm64Sha256 <hash> -Amd64Sha256 <hash>
# Asset names match release-gh.sh output (tag INCLUDED, e.g. universal-ops-v1.5.0-darwin-arm64).
$BaseUrl = "https://github.com/shahriarhaqueabir/UniversalOps/releases/download"

$tag = if ($Version.StartsWith('v')) { $Version } else { "v$Version" }
$ver = $Version.TrimStart('v')

if ([string]::IsNullOrWhiteSpace($Arm64Sha256) -or $Arm64Sha256 -match 'PASTE|PLACEHOLDER|REPLACE') {
    throw "A real ARM64 SHA256 hash is required (-Arm64Sha256). Refusing to emit a placeholder."
}
if ([string]::IsNullOrWhiteSpace($Amd64Sha256) -or $Amd64Sha256 -match 'PASTE|PLACEHOLDER|REPLACE') {
    throw "A real AMD64 SHA256 hash is required (-Amd64Sha256). Refusing to emit a placeholder."
}

$formula = @"
class UniversalOps < Formula
  desc "High-performance native operations platform for systems, network, and security auditing."
  homepage "https://github.com/shahriarhaqueabir/UniversalOps"
  version "$ver"

  if OS.mac?
    if Hardware::CPU.arm?
      url "$BaseUrl/$tag/universal-ops-$tag-darwin-arm64"
      sha256 "$Arm64Sha256"
    else
      url "$BaseUrl/$tag/universal-ops-$tag-darwin-amd64"
      sha256 "$Amd64Sha256"
    end
  end

  def install
    bin.install Dir["universal-ops-*"].first => "universal-ops"
  end
end
"@

$formula | Out-File "universal-ops.rb" -Encoding utf8
Write-Host "Wrote universal-ops.rb (version $ver)"
