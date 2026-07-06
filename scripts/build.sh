#!/bin/bash
# Build script for Hawkward on Linux/macOS
# Requires Go 1.26.4+

set -e

echo "Building Hawkward..."
go build -ldflags="-s -w" -o hawkward ./cmd/hawkward/
echo "Build successful: hawkward"
