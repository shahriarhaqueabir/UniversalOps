#!/bin/bash
set -e
VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: ./release-gh.sh <version>"
    exit 1
fi

echo "Building Hawkward $VERSION..."

wails build -platform windows/amd64 -o hawkward-$VERSION-windows-amd64.exe
wails build -platform linux/amd64 -o hawkward-$VERSION-linux-amd64
wails build -platform darwin/amd64 -o hawkward-$VERSION-darwin-amd64
wails build -platform darwin/arm64 -o hawkward-$VERSION-darwin-arm64

echo "Creating GitHub release $VERSION..."
gh release create "$VERSION" \
    --title "Hawkward $VERSION" \
    --notes "Release $VERSION" \
    hawkward-$VERSION-windows-amd64.exe \
    hawkward-$VERSION-linux-amd64 \
    hawkward-$VERSION-darwin-amd64 \
    hawkward-$VERSION-darwin-arm64

echo "Release $VERSION completed!"
