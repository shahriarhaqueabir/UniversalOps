const http = require('http');
const fs = require('fs');

const prompt = fs.readFileSync('graphify-prompt.txt', 'utf8');
const data = JSON.stringify({ model: 'graphify:latest', prompt: prompt, stream: false });

const req = http.request({
  hostname: 'localhost',
  port: 11434,
  path: '/api/generate',
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) }
}, (res) => {
  let body = '';
  res.on('data', (chunk) => { body += chunk; });
  res.on('end', () => {
    try {
      const parsed = JSON.parse(body);
      const out = parsed.response || 'NO RESPONSE';
      fs.writeFileSync('graphify-out.json', out, 'utf8');
      console.log('Graphify output saved to graphify-out.json (' + out.length + ' chars)');
      console.log(out.slice(0, 500));
    } catch (e) {
      fs.writeFileSync('graphify-out.json', body, 'utf8');
      console.log('Raw output saved (' + body.length + ' chars)');
      console.log(body.slice(0, 500));
    }
  });
});
req.on('error', (e) => { console.log('ERROR: ' + e.message); });
req.write(data);
req.end();
