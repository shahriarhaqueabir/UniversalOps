#!/bin/bash
# UniversalOps Installation Script for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/shahriarhaqueabir/UniversalOps/main/install.sh | bash

# Fetch latest version from GitHub releases
API_URL="https://api.github.com/repos/shahriarhaqueabir/UniversalOps/releases/latest"
VERSION=$(curl -fsSL "$API_URL" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Could not fetch latest version. Falling back to default."
    VERSION="1.6.0"
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" == "darwin" ]; then
    URL="https://github.com/shahriarhaqueabir/UniversalOps/releases/download/v$VERSION/UniversalOps-v$VERSION-darwin-universal"
else
    URL="https://github.com/shahriarhaqueabir/UniversalOps/releases/download/v$VERSION/UniversalOps-v$VERSION-linux-amd64"
fi

DEST="/usr/local/bin/UniversalOps"

echo "Installing UniversalOps v$VERSION..."

curl -L "$URL" -o UniversalOps
chmod +x UniversalOps

if [ "$EUID" -ne 0 ]; then
    echo "Requesting sudo to install to $DEST"
    sudo mv UniversalOps "$DEST"
else
    mv UniversalOps "$DEST"
fi

echo "Installation complete. Type 'UniversalOps' to start."
