'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {jobBody, requireCompleteWindowsNeeds, requireJobNeeds} = require('./workflow-needs.cjs');

// Independent of the auditor and production instruction maps.
const roots = ['TestCrossLinuxRuntimeMatrixX86BlendMinposOriginalFaultsAMD64',
  'TestCrossLinuxRuntimeMatrixX86BlendMinposOriginalFaults386'];
const axes = ['backend_O0', 'backend_O2', 'IR_O2_backend_O2'];
const forms = [['BLENDVPS', 'X'], ['VBLENDVPS', 'X'], ['VBLENDVPS', 'Y'],
  ['BLENDVPD', 'X'], ['VBLENDVPD', 'X'], ['VBLENDVPD', 'Y'],
  ['PBLENDVB', 'X'], ['VPBLENDVB', 'X'], ['VPBLENDVB', 'Y'],
  ['PHMINPOSUW', 'X'], ['VPHMINPOSUW', 'X']];
const parents = roots.flatMap(root => axes.map(axis => root + '/' + axis));
const leaves = parents.flatMap(parent => forms.flatMap(([op, bank]) => ['live', 'dead'].flatMap(role =>
  [0, 1].map(zero => `${parent}/${op}/${bank}/${role}/maskzero=${zero}`))));
const names = [...roots, ...parents, ...leaves];
const inventory = roots.join('\n') + '\nok\tgithub.com/xgo-dev/plan9asm\n';
const digest = 'a'.repeat(64);
const host = `IA32 direct host: linux/amd64 kernel=x86_64 CPUsha=${digest} mapsha=${digest}\n`;
const preflight = 'blend/minimum actual native SSE4.1=1 AVX2=1\n';
const marker = 'blend/minimum original Go bytes and LLVM native alignment/access/CF/MXCSR=0x1f80/canary probes 20 failures 0\n';
const image = `native oracle static IA32 ELF bytes=4096 sha256=${digest}\n`;
const sdk = arch => `actual LEA source/object receipt: SDK=go1.27.2 target=linux/${arch} source=${digest} driver=${digest} object=${digest}\n`;
const event = (Action, Test, Output) => ({Action, Package: 'github.com/xgo-dev/plan9asm',
  ...(Test ? {Test} : {}), ...(Output !== undefined ? {Output} : {})});
const serialize = events => events.map(value => JSON.stringify(value)).join('\n');

function completeEvents() {
  const events = [event('start')];
  for (const root of roots) {
    const arch = root.endsWith('386') ? '386' : 'amd64';
    events.push(event('run', root), event('output', root, host + preflight + sdk(arch)));
    for (const axis of axes) {
      const parent = root + '/' + axis;
      events.push(event('run', parent));
      if (arch === '386') events.push(event('output', parent, image));
      for (const leaf of leaves.filter(name => name.startsWith(parent + '/'))) {
        events.push(event('run', leaf), event('output', leaf, marker), event('pass', leaf));
      }
      events.push(event('pass', parent));
    }
    events.push(event('output', root, host), event('pass', root));
  }
  events.push(event('pass'));
  return events;
}

function audit(events = completeEvents(), discovered = inventory) {
  assert.ok(fs.existsSync('.github/scripts/audit-blend-minpos-native.cjs'),
    'original blend/minimum native scope lacks an explicit strict auditor');
  return require('./audit-blend-minpos-native.cjs').auditBlendMinposNative(discovered, serialize(events));
}

test('blend/minimum requires two roots, 264 complete native kernels and 5280 paired probes', () => {
  assert.deepEqual(audit(), {roots: 2, children: 270, passed: 272, skipped: 0, incomplete: 0,
    nativeHosts: 4, sdkCaptures: 2, isaPreflights: 2, staticIA32Images: 3,
    kernelExecutions: 264, pairedProbes: 5280, directProbeProcesses: 10560});
});

test('blend/minimum rejects every missing duplicate failed skipped or incomplete lifecycle', () => {
  audit();
  for (const name of names) {
    for (const Action of ['fail', 'skip']) {
      assert.throws(() => audit(completeEvents().map(value => value.Test === name && value.Action === 'pass'
        ? {...value, Action} : value)), name + Action);
    }
    for (const Action of ['run', 'pass']) {
      const original = completeEvents();
      const at = original.findIndex(value => value.Test === name && value.Action === Action);
      assert.throws(() => audit(original.filter((_, index) => index !== at)), name + ' missing ' + Action);
      original.splice(at, 0, original[at]);
      assert.throws(() => audit(original), name + ' duplicate ' + Action);
    }
  }
});

test('blend/minimum binds all residues and zero-mask dead reads to their own actual completion', () => {
  audit();
  for (const leaf of leaves) {
    for (const replacement of ['', marker + marker, marker.replace('probes 20', 'probes 19')]) {
      assert.throws(() => audit(completeEvents().map(value => value.Test === leaf && value.Output
        ? {...value, Output: replacement} : value)), leaf);
    }
  }
  for (const replacement of [marker.replace('failures 0', 'failures 1'), marker.replace('0x1f80', 'inherited'),
    marker.replace('0x1f80', '0x21f80'), host + marker, preflight + marker, sdk('amd64') + marker, image + marker]) {
    assert.throws(() => audit(completeEvents().map(value => value.Test === leaves[0] && value.Output
      ? {...value, Output: replacement} : value)));
  }
  for (const target of [roots[0], parents[0]]) {
    assert.throws(() => audit(completeEvents().map(value => value.Test === leaves[0] && value.Output
      ? {...value, Test: target} : value)), 'summary is not actual leaf execution');
  }
});

test('blend/minimum requires native host ISA and exact original Go source before every axis', () => {
  audit();
  for (const root of roots) {
    const arch = root.endsWith('386') ? '386' : 'amd64';
    for (const needle of [host, preflight, sdk(arch)]) {
      for (const replacement of ['', needle + needle, needle.replace(digest, 'invalid'),
        needle.replace('kernel=x86_64', 'kernel=aarch64'), needle.replace('SSE4.1=1', 'SSE4.1=0'),
        needle.replace('AVX2=1', 'AVX2=0'), needle.replace('go1.27.2', 'go1.20.14'),
        needle.replace('linux/' + arch, 'linux/arm64')]) {
        if (replacement === needle) continue;
        assert.throws(() => audit(completeEvents().map(value => value.Test === root && value.Output
          ? {...value, Output: value.Output.replace(needle, replacement)} : value)), root + needle);
      }
    }
    const lateSDK = completeEvents();
    const at = lateSDK.findIndex(value => value.Test === root && value.Output?.includes(preflight));
    lateSDK[at] = event('output', root, host + preflight);
    lateSDK.splice(at + 2, 0, event('output', root, sdk(arch)));
    assert.throws(() => audit(lateSDK), 'source capture after axis start');
    const earlyPost = completeEvents();
    const post = earlyPost.findLastIndex(value => value.Test === root && value.Output === host);
    const [observation] = earlyPost.splice(post, 1);
    const terminal = earlyPost.findIndex(value => value.Test === root + '/' + axes[2] && value.Action === 'pass');
    earlyPost.splice(terminal, 0, observation);
    assert.throws(() => audit(earlyPost), 'host bracket before final axis');
  }
  const otherSDK = completeEvents().map(value => value.Test === roots[1] && value.Output?.includes(' driver=')
    ? {...value, Output: value.Output.replace(' driver=' + digest, ' driver=' + 'b'.repeat(64))} : value);
  assert.throws(() => audit(otherSDK), 'the two owners must use the same actual Go driver');
});

test('blend/minimum IA32 images precede every actual leaf and cannot be emulated summaries', () => {
  audit();
  for (const parent of parents.filter(name => name.startsWith(roots[1] + '/'))) {
    for (const replacement of ['', image + image, image.replace('4096', '0'), image.replace(digest, 'invalid'),
      image.replace('static IA32', 'QEMU translated')]) {
      assert.throws(() => audit(completeEvents().map(value => value.Test === parent && value.Output
        ? {...value, Output: replacement} : value)), parent);
    }
    const late = completeEvents();
    const at = late.findIndex(value => value.Test === parent && value.Output === image);
    const [observation] = late.splice(at, 1);
    late.splice(at + 2, 0, observation);
    assert.throws(() => audit(late), 'image inspected after first execution');
  }
});

test('blend/minimum rejects foreign inventories events packages and out-of-lifecycle output', () => {
  audit();
  for (const discovered of ['', inventory + roots[0] + '\n', inventory.replace(roots[0], 'TestObjectsOnly')]) {
    assert.throws(() => audit(completeEvents(), discovered));
  }
  for (const events of [completeEvents().slice(1), completeEvents().slice(0, -1),
    [event('start'), ...completeEvents()], [...completeEvents(), event('pass')],
    completeEvents().map(value => ({...value, Package: 'foreign/package'})),
    [event('output', undefined, '264 kernels passed')]]) assert.throws(() => audit(events));
  for (const name of [roots[0], parents[0], leaves[0]]) {
    const late = completeEvents();
    const at = late.findIndex(value => value.Test === name && value.Action === 'pass');
    late.splice(at + 1, 0, event('output', name, marker));
    assert.throws(() => audit(late));
  }
});

test('blend/minimum fixed leaves retain their own package and live root/axis lifecycle', () => {
  audit();
  for (const leaf of [leaves[0], leaves.at(-1)]) {
    for (const Action of ['run', 'output', 'pass']) {
      assert.throws(() => audit(completeEvents().map(value => value.Test === leaf && value.Action === Action
        ? {...value, Package: 'foreign/package'} : value)), 'foreign leaf ' + Action);
      assert.throws(() => audit(completeEvents().map(value => value.Test === leaf && value.Action === Action
        ? {...value, Action: 'unrecognized'} : value)), 'unknown leaf ' + Action);
    }
    const parent = leaf.split('/').slice(0, 2).join('/');
    for (const before of [leaf.split('/')[0], parent]) {
      const events = completeEvents();
      const own = events.filter(value => value.Test === leaf);
      const others = events.filter(value => value.Test !== leaf);
      const at = others.findIndex(value => value.Test === before && value.Action === 'run');
      others.splice(at, 0, ...own);
      assert.throws(() => audit(others), 'leaf before actual parent ' + before);
    }
    const events = completeEvents();
    events.push(event('output', leaf, marker));
    assert.throws(() => audit(events), 'leaf after completed package');
  }
  const foreign = completeEvents();
  foreign.splice(4, 0, event('run', parents[0] + '/unknown/opcode'));
  assert.throws(() => audit(foreign), 'unknown leaf cannot evade fixed ancestry');
});

function checkWorkflow(workflow) {
  const job = jobBody(workflow, 'native-x86-oracles');
  assert.match(job, /runs-on: ubuntu-latest/);
  assert.match(job, /timeout-minutes: 25/);
  assert.doesNotMatch(job, /continue-on-error:/, 'native owner failure must remain a failed job');
  assert.match(job, /install-ci-llvm22-ubuntu\.sh gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.ok(job.includes("go-version: '1.27.2'"));
  const title = '      - name: Execute original blend minimum alignment and zero-mask faults';
  assert.equal(job.split(title).length - 1, 1, 'exactly one explicit required blend/minimum owner');
  const start = job.indexOf(title), end = job.indexOf('      - name:', start + title.length);
  const block = job.slice(start, end);
  for (const root of roots) assert.equal(block.split(root).length - 1, 1, root);
  for (const command of ['set -euo pipefail', 'directory=_out/native-x86-oracles/blend-minpos',
    'go test . -list "^($roots)$" > "$directory/inventory.log"', 'grep -Fx "$root" "$directory/inventory.log"',
    'go test . -run "^($roots)$" -count=1 -timeout=5m -json > "$directory/raw.jsonl" 2> "$directory/stderr.log"',
    'node .github/scripts/audit-blend-minpos-native.cjs "$directory"',
    "if: ${{ !cancelled() && steps.native_toolchain.outcome == 'success' && steps.native_go.outcome == 'success' }}"]) {
    assert.ok(block.includes(command), command);
  }
  assert.ok(start < job.indexOf('      - name: Execute original complete floating programs with actual EVEX DQ state'),
    'AVX2 owner must not inherit an AVX512 prerequisite');
  assert.doesNotMatch(block, /continue-on-error|\|\| true|qemu|objects-only|success\(\)/);
  const integer = job.slice(job.indexOf('      - name: Execute original native integer stores align and division'));
  for (const root of ['TestCrossLinuxRuntimeMatrixX86MOVNTIOriginalFlagsAndFaults',
    'TestCrossLinuxRuntimeMatrixX86PackedAlignLegacyNativeFaults',
    'TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNativeAMD64',
    'TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNative386',
    'TestCrossLinuxRuntimeMatrixX86SegmentQueryOriginalFlagsAliasesAndFaults',
    'TestCrossLinuxRuntimeMatrixX86MoveHalfOriginalGoFlagsDFAliases',
    'TestCrossLinuxRuntimeMatrixX86ImplicitDivideLogicalFPOriginalGoAndNativeAMD64',
    'TestCrossLinuxRuntimeMatrixX86ImplicitDivideLogicalFPOriginalGoAndNative386']) {
    assert.equal(integer.split(root).length - 1, 1, 'old integer owner disappeared: ' + root);
  }
  requireCompleteWindowsNeeds(workflow);
  requireJobNeeds(workflow, 'diagnostic-qualification', ['native-x86-oracles', 'complete-windows']);
  const aggregate = jobBody(workflow, 'diagnostic-qualification');
  assert.ok(aggregate.includes('DIAGNOSTIC_NEEDS_JSON: ${{ toJSON(needs) }}'));
  assert.ok(aggregate.includes('run: node .github/scripts/diagnostic-qualification.cjs'));
}

test('blend/minimum CI is a separate required scope without shrinking old eight-root obligations', () => {
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  checkWorkflow(workflow);
  for (const needle of [...roots, 'node .github/scripts/audit-blend-minpos-native.cjs "$directory"']) {
    assert.throws(() => checkWorkflow(workflow.replace(needle, 'removed')), 'removed owner component: ' + needle);
  }
  for (const [old, replacement] of [
    ['directory=_out/native-x86-oracles/blend-minpos', 'directory=_out/unbound'],
    ['Execute original blend minimum alignment and zero-mask faults', 'unowned math only'],
  ]) assert.throws(() => checkWorkflow(workflow.replace(old, replacement)));
  const title = '      - name: Execute original blend minimum alignment and zero-mask faults';
  const start = workflow.indexOf(title), end = workflow.indexOf('      - name:', start + title.length);
  const block = workflow.slice(start, end);
  for (const [old, replacement] of [
    ['-timeout=5m', '-timeout=10m'],
    ['-run "^($roots)$"', '-run "^TestObjectsOnly$"'],
    ['!cancelled()', 'success()'],
    ["steps.native_toolchain.outcome == 'success'", 'true'],
    ["steps.native_go.outcome == 'success'", 'true'],
    ['        env:', '        continue-on-error: true\n        env:'],
  ]) {
    assert.ok(block.includes(old));
    assert.throws(() => checkWorkflow(workflow.replace(block, block.replace(old, replacement))), old);
  }
  assert.throws(() => checkWorkflow(workflow.replace('    timeout-minutes: 25', '    timeout-minutes: 30')));
  assert.throws(() => checkWorkflow(workflow.replace('      - name: Execute original native integer stores align and division',
    '      - name: Execute original native integer stores align and division\n        continue-on-error: true')));
  assert.throws(() => checkWorkflow(workflow.replace('DIAGNOSTIC_NEEDS_JSON: ${{ toJSON(needs) }}',
    'DIAGNOSTIC_NEEDS_JSON: untrusted')));
});
