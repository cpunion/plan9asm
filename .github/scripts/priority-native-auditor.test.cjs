const assert = require('node:assert/strict');
const path = require('node:path');
const test = require('node:test');
const { auditNativePriority } = require('./audit-native-priority.cjs');

// Synthetic machine-event controls are not native execution receipts.
const inventories = {
  alias: ['TestCrossLinuxRuntimeMatrixX86MOVQSourceAliasesOriginalAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86MOVQSourceAliasesCountedProvenance'],
  bool: ['TestCrossLinuxRuntimeMatrixAMD64RegisterCanonicalBoolInstructionBytes',
    'TestCrossLinuxRuntimeMatrixGoAMD64CanonicalBoolByteRepresentation'],
  caller: ['TestCrossLinuxRuntimeMatrixX86CallerAddressInstructionBytesAndLLVM',
    'TestCrossLinuxRuntimeMatrixGoX86CallerAddressABI0Representation'],
  'caller-cmp': ['TestCrossLinuxRuntimeMatrixX86CallerAddressCMPInstructionBytesAndLLVM',
    'TestCrossLinuxRuntimeMatrixGoX86CallerAddressCMPABI0Representation'],
  numerical: ['TestCrossLinuxRuntimeMatrixX86NumericalSBReadMutableAMD64',
    'TestCrossLinuxRuntimeMatrixX86NumericalSBReadMutableLocalAMD64'],
  'numerical-lea': ['TestCrossLinuxRuntimeMatrixX86NumericalLEAOriginalBytesAndLLVM'],
  'direct-tail': ['TestCrossLinuxRuntimeMatrixAMD64DirectTailWholeSourceOriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86DirectTail386OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86DirectTailPCResultAliasesAndExternal'],
  'gp-simd-sparse': ['TestCrossLinuxRuntimeMatrixX86GPSIMDSparseNativeAMD64OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86GPSIMDSparseNative386OriginalGoAndLLVM'],
  'bit-scan': ['TestCrossLinuxRuntimeMatrixX86BitScanNativeAMD64OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86BitScanNative386OriginalGoAndLLVM'],
  'bool-386': ['TestCrossLinuxRuntimeMatrixX86ABI0Mode32CanonicalBoolOriginalGoAndLLVM'],
  'packed-string': ['TestCrossLinuxRuntimeMatrixX86PackedStringImplicitVectorNativeAVX'],
  'independent-fp': ['TestCrossLinuxRuntimeMatrixX86InsertExtractIndependentFPFields',
    'TestCrossLinuxRuntimeMatrixX86LegacyPSIndependentFPFields'],
  'source-stack': ['TestCrossLinuxRuntimeMatrixX86RawSourceStackNumerical'],
  continuation: ['TestCrossLinuxRuntimeMatrixX86ContinuationWholeCFGOriginalGoAndLLVM'],
  'packed-string-fault': ['TestCrossLinuxRuntimeMatrixX86PackedStringOriginalNativeMemoryFaults'],
};
const sourceStack = inventories['source-stack'][0];
const continuation = inventories.continuation[0];
const packedStringFault = inventories['packed-string-fault'][0];
const requiredChildren = [sourceStack + '/amd64', sourceStack + '/386',
  ...['amd64', '386'].flatMap(arch => ['memory', 'register'].map(dispatch =>
    continuation + '/' + arch + '/' + dispatch)),
  packedStringFault + '/-O0', packedStringFault + '/-O2'];

function fixture() {
  const files = new Map();
  for (const [name, roots] of Object.entries(inventories)) {
    files.set(name + '-inventory.log', roots.join('\n') + '\n');
  }
  const roots = Object.values(inventories).flat();
  for (let index = 0; index < 38; index++) roots.push('TestSyntheticExtraOwner' + index);
  const events = [];
  const emit = (Action, Test) => events.push({
    Action, Package: 'github.com/xgo-dev/plan9asm', ...(Test ? { Test } : {}), Elapsed: 1,
  });
  for (const root of roots) {
    emit('run', root);
    for (const child of requiredChildren.filter(name => name.startsWith(root + '/'))) {
      emit('run', child);
      emit('pass', child);
    }
    if (root === 'TestSyntheticExtraOwner0') {
      // A literal slash in one t.Run name does not create intermediate parents.
      emit('run', root + '/one/literal/name');
      emit('pass', root + '/one/literal/name');
    }
    emit('pass', root);
  }
  events.push({ Action: 'output', Package: 'github.com/xgo-dev/plan9asm',
    Output: '--- PASS: TestOriginal (1.00s)\n{"Action":"pass","Test":"TestForgedPayload"}\n' });
  emit('pass');
  files.set('all-roots.stdout', roots.join('\n') + '\n');
  files.set('native.jsonl', events.map(JSON.stringify).join('\n') + '\n');
  files.set('native.stderr', '');
  return files;
}

function events(files) {
  return files.get('native.jsonl').trim().split('\n').map(JSON.parse);
}

function replaceEvents(files, rows) {
  files.set('native.jsonl', rows.map(JSON.stringify).join('\n') + '\n');
}

function audit(files) {
  return auditNativePriority('synthetic-fixture', {
    readdirSync() { return [...files.keys()]; },
    readFileSync(filename) {
      const name = path.basename(filename);
      assert.ok(files.has(name), 'missing synthetic file: ' + name);
      return files.get(name);
    },
  });
}

test('synthetic63 requires fifteen inventories,25 owners and eight source/ISA/fault children', () => {
  const result = audit(fixture());
  assert.equal(result.protocol, 'go_test_json_native63');
  assert.equal(result.roots, 63);
  assert.equal(result.inventories, 15);
  assert.equal(result.mandatory.length, 25);
  assert.equal(result.requiredChildren.length, 8);
  assert.equal(result.children, 9);
});

test('fresh full discovery rejects same-count nonmandatory root substitution', () => {
  const files = fixture();
  files.set('native.jsonl', files.get('native.jsonl').replaceAll(
    'TestSyntheticExtraOwner0', 'TestUnrelatedSameCountOwner'));
  assert.throws(() => audit(files), /discovery/);
});

test('missing, duplicate or renamed independent inventory cannot certify execution', () => {
  for (const inventory of Object.keys(inventories)) {
    const absent = fixture();
    absent.delete(inventory + '-inventory.log');
    assert.throws(() => audit(absent), /inventories/);
    const renamed = fixture();
    renamed.set(inventory + '-inventory.log', 'TestUnrelatedNativeOwner\n');
    assert.throws(() => audit(renamed), /inventory/);
    const duplicate = fixture();
    duplicate.set(inventory + '-inventory.log', duplicate.get(inventory + '-inventory.log').repeat(2));
    assert.throws(() => audit(duplicate), /inventory/);
  }
});

test('each required source-stack ISA, continuation dispatch and fault axis must finish', () => {
  for (const child of requiredChildren) {
    const files = fixture();
    replaceEvents(files, events(files).filter(event => event.Test !== child));
    assert.throws(() => audit(files), /required child/);
  }
});

test('root or child SKIP and FAIL are never PASS', () => {
  for (const name of [...Object.values(inventories).flat(), ...requiredChildren]) {
    for (const action of ['skip', 'fail']) {
      const files = fixture();
      const rows = events(files);
      rows.find(event => event.Action === 'pass' && event.Test === name).Action = action;
      replaceEvents(files, rows);
      assert.throws(() => audit(files), /execute successfully/);
    }
  }
});

test('fault children cannot finish after their successful root terminal', () => {
  for (const child of requiredChildren.filter(name => name.startsWith(packedStringFault + '/'))) {
    const files = fixture();
    const rows = events(files);
    const index = rows.findIndex(event => event.Action === 'pass' && event.Test === child);
    const [terminal] = rows.splice(index, 1);
    const parent = rows.findIndex(event => event.Action === 'pass' && event.Test === packedStringFault);
    rows.splice(parent + 1, 0, terminal);
    replaceEvents(files, rows);
    assert.throws(() => audit(files), /child.*root terminal/);
  }
});

test('duplicate or missing root and child events cannot close the protocol', () => {
  for (const name of [sourceStack, requiredChildren[0]]) {
    for (const action of ['run', 'pass']) {
      const duplicate = fixture();
      const rows = events(duplicate);
      const index = rows.findIndex(event => event.Action === action && event.Test === name);
      rows.splice(index + 1, 0, { ...rows[index] });
      replaceEvents(duplicate, rows);
      assert.throws(() => audit(duplicate), /duplicate/);
      const missing = fixture();
      replaceEvents(missing, events(missing).filter(event => event.Action !== action || event.Test !== name));
      assert.throws(() => audit(missing));
    }
  }
});

test('tool-output text cannot replace missing machine events', () => {
  const files = fixture();
  const rows = events(files).filter(event => event.Action !== 'pass' || event.Test !== requiredChildren[1]);
  rows.splice(rows.length - 1, 0, { Action: 'output', Package: 'github.com/xgo-dev/plan9asm',
    Output: JSON.stringify({ Action: 'pass', Test: requiredChildren[1] }) + '\n' });
  replaceEvents(files, rows);
  assert.throws(() => audit(files), /child terminals/);
});

test('missing package pass, failed package and malformed JSON remain failures', () => {
  const incomplete = fixture();
  replaceEvents(incomplete, events(incomplete).slice(0, -1));
  assert.throws(() => audit(incomplete), /package terminal/);
  const failed = fixture();
  const rows = events(failed);
  rows[rows.length - 1].Action = 'fail';
  replaceEvents(failed, rows);
  assert.throws(() => audit(failed), /execute successfully/);
  const malformed = fixture();
  malformed.set('native.jsonl', malformed.get('native.jsonl') + '{truncated\n');
  assert.throws(() => audit(malformed));
});

test('foreign package, unstarted child or duplicate package terminals are rejected', () => {
  const foreign = fixture();
  foreign.set('native.jsonl', foreign.get('native.jsonl').replaceAll(
    'github.com/xgo-dev/plan9asm', 'example.invalid/foreign'));
  assert.throws(() => audit(foreign), /package/);
  const child = fixture();
  const rows = events(child);
  rows.splice(rows.length - 1, 0, { Action: 'pass', Package: 'github.com/xgo-dev/plan9asm',
    Test: 'TestUnstarted/child' });
  replaceEvents(child, rows);
  assert.throws(() => audit(child), /unstarted child/);
  const duplicate = fixture();
  const repeated = events(duplicate);
  repeated.push({ ...repeated[repeated.length - 1] });
  replaceEvents(duplicate, repeated);
  assert.throws(() => audit(duplicate), /duplicate package/);
});

test('package completion requires its own exact package identity', () => {
  for (const identity of [undefined, null, '']) {
    const files = fixture();
    const rows = events(files);
    rows[rows.length - 1].Package = identity;
    replaceEvents(files, rows);
    assert.throws(() => audit(files), /package terminal.*identity/);
  }
});
