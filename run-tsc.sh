#!/bin/sh
cd /e/Projects/projectx/AllOpsFull
./cmd/hawkward-gui/frontend/node_modules/.bin/tsc --noEmit -p ./cmd/hawkward-gui/frontend/tsconfig.json > /tmp/tsc-output.txt 2>&1
echo "TSC_EXIT=$?" > /tmp/tsc-exit.txt
