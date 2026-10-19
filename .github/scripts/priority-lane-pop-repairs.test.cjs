'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const roots = [
  'TestX86ScalarLaneNativeStoresSelectedBytesAndOrigins',
  'TestX86ScalarLaneNativeStoresFiveTargets',
  'TestX86ScalarLaneNativeStoreVoidRequiresSelectedInputs',
  'TestX86ScalarLaneNativeResultsKeepWholeSourceBoundaries',
  'TestX86PrivatePOPNativeStoresFiveTargets',
  'TestX86PrivatePOPNativeStoreExactBitsAndOrigins',
  'TestX86ResultValueStackCompleteWidthsAndBoundaries',
  'TestX86ResultValueStackPackedBitDefinedness',
  'TestAMD64ExtractPSRuntimeSemantics',
  'TestAMD64StackWidthRuntimeValuesAndFlags',
  'TestCrossLinuxRuntimeMatrixX86ScalarLaneCompleteBytes',
  'TestCrossLinuxRuntimeMatrixX86InsertExtractByteSemantics',
  'TestTranslateX86ExtractPSCompleteGo127Forms',
  'TestTranslateAMD64VExtractPSCompleteGo127Forms',
  'TestTranslateAMD64VectorLaneExtractCompleteGoAssemblerForms',
  'TestTranslateAMD64PackedScalarExtractCompleteGoAssemblerForms',
];
const lane = 'scalar-lane-pop-repairs';
const block = `- lane: ${lane}\n            module: '.'\n            roots:\n` +
  roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function validate(source) {
  assert.equal(roots.length, 16);
  assert.equal(new Set(roots).size, 16);
  assert.ok(source.includes(block), 'complete typed family, missing-input and original execution repairs');
  assert.equal(source.split(`- lane: ${lane}\n`).length - 1, 1);
  for (const root of roots) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
}

test('lane and POP repairs execute complete original failures before full Windows ownership', () => {
  validate(fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8'));
});

test('lane and POP priority rejects absent, duplicate, narrowed or enlarged owners', () => {
  validate(block);
  for (const changed of [
    '', block + '\n' + block,
    block.replace(lane, 'other-lane'),
    block.replace("module: '.'", 'module: cmd/plan9asmll'),
    block.replace('timeout: 5m', 'timeout: 6m'),
    ...roots.flatMap(root => [
      block.replace(`              - ${root}\n`, ''),
      block + `\n              - ${root}\n`,
      block.replace(root, root + '/subset'),
    ]),
  ]) assert.throws(() => validate(changed));
});
