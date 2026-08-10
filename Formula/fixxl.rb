# fixxl Homebrew formula.
#
# Release flow:
#   1. Tag v0.1.0 (or bump) and let the GitHub release workflow build
#      dist binaries + SHA256SUMS.
#   2. Replace the two sha256 placeholders with the values published in that
#      release for fixxl-darwin-arm64 / fixxl-darwin-amd64.
#   3. Ship this file as Formula/fixxl.rb in your tap
#      (github.com/krishnasureshcpa/homebrew-fixxl).
#   4. `brew install krishnasureshcpa/fixxl/fixxl`
class Fixxl < Formula
  desc "converts spreadsheet files into clean clones — the source is never written"
  homepage "https://github.com/krishnasureshcpa/fixxl"
  version "0.1.0"

  on_arm do
    url "https://github.com/krishnasureshcpa/fixxl/releases/download/v0.1.0/fixxl-darwin-arm64"
    sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
  end
  on_intel do
    url "https://github.com/krishnasureshcpa/fixxl/releases/download/v0.1.0/fixxl-darwin-amd64"
    sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
  end

  def install
    bin.install Dir["fixxl-darwin-*"].first => "fixxl"
  end

  test do
    assert_match "fixxl", shell_output("#{bin}/fixxl -h")
  end
end