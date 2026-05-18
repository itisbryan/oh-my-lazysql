#!/usr/bin/env bash
# install.sh — Install OhMyLazySQL binary for macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/<user>/oh-my-lazysql/main/install.sh | bash

set -e

REPO="itisbryan/oh-my-lazysql"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
DEST="$INSTALL_DIR/oh-my-lazysql"

detect_arch() {
  case "$(uname -m)" in
    arm64|aarch64) echo "Darwin_arm64" ;;
    x86_64)       echo "Darwin_x86_64" ;;
    *)            echo "Darwin_$(uname -m)" ;;
  esac
}

detect_tag() {
  local tag
  tag=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$tag" ]; then
    echo "ERROR: Could not fetch latest release tag. Is the repository public with a release?" >&2
    exit 1
  fi
  echo "$tag"
}

main() {
  local arch tag url

  echo "Installing OhMyLazySQL..."
  arch=$(detect_arch)
  tag=$(detect_tag)
  url="https://github.com/${REPO}/releases/download/${tag}/oh-my-lazysql_${arch}.tar.gz"

  echo "  Release: $tag"
  echo "  Arch:    $arch"
  echo "  URL:     $url"
  echo "  Dest:    $DEST"

  local tmpfile
  tmpfile=$(mktemp)
  curl -fsSL "$url" -o "$tmpfile" || { echo "ERROR: Download failed." >&2; exit 1; }
  tar -xzf "$tmpfile" -C "$INSTALL_DIR" || { echo "ERROR: Could not extract archive." >&2; exit 1; }
  rm -f "$tmpfile"
  chmod +x "$DEST"

  echo "Installed: $DEST"
  "$DEST" --version
}

main "$@"