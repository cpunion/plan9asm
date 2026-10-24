const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

// All three failed roots in run 38853831617 and independent class boundaries.
const roots = [
  'TestX86StaticIntegerSIMDSBStoresCompleteMoveFamily',
  'TestX86StaticIntegerSIMDSBStoresRetainInputResultAndExtentBoundaries',
  'TestX86ScalarFPLogicalNumericalSBStoreKeepsIndependentResult',
  'TestX86StaticIntegerResultDiscardedSBAddresses',
  'TestX86StaticIntegerSBAddressesDoNotAcquireNumericalOrStackRoles',
  'TestX86SymbolIndexedSBThreeAPIsFiveLLVMObjects',
  'TestTranslateAMD64MOVQFromXRegisterToFrameResult',
];
const lane = 'static-result-effects';
const block = `- lane: ${lane}\n            module: '.'\n            roots:\n` +
  roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function validate(source) {
  assert.equal(new Set(roots).size, 7);
  assert.ok(source.includes(block), 'all complete failed roots and class regressions');
  assert.equal(source.split(`- lane: ${lane}\n`).length - 1, 1);
  for (const root of roots) {
    assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  }
}

test('static result priority retains all failed roots and independent boundaries', () => {
  validate(fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8'));
});

test('static result priority rejects missing, duplicate, partial or widened owners', () => {
  validate(block);
  for (const changed of [
    '', block + '\n' + block,
    block.replace(lane, 'unrelated-lane'),
    block.replace("module: '.'", 'module: cmd/plan9asmll'),
    block.replace('timeout: 5m', 'timeout: 6m'),
    block.replace('            timeout: 5m',
      '              - TestUnrelatedOwner\n            timeout: 5m'),
    ...roots.flatMap(root => [
      block.replace(`              - ${root}\n`, ''),
      block + `\n              - ${root}\n`,
      block.replace(root, root + '/partial'),
      block.replace(root, 'TestUnrelatedOwner'),
    ]),
  ]) assert.throws(() => validate(changed));
});
