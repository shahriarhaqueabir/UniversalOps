#!/bin/bash

# Simple GitHub release script using the gh CLI

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: ./release-gh.sh <version>"
    exit 1
fi

echo "Building Hawkward $VERSION..."

# Build for major platforms
GOOS=windows GOARCH=amd64 go build -o hawkward-windows-amd64.exe ./cmd/hawkward
GOOS=linux GOARCH=amd64 go build -o hawkward-linux-amd64 ./cmd/hawkward
GOOS=darwin GOARCH=amd64 go build -o hawkward-darwin-amd64 ./cmd/hawkward
GOOS=darwin GOARCH=arm64 go build -o hawkward-darwin-arm64 ./cmd/hawkward

echo "Creating GitHub release $VERSION..."
gh release create "$VERSION" \
    --title "Hawkward $VERSION" \
    --notes "Release $VERSION" \
    hawkward-windows-amd64.exe \
    hawkward-linux-amd64 \
    hawkward-darwin-amd64 \
    hawkward-darwin-arm64

echo "Release $VERSION completed!"
