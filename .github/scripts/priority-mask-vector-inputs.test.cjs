'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const owners = {
  'mask-move-complete-values': {module: '.', roots: [
    'TestX86MaskMoveGPResultFullWordDefinedness', 'TestX86MaskMoveGPResultFiveTargetsThreeAPIs',
    'TestX86MaskMoveGPResultCannotInventInputOrBoolCanonicality', 'TestX86MaskMoveLogicalFPExactReadAndWriteWidths',
    'TestX86MaskMoveLogicalFPExactBytesPortableRuntime', 'TestX86NumericalEffectConsumesOnlyCheckedCallerContinuation',
  ]},
  'mask-arithmetic-complete-values': {module: '.', roots: [
    'TestX86MaskArithmeticCompletePortableValues', 'TestX86MaskArithmeticNumericalInputsAndState',
    'TestX86MaskArithmeticRejectsNonIntegerImmediateConsumers', 'TestX86MaskArithmeticVoidCannotInventInputs',
    'TestX86MaskArithmeticCompleteResultFiveTargetsThreeAPIs',
  ]},
  'mask-native-and-predicate-inputs': {module: '.', roots: [
    'TestX86MaskNativeEmitterPreservesExactVolatileWidths', 'TestX86MaskNativeCompleteFiveTargetsThreeAPIs',
    'TestX86MaskNativeCannotSupplyDeclarationOrMissingInputs', 'TestX86NativeConditionalSetExactFlagsAndState',
    'TestX86NativeConditionalSetCompleteFiveTargetsThreeAPIs', 'TestX86NativeConditionalSetCannotInventFlagsOrAddress',
    'TestX86ConditionalSetVoidRequiresActualPredicateFlags',
  ]},
  'vector-native-memory-complete': {module: '.', roots: [
    'TestX86VectorValueNativeCompleteFiveTargetsThreeAPIs', 'TestX86VectorValueNativeMissingInputsStayContext',
    'TestX86VectorValueNativeDeadEmitterKeepsInputRead', 'TestX86VectorValueVoidInputsCannotBorrowImplementationSlots',
  ]},
  'immediate-control-and-frame-hashing': {module: '.', roots: [
    'TestX86ImmediateControlHelpersRetainIntegerBoundaries', 'TestX86ImmediateControlHelpersRejectNonIntegerRoles',
    'TestX86GFNIImmediateFloatZeroRejectedLikeGo', 'TestGoAMD64FrameInputVerificationHasBoundedAllocations',
    'TestGoAMD64FrameInputVerificationReadsTailDespiteSameMetadata', 'TestGoAMD64FrameInputVerificationRetainsInputRole',
  ]},
  'sdk-gc-filter-mask-width': {module: 'cmd/plan9asmll', roots: [
    'TestCompileOneX86SDKGCFilterMaskWidthExistingObjects',
  ]},
  'typed-value-input-coverage': {module: 'cmd/plan9asmscan', roots: [
    'TestMaskValueCoverageRequiresCompleteInputs', 'TestVectorValueCoverageRequiresCompleteInputs',
  ]},
};
const dependencies = 'needs: [ownership, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, vector-value-fault-priority]';
const block = (lane, owner) => `- lane: ${lane}\n            module: '${owner.module}'\n            roots:\n` +
  owner.roots.map(root => `              - ${root}\n`).join('') + '            timeout: 5m';

function requireInputFamilies(source) {
  for (const [lane, owner] of Object.entries(owners)) {
    assert.ok(source.includes(block(lane, owner)), `complete ${lane} ownership`);
    assert.equal(source.split(`- lane: ${lane}\n`).length-1, 1);
    for (const root of owner.roots) assert.equal(source.split(`- ${root}\n`).length-1, 1);
  }
  assert.ok(source.includes(dependencies));
  assert.doesNotMatch(source, /max-parallel:/);
  const marker = '\n  vector-value-fault-priority:\n';
  assert.equal(source.split(marker).length-1, 1);
  const start = source.indexOf(marker);
  const end = source.indexOf('\n  ', start+marker.length);
  const job = source.slice(start, end < 0 ? undefined : end);
  assert.match(job, /runs-on: ubuntu-latest/);
  assert.match(job, /scripts\/install-ci-llvm22-ubuntu.sh/);
  assert.match(job, /go-version: '1\.27\.2'/);
  assert.match(job, /roots='TestX86VectorValueNativePortableFaults'/);
  assert.match(job, /roots\+='\|TestGoX86NumericalSBReadRuntimeFaultCorePolicy'/);
  assert.match(job, /-run "\^\(\$roots\)\$" -count=1 -timeout=5m -json/);
  assert.match(job, /audit-value-fault-priority\.cjs/);
  assert.doesNotMatch(job, /continue-on-error:|allow-skips|\|\| true/);
}

test('complete mask, predicate, immediate, frame and vector inputs have bounded first-party owners', () => {
  requireInputFamilies(workflow);
});

test('input-family priority rejects omissions, duplicates, sampling and larger command budgets', () => {
  requireInputFamilies(workflow);
  for (const [lane, owner] of Object.entries(owners)) {
    const original = block(lane, owner);
    for (const root of owner.roots) {
      const line = `              - ${root}\n`;
      for (const changed of [original.replace(line, ''), original.replace(root, root+'/sample'),
        original.replace(line, line+line)]) {
        assert.throws(() => requireInputFamilies(workflow.replace(original, changed)));
      }
    }
    for (const changed of ['', original.replace('timeout: 5m', 'timeout: 6m'),
      original.replace(`module: '${owner.module}'`, 'module: other'),
      original.replace('            timeout:', '              - TestUnrelatedOwner\n            timeout:')]) {
      assert.throws(() => requireInputFamilies(workflow.replace(original, changed)));
    }
    assert.throws(() => requireInputFamilies(workflow+'\n          '+original));
  }
  for (const changed of [workflow.replace(', vector-value-fault-priority]', ']'),
    workflow.replace('roots=\'TestX86VectorValueNativePortableFaults\'', 'roots=\'TestUnrelatedFault\''),
    workflow.replace('roots+=\'|TestGoX86NumericalSBReadRuntimeFaultCorePolicy\'', ''),
    workflow+'\n    max-parallel: 8\n']) {
    assert.throws(() => requireInputFamilies(changed));
  }
});
