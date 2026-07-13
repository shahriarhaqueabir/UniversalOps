const { execSync } = require('child_process');
const path = require('path');

process.chdir(path.join(__dirname));

try {
  execSync('node node_modules/typescript/lib/tsc.js --noEmit', {
    stdio: 'inherit',
    timeout: 120000
  });
  console.log('TSC: PASS');
} catch (e) {
  console.log('TSC: FAIL (exit', e.status, ')');
}
