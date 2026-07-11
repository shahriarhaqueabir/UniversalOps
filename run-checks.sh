#!/bin/sh
cd cmd/hawkward-gui/frontend
node node_modules/typescript/lib/tsc.js --noEmit 2>&1
echo "TSC_EXIT=$?"
node node_modules/eslint/bin/eslint.js src/pages/Logs.tsx 2>&1
echo "LINT_EXIT=$?"
