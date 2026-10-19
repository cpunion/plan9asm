'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

function job(source, name) {
  const begin = source.indexOf('\n  ' + name + ':\n');
  assert.ok(begin >= 0, 'missing required complete native reduction job: ' + name);
  const end = source.slice(begin + 1).search(/\n  [a-z][a-z0-9-]*:\n/);
  return source.slice(begin, end < 0 ? undefined : begin + 1 + end);
}

test('all native reductions have four non-cancelling real Linux owner pairs', () => {
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  const native = job(workflow, 'packed-reduction-native-priority');
  assert.match(native, /runs-on: ubuntu-latest/);
  assert.match(native, /fail-fast: false/);
  assert.match(native, /owner: \[0, 1, 2, 3\]/);
  assert.ok(!/continue-on-error|qemu|sde|rosetta/i.test(native), 'real reduction execution cannot be replaced or ignored');
  assert.ok(native.includes('go test . -list "^($roots)$"'));
  assert.ok(native.includes('go test . -run "^($roots)$" -count=1'));
  assert.ok(native.includes('-json'));
  assert.ok(native.includes('node .github/scripts/audit-packed-reduction-native.cjs owner'));
  assert.ok(native.includes('if-no-files-found: error'));
  const aggregate = job(workflow, 'packed-reduction-native-complete');
  assert.match(aggregate, /needs: packed-reduction-native-priority/);
  assert.ok(aggregate.includes('node .github/scripts/audit-packed-reduction-native.cjs complete'));
  assert.ok(aggregate.includes('if-no-files-found: error'));
});

test('native reduction receipts require all current source execution and CPU evidence', () => {
  const path = '.github/scripts/audit-packed-reduction-native.cjs';
  assert.ok(fs.existsSync(path), 'missing complete native reduction receipt auditor');
  const source = fs.readFileSync(path, 'utf8');
  for (const required of [
    'auditNativeOwners', 'SSE3=1 SSE41=1 AVX=1 OSXSAVE=1',
    'PACKED_REDUCTION_NATIVE_SSE3_SSE41_AVX_OK', '6n',
    'original_go_named', 'original_go_raw', 'original_go_opcode_reference',
    'backend_O0', 'backend_O2', 'IR_O2_backend_O2',
  ]) {
    assert.ok(source.includes(required), 'missing strict native reduction guard: ' + required);
  }
});
