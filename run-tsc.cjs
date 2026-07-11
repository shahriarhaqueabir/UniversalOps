const { execSync } = require('child_process');
const fs = require('fs');

try {
  execSync('cmd/hawkward-gui/frontend/node_modules/.bin/tsc --noEmit -p cmd/hawkward-gui/frontend/tsconfig.json', {
    stdio: 'pipe',
    timeout: 120000,
    cwd: __dirname
  });
  fs.writeFileSync('/tmp/tsc-result.txt', 'PASS\n');
} catch (e) {
  const out = (e.stdout || '').toString();
  const err = (e.stderr || '').toString();
  fs.writeFileSync('/tmp/tsc-result.txt', 'FAIL\n' + out + err);
}
