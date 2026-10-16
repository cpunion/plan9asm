const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const sdk = workflow.slice(workflow.indexOf('\n  sdk-priority:'), workflow.indexOf('\n  ownership:'));
const sdkLaneCount = 55;

const crc32ProfileOwners = [
  ['llvm-crc32', 'TestCompileOneX86SDKOriginalCRC32NumericalDataJoin'],
  ['llvm-crc32-effective-defaults', 'TestCompileOneX86SDKOriginalCRC32NumericalDataJoinEffectiveDefaults'],
];
const crc32ProfileBlocks = crc32ProfileOwners.map(([lane, root]) =>
  `- lane: ${lane}\n            module: cmd/plan9asmll\n            roots:\n              - ${root}\n            timeout: 5m`);

function checkCRC32ProfileOwners(text) {
  for (let index = 0; index < crc32ProfileOwners.length; index++) {
    const [lane, root] = crc32ProfileOwners[index];
    assert.ok(text.includes(crc32ProfileBlocks[index]), 'complete independent CRC32 profile: ' + lane);
    assert.equal(text.split(`- lane: ${lane}\n`).length - 1, 1, 'one CRC32 profile lane: ' + lane);
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete CRC32 profile owner: ' + root);
  }
}

test('CRC32 actual profiles each retain an independent original five-minute owner', () => {
  checkCRC32ProfileOwners(sdk);
});

test('CRC32 profile priority rejects missing, shared, duplicate, substituted or enlarged owners', () => {
  const valid = crc32ProfileBlocks.join('\n          ');
  checkCRC32ProfileOwners(valid);
  for (let index = 0; index < crc32ProfileOwners.length; index++) {
    const [lane, root] = crc32ProfileOwners[index];
    const block = crc32ProfileBlocks[index];
    for (const changed of [
      valid.replace(block, ''),
      valid + '\n          ' + block,
      valid.replace(`- lane: ${lane}\n`, '- lane: unrelated-crc32-profile\n'),
      valid + `\n              - ${root}\n`,
      valid.replace(root, root + 'Subset'),
      valid.replace(root, root + '/effective-defaults'),
      valid.replace(block, block.replace('module: cmd/plan9asmll', "module: '.'")),
      valid.replace(block, block.replace('timeout: 5m', 'timeout: 6m')),
      valid.replace(block, block.replace('            timeout: 5m',
        '              - TestUnrelatedAdditionalOwner\n            timeout: 5m')),
      valid.replace(block, block.replace(root, crc32ProfileOwners[(index + 1) % crc32ProfileOwners.length][1])),
    ]) assert.throws(() => checkCRC32ProfileOwners(changed));
  }
});

const crc32TypedRoots = [
  'TestX86CRC32WholeResultNUMThreeAPIs',
  'TestX86CRC32PairedNativeReadVolatile',
  'TestX86CRC32FPExactSourceBytes',
  'TestX86CRC32SparseNUMFlagsAndSourceBoundaries',
  'TestX86CRC32GrammarBanksAndOriginalIdentity',
  'TestX86CRC32PhysicalByteNormalizationAndScanner',
  'TestX86CRC32UnownedMemoryAndProvenanceThreeAPIs',
  'TestX86CRC32CompleteOfficialGoRows',
  'TestX86CRC32OriginalGoByteViews',
  'TestX86IncomingPointerResidueMasksRolesAndByteAliases',
  'TestX86IncomingPointerResidueWholeSourceThreeAPIs',
  'TestX86IncomingPointerUnaryAdvanceWholeSourceThreeAPIs',
  'TestX86IncomingPointerUnaryAdvanceRolesFlagsAndOrigins',
  'TestX86CRC32CompleteOriginalGoRegisterBanks',
  'TestX86CRC32FiveTargetOriginalGoAndLLVMObjects',
  'TestX86TypedNumericalMixedBoolCanonicalCompositionThreeAPIs',
  'TestX86TypedNumericalMixedBoolRequiresLiveCachedCanonicalRole',
];
const crc32TypedLane = '- lane: crc32-typed-effects\n            module: \'.\'\n            roots:\n' +
  crc32TypedRoots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function checkCRC32TypedPriority(text) {
  assert.equal(crc32TypedRoots.length, 17);
  assert.equal(new Set(crc32TypedRoots).size, 17);
  assert.ok(text.includes(crc32TypedLane), 'complete CRC32 typed effects, pointer composition and actual-object owners');
  assert.equal(text.split('- lane: crc32-typed-effects\n').length - 1, 1,
    'one independent CRC32 typed-effects priority lane');
  for (const root of crc32TypedRoots) {
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete literal CRC32 typed owner: ' + root);
  }
}

test('CRC32 typed effects priority retains seventeen complete semantic and object owners within five minutes', () => {
  checkCRC32TypedPriority(sdk);
});

test('CRC32 typed effects priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkCRC32TypedPriority(crc32TypedLane);
  for (const changed of [
    '',
    crc32TypedLane.replace('crc32-typed-effects', 'unrelated-crc32-effects'),
    crc32TypedLane + '\n          ' + crc32TypedLane,
    crc32TypedLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    crc32TypedLane.replace('timeout: 5m', 'timeout: 6m'),
    crc32TypedLane.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
    ...crc32TypedRoots.flatMap(root => [
      crc32TypedLane.replace(`              - ${root}\n`, ''),
      crc32TypedLane + `\n              - ${root}\n`,
      crc32TypedLane.replace(root, 'TestUnrelatedSubstituteOwner'),
      crc32TypedLane.replace(root, root + '/amd64'),
    ]),
  ]) assert.throws(() => checkCRC32TypedPriority(changed));
});

const scalarFPByteRoots = [
  'TestX86ScalarFPWideTupleStoresPreserveEveryField',
  'TestX86ScalarFPWideTupleReadsPreserveEveryField',
  'TestX86ScalarFPIntegerPartialAndInteriorBytesThreeAPIs',
  'TestX86IntegerFPByteViewsKeepIndependentRoles',
  'TestX86IntegerFPByteViewsMode32QwordKeepsVectorConsumer',
  'TestX86IntegerFPByteViewsPreserveStatusAndDirection',
  'TestX86IntegerFPByteViewsDescriptorCannotCreateStorage',
  'TestX86IntegerFPByteViewsReadonlyInputsAndFloatingBits',
  'TestX86IntegerFPByteViewsSparseMustInitialization',
  'TestX86IntegerFPByteViewsMOVQSourceAliases',
  'TestX86OrdinaryIntegerFPStoreWidthsRequireWholeSource',
  'TestX86OrdinaryIntegerFPDefaultArchitectureCannotBypassProof',
  'TestX86OrdinaryIntegerFPStoreKeepsIndependentStorageConsumers',
  'TestX86ScalarFPWideTupleSpanOriginalGoAssembly',
  'TestX86IntegerFPByteViewsFiveTargetOriginalAndLLVMObjects',
  'TestX86ImplicitMultiplyWholeSourceThreeAPIs',
  'TestX86ImplicitMultiplyNUMBytesAndStatus',
  'TestX86ImplicitMultiplyDoesNotRegisterRawEntryOrDivide',
  'TestX86ImplicitMultiplyFPSourceBytesThreeAPIs',
  'TestX86ImplicitMultiplyMissingFPResultsThreeAPIs',
  'TestX86ImplicitMultiplyCompleteOfficialGoRows',
  'TestX86ImplicitMultiplyGrammarBanksAndMemoryBoundaries',
  'TestX86ImplicitMultiplyAliasesReadBeforeWriteAndFlags',
  'TestX86ImplicitMultiplyOriginalGoGrammarAndEncoding',
  'TestX86ImplicitMultiplyFiveTargetOriginalGoAndLLVMObjects',
  'TestX86ImplicitMultiplyIndependentStorageCannotBypassNUM',
  'TestTranslate386InstructionFamilies',
];
const scalarFPByteLane = '- lane: scalar-fp-byte-spans\n            module: \'.\'\n            roots:\n' +
  scalarFPByteRoots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function checkScalarFPBytePriority(text) {
  assert.equal(scalarFPByteRoots.length, 27);
  assert.equal(new Set(scalarFPByteRoots).size, 27);
  assert.ok(text.includes(scalarFPByteLane), 'complete FP byte, implicit multiply and whole386 semantic/control and actual-object owners');
  assert.equal(text.split('- lane: scalar-fp-byte-spans\n').length - 1, 1,
    'one bounded FP byte priority lane');
  for (const root of scalarFPByteRoots) {
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete literal FP byte owner: ' + root);
  }
}

test('Scalar FP byte priority retains twenty-seven complete semantic and object owners within five minutes', () => {
  checkScalarFPBytePriority(sdk);
});

test('Scalar FP byte priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkScalarFPBytePriority(scalarFPByteLane);
  for (const changed of [
    '',
    scalarFPByteLane.replace('scalar-fp-byte-spans', 'unrelated-fp-byte-spans'),
    scalarFPByteLane + '\n          ' + scalarFPByteLane,
    scalarFPByteLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    scalarFPByteLane.replace('timeout: 5m', 'timeout: 6m'),
    scalarFPByteLane.replace('            timeout: 5m',
      '              - TestX86ABI0ResultTupleScalarTypesAndObjects\n            timeout: 5m'),
    ...scalarFPByteRoots.flatMap(root => [
      scalarFPByteLane.replace(`              - ${root}\n`, ''),
      scalarFPByteLane + `\n              - ${root}\n`,
      scalarFPByteLane.replace(root, 'TestUnrelatedSubstituteOwner'),
      scalarFPByteLane.replace(root, root + '/386'),
    ]),
  ]) assert.throws(() => checkScalarFPBytePriority(changed));
});

const stringLODSRoots = [
  'TestX86StringImplicitMemoryIndependentFamilyEffects',
  'TestX86StringImplicitHelpersKeepObservableNativeAccesses',
  'TestX86StringImplicitHelpersGuardZeroCountBeforeMemory',
  'TestX86StringImplicitMemoryDoesNotGrantResultOrPointerAuthority',
  'TestX86StringRepeatCFGAcceptsOnlyUnambiguousAdjacency',
  'TestX86StringRepeatCFGRejectsBypassedOrAmbiguousPrefix',
  'TestX86StringRepeatCFGFixtureKeepsActualBranchAndPrefixNodes',
  'TestX86StringRepeatCFGPortGrammarRemainsIndependent',
  'TestX86StringImplicitMemoryOriginalGoAndLLVMO0O2',
  'TestX86StringRepeatCFGOriginalGoKeepsCompleteFixtures',
  'TestX86LODSStateCompleteIndependentCases',
  'TestX86LODSZeroCountDoesNotCreateResultFacts',
  'TestX86LODSZeroStoreObserverRejectsEveryFieldAndUnknownEffect',
  'TestX86LODSZeroStoreObserverVolatileAndLineEndingForms',
  'TestX86LODSZeroCountOriginalGoAndLLVMWholeState',
  'TestX86StringMemoryIRReaderEmptyZeroCount',
  'TestX86StringMemoryIRReaderLineEndingsAndMetadata',
  'TestX86StringMemoryIRReaderRejectsMemoryMutations',
];
const stringLODSLane = '- lane: string-memory-lods-state\n            module: \'.\'\n            roots:\n' +
  stringLODSRoots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function checkStringLODSPriority(text) {
  assert.equal(stringLODSRoots.length, 18);
  assert.equal(new Set(stringLODSRoots).size, 18);
  assert.ok(text.includes(stringLODSLane), 'all new string memory/CFG and LODS state owners');
  assert.equal(text.split('- lane: string-memory-lods-state\n').length - 1, 1,
    'one independent string/LODS priority lane');
  for (const root of stringLODSRoots) {
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete literal string/LODS owner: ' + root);
  }
}

test('String LODS priority retains eighteen complete source, state and reader owners within five minutes', () => {
  checkStringLODSPriority(sdk);
});

test('String LODS priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkStringLODSPriority(stringLODSLane);
  for (const changed of [
    '',
    stringLODSLane.replace('string-memory-lods-state', 'unrelated-string-state'),
    stringLODSLane + '\n          ' + stringLODSLane,
    stringLODSLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    stringLODSLane.replace('timeout: 5m', 'timeout: 6m'),
    stringLODSLane.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
    ...stringLODSRoots.flatMap(root => [
      stringLODSLane.replace(`              - ${root}\n`, ''),
      stringLODSLane + `\n              - ${root}\n`,
      stringLODSLane.replace(root, 'TestUnrelatedSubstituteOwner'),
      stringLODSLane.replace(root, root + '/386'),
    ]),
  ]) assert.throws(() => checkStringLODSPriority(changed));
});

const ordinaryBPRoots = [
  'TestX86LegacyPackedFloatingNativeMemoryUnboundSourcesFiveTargets',
  'TestX86LegacyLogicalMemoryVirtualStackNeedsCertificate',
  'TestX86WordMultiplyAddLegacyVirtualMemoryCannotProveAlignment',
  'TestX86PackedWordMultiplyVirtualLegacyMemoryRemainsContext',
  'TestX86PackedLowMultiplyVirtualLegacyStorageRemainsContext',
  'TestX86PackedConversionOriginalNativeAddressRequired',
  'TestX86ByteShuffleVirtualLegacyStorageRemainsContext',
  'TestX86SignedPackedFloatAllWidthsZeroMaskStorageContext',
  'TestX86CMOVFamilyNativeStackAndFPAuthorityUnchanged',
  'TestX86OrdinaryBPNativeMemoryCompleteFamiliesGoAndLLVMObjects',
  'TestX86OrdinaryBPCMOVCompleteFamilyGoAndLLVMObjects',
];
const ordinaryBPLane = '- lane: ordinary-bp-native-read\n            module: \'.\'\n            roots:\n' +
  ordinaryBPRoots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function checkOrdinaryBPPriority(text) {
  assert.equal(ordinaryBPRoots.length, 11);
  assert.equal(new Set(ordinaryBPRoots).size, 11);
  assert.ok(text.includes(ordinaryBPLane), 'all nine original BP fixtures and both complete object owners');
  assert.equal(text.split('- lane: ordinary-bp-native-read\n').length - 1, 1,
    'one bounded BP family priority lane, not eleven new jobs');
  for (const root of ordinaryBPRoots) {
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete owner per literal root: ' + root);
  }
}

test('Ordinary BP priority keeps the complete eleven-owner family in its original five-minute budget', () => {
  checkOrdinaryBPPriority(sdk);
});

test('Ordinary BP priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkOrdinaryBPPriority(ordinaryBPLane);
  for (const changed of [
    '',
    ordinaryBPLane.replace('ordinary-bp-native-read', 'unrelated-native-read'),
    ordinaryBPLane + '\n          ' + ordinaryBPLane,
    ordinaryBPLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    ordinaryBPLane.replace('timeout: 5m', 'timeout: 6m'),
    ordinaryBPLane.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
    ...ordinaryBPRoots.flatMap(root => [
      ordinaryBPLane.replace(`              - ${root}\n`, ''),
      ordinaryBPLane + `\n              - ${root}\n`,
      ordinaryBPLane.replace(root, 'TestUnrelatedSubstituteOwner'),
      ordinaryBPLane.replace(root, root + '/i386-unknown-linux-gnu'),
    ]),
  ]) assert.throws(() => checkOrdinaryBPPriority(changed));
});

const tupleRoot = 'TestX86ABI0ResultTupleScalarTypesAndObjects';
const tupleLane = '- lane: abi0-result-tuples\n            module: \'.\'\n' +
  `            roots:\n              - ${tupleRoot}\n            timeout: 5m`;

function checkTuplePriority(text) {
  assert.ok(text.includes(tupleLane), 'complete ABI0 tuple object owner with its original budget');
  assert.equal(text.split('- lane: abi0-result-tuples\n').length - 1, 1,
    'exactly one independent tuple object priority lane');
  assert.equal(text.split(`- ${tupleRoot}\n`).length - 1, 1, 'exactly one complete tuple owner');
}

test('Tuple priority retains the complete five-target owner before full Windows modules', () => {
  checkTuplePriority(sdk);
});

test('Tuple priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkTuplePriority(tupleLane);
  for (const changed of [
    '',
    tupleLane.replace('abi0-result-tuples', 'unrelated-result-tuples'),
    tupleLane + '\n          ' + tupleLane,
    tupleLane + `\n              - ${tupleRoot}\n`,
    tupleLane.replace(tupleRoot, 'TestX86ABI0ResultTupleEveryByteAndPath'),
    tupleLane.replace(tupleRoot, tupleRoot + '/i386-unknown-linux-gnu'),
    tupleLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    tupleLane.replace('timeout: 5m', 'timeout: 6m'),
    tupleLane.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
  ]) assert.throws(() => checkTuplePriority(changed));
});

const compareRoot = 'TestGoX86SDKAMD64CompareTypedArgumentsWholeSource';
const compareLane = '- lane: typed-arguments-compare\n            module: \'.\'\n' +
  `            roots:\n              - ${compareRoot}\n            timeout: 5m`;

function checkTypedArgumentsComparePriority(text) {
  assert.ok(text.includes(compareLane), 'complete original Compare owner with its original budget');
  assert.equal(text.split('- lane: typed-arguments-compare\n').length - 1, 1,
    'exactly one independent Compare priority lane');
  assert.equal(text.split(`- ${compareRoot}\n`).length - 1, 1, 'exactly one complete Compare owner');
}

test('Complete typed-argument Compare objects precede full Windows modules with the original budget', () => {
  checkTypedArgumentsComparePriority(sdk);
});

test('Compare priority rejects missing, duplicate, substituted, partial or enlarged owners', () => {
  checkTypedArgumentsComparePriority(compareLane);
  for (const changed of [
    compareLane.replace('typed-arguments-compare', 'unrelated-compare'),
    compareLane + '\n          ' + compareLane,
    compareLane + `\n              - ${compareRoot}\n`,
    compareLane.replace(compareRoot, 'TestGoX86SDKAMD64EqualTypedArgumentsWholeSource'),
    compareLane.replace(compareRoot, compareRoot + '/text'),
    compareLane.replace("module: '.'", 'module: cmd/plan9asmll'),
    compareLane.replace('timeout: 5m', 'timeout: 6m'),
    compareLane.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
  ]) assert.throws(() => checkTypedArgumentsComparePriority(changed));
});

const compareAPIPartitions = [
  ['typed-arguments-compare', 'TestGoX86SDKAMD64CompareTypedArgumentsWholeSource'],
  ['typed-arguments-compare-module', 'TestGoX86SDKAMD64CompareTypedArgumentsWholeSourceModule'],
  ['typed-arguments-compare-go-module', 'TestGoX86SDKAMD64CompareTypedArgumentsWholeSourceGoModule'],
];
const compareAPIBlocks = compareAPIPartitions.map(([lane, root]) =>
  `- lane: ${lane}\n            module: '.'\n            roots:\n              - ${root}\n            timeout: 5m`);

function checkCompareAPIPartitions(text) {
  for (let index = 0; index < compareAPIPartitions.length; index++) {
    const [lane, root] = compareAPIPartitions[index];
    assert.ok(text.includes(compareAPIBlocks[index]), 'complete original Compare API partition: ' + lane);
    assert.equal(text.split(`- lane: ${lane}\n`).length - 1, 1, 'one independent API lane: ' + lane);
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, 'one complete original-source API owner: ' + root);
  }
}

test('Complete Compare API partitions each retain an independent original five-minute owner', () => {
  checkCompareAPIPartitions(sdk);
});

test('Compare API partition policy rejects missing, shared, substituted or enlarged owners', () => {
  const valid = compareAPIBlocks.join('\n          ');
  checkCompareAPIPartitions(valid);
  for (let index = 0; index < compareAPIPartitions.length; index++) {
    const [lane, root] = compareAPIPartitions[index];
    const block = compareAPIBlocks[index];
    for (const changed of [
      valid.replace(block, ''),
      valid + '\n          ' + block,
      valid.replace(`- lane: ${lane}\n`, '- lane: unrelated-compare-api\n'),
      valid + `\n              - ${root}\n`,
      valid.replace(root, compareRoot + 'Subset'),
      valid.replace(root, root + '/text'),
      valid.replace(block, block.replace("module: '.'", 'module: cmd/plan9asmll')),
      valid.replace(block, block.replace('timeout: 5m', 'timeout: 6m')),
      valid.replace(block, block.replace('            timeout: 5m',
        '              - TestUnrelatedAdditionalOwner\n            timeout: 5m')),
      valid.replace(block, block.replace(root, compareAPIPartitions[(index + 1) % compareAPIPartitions.length][1])),
    ]) assert.throws(() => checkCompareAPIPartitions(changed));
  }
});

function checkBytealgPriority(text) {
  for (const [lane, root] of [
    ['llvm-bytealg-count', 'TestCompileOneX86SDKBytealgCountExistingObjects'],
    ['llvm-bytealg-indexbyte', 'TestCompileOneX86SDKBytealgIndexByteExistingObjects'],
  ]) {
    const block = `- lane: ${lane}\n            module: cmd/plan9asmll\n            roots:\n              - ${root}\n            timeout: 5m`;
    assert.ok(text.includes(block), lane);
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, root);
  }
  assert.doesNotMatch(text, /- TestCompileOneX86SDKBytealgExistingObjects\n/);
}

test('Bytealg original Count and IndexByte each precede full modules with their own budget', () => {
  checkBytealgPriority(sdk);
});

test('Bytealg priority rejects a missing, duplicate, substituted or enlarged owner', () => {
  for (const [before, after] of [
    ['- lane: llvm-bytealg-count\n', '- lane: missing-bytealg-count\n'],
    ['TestCompileOneX86SDKBytealgCountExistingObjects', 'TestCompileOneX86SDKBytealgExistingObjects'],
    ['TestCompileOneX86SDKBytealgIndexByteExistingObjects', 'TestCompileOneX86SDKBytealgCountExistingObjects'],
    ['TestCompileOneX86SDKBytealgIndexByteExistingObjects\n            timeout: 5m',
      'TestCompileOneX86SDKBytealgIndexByteExistingObjects\n            timeout: 6m'],
  ]) assert.throws(() => checkBytealgPriority(sdk.replace(before, after)));
});

const carrierRoot = 'TestARM64ByteCarrierTwoScalarNativeWidthConsumers';
const carrierBlock = '- lane: arm64-carrier-widths\n            module: \'.\'\n' +
  `            roots:\n              - ${carrierRoot}\n            timeout: 5m`;

function checkCarrierPriority(text) {
  assert.ok(text.includes(carrierBlock), 'complete failed ARM64 carrier owner');
  assert.equal(text.split(`- ${carrierRoot}\n`).length - 1, 1);
}

test('Failed ARM64 native-width source matrix precedes complete Windows modules', () => {
  checkCarrierPriority(sdk);
});

test('Carrier priority rejects missing, duplicate, substituted or enlarged owners', () => {
  checkCarrierPriority(carrierBlock);
  for (const changed of [
    carrierBlock.replace('arm64-carrier-widths', 'missing-carrier-widths'),
    carrierBlock + `\n              - ${carrierRoot}\n`,
    carrierBlock.replace(carrierRoot, carrierRoot + 'Subset'),
    carrierBlock.replace('timeout: 5m', 'timeout: 6m'),
  ]) assert.throws(() => checkCarrierPriority(changed));
});

test('New panic observations retain both complete actual compiler and linker owners', () => {
  const block = '- lane: panic-observation\n            module: \'.\'\n            roots:\n' +
    '              - TestX86GoPanicTerminal386OriginalABICompileAndLink\n' +
    '              - TestX86GoPanicTerminalOriginal386CompilerAndCALLAssociations\n            timeout: 5m';
  assert.ok(sdk.includes(block));
});

test('SDK priority runs each previous simple root in its own bounded lane', () => {
  const roots = [
    'TestTranslateOriginalSDKCRC32NumericalSource',
    'TestTranslateOriginalSDKCRC32NumericalSourceNonLinux',
    'TestSimpleCLIOriginalSDKCRC32PublicationAndRejection',
    'TestSimpleCLIOriginalSDKMapsScalarRegisterSource',
  ];
  assert.match(sdk, /fail-fast: false/);
  assert.match(sdk, /name: SDK priority \(\$\{\{ matrix.lane \}\}\)/);
  for (const root of roots) {
    assert.match(sdk, new RegExp(`module: cmd/plan9asm\\n            roots:\\n              - ${root}\\n            timeout: 8m`));
  }
  assert.match(sdk, /module: cmd\/plan9asmll\n[\s\S]*?timeout: 5m/);
  assert.match(sdk, /module: '\.'\n[\s\S]*?timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
  assert.match(sdk, /- lane: frame-context\n[\s\S]*?TestGoAMD64SDKFrameBindContextCallbacksMayClose/);
  assert.match(sdk, /- lane: frame-context\n[\s\S]*?TestGoX86NumericalSBReadFrameProducerObserverCancellationCause/);
  assert.match(sdk, /- lane: source\n[\s\S]*?TestGoX86SDKNumericalNativeOriginalMixedTextOrder/);
  assert.match(sdk, /- lane: source\n[\s\S]*?TestGoX86SDKNumericalNativeTextOrderLiveRevocation/);
});

test('Expensive CRC32 and the complete failed selection class have independent original budgets', () => {
  const llvm = sdk.slice(sdk.indexOf('- lane: llvm\n'), sdk.indexOf('- lane: llvm-crc32\n'));
  assert.doesNotMatch(llvm, /TestCompileOneX86SDKOriginalCRC32NumericalDataJoin/);
  const crc32 = sdk.slice(sdk.indexOf('- lane: llvm-crc32\n'), sdk.indexOf('- lane: assembly-selection\n'));
  assert.match(crc32, /module: cmd\/plan9asmll/);
  assert.match(crc32, /roots:\n              - TestCompileOneX86SDKOriginalCRC32NumericalDataJoin\n            timeout: 5m/);
  const selection = sdk.slice(sdk.indexOf('- lane: assembly-selection\n'), sdk.indexOf('- lane: crc32-linux\n'));
  assert.match(selection, /module: cmd\/plan9asmll/);
  assert.match(selection, /timeout: 5m/);
  for (const root of [
    'TestActualAssemblyMemberExactSelection',
    'TestActualAssemblyMemberOriginalAliasLifetime',
    'TestActualAssemblyMemberFinalSymlinkIdentity',
    'TestARM64ABI0TailActualSelectedMemberDirectoryAlias',
    'TestARM64ABI0TailActualSelectedFinalSymlink',
    'TestARM64ABI0TailOriginalAliasPostGuard',
    'TestCompileOneARM64ABI0TailActualDeclaration',
    'TestCompileOneARM64ABI0TailCompleteAndChunks',
    'TestCompileOneARM64ABI0TailSourcePostGuard',
    'TestReflectDynamicStubNativeObjectsRetainContext',
    'TestReflectDynamicStubSelectedAssemblyIdentity',
    'TestReflectDynamicStubSelectedAssemblyRejectsUnboundIdentity',
  ]) {
    assert.match(selection, new RegExp(`- ${root}\\n`));
  }
});

test('Every lane retains exact inventory and an independent artifact', () => {
  assert.match(sdk, /SDK_PRIORITY_ROOTS: \$\{\{ join\(matrix.roots, '\|'\) \}\}/);
  assert.match(sdk, /SDK_PRIORITY_TIMEOUT: \$\{\{ matrix.timeout \}\}/);
  assert.match(sdk, /working-directory: \$\{\{ matrix.module \}\}/);
  assert.match(sdk, /grep -Fx "\$root" "\$report\/inventory.log"/);
  assert.match(sdk, /-timeout="\$SDK_PRIORITY_TIMEOUT"/);
  assert.match(sdk, /name: windows-sdk-priority-\$\{\{ matrix.lane \}\}/);
  assert.match(workflow, /needs: \[ownership, native-x86-oracles, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, unlocked-rmw-native-priority, vector-value-fault-priority\]/);
  assert.doesNotMatch(sdk, /continue-on-error|max-parallel|go test[^\n]* -skip/);
});

test('Source priority includes the complete reader and child-harness class without another lane', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  assert.match(source, /module: '\.'/);
  assert.match(source, /timeout: 5m/);
  for (const root of [
    'TestReadGoAssemblySourceInputsShareExactExpansion',
    'TestReadGoAssemblySourceInputsRetainRequestedAliases',
    'TestReadGoAssemblySourceReportSeparatesFallbackSearchMiss',
    'TestReadGoAssemblySourceReportErrorReturnsNoInventory',
    'TestReadGoAssemblySourceExpandsToolchainHeadersInTrimpathBinary',
    'TestReadGoAssemblySourceRejectsEscapedSourceAlias',
    'TestReadGoAssemblySourceExpandsNestedLocalHeaders',
    'TestReadGoAssemblySourceRejectsSymlinkIncludeOutsideRoot',
    'TestAssemblyIncludeSymlinkRegressionIsAvailableOnSupportedHosts',
    'TestFrameworkChildGoBuildRetainsParentLLVMMode',
    'TestX86ResultGPSIMDPlainUnknownReturnNeedsWholeFunctionProof',
    'TestX86ResultGPSIMDAdmissionScalarReturnFamilyThreeAPIs',
    'TestX86ResultGPSIMDAdmissionCompleteOfficialRegisterRows',
    'TestX86ResultGPSIMDAdmissionOrdinaryRolesAndCachedProof',
    'TestX86ResultGPSIMDAdmissionWholeSourcePathsWithoutProbe',
  ]) {
    assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  }
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('New source mechanisms have their own exact priority lanes', () => {
  const additions = [
    ['literal-minimum-simple', 'cmd/plan9asm', 'TestSimpleCLIOriginalSDKMapsLiteralMinimumSource', '8m'],
    ['literal-minimum-llvm-linux', 'cmd/plan9asmll', 'TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceLinux', '5m'],
    ['literal-minimum-llvm-linux-public-cli', 'cmd/plan9asmll', 'TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceLinuxPublicCLI', '5m'],
    ['literal-minimum-llvm-darwin', 'cmd/plan9asmll', 'TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceDarwin', '5m'],
    ['literal-minimum-llvm-windows', 'cmd/plan9asmll', 'TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceWindows', '5m'],
    ['metadata-only', 'internal/sdkfrontend', 'TestSharedSDKMetadataOriginalEmptyDATACompleteThreeOS', '10m'],
    ['nodata-count-text', 'internal/sdkfrontend', 'TestSharedSDKNoDATACountText', '5m'],
    ['nodata-count-module', 'internal/sdkfrontend', 'TestSharedSDKNoDATACountModule', '5m'],
    ['nodata-count-go-module', 'internal/sdkfrontend', 'TestSharedSDKNoDATACountGoModule', '5m'],
    ['nodata-indexbyte-text', 'internal/sdkfrontend', 'TestSharedSDKNoDATAIndexByteText', '5m'],
    ['nodata-indexbyte-module', 'internal/sdkfrontend', 'TestSharedSDKNoDATAIndexByteModule', '5m'],
    ['nodata-indexbyte-go-module', 'internal/sdkfrontend', 'TestSharedSDKNoDATAIndexByteGoModule', '5m'],
  ];
  for (const [lane, module, root, timeout] of additions) {
    assert.ok(sdk.includes(`- lane: ${lane}\n            module: ${module}\n            roots:\n              - ${root}\n            timeout: ${timeout}`), lane);
  }
  const publication = sdk.slice(sdk.indexOf('- lane: publication\n'), sdk.indexOf('- lane: crc32-linux\n'));
  assert.match(publication, /module: cmd\/plan9asmll/);
  assert.match(publication, /timeout: 5m/);
  for (const root of [
    'TestCompileOneCallerLateSourceGuardDoesNotPublish',
    'TestCompileOneARM64TailLateSourceGuardDoesNotPublish',
    'TestCompileOneFixedLocalLateOwnerGuardDoesNotPublish',
    'TestCompileOneContinuationLateReferenceDoesNotPublish',
    'TestCompileOutputTransactionRollbackAndRestoreFailure',
    'TestCompileOutputTransactionNonretainedObjectPolicy',
  ]) assert.ok(publication.includes(`- ${root}\n`), root);
});

test('Complete maps target owners are disjoint and retain the original budget', () => {
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
  for (const goos of ['linux', 'darwin', 'windows']) {
    const suffix = goos[0].toUpperCase() + goos.slice(1);
    const root = `TestCompileOneX86SDKOriginalMapsLiteralMinimumSource${suffix}`;
    assert.equal(sdk.split(`- ${root}\n`).length - 1, 1, root);
    assert.ok(sdk.includes(`- lane: literal-minimum-llvm-${goos}\n            module: cmd/plan9asmll\n            roots:\n              - ${root}\n            timeout: 5m`));
  }
  assert.doesNotMatch(sdk, /- TestCompileOneX86SDKOriginalMapsLiteralMinimumSource\n/);
  assert.equal(sdk.split('- TestCompileOneX86SDKOriginalMapsLiteralMinimumSourceLinuxPublicCLI\n').length - 1, 1);
  assert.equal(sdk.split('- TestMapsLiteralMinimumTargetsHaveIndependentCompleteOwners\n').length - 1, 1);
});

test('Typed MXCSR conversions and authenticated proof joins have separate bounded lanes', () => {
  const required = {
    'typed-matrix': [
      'TestX86MXCSRCompleteOriginalGoRows',
      'TestX86MXCSRByteEffectHasNoIndependentMemoryAuthority',
      'TestX86MXCSRFullMemoryWidthAndOperandGrammar',
      'TestX86MXCSRPlainAUTOReturnNeedsReachingBytes',
      'TestX86MXCSRAUTOExtentKeepsCompleteSourceAndBounds',
      'TestX86SignedIntegerFloatResultPreservesSparseMergeBytes',
      'TestX86SignedIntegerFloatUnknownUpperResultThreeAPIs',
      'TestX86SignedIntegerFloatPlainMXCSRReturnRequiresWholeSource',
      'TestX86SignedIntegerFloatPairedNativeReadRetainsExactDeadWidth',
      'TestX86SignedIntegerFloatOriginalFPFixtureWholeThreeAPIs',
      'TestX86MXCSRConversionCompleteFiveTargetOriginalGoAndLLVMObjects',
    ],
    'proof-join-guards': [
      'TestX86ResultAUTONumericalJoinHasIndependentCompleteOwners',
      'TestX86ResultAUTONumericalJoinOwnerPolicyRejectsMissingOrSharedAxes',
      'TestX86ResultAUTOReplayDoesNotWashUnboundSourceRoles',
      'TestX86FixedLocalFeatureProjectionUsesCompleteSourceOrder',
      'TestGoAMD64FixedLocalResultConsumerExactProofHandoff',
      'TestGoX86ResultAUTONumericalReplayExactOriginalNodeGuards',
    ],
    'proof-join-original': ['TestGoX86ResultAUTOCompleteActualFixedLocalNumericalJoinOriginalSourceShape'],
    'proof-join-text': ['TestGoX86ResultAUTOCompleteActualFixedLocalNumericalJoinText'],
    'proof-join-context': ['TestGoX86ResultAUTOCompleteActualFixedLocalNumericalJoinContext'],
    'proof-join-go-module': ['TestGoX86ResultAUTOCompleteActualFixedLocalNumericalJoinGoModule'],
  };
  for (const [lane, roots] of Object.entries(required)) {
    const start = sdk.indexOf(`- lane: ${lane}\n`);
    assert.notEqual(start, -1, lane);
    const end = sdk.indexOf('- lane:', start + 1);
    const block = sdk.slice(start, end === -1 ? sdk.length : end);
    assert.match(block, /module: '\.'/);
    assert.match(block, /timeout: 5m/);
    const actual = [...block.matchAll(/- (Test\w+)\n/g)].map(match => match[1]);
    assert.deepEqual(actual.sort(), roots.sort());
    for (const root of roots) assert.equal(sdk.split(`- ${root}\n`).length - 1, 1, root);
  }
});

test('Native source policies retain CRLF equivalence and all independent owners', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestNativeOwnerPoliciesKeepCRLFSourceAndMutations',
    'TestNativeOwnerPolicyNormalizationIsLimitedToCRLF',
    'TestX86ResultGPSIMDSparseRuntimeNativeOwnerCannotEmulateOrSkip',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
});

test('Guarded bit-scan and typed predicate rule classes are exact source priorities', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ScalarBitScanCompleteOfficialRowsAndOperandGrammar',
    'TestX86ABI0ResultBitScanZeroBranchDefinesOnlyNonzeroEdge',
    'TestX86ABI0ResultBitScanUndefinedValuesAndFlagsRemainContext',
    'TestX86ABI0ResultBitScanGuardedByteAliasesAndCFGJoins',
    'TestX86ScalarBitScanWholeSourceThreeAPIAdmission',
    'TestX86ABI0ResultBitScanPendingStateCannotLeakAuthority',
    'TestX86ABI0ResultBitScanPortableStatusFlagConsumers',
    'TestX86ResultPredicateNonzeroComparisonMaskBitScan',
    'TestX86ResultPredicateNonzeroCannotBorrowOtherValueOrWidth',
    'TestX86ResultPredicateLegacyAndVEXCompleteComparisonFamilies',
    'TestX86ResultPredicateWholeSourceThreeAPIsAndMode32',
    'TestX86ResultPredicateBoundedStateAndRawRevocation',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Caller-address differences retain their exact source-origin priority controls', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ResultCallerAddressDifferenceCancelsSameTypedOrigin',
    'TestX86ResultCallerAddressDifferenceUnknownOrMixedOriginsRemainContext',
    'TestX86ResultCallerAddressDifferenceThreeAPIWholeSourceAdmission',
    'TestX86ResultCallerAddressDifferenceCFGAndRevocation',
    'TestX86ResultCallerAddressOriginIsBoundedIndependentMustState',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Bit-scan zero semantics and complete object cohort cannot wait behind full modules', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ScalarBitScanNativeHelperShapesAndTargetBoundaries',
    'TestX86ScalarBitScanNarrowRegisterAliasRetainsUnknownUpperBytes',
    'TestX86ScalarBitScanNamespaceAndMissingCompilerBridge',
    'TestX86ResultBitScanZeroPreservesOnlyPriorNumericalBytes',
    'TestX86ScalarBitScanInitializedPriorThreeAPIs',
    'TestX86ScalarBitScanUnknownPriorStillNeedsZeroEdgeContext',
    'TestX86ScalarBitScanOriginalGoHexWordOpsFunctionThreeAPIs',
    'TestX86ScalarBitScanOriginalGoHexWordOpsSourceLineEndings',
    'TestX86ScalarBitScanCompleteObjectFixtureNumericalRoles',
    'TestX86ScalarBitScanCompleteOriginalGoAndLLVMFiveTargetObjects',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Canonical bool calls retain complete source, byte, lifetime and budget priorities', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ABI0Mode32CanonicalBoolCallOriginalIdentity',
    'TestX86ABI0Mode32CanonicalBoolCallPlanStateAndCache',
    'TestX86ABI0Mode32CanonicalBoolSameFileCallThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolCallFinalCallbacks',
    'TestX86ABI0Mode32CanonicalBoolCallKeepsWholeSource',
    'TestX86ABI0Mode32CanonicalBoolCallExistingFloatPredicate',
    'TestX86ABI0Mode32CanonicalBoolCallRevokesAllDerivedDomains',
    'TestX86ABI0Mode32CanonicalBoolCallDerivedDomainsThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolExistingWholeFrameThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolTypedInputAndPrivateAlias',
    'TestX86ABI0Mode32CanonicalBoolParameterRawBytesRemainPhysical',
    'TestX86ABI0Mode32CanonicalBoolAliasCloneJoinAndBudget',
    'TestX86ABI0Mode32CanonicalBoolByteResult',
    'TestX86ABI0Mode32BoolThreeAPIsCannotBypassCanonicalProof',
    'TestX86ABI0Mode32BoolIntermediateBytesSurviveThreeAPIs',
    'TestX86ABI0Mode32BoolWholeCFGAndOverwriteThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolRoleAndSourceIdentity',
    'TestX86ABI0Mode32CanonicalFPCloneJoinAndBudget',
    'TestX86ABI0Mode32BoolPreflightUsesOneBoundedWholeProof',
    'TestX86ABI0Mode32InitializedByteIsNotCanonicalBool',
    'TestX86ABI0Mode32CanonicalFactDomainIsIndependent',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Complete bit-scan native axes and their authority controls are mandatory', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  for (const root of [
    'TestCrossLinuxRuntimeMatrixX86BitScanNativeAMD64OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86BitScanNative386OriginalGoAndLLVM',
  ]) assert.ok(native.includes(root), root);
  assert.match(native, /go test \. -list "\^\(\$bit_scan_roots\)\$"/);
  assert.match(native, /grep -Fx "\$root" _out\/native-x86-oracles\/bit-scan-inventory.log/);
  assert.match(native, /roots\+="\|\$bit_scan_roots"/);
  assert.match(native, /-timeout=8m/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ScalarBitScanExecutionOracleDoesNotInventZeroOrFlags',
    'TestX86ScalarBitScanExecutionOracleCompleteWidthsAliasesAndRecords',
    'TestX86ScalarBitScanExecutionCompleteOriginalGoLinksAndLLVMObjects',
    'TestX86ScalarBitScanExecutionPhysicalStatusShimKeepsPublicRoles',
    'TestX86ScalarBitScanExecutionStatusObservationDoesNotAuthorizePublicFlags',
    'TestX86ScalarBitScanRuntimeNativeOwnersCannotEmulateOrSkip',
    'TestX86ScalarBitScanRuntimeNativePolicyRejectsMutation',
    'TestX86ABI0Mode32CanonicalBoolCallSelectionKeepsNonboolGoFilter',
    'TestX86ABI0Mode32CanonicalBoolCallSelectionKeepsNonboolWholeFiles',
    'TestX86ABI0Mode32CanonicalBoolCallSelectionRetainsTrueBoolLimit',
    'TestX86ABI0Mode32CanonicalBoolCallSelectionRejectsChangingQuery',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('ABI0 mode32 bool must execute its sealed native owner', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  assert.match(native, /bool_386_root='TestCrossLinuxRuntimeMatrixX86ABI0Mode32CanonicalBoolOriginalGoAndLLVM'/);
  assert.match(native, /go test \. -list "\^\(\$bool_386_root\)\$"/);
  assert.match(native, /grep -Fx "\$bool_386_root" _out\/native-x86-oracles\/bool-386-inventory.log/);
  assert.match(native, /roots\+="\|\$bool_386_root"/);
  assert.match(native, /-timeout=8m/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ABI0Mode32BoolRuntimeNativeOwnerCannotEmulateOrSkip',
    'TestX86ABI0Mode32BoolRuntimeNativePolicyRejectsMutation',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Packed string result effects retain their independent native AVX owner', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  const root = 'TestCrossLinuxRuntimeMatrixX86PackedStringImplicitVectorNativeAVX';
  assert.ok(native.includes(`packed_string_root='${root}'`));
  assert.ok(native.includes('go test . -list "^($packed_string_root)$" > _out/native-x86-oracles/packed-string-inventory.log'));
  assert.ok(native.includes('grep -Fx "$packed_string_root" _out/native-x86-oracles/packed-string-inventory.log'));
  assert.ok(native.includes('roots+="|$packed_string_root"'));
  assert.match(native, /-timeout=8m/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
});

test('Typed bool terminals and distinct custom register proofs are mandatory source priorities', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ABI0Mode32CanonicalBoolTypedTailThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolTypedTailSourceReplayAndLayouts',
    'TestX86ABI0Mode32CanonicalBoolTypedTailCannotBorrowGoCaller',
    'TestX86ABI0Mode32CanonicalBoolTypedTailCacheAndKeepSource',
    'TestX86ABI0Mode32CanonicalBoolTypedTailFinalCallbacks',
    'TestX86Mode32CustomRegisterBoolCanonicalThreeAPIs',
    'TestX86Mode32CustomRegisterBoolCompareExchangeEffects',
    'TestX86Mode32CustomRegisterBoolRoleAndByteViews',
    'TestX86Mode32CustomRegisterBoolPairedEffectHasNoNewAuthority',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Sparse GP SIMD must execute both actual native architectures', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  for (const root of [
    'TestCrossLinuxRuntimeMatrixX86GPSIMDSparseNativeAMD64OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86GPSIMDSparseNative386OriginalGoAndLLVM',
  ]) assert.ok(native.includes(root), root);
  assert.match(native, /go test \. -list "\^\(\$gp_simd_sparse_roots\)\$"/);
  assert.match(native, /grep -Fx "\$root" _out\/native-x86-oracles\/gp-simd-sparse-inventory.log/);
  assert.match(native, /roots\+="\|\$gp_simd_sparse_roots"/);
  assert.match(native, /-timeout=8m/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
});

test('Void calls and text context ownership regressions run before complete modules', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86ABI0Mode32CanonicalBoolVoidCallThreeAPIs',
    'TestX86ABI0Mode32CanonicalBoolVoidCallHasNoResultOrOutgoingBytes',
    'TestX86ABI0Mode32CanonicalBoolActualAtomicCompleteSourceNextContext',
    'TestTranslateTextOwnsRequestLLVMContext',
    'TestTranslateTextConcurrentRequestsKeepOriginalFile',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Complete direction control and string-consumer regressions are mandatory', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86DirectionControlNumericalAndCallerThreeAPIs',
    'TestX86DirectionControlCompleteGoGrammarAndRawSeparation',
    'TestX86DirectionStateCloneJoinPreservationAndRevocation',
    'TestX86DirectionWholeSourceCFGJoinsAndFinalEffectRevocation',
    'TestX86DirectionMalformedDeadSourceCannotBorrowNumericalProof',
    'TestX86DirectionLoweredStringCompleteFamilyConsumesDirection',
    'TestX86DirectionControlPreservesSixStatusButGrantsNoEntryFacts',
    'TestX86DirectionOriginalSDKIndexByte386ThreeAPIsNextContext',
    'TestX86DirectionOriginalGo127CompleteYnoneRows',
    'TestX86ABI0ResultInitializationCompleteStateByteEnvelope',
    'TestX86ABI0Mode32CanonicalBoolDirectCallRevokesDirection',
    'TestX86DirectionTypedBoolAndVoidWholeCallThreeAPIs',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Ordinary result aliases require the complete class and target objects', () => {
  const source = sdk.slice(sdk.indexOf('- lane: source\n'), sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestX86OrdinaryResultAddressWholeSource',
    'TestX86OrdinaryResultAddressNoProbeRejectsUnknown',
    'TestX86OrdinaryResultAddressScalarTransferWidths',
    'TestX86OrdinaryResultAddressPrivateProofNotPublicAuthority',
    'TestX86OrdinaryResultAddressNativeAliasOrdering',
    'TestX86OrdinaryResultAddressPhysicalBanksAndFieldBytes',
    'TestX86OrdinaryResultAddressKeepsUnclosedRoles',
    'TestX86OrdinaryResultAddressLayoutAndBudgetRemainExact',
    'TestX86OrdinaryResultAddressDistinctFieldsAndRevocation',
    'TestX86OrdinaryResultAddressConcurrentOriginalFile',
    'TestX86OrdinaryResultAddressOriginalGoAndLLVMFiveTargetObjects',
    'TestX86ResultSlotMoveWidthKeepsScalarMode32Boundary',
    'TestX86ResultSlotMoveWidthWholeProofAndThreeAPIs',
  ]) assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  assert.match(source, /timeout: 5m/);
  assert.equal((sdk.match(/- lane:/g) || []).length, sdkLaneCount);
});

test('Actual failed SDK frame paths run before the complete-module barrier', () => {
  const llvm = sdk.slice(sdk.indexOf('- lane: llvm'), sdk.indexOf('- lane: crc32-linux'));
  for (const root of [
    'TestRunOneTargetSDKFixedLocalPreservesActualDefaultProfile',
    'TestRunOneTargetSDKFixedLocalCanonicalCommandPaths',
  ]) {
    assert.match(llvm, new RegExp(`- ${root}\\n`));
  }
  assert.match(llvm, /timeout: 5m/);
  const shared = sdk.slice(sdk.indexOf('- name: Check shared command-path identities'), sdk.indexOf('- name: Check exact SDK priority inventory'));
  assert.match(shared, /if: \$\{\{ matrix.lane == 'llvm' \}\}/);
  assert.match(shared, /working-directory: internal\/sdkfrontend/);
  for (const root of [
    'TestOrdinaryEnvironmentActualCommandPathIdentity',
    'TestOrdinaryEnvironmentCommandPathRejectionPreservesInputs',
    'TestOrdinaryEnvironmentRetainsGOPATHComponentSelection',
  ]) {
    assert.match(shared, new RegExp(root));
  }
  assert.match(shared, /grep -Fx "\$root" "\$report\/shared-path-inventory.log"/);
  assert.match(shared, /-timeout=5m/);
});

test('Caller-address CMP requires both independently inventoried native axes', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  for (const root of [
    'TestCrossLinuxRuntimeMatrixX86CallerAddressCMPInstructionBytesAndLLVM',
    'TestCrossLinuxRuntimeMatrixGoX86CallerAddressCMPABI0Representation',
  ]) {
    assert.match(native, new RegExp(root));
  }
  assert.match(native, /go test \. -list "\^\(\$caller_cmp_roots\)\$"/);
  assert.match(native, /grep -Fx "\$root" _out\/native-x86-oracles\/caller-cmp-inventory.log/);
  assert.match(native, /roots\+="\|\$caller_cmp_roots"/);
});

test('Numerical LEA requires its actual native owner, not an emulator overlay', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  assert.match(native, /numerical_lea_root='TestCrossLinuxRuntimeMatrixX86NumericalLEAOriginalBytesAndLLVM'/);
  assert.match(native, /go test \. -list "\^\(\$numerical_lea_root\)\$"/);
  assert.match(native, /grep -Fx "\$numerical_lea_root" _out\/native-x86-oracles\/numerical-lea-inventory.log/);
  assert.match(native, /roots\+="\|\$numerical_lea_root"/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
});

test('Direct-tail state executes both native x86 modes with exact mandatory roots', () => {
  const native = workflow.slice(workflow.indexOf('\n  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  for (const root of [
    'TestCrossLinuxRuntimeMatrixAMD64DirectTailWholeSourceOriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86DirectTail386OriginalGoAndLLVM',
    'TestCrossLinuxRuntimeMatrixX86DirectTailPCResultAliasesAndExternal',
  ]) assert.ok(native.includes(root), root);
  assert.match(native, /go test \. -list "\^\(\$direct_tail_roots\)\$"/);
  assert.match(native, /grep -Fx "\$root" _out\/native-x86-oracles\/direct-tail-inventory.log/);
  assert.match(native, /roots\+="\|\$direct_tail_roots"/);
  assert.match(native, /run: bash scripts\/install-ci-llvm22-ubuntu\.sh gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.doesNotMatch(native, /sudo apt-get update|sudo apt-get install/);
  assert.doesNotMatch(native, /qemu|overlay|continue-on-error|max-parallel/);
});

test('Input publication and terminal context counterexamples run before full modules', () => {
  const publication = sdk.slice(sdk.indexOf('- lane: publication\n'), sdk.indexOf('- lane: crc32-linux\n'));
  for (const root of [
    'TestCompileOutputRejectsSourceExtensionCaseVariants',
    'TestCompileOneOutputCannotReplaceCaseVariantSelectedGoSource',
    'TestCompileOneOutputCannotReplaceOriginalModuleMetadata',
    'TestCompileOneOutputCannotCreateAbsentMetadataAlias',
    'TestCompileOutputMetadataNameProtectionUsesActualParent',
    'TestCompileOneOutputCannotReplaceActualSelectedEmbed',
    'TestCompileOneLateSelectedEmbedReplacementDoesNotPublish',
    'TestCompileOutputInputDirectoryAliasLifetime',
    'TestCompileOutputRechecksOriginalInputsDuringPublication',
    'TestCompileOutputInputIdentityExcludesEquivalentExistingName',
    'TestCompileOutputInputProtectionAccumulatesAttemptMembers',
    'TestCompileOutputSuccessfullyReadInputCannotBecomeMissingRole',
    'TestCompileOutputInputAliasLifetime',
    'TestCompileOutputSearchMissProtection',
    'TestCompileOutputSearchMissCannotBePublicDestination',
    'TestCompileOneLateEqualByteIncludeRedirectDoesNotPublish',
    'TestCompileOneTaskBuilderOutputCannotReplaceActualInclude',
    'TestFeatureCPPReadReportRetainsOnlyActiveSearchMisses',
  ]) assert.ok(publication.includes(`- ${root}\n`), root);
  const contexts = sdk.slice(sdk.indexOf('- lane: frame-context\n'));
  for (const root of [
    'TestGoSourceContextTerminalSuccessHasNoLaterCallbacks',
    'TestGoSourceContextTerminalLostBound',
    'TestGoSourceContextTerminalCaptureCannotFreezeStaleLifetime',
    'TestGoSourceContextTerminalOriginalErrCannotChangeLifetime',
    'TestGoSourceContextTerminalJoinedOwnerCallbackCannotChangeLifetime',
    'TestGoSourceContextTerminalCancellationSignalsAndCauses',
    'TestGoSourceContextTerminalClosedDoneCannotReturnSuccess',
    'TestGoSourceContextTerminalSameContextRolesAreBounded',
    'TestGoSourceContextTerminalFrameDoneCannotChangeInLastErr',
    'TestGoSourceContextTerminalPrivateBridgeOnlyTightensAndStops',
    'TestGoSourceContextTerminalIdentityAndBoundedRoleGuards',
  ]) assert.ok(contexts.includes(`- ${root}\n`), root);
  assert.match(publication, /timeout: 5m/);
  assert.match(contexts, /timeout: 5m/);
});

test('Caller selected-member lifetime also runs before complete Windows modules', () => {
  const selection = sdk.slice(sdk.indexOf('- lane: assembly-selection\n'), sdk.indexOf('- lane: crc32-linux\n'));
  for (const root of [
    'TestX86CallerCLIActualSelectedMemberDirectoryAlias',
    'TestX86CallerCLIOriginalAliasFinalSourceGuard',
    'TestX86CallerCLIActualSelectedFinalSymlink',
    'TestX86CallerCLIOriginalAliasFiveTargetObjects',
  ]) {
    assert.match(selection, new RegExp(`- ${root}\\n`));
  }
  assert.match(selection, /timeout: 5m/);
});

test('Both actual Windows SDKs execute the complete shared path class', () => {
  const ownership = workflow.slice(workflow.indexOf('\n  ownership:'));
  assert.match(ownership, /go: \['1.20.14', '1.27.2'\]/);
  const paths = ownership.slice(ownership.indexOf('- name: Check actual shared command paths on both SDKs'), ownership.indexOf('- name: Real process-tree'));
  assert.match(paths, /go test -tags=byollvm \.\/internal\/sdkfrontend -list/);
  assert.match(paths, /grep -Fx "\$root" _out\/windows-ownership\/command-path-inventory.log/);
  assert.match(paths, /-timeout=5m/);
  for (const root of [
    'TestGoEnvFileSelection',
    'TestOrdinaryEnvironmentActualCommandPathIdentity',
    'TestOrdinaryEnvironmentCommandPathRejectionPreservesInputs',
    'TestOrdinaryEnvironmentRetainsGOPATHComponentSelection',
  ]) {
    assert.match(paths, new RegExp(root));
  }
});

test('Paired physical directories retain SDK membership and outside-SDK rejection', () => {
  const llvm = sdk.slice(sdk.indexOf('- lane: llvm'), sdk.indexOf('- lane: crc32-linux'));
  for (const root of [
    'TestOrdinaryPackageLoadSDKWorkingDirectoryIdentity',
    'TestOrdinaryPackageLoadOutsideSDKRetainsInternalImportRejection',
  ]) {
    assert.match(llvm, new RegExp(`- ${root}\\n`));
  }
  const shared = sdk.slice(sdk.indexOf('- name: Check shared command-path identities'), sdk.indexOf('- name: Check exact SDK priority inventory'));
  const ownership = workflow.slice(workflow.indexOf('- name: Check actual shared command paths on both SDKs'), workflow.indexOf('- name: Real process-tree'));
  for (const root of [
    'TestWorkingDirectoryActualIdentity',
    'TestWorkingDirectoryRejectsMissingOrNonDirectory',
  ]) {
    assert.match(shared, new RegExp(root));
    assert.match(ownership, new RegExp(root));
  }
});
