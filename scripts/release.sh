#!/bin/sh
set -eu

VERSION="${1:-dev}"
OUTPUT_DIR="${2:-dist}"

mkdir -p "$OUTPUT_DIR"

build_target() {
  goos="$1"
  goarch="$2"
  ext="$3"
  name="hawkward-${VERSION}-${goos}-${goarch}${ext}"
  echo "Building ${name}"
  GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "${OUTPUT_DIR}/${name}" ./cmd/hawkward/
}

build_target windows amd64 .exe
build_target linux amd64 ""
build_target darwin amd64 ""
build_target darwin arm64 ""

(
  cd "$OUTPUT_DIR"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum hawkward-"${VERSION}"-* > checksums.txt
  else
    shasum -a 256 hawkward-"${VERSION}"-* > checksums.txt
  fi
)

echo "Release artifacts written to ${OUTPUT_DIR}"
echo "Checksums written to ${OUTPUT_DIR}/checksums.txt"
