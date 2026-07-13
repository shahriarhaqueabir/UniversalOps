#!/bin/sh
cd /e/Projects/projectx/AllOpsFull/cmd/hawkward-gui/frontend
node node_modules/typescript/lib/tsc.js --noEmit > /tmp/tsc-stdout.txt 2> /tmp/tsc-stderr.txt
echo "EXIT_CODE=$?" > /tmp/tsc-result.txt
cat /tmp/tsc-stderr.txt >> /tmp/tsc-result.txt
