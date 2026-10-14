const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const sdk = workflow.slice(workflow.indexOf('\n  sdk-priority:'), workflow.indexOf('\n  ownership:'));

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
  assert.equal((sdk.match(/- lane:/g) || []).length, 13);
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
  assert.match(workflow, /needs: \[ownership, sdk-priority\]/);
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
    'TestFrameworkChildGoBuildRetainsParentLLVMMode',
    'TestX86ResultGPSIMDPlainUnknownReturnNeedsWholeFunctionProof',
    'TestX86ResultGPSIMDAdmissionScalarReturnFamilyThreeAPIs',
    'TestX86ResultGPSIMDAdmissionCompleteOfficialRegisterRows',
    'TestX86ResultGPSIMDAdmissionOrdinaryRolesAndCachedProof',
    'TestX86ResultGPSIMDAdmissionWholeSourcePathsWithoutProbe',
  ]) {
    assert.equal(source.split(`- ${root}\n`).length - 1, 1, root);
  }
  assert.equal((sdk.match(/- lane:/g) || []).length, 13);
});

test('New source mechanisms have their own exact priority lanes', () => {
  const additions = [
    ['literal-minimum-simple', 'cmd/plan9asm', 'TestSimpleCLIOriginalSDKMapsLiteralMinimumSource', '8m'],
    ['literal-minimum-llvm', 'cmd/plan9asmll', 'TestCompileOneX86SDKOriginalMapsLiteralMinimumSource', '5m'],
    ['metadata-only', 'internal/sdkfrontend', 'TestSharedSDKMetadataOriginalEmptyDATACompleteThreeOS', '10m'],
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
