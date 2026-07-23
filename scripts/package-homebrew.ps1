param([string]$Version)
# Generates a Homebrew formula for Universal-Ops
$BaseUrl = "https://github.com/shahriarhaqueabir/AllOpsFull/releases/download"

$formula = @"
class UniversalOps < Formula
  desc "High-performance native operations platform for systems, network, and security auditing."
  homepage "https://github.com/shahriarhaqueabir/AllOpsFull"
  version "$Version"

  if OS.mac?
    if Hardware::CPU.arm?
      url "$BaseUrl/v$Version/universal-ops-$Version-darwin-arm64"
      sha256 "PASTE_ARM64_HASH"
    else
      url "$BaseUrl/v$Version/universal-ops-$Version-darwin-amd64"
      sha256 "PASTE_AMD64_HASH"
    end
  end

  def install
    bin.install Dir["universal-ops-*"].first => "universal-ops"
  end
end
"@

$formula | Out-File "universal-ops.rb" -Encoding utf8
