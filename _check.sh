#!/bin/bash
cd "E:/Projects/projectx/AllOpsFull"
echo "=== COMPILING common ===" 
go build ./internal/common/ 2>&1 || true
echo "=== COMMON DONE ==="
echo "=== COMPILING app ==="
go build ./internal/app/ 2>&1 || true  
echo "=== APP DONE ==="
echo "=== COMPILING all ==="
go build ./... 2>&1 || true
echo "=== ALL DONE ==="
