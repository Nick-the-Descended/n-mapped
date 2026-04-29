#!/usr/bin/env sh
# n-mapped one-line installer.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/nick-the-descended/n-mapped/main/scripts/install.sh | sh
#
# Detects OS/arch, downloads the matching release tarball from GitHub, verifies
# the SHA256 checksum, and drops the binary into ~/.local/bin (or $PREFIX/bin
# if PREFIX is set and writable). Adds nothing to PATH; the script reminds the
# user if ~/.local/bin isn't on it.

set -eu

REPO="nick-the-descended/n-mapped"
PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT INT TERM

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '\033[1;34m::\033[0m %s\n' "$*"; }

uname_s="$(uname -s)"
case "$uname_s" in
  Linux)  os=linux ;;
  Darwin) os=darwin ;;
  *)      err "unsupported OS: $uname_s. Build from source: git clone https://github.com/$REPO" ;;
esac

uname_m="$(uname -m)"
case "$uname_m" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) err "unsupported arch: $uname_m" ;;
esac

if ! command -v curl >/dev/null 2>&1; then err "curl is required"; fi
if ! command -v tar  >/dev/null 2>&1; then err "tar is required"; fi

# Resolve the latest version via GitHub's redirect on /releases/latest.
info "resolving latest release..."
latest_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")"
version="${latest_url##*/}"
case "$version" in v*) ;; *) err "could not detect latest tag from $latest_url" ;; esac
ver_no_v="${version#v}"

archive="n-mapped_${ver_no_v}_${os}_${arch}.tar.gz"
checksums="SHA256SUMS"
base="https://github.com/$REPO/releases/download/$version"

info "downloading $archive"
curl -fL --progress-bar -o "$TMPDIR/$archive" "$base/$archive"
info "downloading checksums"
curl -fsSL -o "$TMPDIR/$checksums" "$base/$checksums"

info "verifying checksum"
expected="$(grep "  $archive$" "$TMPDIR/$checksums" || true)"
[ -n "$expected" ] || err "no checksum for $archive in SHA256SUMS"
if command -v sha256sum >/dev/null 2>&1; then
  ( cd "$TMPDIR" && printf '%s\n' "$expected" | sha256sum -c - >/dev/null )
elif command -v shasum >/dev/null 2>&1; then
  ( cd "$TMPDIR" && printf '%s\n' "$expected" | shasum -a 256 -c - >/dev/null )
else
  err "no sha256sum / shasum available; cowardly refusing to install unverified"
fi

info "extracting"
mkdir -p "$BIN_DIR"
tar -xzf "$TMPDIR/$archive" -C "$TMPDIR"
install -m 0755 "$TMPDIR/n-mapped" "$BIN_DIR/n-mapped"

if ! command -v nmap >/dev/null 2>&1; then
  printf '\nnote: \033[1;33mnmap is not installed\033[0m. Install it before running scans:\n  Linux:  sudo apt install nmap   (or your distro equivalent)\n  macOS:  brew install nmap\n'
fi

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) printf '\nnote: %s is not on your $PATH. Add it to your shell rc, e.g.:\n  export PATH="%s:$PATH"\n' "$BIN_DIR" "$BIN_DIR" ;;
esac

printf '\n\033[1;32mInstalled\033[0m %s -> %s/n-mapped\n' "$version" "$BIN_DIR"
printf 'Run %sn-mapped%s to launch (add %s--privileged%s for raw-socket scans).\n' '`' '`' '`' '`'
