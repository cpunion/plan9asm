const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const owners = {
  'compact-native-memory': [
    'TestX86CompactNativeMemorySequencesThreeAPIs',
    'TestX86CompactNativeMemoryCannotInventSourceOrDestinationFacts',
    'TestX86CompactNativeExpandClearsNonzeroUpperAliases',
    'TestX86PackedMoveInitializedNativeStoresWholeFamily',
    'TestAMD64PackedCompressRuntimeSemantics',
    'TestAMD64PackedExpandRuntimeSemantics',
    'TestX86PackedMoveInvalidStoreCannotAcquireCallerMemoryCertificate',
    'TestX86PackedZeroMemoryStorePreservesUndefinedOpcode',
    'TestX86PackedZeroMemoryStoreFaultPrecedesFPBoundsAndMaskReads',
  ],
  'complete-vector-value-families': [
    'TestX86IndexedPermuteInitializedResultWholeFamily',
    'TestX86IndexedPermuteResultInputsStayIndependent',
    'TestX86IndexedPermuteResultRequiresCompleteTablesIndicesAndMerge',
    'TestX86GFNIInitializedResultWholeFamily',
    'TestX86PackedBitCountInitializedResultWholeFamily',
    'TestX86FixedLaneValueResultInputsStayIndependent',
    'TestX86VectorValueFamiliesClearNonzeroUpperAliases',
    'TestX86VectorValueUpperAliasesPortableRuntime',
    'TestX86VectorValueMode32HighGPAddressesNeedPhysicalContext',
    'TestX86VectorValueMode32AddressContextDoesNotHideInvalidFormats',
    'TestTranslateAMD64AVX512Go126Families',
    'TestX86Go126FamiliesCannotBorrowUninitializedVectorInputs',
    'TestAMD64IndexedPermuteRuntimeSemantics',
    'TestAMD64GFNIRuntimeSemantics',
    'TestAMD64PackedPopcountRuntimeSemantics',
    'TestAMD64PackedLeadingZeroCountRuntimeSemantics',
  ],
};
const block = (lane, roots) => `- lane: ${lane}\n            module: '.'\n            roots:\n` +
  roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function requireCompleteFamilies(source) {
  for (const [lane, roots] of Object.entries(owners)) {
    assert.ok(source.includes(block(lane, roots)), `complete ${lane} source and runtime owners`);
    assert.equal(source.split(`- lane: ${lane}\n`).length - 1, 1);
    for (const root of roots) assert.equal(source.split(`- ${root}\n`).length - 1, 1);
  }
  assert.match(source, /needs: \[ownership, native-x86-oracles, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, unlocked-rmw-native-priority, vector-value-fault-priority\]/);
}

test('complete compact-memory and vector-value families run before full modules', () => {
  requireCompleteFamilies(workflow);
});

test('complete family priorities reject omitted, sampled, duplicate or expanded owners', () => {
  requireCompleteFamilies(workflow);
  for (const [lane, roots] of Object.entries(owners)) {
    const original = block(lane, roots);
    for (const root of roots) {
      const line = `              - ${root}\n`;
      for (const changed of [
        original.replace(line, ''),
        original.replace(root, root + '/sample'),
        original.replace(line, line + line),
      ]) assert.throws(() => requireCompleteFamilies(workflow.replace(original, changed)));
    }
    for (const changed of [
      '',
      original.replace("module: '.'", 'module: cmd/plan9asmll'),
      original.replace('timeout: 5m', 'timeout: 6m'),
      original.replace('            timeout:', '              - TestUnrelatedOwner\n            timeout:'),
    ]) assert.throws(() => requireCompleteFamilies(workflow.replace(original, changed)));
    assert.throws(() => requireCompleteFamilies(workflow + '\n          ' + original));
  }
});
