#!/bin/sh
# install.sh — installer for zeus
# Usage: curl -fsSL https://raw.githubusercontent.com/smitt14ua/zeus/main/install.sh | sh
set -e

REPO="smitt14ua/zeus"
INSTALL_DIR="/usr/local/bin"
BIN_NAME="zeus"

# ── OS detection ────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS_NAME="linux"  ;;
  Darwin) OS_NAME="darwin" ;;
  *)
    printf 'Error: unsupported OS "%s"\n' "$OS" >&2
    exit 1
    ;;
esac

# ── Architecture detection ───────────────────────────────────────────────────
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

# ── Download helper ──────────────────────────────────────────────────────────
_download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    # --proto '=https' prevents redirect to non-HTTPS; --tlsv1.2 prevents downgrade
    curl -fsSL --proto '=https' --tlsv1.2 -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    printf 'Error: curl or wget is required.\n' >&2
    exit 1
  fi
}

_fetch_text() {
  url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --proto '=https' --tlsv1.2 "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url"
  else
    printf 'Error: curl or wget is required.\n' >&2
    exit 1
  fi
}

# ── Resolve download URL from latest release ─────────────────────────────────
printf 'Detecting latest release...\n'

LATEST_TAG="$(_fetch_text "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')"

if [ -z "$LATEST_TAG" ]; then
  printf 'Error: could not determine latest release tag.\n' >&2
  exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET_NAME}"

printf 'Installing zeus %s (%s/%s)...\n' "$LATEST_TAG" "$OS_NAME" "$ARCH_NAME"

# ── Download to temp file ────────────────────────────────────────────────────
TMP_FILE="$(mktemp)"
# Always remove the temp file on exit (even if mv succeeded — rm -f is a no-op on missing files)
trap 'rm -f "$TMP_FILE"' EXIT

_download "$DOWNLOAD_URL" "$TMP_FILE"

# Verify we got a non-empty file
if [ ! -s "$TMP_FILE" ]; then
  printf 'Error: downloaded file is empty. Check that the release asset "%s" exists.\n' "$ASSET_NAME" >&2
  exit 1
fi

# ── Checksum verification ────────────────────────────────────────────────────
TMP_SUM="$(mktemp)"
trap 'rm -f "$TMP_FILE" "$TMP_SUM"' EXIT

_download "${DOWNLOAD_URL}.sha256" "$TMP_SUM"
EXPECTED="$(cat "$TMP_SUM")"

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMP_FILE" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "$TMP_FILE" | awk '{print $1}')"
else
  printf 'Warning: sha256sum/shasum not found — skipping checksum verification.\n' >&2
  ACTUAL="$EXPECTED"
fi

if [ "$ACTUAL" != "$EXPECTED" ]; then
  printf 'Error: checksum mismatch.\n  expected: %s\n  got:      %s\n' "$EXPECTED" "$ACTUAL" >&2
  exit 1
fi
printf 'Checksum OK.\n'

chmod +x "$TMP_FILE"

# ── Install ──────────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/${BIN_NAME}"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "$DEST"
else
  printf 'Installing to %s requires sudo...\n' "$INSTALL_DIR"
  sudo mv "$TMP_FILE" "$DEST"
fi

printf 'zeus installed to %s\n' "$DEST"
printf 'Run "zeus --version" to verify.\n'
