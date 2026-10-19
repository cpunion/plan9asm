'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

function job(source, name) {
  const begin = source.indexOf('\n  ' + name + ':\n');
  assert.ok(begin >= 0, 'missing required complete native conversion job: ' + name);
  const end = source.slice(begin + 1).search(/\n  [a-z][a-z0-9-]*:\n/);
  return source.slice(begin, end < 0 ? undefined : begin + 1 + end);
}

test('complete native conversions have eight non-cancelling real Linux execution owners', () => {
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  const native = job(workflow, 'packed-conversion-native-priority');
  assert.match(native, /runs-on: ubuntu-latest/);
  assert.match(native, /fail-fast: false/);
  assert.match(native, /owner: \[0, 1, 2, 3, 4, 5, 6, 7\]/);
  assert.ok(!/continue-on-error|qemu|sde|rosetta/i.test(native), 'native execution cannot be replaced or ignored');
  assert.ok(native.includes('go test . -list "^($roots)$"'));
  assert.ok(native.includes('go test . -run "^($roots)$" -count=1'));
  assert.ok(native.includes('-json'));
  assert.ok(native.includes('node .github/scripts/audit-packed-conversion-native.cjs owner'));
  assert.ok(native.includes('if-no-files-found: error'));
  const aggregate = job(workflow, 'packed-conversion-native-complete');
  assert.match(aggregate, /needs: packed-conversion-native-priority/);
  assert.ok(aggregate.includes('node .github/scripts/audit-packed-conversion-native.cjs complete'));
  assert.ok(aggregate.includes('if-no-files-found: error'));
});

test('native conversion receipts require a strict current-head owner and complete-union auditor', () => {
  const path = '.github/scripts/audit-packed-conversion-native.cjs';
  assert.ok(fs.existsSync(path), 'missing complete native conversion receipt auditor');
  const source = fs.readFileSync(path, 'utf8');
  for (const guard of ['auditNativeOwners', 'AVX512F=1 AVX512DQ=1 AVX512BW=1 AVX512VL=1 OSXSAVE=1',
    'PACKED_CONVERSION_NATIVE_F_DQ_BW_VL_OK', '0xe6n', 'original_go_named', 'original_go_raw']) {
    assert.ok(source.includes(guard), 'missing strict native conversion guard: ' + guard);
  }
});
