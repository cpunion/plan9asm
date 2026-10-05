const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/go-ci.yml', 'utf8');
function job(name) {
  const marker = `\n  ${name}:\n`;
  const start = workflow.indexOf(marker);
  assert.notEqual(start, -1, `missing ${name}`);
  const body = workflow.slice(start + marker.length);
  const end = body.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  return end < 0 ? body : body.slice(0, end);
}

test('required cross-runtime inventory has sixteen exhaustive owners, not a concurrency cap', () => {
  const body = job('cross-runtime');
  assert.match(body, /fail-fast: false/);
  assert.doesNotMatch(body, /max-parallel:|continue-on-error:/);
  assert.match(body, /owner: \[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15\]/);
  assert.match(body, /-ci-isolate-root TestCrossLinuxRuntimeMatrix/);
  assert.match(body, /-ci-owner \$\{\{ matrix.owner \}\}/);
  assert.match(body, /-shards 16 -timeout 20m/);
  assert.match(body, /\^\(TestCrossLinuxRuntimeMatrix\|TestARM64Conformance\|TestARM64RawSQSHLURuntimeSemantics\|TestARMIntegerMemoryRuntimeCross\|TestARMRawSaturationRuntimeCross\)/);
  assert.match(body, /PLAN9ASM_CROSS_EXEC: '1'/);
  assert.match(body, /bash scripts\/install-ci-qemu.sh/);
  assert.match(body, /actions\/upload-artifact@v6/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
});

test('strict completion runs even after a failed owner and keeps original tool requirements', () => {
  const body = job('cross-runtime-verify');
  assert.match(body, /needs: cross-runtime/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
  assert.match(body, /-verify-ci-partitions _out\/cross-runtime-partitions/);
  assert.match(body, /-shards 16 -timeout 20m/);
  assert.match(body, /-ci-isolate-root TestCrossLinuxRuntimeMatrix/);
  assert.match(body, /actions\/download-artifact@v8/);
  assert.match(body, /gcc-i686-linux-gnu gcc-arm-linux-gnueabihf/);
  assert.match(body, /gcc-aarch64-linux-gnu/);
  assert.match(body, /bash scripts\/install-ci-qemu.sh/);
  assert.doesNotMatch(body, /continue-on-error:|PLAN9ASM_CROSS_EXEC: '0'/);
});
