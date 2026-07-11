#!/bin/bash
cd "$(dirname "$0")"
go build ./internal/common/ 2>_build_err.txt
echo "BUILD_EXIT=$?"
cat _build_err.txt
go build ./internal/app/ 2>_build_err2.txt
echo "APP_EXIT=$?"
cat _build_err2.txt
rm -f _build_err.txt _build_err2.txt _buildcheck.sh
