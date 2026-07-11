#!/bin/bash
set -e
cd cmd/hawkward-gui/frontend
npx tsc --noEmit 2>&1
echo "---TSC_DONE---"
npm run lint 2>&1
echo "---LINT_DONE---"
