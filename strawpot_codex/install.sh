#!/bin/sh
set -e

REPO="strawpot/strawpot_codex_cli"
BINARY_NAME="strawpot_codex"

# Detect OS
OS=$(uname -s)
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Default install directory
if [ -z "$INSTALL_DIR" ]; then
  INSTALL_DIR="/usr/local/bin"
fi

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="${BINARY_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading ${ASSET}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${BINARY_NAME}" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "${BINARY_NAME}" "$URL"
else
  echo "Error: curl or wget is required" >&2
  exit 1
fi

chmod +x "${BINARY_NAME}"
mkdir -p "$INSTALL_DIR"

if [ -w "$INSTALL_DIR" ]; then
  mv "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Moving to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
