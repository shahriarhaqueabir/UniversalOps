#!/bin/bash
# Build script for OpsForAll GUI on Linux/macOS

# Prerequisites: Go, Node.js, npm, Wails CLI
echo "Building OpsForAll GUI..."
wails build -o opsforall-gui
if [ $? -eq 0 ]; then
    echo "Build successful: build/bin/opsforall-gui"
else
    echo "Build failed."
    exit 1
fi
