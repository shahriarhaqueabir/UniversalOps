const { execSync } = require('child_process');
const fs = require('fs');
const r = [];
const cwd = 'E:/Projects/projectx/AllOpsFull';

function run(label, cmd) {
  try {
    execSync(cmd, { cwd, encoding: 'utf8', timeout: 60000 });
    r.push(label + ': PASS');
  } catch (e) {
    const stderr = (e.stderr || '').replace(/[`]/g, "'").trim().slice(0, 500);
    r.push(label + ': FAIL');
    if (stderr) r.push(stderr);
  }
}

run('internal/common', 'go build ./internal/common/');
run('internal/app', 'go build ./internal/app/');
run('networkdesign vet', 'go vet ./internal/networkdesign/');
run('networkdesign test', 'go test ./internal/networkdesign/ -count=1');
run('netops', 'go build ./internal/netops/');
run('secops', 'go build ./internal/secops/');
run('devops', 'go build ./internal/devops/');

fs.writeFileSync(cwd + '/_buildcheck.txt', r.join('\n'));
process.stdout.write('done\n');
