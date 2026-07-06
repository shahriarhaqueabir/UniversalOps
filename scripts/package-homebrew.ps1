param(
    [Parameter(Mandatory = $true)][string]$Version,
    [Parameter(Mandatory = $true)][string]$DarwinAmd64Sha,
    [Parameter(Mandatory = $true)][string]$DarwinArm64Sha,
    [string]$BaseUrl = "https://example.com/hawkward/releases/download"
)

$formula = @"
class Hawkward < Formula
  desc "Operations platform for the terminal"
  homepage "https://example.com/hawkward"
  version "$Version"

  on_macos do
    if Hardware::CPU.arm?
      url "$BaseUrl/v$Version/hawkward-$Version-darwin-arm64"
      sha256 "$DarwinArm64Sha"
    else
      url "$BaseUrl/v$Version/hawkward-$Version-darwin-amd64"
      sha256 "$DarwinAmd64Sha"
    end
  end

  def install
    bin.install Dir["hawkward-*"].first => "hawkward"
  end
end
"@

$formula
