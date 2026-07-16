#!/bin/bash
# OpsForAll Installation Script for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/shahriarhaqueabir/AllOpsFull/main/install.sh | bash

VERSION="1.3.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" == "darwin" ]; then
    URL="https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$VERSION/opsforall-v$VERSION-darwin-universal"
else
    URL="https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$VERSION/opsforall-v$VERSION-linux-amd64"
fi

DEST="/usr/local/bin/opsforall"

echo "Installing OpsForAll v$VERSION..."

curl -L "$URL" -o opsforall
chmod +x opsforall

if [ "$EUID" -ne 0 ]; then
    echo "Requesting sudo to install to $DEST"
    sudo mv opsforall "$DEST"
else
    mv opsforall "$DEST"
fi

echo "Installation complete. Type 'opsforall' to start."
