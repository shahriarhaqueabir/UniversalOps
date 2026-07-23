#!/bin/bash
# Build script for Universal-Ops on Linux/macOS

# Prerequisites: Go, Node.js, npm, Wails CLI
echo "Building Universal-Ops..."
wails build -o universal-ops
if [ $? -eq 0 ]; then
    echo "Build successful: build/bin/universal-ops"
else
    echo "Build failed."
    exit 1
fi
