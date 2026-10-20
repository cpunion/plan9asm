'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {jobBody, requireCompleteWindowsNeeds} = require('./workflow-needs.cjs');

// Independent complete inventory, not a prefix that could silently select a
// subset. Keep the original official implicit-multiply owners in their lane.
const nativeRoots = [
  'TestX86ImplicitMultiplyNativeMemoryIndependentGoEncoding',
  'TestX86ImplicitMultiplyNativeMemoryInitializedThreeAPIs',
  'TestX86ImplicitMultiplyNativeMemoryDeadAccessAndExactWidths',
  'TestX86ImplicitMultiplyNativeMemoryInputsStatusAndAuthority',
  'TestX86ImplicitMultiplyNativeMemoryEveryByteAndExactOutputState',
  'TestX86ImplicitMultiplyNativeMemoryLinuxAMD64Objects',
  'TestX86ImplicitMultiplyNativeMemoryDarwinAMD64Objects',
  'TestX86ImplicitMultiplyNativeMemoryWindowsAMD64Objects',
  'TestX86ImplicitMultiplyNativeMemoryLinux386Objects',
  'TestX86ImplicitMultiplyNativeMemoryWindows386Objects',
];
const cliRoots = [
  'TestCompileOneOriginalSDKMathExpScalarShiftAndNumericalSource',
  'TestCompileOneOriginalSDKFieldImplicitMultiplySource',
];
const groups = [
  ['implicit-multiply-native-inputs', "'.'", nativeRoots],
  ['llvm-original-math-field', 'cmd/plan9asmll', cliRoots],
];
const blocks = groups.map(([lane, module, roots]) =>
  `          - lane: ${lane}\n            module: ${module}\n            roots:\n` +
  roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m');
const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');

function checkOwners(text) {
  assert.deepEqual(groups.map(([, , roots]) => roots.length), [10, 2]);
  const roots = groups.flatMap(([, , roots]) => roots);
  assert.equal(new Set(roots).size, 12, 'disjoint exact native and original-file roots');
  for (let index = 0; index < groups.length; index++) {
    const [lane, , owners] = groups[index];
    const marker = `          - lane: ${lane}\n`;
    assert.equal(text.split(marker).length - 1, 1, 'one independent bounded lane: ' + lane);
    const after = text.slice(text.indexOf(marker));
    const end = after.slice(marker.length).search(/^          - lane:|^    steps:/m);
    const block = end < 0 ? after : after.slice(0, marker.length + end);
    assert.equal(block.trimEnd(), blocks[index], 'complete literal lane, module and five-minute budget: ' + lane);
    for (const root of owners) {
      assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete owner: ' + root);
    }
  }
}

function testRoots(file) {
  const source = fs.readFileSync(file, 'utf8');
  return [...source.matchAll(/^func (Test\w+)\(t \*testing\.T\)/gm)].map(match => match[1]);
}

test('native multiply priority exactly covers all ten new source state and five-target object roots', () => {
  const actual = ['input', 'effect', 'objects'].flatMap(role =>
    testRoots(`x86_implicit_multiply_native_${role}_test.go`));
  assert.equal(actual.length, 10);
  assert.equal(new Set(actual).size, 10);
  assert.deepEqual(actual.sort(), [...nativeRoots].sort());
  const original = testRoots('cmd/plan9asmll/x86_sdk_math_shift_go127_test.go');
  assert.deepEqual(original.sort(), [...cliRoots].sort());
});

test('native multiply and both original rich CLI files have independent complete five-minute owners', () => {
  const sdk = jobBody(workflow, 'sdk-priority');
  checkOwners(sdk);
  const previous = sdk.slice(sdk.indexOf('          - lane: scalar-fp-byte-spans\n'));
  assert.ok(previous.includes('              - TestX86ImplicitMultiplyCompleteOfficialGoRows\n'));
  requireCompleteWindowsNeeds(workflow);
});

test('multiply and original-file priorities reject missing duplicate substituted partial module and budget changes', () => {
  const fixture = blocks.join('\n');
  checkOwners(fixture);
  for (let index = 0; index < groups.length; index++) {
    const [lane, module, roots] = groups[index];
    const block = blocks[index];
    for (const changed of [
      '',
      block + '\n' + block,
      block.replace(lane, 'unrelated-owner'),
      block.replace(`module: ${module}`, `module: ${index === 0 ? 'cmd/plan9asmll' : "'.'"}`),
      block.replace('timeout: 5m', 'timeout: 6m'),
      block.replace('            timeout: 5m',
        '              - TestUnrelatedAdditionalRoot\n            timeout: 5m'),
      ...roots.flatMap(root => [
        block.replace(`              - ${root}\n`, ''),
        block.replace(`              - ${root}\n`, `              - ${root}\n              - ${root}\n`),
        block.replace(root, 'TestUnrelatedSubstituteRoot'),
        block.replace(root, root + '/partial'),
      ]),
    ]) {
      assert.throws(() => checkOwners(fixture.replace(block, changed)), lane);
    }
  }
});
