#!/bin/sh
cd /c/Projects/projectx/AllOpsFull 2>/dev/null || cd "E:/Projects/projectx/AllOpsFull" 2>/dev/null || true
go build ./internal/common/ > _build_common.txt 2>&1
echo "common: $?" > _build_result.txt
go build ./internal/app/ > _build_app.txt 2>&1
echo "app: $?" >> _build_result.txt
go vet ./internal/networkdesign/ > _build_netdesign.txt 2>&1
echo "netdesign vet: $?" >> _build_result.txt
go test ./internal/networkdesign/ -count=1 > _build_test.txt 2>&1
echo "netdesign test: $?" >> _build_result.txt
echo "done" >> _build_result.txt
