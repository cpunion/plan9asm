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
];
const block = '- lane: completed-ci-repair\n            module: \'.\'\n            roots:\n' +
  roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';
const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');

function validate(source) {
  assert.equal(roots.length, 23);
  assert.equal(new Set(roots).size, 23);
  assert.ok(source.includes(block), 'all fourteen original failures, eight new class roots and retained load/store owner must run first');
  assert.equal(source.split('- lane: completed-ci-repair\n').length - 1, 1);
}

test('complete old failures and independent fixes have one mandatory bounded priority', () => {
  validate(workflow);
});

test('failed-root priority rejects removal, duplication, substitution, partial roots and widened budgets', () => {
  for (const changed of [
    block.replace("module: '.'", 'module: cmd/plan9asmll'),
    block.replace('timeout: 5m', 'timeout: 6m'),
    block + '\n' + block,
    ...roots.flatMap(root => [
      block.replace(`              - ${root}\n`, ''),
      block.replace(root, root + '/partial'),
      block.replace(root, 'TestUnrelatedFailure'),
    ]),
  ]) assert.throws(() => validate(changed));
});
