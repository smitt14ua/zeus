#!/bin/sh
# install.sh — installer for zeus
# Usage: curl -fsSL https://raw.githubusercontent.com/smitt14ua/zeus/main/install.sh | sh
set -e

REPO="smitt14ua/zeus"
INSTALL_DIR="/usr/local/bin"
BIN_NAME="zeus"

# ── OS detection ─────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS_NAME="linux"  ;;
  Darwin) OS_NAME="darwin" ;;
  *)
    printf 'Error: unsupported OS "%s"\n' "$OS" >&2
    exit 1
    ;;
esac

# ── Architecture detection ────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *)
    printf 'Error: unsupported architecture "%s"\n' "$ARCH" >&2
    exit 1
    ;;
esac

ASSET_NAME="${BIN_NAME}-${OS_NAME}-${ARCH_NAME}"

# ── Network helpers ───────────────────────────────────────────────────────────
_fetch_text() {
  url="$1"
  if command -v curl >/dev/null 2>&1; then
    # --proto '=https' prevents redirect to non-HTTPS; --tlsv1.2 prevents downgrade
    curl -fsSL --proto '=https' --tlsv1.2 "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
  else
    printf 'Error: curl or wget is required.\n' >&2
    exit 1
  fi
}

_download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --proto '=https' --tlsv1.2 -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    printf 'Error: curl or wget is required.\n' >&2
    exit 1
  fi
}

# ── Fetch release metadata (single API call) ──────────────────────────────────
printf 'Detecting latest release...\n'

API_RESPONSE="$(_fetch_text "https://api.github.com/repos/${REPO}/releases/latest")"

LATEST_TAG="$(printf '%s' "$API_RESPONSE" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"

if [ -z "$LATEST_TAG" ]; then
  printf 'Error: could not determine latest release tag.\n' >&2
  exit 1
fi

# ── Extract SHA256 digest from the API response ───────────────────────────────
# GitHub automatically computes digests for all release assets (format: sha256:<hex>).
# GitHub API pretty-prints JSON; "name" appears before "digest" within each asset
# object. Match the exact asset name, then grab the first digest line and exit.
EXPECTED_HASH="$(printf '%s' "$API_RESPONSE" | awk '
  /"name":[[:space:]]*"'"${ASSET_NAME}"'"/ { in_asset=1 }
  in_asset && /"digest":[[:space:]]*"sha256:/ {
    sub(/.*"sha256:/, "")
    sub(/".*/, "")
    print
    exit
  }
')"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET_NAME}"

printf 'Installing zeus %s (%s/%s)...\n' "$LATEST_TAG" "$OS_NAME" "$ARCH_NAME"

# ── Download to temp file ─────────────────────────────────────────────────────
TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

_download "$DOWNLOAD_URL" "$TMP_FILE"

if [ ! -s "$TMP_FILE" ]; then
  printf 'Error: downloaded file is empty. Check that the release asset "%s" exists.\n' "$ASSET_NAME" >&2
  exit 1
fi

# ── Verify checksum ───────────────────────────────────────────────────────────
if [ -n "$EXPECTED_HASH" ]; then
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_HASH="$(sha256sum "$TMP_FILE" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_HASH="$(shasum -a 256 "$TMP_FILE" | awk '{print $1}')"
  else
    printf 'Warning: sha256sum/shasum not found — skipping checksum verification.\n' >&2
    ACTUAL_HASH="$EXPECTED_HASH"
  fi

  if [ "$ACTUAL_HASH" != "$EXPECTED_HASH" ]; then
    printf 'Error: checksum mismatch.\n  expected: %s\n  got:      %s\n' "$EXPECTED_HASH" "$ACTUAL_HASH" >&2
    exit 1
  fi
  printf 'Checksum OK.\n'
else
  printf 'Warning: digest not found in API response — skipping verification.\n' >&2
fi

chmod +x "$TMP_FILE"

# ── Install ───────────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/${BIN_NAME}"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "$DEST"
else
  printf 'Installing to %s requires sudo...\n' "$INSTALL_DIR"
  sudo mv "$TMP_FILE" "$DEST"
fi

printf 'zeus installed to %s\n' "$DEST"
printf 'Run "zeus --version" to verify.\n'
