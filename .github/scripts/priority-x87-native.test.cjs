const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const roots = [
  'TestCrossLinuxRuntimeMatrixX86X87SignedI64IdentityOriginalGoAndLLVM',
  'TestAMD64X87ConditionalMoveRuntimeConditions',
  'TestX86RawX87ClearExceptionRuntimeSemantics',
  'TestX86RawX87SoftwareClearPreservesAllStatusFields',
];

function requireX87NativeJob(source) {
  const marker = '\n  x87-native-priority:\n';
  const start = source.indexOf(marker);
  assert.ok(start >= 0, 'missing independent native x87 execution owner');
  let body = source.slice(start + marker.length);
  const end = body.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  if (end >= 0) body = body.slice(0, end);
  assert.match(body, /runs-on: ubuntu-latest/);
  assert.match(body, /gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.match(body, /go-version: '1\.27\.2'/);
  assert.match(body, /set -euo pipefail/);
  assert.match(body, /go test -p=1 \. -list "\^\(\$roots\)\$"/);
  assert.match(body, /go test -p=1 \. -run "\^\(\$roots\)\$" -count=1 -timeout=5m -json/);
  for (const root of roots) assert.equal(body.split(root).length - 1, 1, 'exact complete native owner: ' + root);
  assert.match(body, /2> _out\/native-x87\/stderr\.log \| tee _out\/native-x87\/raw\.jsonl/);
  assert.match(body, /node \.github\/scripts\/audit-x87-native\.cjs _out\/native-x87/);
  assert.match(body, /actions\/upload-artifact@v6/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
  assert.doesNotMatch(body, /continue-on-error|max-parallel|qemu|rosetta|PLAN9ASM_CROSS_EXEC|overlay/);
  return body;
}

test('x87 complete native hardware and original-Go owners cannot be replaced by host skips or objects', () => {
  requireX87NativeJob(workflow);
});

module.exports = {requireX87NativeJob, roots};
