#!/bin/sh
set -e

REPO="luuuc/council-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="council"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get latest version
echo "Fetching latest release..."
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version"
  exit 1
fi

echo "Latest version: $VERSION"

# Build download URL
EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

URL="https://github.com/$REPO/releases/download/$VERSION/council-${OS}-${ARCH}${EXT}"

echo "Downloading from: $URL"

# Download
TMP_FILE=$(mktemp)
if ! curl -fsSL "$URL" -o "$TMP_FILE"; then
  echo "Download failed"
  rm -f "$TMP_FILE"
  exit 1
fi

# Install
chmod +x "$TMP_FILE"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
fi

echo ""
echo "council $VERSION installed to $INSTALL_DIR/$BINARY_NAME"
echo ""
echo "Get started:"
echo "  council init           Initialize council directory"
echo "  council setup --apply  Create council with AI assistance"
echo "  council sync           Sync to AI tool configs"
