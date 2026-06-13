#!/usr/bin/env bash
# Install theirtime from GitHub Releases (macOS universal binary).
set -euo pipefail

REPO="${THEIRTIME_REPO:-haritabh17/theirtime}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${THEIRTIME_VERSION:-latest}"

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "theirtime is macOS-only (detected: $(uname -s))" >&2
  exit 1
fi

arch="$(uname -m)"
case "$arch" in
  arm64|x86_64) ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

api_base="https://api.github.com/repos/${REPO}/releases"
if [[ "$VERSION" == "latest" ]]; then
  release_url="${api_base}/latest"
else
  release_url="${api_base}/tags/${VERSION}"
fi

echo "Fetching release metadata…"
release_json="$(curl -fsSL "$release_url")"

read -r tarball_url checksums_url <<<"$(RELEASE_JSON="$release_json" python3 - <<'PY'
import json, os, sys

data = json.loads(os.environ["RELEASE_JSON"])
tarball = checksums = None
for asset in data.get("assets", []):
    name = asset.get("name", "")
    url = asset.get("browser_download_url", "")
    if name.endswith("_darwin_all.tar.gz"):
        tarball = url
    elif name == "checksums.txt":
        checksums = url

if not tarball:
    sys.stderr.write("error: no darwin_all tarball in release assets\n")
    sys.exit(1)
if not checksums:
    sys.stderr.write("error: no checksums.txt in release assets\n")
    sys.exit(1)

print(tarball, checksums)
PY
)"

tarball_name="$(basename "$tarball_url")"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading checksums…"
curl -fsSL "$checksums_url" -o "$tmp/checksums.txt"

echo "Downloading theirtime (universal macOS binary)…"
curl -fsSL "$tarball_url" -o "$tmp/$tarball_name"

(
  cd "$tmp"
  grep -F "$tarball_name" checksums.txt | shasum -a 256 -c -
)

tar -xzf "$tmp/$tarball_name" -C "$tmp"
bin="$(find "$tmp" -name theirtime -type f | head -1)"
if [[ -z "$bin" ]]; then
  echo "error: theirtime binary not found in archive" >&2
  exit 1
fi

if [[ ! -d "$INSTALL_DIR" ]]; then
  if [[ -w "$(dirname "$INSTALL_DIR")" ]]; then
    mkdir -p "$INSTALL_DIR"
  fi
fi

if [[ -w "$INSTALL_DIR" ]]; then
  install -m 755 "$bin" "$INSTALL_DIR/theirtime"
else
  echo "Installing to $INSTALL_DIR (may prompt for password)…"
  sudo install -m 755 "$bin" "$INSTALL_DIR/theirtime"
fi

echo "Installed: $($INSTALL_DIR/theirtime version)"
echo "Next: theirtime onboard"
