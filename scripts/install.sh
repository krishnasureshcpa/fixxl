#!/bin/sh
# fixxl install — downloads the binary from GitHub Releases and verifies its
# SHA256 against the published SHA256SUMS before installing.
#
#   curl -fsSL https://raw.githubusercontent.com/krishnasureshcpa/fixxl/main/scripts/install.sh | sh
#
# Overrides (useful for testing before the first release exists):
#   FIXXL_VERSION   tag, e.g. v0.1.0   (default: latest)
#   FIXXL_BASE      release base URL   (default: GitHub releases)
#   FIXXL_PREFIX    install directory  (default: $HOME/.local/bin)
set -eu

REPO="${FIXXL_REPO:-krishnasureshcpa/fixxl}"
BASE="${FIXXL_BASE:-https://github.com/$REPO/releases/download}"
VERSION="${FIXXL_VERSION:-latest}"

# --- platform detection -----------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$os" in
  darwin) ;;
  linux) ;;
  *) echo "fixxl: unsupported platform: $os" >&2; exit 1 ;;
esac
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "fixxl: unsupported arch: $arch" >&2; exit 1 ;;
esac

# --- resolve latest to a concrete tag ------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')"
fi

asset="fixxl-$os-$arch"
url="$BASE/$VERSION/$asset"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "fixxl: $VERSION ($os/$arch)"
echo "  download  $url"

curl -fsSL -o "$tmp/$asset" "$url"
curl -fsSL -o "$tmp/SHA256SUMS" "$BASE/$VERSION/SHA256SUMS"

want="$(awk -v f="$asset" '$2==f {print $1}' "$tmp/SHA256SUMS")"
[ -n "$want" ] || { echo "fixxl: no checksum for $asset" >&2; exit 1; }
got="$(shasum -a 256 "$tmp/$asset" | awk '{print $1}')"
[ "$got" = "$want" ] || { echo "fixxl: checksum mismatch — aborting" >&2; exit 1; }

prefix="${FIXXL_PREFIX:-$HOME/.local/bin}"
mkdir -p "$prefix"
install -m 0755 "$tmp/$asset" "$prefix/fixxl"
echo "fixxl: installed $prefix/fixxl"
echo "  next: export PATH=\"$prefix:\$PATH\""
echo "  run:  fixxl demo"