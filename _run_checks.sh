#!/bin/sh
cd "E:/Projects/projectx/AllOpsFull"
> _buildcheck.txt

go build ./internal/common/ >/dev/null 2>/dev/null
echo "common: $?" >> _buildcheck.txt

go build ./internal/app/ >/dev/null 2>/dev/null
echo "app: $?" >> _buildcheck.txt

go build ./internal/netops/ >/dev/null 2>/dev/null
echo "netops: $?" >> _buildcheck.txt

go build ./internal/secops/ >/dev/null 2>/dev/null
echo "secops: $?" >> _buildcheck.txt

go build ./internal/devops/ >/dev/null 2>/dev/null
echo "devops: $?" >> _buildcheck.txt

go vet ./internal/networkdesign/ >/dev/null 2>/dev/null
echo "netdesign-vet: $?" >> _buildcheck.txt

go test ./internal/networkdesign/ -count=1 >/dev/null 2>/dev/null
echo "netdesign-test: $?" >> _buildcheck.txt

echo "done" >> _buildcheck.txt
