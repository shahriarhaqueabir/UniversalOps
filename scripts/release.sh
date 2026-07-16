#!/bin/bash
VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh <version>"
  exit 1
fi

PLATFORMS=("windows/amd64" "linux/amd64" "darwin/universal")

for p in "${PLATFORMS[@]}"; do
  IFS="/" read -r goos goarch <<< "$p"
  ext=""
  if [ "$goos" == "windows" ]; then ext=".exe"; fi
  name="opsforall-${VERSION}-${goos}-${goarch}${ext}"

  echo "Building for $p..."
  wails build -platform "$p" -o "$name"

  if [ $? -ne 0 ]; then
    echo "Build failed for $p"
    exit 1
  fi
done

# Generate checksums
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  sha256sum opsforall-"${VERSION}"-* > checksums.txt
else
  shasum -a 256 opsforall-"${VERSION}"-* > checksums.txt
fi

echo "Builds and checksums complete."
