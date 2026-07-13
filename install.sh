#!/bin/bash
# OpsForAll Linux/macOS Installer
# One-liner: curl -fsSL https://opsforall.app/install.sh | sh

set -e

OWNER="shahriarhaqueabir"
REPO="AllOpsFull"
BIN_NAME="opsforall"

echo "--- OpsForAll Installer ---"

# 1. Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) PLATFORM="linux-$ARCH" ;;
    darwin) PLATFORM="darwin-$ARCH" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# 2. Get latest release version
echo "Checking for latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$OWNER/$REPO/releases/latest")
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest release version."
    exit 1
fi

echo "Found version: $VERSION"

# 3. Download binary
ASSET_NAME="OpsForAll-$VERSION-$PLATFORM"
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)

# Fallback for dev naming
if [ -z "$DOWNLOAD_URL" ]; then
    CLEAN_VERSION=$(echo "$VERSION" | sed 's/^v//')
    ASSET_NAME="OpsForAll-$CLEAN_VERSION-$PLATFORM"
    DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Could not find asset $ASSET_NAME in release $VERSION"
    exit 1
fi

INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"
DEST_PATH="$INSTALL_DIR/$BIN_NAME"

echo "Downloading OpsForAll to $DEST_PATH..."
curl -L -o "$DEST_PATH" "$DOWNLOAD_URL"
chmod +x "$DEST_PATH"

# 4. Check PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Adding $INSTALL_DIR to PATH in ~/.bashrc and ~/.zshrc..."
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
    [ -f "$HOME/.zshrc" ] && echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
    echo "Please restart your terminal or run 'source ~/.bashrc' to use $BIN_NAME"
fi

echo -e "\nSuccessfully installed OpsForAll $VERSION!"
echo "You can now run '$BIN_NAME' from your terminal."
