#!/usr/bin/env node
const { execSync } = require('child_process');
const dir = require('path').join(__dirname, 'cmd', 'hawkward-gui', 'frontend');
process.chdir(dir);
try {
  execSync('node node_modules/typescript/lib/tsc.js --noEmit', { stdio: 'pipe', timeout: 120000 });
  console.log('TSC: PASS');
} catch (e) {
  if (e.stdout) process.stdout.write(e.stdout);
  if (e.stderr) process.stderr.write(e.stderr);
  console.log('TSC: FAIL exit=' + e.status);
  process.exit(1);
}
