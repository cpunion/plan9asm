const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/go-ci.yml', 'utf8');
function job(name, source = workflow) {
  const marker = `\n  ${name}:\n`;
  const start = source.indexOf(marker);
  assert.notEqual(start, -1, `missing ${name}`);
  const body = source.slice(start + marker.length);
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

function requireARM64FallbackOwner(source) {
  const owner = job('cross-runtime', source);
  assert.match(owner, /- name: Execute rich CLI ARM64 fallback oracle\n        if: \$\{\{ !cancelled\(\) && matrix\.owner == 15 \}\}\n        run: bash scripts\/test-arm64-fallback-ci\.sh execute/);
  assert.equal((source.match(/- name: Execute rich CLI ARM64 fallback oracle/g) || []).length, 1);
  const reader = job('cross-runtime-verify', source);
  assert.match(reader, /- name: Verify rich CLI ARM64 fallback execution\n        if: \$\{\{ !cancelled\(\) \}\}\n        run: bash scripts\/test-arm64-fallback-ci\.sh complete/);
}

test('rich CLI ARM64 execution has a unique cross owner and a fresh completion reader', () => {
  requireARM64FallbackOwner(workflow);
});

test('ARM64 execution cannot lose its unique owner, success guard or reader', () => {
  for (const [before, after] of [
    ['Execute rich CLI ARM64 fallback oracle', 'Removed runtime oracle'],
    ['matrix.owner == 15', 'matrix.owner == 14'],
    ['!cancelled() && matrix.owner == 15', 'success()'],
    ['test-arm64-fallback-ci.sh execute', 'test-arm64-fallback-ci.sh complete'],
    ['Verify rich CLI ARM64 fallback execution', 'Removed execution reader'],
  ]) {
    const changed = workflow.replace(before, after);
    assert.notEqual(changed, workflow);
    assert.throws(() => requireARM64FallbackOwner(changed));
  }
  assert.throws(() => requireARM64FallbackOwner(workflow + '\n      - name: Execute rich CLI ARM64 fallback oracle\n'));
});
