const fs = require('fs');
if (!fs.existsSync('test-map.yaml')) { console.log('SKIP: no test-map.yaml'); process.exit(0); }
const raw = fs.readFileSync('test-map.yaml', 'utf8');
const t = [];
let s = false;
for (const l of raw.split('\n')) {
  if (/^type_tests:/.test(l)) { s = true; continue; }
  if (s && /^\S/.test(l) && !/^type_tests:/.test(l)) { s = false; }
  if (s) { const m = l.match(/^\s+-?\s*type:\s+(\S+)/); if (m) t.push(m[1]); }
}
if (!t.length) { console.log('SKIP: no type_tests declared'); process.exit(0); }
if (!fs.existsSync('test-report.md')) { process.stderr.write('FAIL: test-report.md not found\n'); process.exit(1); }
const report = fs.readFileSync('test-report.md', 'utf8');
const missing = t.filter(x => !report.includes(x));
if (missing.length) { process.stderr.write('FAIL: missing type evidence: ' + missing.join(', ') + '\n'); process.exit(1); }
console.log('OK: ' + t.join(', '));
