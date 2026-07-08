#!/bin/bash
# Build script for Hawkward GUI on Linux/macOS
# Requires: Go 1.26.4+, Node.js, npm, Wails CLI
set -e
echo "Building Hawkward GUI..."
wails build -o hawkward-gui
echo "Build successful: build/bin/hawkward-gui"
