#!/bin/bash
# Universal-Ops Installation Script for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/shahriarhaqueabir/AllOpsFull/main/install.sh | bash

VERSION="1.3.1"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" == "darwin" ]; then
    URL="https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$VERSION/universal-ops-v$VERSION-darwin-universal"
else
    URL="https://github.com/shahriarhaqueabir/AllOpsFull/releases/download/v$VERSION/universal-ops-v$VERSION-linux-amd64"
fi

DEST="/usr/local/bin/universal-ops"

echo "Installing Universal-Ops v$VERSION..."

curl -L "$URL" -o universal-ops
chmod +x universal-ops

if [ "$EUID" -ne 0 ]; then
    echo "Requesting sudo to install to $DEST"
    sudo mv universal-ops "$DEST"
else
    mv universal-ops "$DEST"
fi

echo "Installation complete. Type 'universal-ops' to start."
