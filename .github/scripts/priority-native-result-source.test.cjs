const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const groups = [
  {
    variable: 'independent_fp_roots',
    inventory: 'independent-fp',
    roots: [
      'TestCrossLinuxRuntimeMatrixX86InsertExtractIndependentFPFields',
      'TestCrossLinuxRuntimeMatrixX86LegacyPSIndependentFPFields',
    ],
  },
  {
    variable: 'source_stack_root',
    inventory: 'source-stack',
    roots: ['TestCrossLinuxRuntimeMatrixX86RawSourceStackNumerical'],
  },
  {
    variable: 'continuation_root',
    inventory: 'continuation',
    roots: ['TestCrossLinuxRuntimeMatrixX86ContinuationWholeCFGOriginalGoAndLLVM'],
  },
  {
    variable: 'packed_string_fault_root',
    inventory: 'packed-string-fault',
    roots: ['TestCrossLinuxRuntimeMatrixX86PackedStringOriginalNativeMemoryFaults'],
  },
  {
    variable: 'byte_swap_root',
    inventory: 'byte-swap',
    roots: ['TestCrossLinuxRuntimeMatrixX86ByteSwapOriginalGoAndLLVM'],
  },
  {
    variable: 'crc32_root',
    inventory: 'crc32',
    roots: ['TestAMD64CRC32RuntimeSemanticsAndFlags'],
  },
  {
    variable: 'bit_test_root',
    inventory: 'bit-test',
    roots: ['TestCrossLinuxRuntimeMatrixX86BitTestNumericalAndAliases'],
  },
];

function nativeBody(text) {
  const start = text.indexOf('\n  native-x86-oracles:');
  const end = text.indexOf('\n  complete-windows:', start);
  assert.ok(start >= 0 && end > start, 'missing native job boundary');
  return text.slice(start, end);
}

function validate(text) {
  const body = nativeBody(text);
  const execution = body.indexOf('go test . -run "^($roots|TestCrossLinuxRuntimeMatrixX86($families))$"');
  assert.ok(execution >= 0, 'missing complete native execution');
  const allQuery = 'go test . -list "^($roots|TestCrossLinuxRuntimeMatrixX86($families))$" > _out/native-x86-oracles/all-roots.stdout';
  const allQueryAt = body.indexOf(allQuery);
  assert.ok(allQueryAt >= 0 && allQueryAt < execution, 'missing complete exact66 discovery');
  assert.ok(body.includes('test "$(grep -c \'^Test\' _out/native-x86-oracles/all-roots.stdout)" = 66'),
    'complete fresh discovery must retain exactly66 roots');
  for (const group of groups) {
    const initial = new RegExp('\\b' + group.variable + "='([^']+)'").exec(body);
    assert.ok(initial, 'missing ' + group.inventory + ' registration');
    const names = initial[1].split('|');
    const append = new RegExp('\\b' + group.variable + "\\+='([^']+)'", 'g');
    for (const row of body.matchAll(append)) names.push(...row[1].replace(/^\|/, '').split('|'));
    assert.deepEqual(names, group.roots, 'wrong exact ' + group.inventory + ' roots');
    for (const name of group.roots) {
      assert.equal(body.split(name).length - 1, 1, 'duplicate native registration: ' + name);
    }
    const query = 'go test . -list "^($' + group.variable + ')$" > _out/native-x86-oracles/' + group.inventory + '-inventory.log';
    const queryAt = body.indexOf(query);
    assert.ok(queryAt >= 0 && queryAt < execution, 'missing/belated ' + group.inventory + ' discovery');
    const needle = group.roots.length === 1 ? '$' + group.variable : '$root';
    const guard = 'grep -Fx "' + needle + '" _out/native-x86-oracles/' + group.inventory + '-inventory.log';
    const guardAt = body.indexOf(guard);
    assert.ok(guardAt > queryAt && guardAt < execution, 'missing/belated exact inventory guard');
    if (group.roots.length > 1) {
      assert.ok(body.includes('for root in $' + '{' + group.variable + '//|/ }; do'),
        'every independent FP owner must be checked');
    }
    const join = 'roots+="|$' + group.variable + '"';
    const joinAt = body.indexOf(join);
    assert.ok(joinAt > guardAt && joinAt < execution, 'inventory is not joined before execution');
  }
  const command = body.slice(execution, body.indexOf('\n      - name:', execution));
  assert.match(command, /-json -count=1 -timeout=8m/);
  assert.match(command, /2> _out\/native-x86-oracles\/native\.stderr \| tee _out\/native-x86-oracles\/native\.jsonl/);
  assert.doesNotMatch(command, /2>&1|native\.log|;.*go test/);
  const audit = 'node .github/scripts/audit-native-priority.cjs _out/native-x86-oracles > _out/native-x86-oracles/audit.json';
  assert.ok(body.indexOf(audit) > execution, 'native job must enforce the JSON audit after execution');
  assert.doesNotMatch(body, /PLAN9ASM_CROSS_EXEC|qemu|overlay|continue-on-error|max-parallel/);
  assert.match(body, /set -euo pipefail/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
}

test('native result/source/fault/byte-swap/CRC32/BIT roles retain eight exact independently inventoried owners', () => {
  validate(workflow);
});

test('missing, duplicated or substituted native result/source registration is rejected', () => {
  for (const group of groups) {
    for (const root of group.roots) {
      assert.throws(() => validate(workflow.replace(root, 'TestUnrelatedNativeResultSource')));
      assert.throws(() => validate(workflow.replace(root, root + '|' + root)));
    }
    const query = 'go test . -list "^($' + group.variable + ')$"';
    assert.throws(() => validate(workflow.replace(query, 'printf ignored-inventory')));
    const guard = 'grep -Fx "' + (group.roots.length === 1 ? '$' + group.variable : '$root') +
      '" _out/native-x86-oracles/' + group.inventory + '-inventory.log';
    assert.throws(() => validate(workflow.replace(guard, 'printf ignored-membership')));
    assert.throws(() => validate(workflow.replace('roots+="|$' + group.variable + '"', 'printf ignored-role')));
  }
});

test('new profile cannot merge stderr, expand deadlines or enable a global cross executor', () => {
  assert.throws(() => validate(workflow.replace('-json -count=1 -timeout=8m', '-json -count=1 -timeout=9m')));
  assert.throws(() => validate(workflow.replace('2> _out/native-x86-oracles/native.stderr', '2>&1')));
  assert.throws(() => validate(workflow.replace('set -euo pipefail', 'set -euo pipefail\n          export PLAN9ASM_CROSS_EXEC=1')));
});

test('inventories and joins must precede the one native execution', () => {
  for (const group of groups) {
    const join = 'roots+="|$' + group.variable + '"';
    const changed = workflow.replace(join, '').replace(
      '            -json -count=1 -timeout=8m', '            -json -count=1 -timeout=8m\n          ' + join);
    assert.throws(() => validate(changed));
  }
});

test('the job cannot omit or replace its required machine-event audit', () => {
  assert.throws(() => validate(workflow.replace('node .github/scripts/audit-native-priority.cjs', 'printf skipped-audit')));
});
