const { execSync } = require('child_process');
const fs = require('fs');
const results = [];

function run(label, cmd) {
  try {
    const out = execSync(cmd, { cwd: 'E:/Projects/projectx/AllOpsFull', encoding: 'utf8', timeout: 60000 });
    results.push(`${label}: PASS`);
    if (out.trim()) results.push(`  output: ${out.trim().slice(0, 200)}`);
  } catch (e) {
    results.push(`${label}: FAIL`);
    if (e.stdout) results.push(`  stdout: ${e.stdout.trim().slice(0, 500)}`);
    if (e.stderr) results.push(`  stderr: ${e.stderr.trim().slice(0, 500)}`);
  }
}

run('go build ./internal/common/', 'go build ./internal/common/');
run('go build ./internal/app/', 'go build ./internal/app/');
run('go vet ./internal/networkdesign/', 'go vet ./internal/networkdesign/');
run('go test ./internal/networkdesign/', 'go test ./internal/networkdesign/ -count=1');

fs.writeFileSync('E:/Projects/projectx/AllOpsFull/_buildcheck.txt', results.join('\n'));
console.log('done');
