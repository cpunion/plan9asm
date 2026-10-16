const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

// Independent complete failed-root inventory from run38824961157, followed
// by the newly added class regressions. Never substitute a child or sample.
const roots = [
  'TestTranslateIRTextCoverage',
  'TestX86RawSegmentAbsoluteMemory',
  'TestX86ResultCompareExchangePrefixAndAuthorityBoundaries',
  'TestTranslate386RejectsInvalidInstructionForms',
  'TestX86RawReturnBoundFPAndStaticObjects',
  'TestX86RawReturnActualGoDeclarationContract',
  'TestX86RawAddressOverrideDefaultCallerControlsAndPhysicalSegments',
  'TestX86DirectSymbolMemorySinkObjectsRemainSupported',
  'TestTranslate386DestinationForms',
  'TestX86RawReturnStaticDATAExtents',
  'TestTranslateHelpersCoverage',
  'TestX86AddressClassPreservesActualMemoryAndRelocationObjects',
  'TestTranslateGoModule_X87Mode',
  'TestCrossLinuxRuntimeMatrixX86BitTestNumericalAndAliases',
  'TestX86PoppedFlagWordImmediateMaskThreeAPIs',
  'TestX86NarrowPoppedWordCannotDefineUnwrittenHighBits',
  'TestX86NativeBitTestCompleteReachingInputsThreeAPIs',
  'TestX86NativeBitTestExactStateAndIndependentAuthorities',
  'TestX86BitTestPortableNumericalAndAliases',
  'TestX86NativeGPStoreCompleteWidthsThreeAPIs',
  'TestX86NativeGPStoreCannotBorrowUnknownOrPointerData',
  'TestX86NativeGPStoreOriginalByteAliasesAndUnchangedState',
  'TestX86ScalarMoveNativeRetainsPublicAndSourceBoundaries',
  'TestX86SegmentMOVCannotGrantIndependentAuthority',
  'TestX86SegmentLoadAndNativeStoreHaveIndependentPairedEffects',
  'TestX86PoppedPartialFlagsCannotInventUnmodeledControls',
  'TestX86PoppedMixedWidthsRetainOriginalFlagBitPositions',
  'TestX86StackRefinementCloneJoinAndReuse',
];
const groups = [
  ['completed-ci-repair', roots.slice(0, 14)],
  ['native-memory-effects', roots.slice(14)],
];
const blocks = groups.map(([lane, owners]) => `- lane: ${lane}\n            module: '.'\n            roots:\n` +
  owners.map(root => `              - ${root}\n`).join('') + '            timeout: 5m');
const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');

function validate(source) {
  assert.equal(roots.length, 28);
  assert.equal(new Set(roots).size, 28);
  assert.deepEqual(groups.map(([, owners]) => owners.length), [14, 14]);
  for (let index = 0; index < groups.length; index++) {
    assert.ok(source.includes(blocks[index]), 'complete independent old-failure and native-effect priorities');
    assert.equal(source.split(`- lane: ${groups[index][0]}\n`).length - 1, 1);
  }
}

test('complete old failures and independent fixes have two disjoint bounded priorities', () => {
  validate(workflow);
});

test('failed-root priority rejects removal, duplication, substitution, partial roots and widened budgets', () => {
  const fixture = blocks.join('\n');
  validate(fixture);
  for (let index = 0; index < groups.length; index++) {
    const block = blocks[index];
    for (const changed of [
      block.replace("module: '.'", 'module: cmd/plan9asmll'),
      block.replace('timeout: 5m', 'timeout: 6m'),
      block + '\n' + block,
      ...groups[index][1].flatMap(root => [
        block.replace(`              - ${root}\n`, ''),
        block.replace(root, root + '/partial'),
        block.replace(root, 'TestUnrelatedFailure'),
      ]),
    ]) assert.throws(() => validate(fixture.replace(block, changed)));
  }
});
