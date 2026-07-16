#!/bin/bash
# Builds and uploads a release to GitHub using the gh CLI.
# Usage: ./release-gh.sh v1.3.0

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: ./release-gh.sh <version>"
    exit 1
fi

echo "Building OpsForAll $VERSION..."
# Build all platforms
wails build -platform windows/amd64 -o opsforall-$VERSION-windows-amd64.exe
wails build -platform linux/amd64 -o opsforall-$VERSION-linux-amd64
wails build -platform darwin/amd64 -o opsforall-$VERSION-darwin-amd64
wails build -platform darwin/arm64 -o opsforall-$VERSION-darwin-arm64

echo "Creating GitHub release $VERSION..."
gh release create "$VERSION" \
    --title "OpsForAll $VERSION" \
    --notes "Release hardening and technical substrate updates." \
    opsforall-$VERSION-windows-amd64.exe \
    opsforall-$VERSION-linux-amd64 \
    opsforall-$VERSION-darwin-amd64 \
    opsforall-$VERSION-darwin-arm64

echo "Release $VERSION uploaded successfully."
