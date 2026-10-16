const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const roots = ['TestCrossLinuxRuntimeMatrixX86LockScalarRMWNativeAMD64',
  'TestCrossLinuxRuntimeMatrixX86LockScalarRMWNative386'];
const dispatch = 'TestX86LockScalarRMWPublicBodiesCallCompleteNativeHelpers';
const priority = "- lane: lock-scalar-dispatch\n            module: '.'\n            roots:\n" +
  `              - ${dispatch}\n            timeout: 5m`;

function requireLockOwners(source) {
  const marker = '\n  lock-native-priority:\n';
  assert.equal(source.split(marker).length - 1, 1, 'one required native LOCK job');
  let body = source.slice(source.indexOf(marker) + marker.length);
  const end = body.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  if (end >= 0) body = body.slice(0, end);
  assert.match(body, /runs-on: ubuntu-latest/);
  assert.match(body, /gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.match(body, /scripts\/install-ci-llvm22-ubuntu\.sh/);
  assert.match(body, /go-version: '1\.27\.2'/);
  assert.match(body, /set -euo pipefail/);
  assert.match(body, /go test -p=1 \. -list "\^\(\$roots\)\$"/);
  assert.match(body, /go test -p=1 \. -run "\^\(\$roots\)\$" -count=1 -timeout=5m -json/);
  for (const root of roots) assert.equal(body.split(root).length - 1, 1);
  assert.ok(body.includes(`roots='${roots[0]}'\n`), 'complete AMD64 root assignment');
  assert.ok(body.includes(`roots+='|${roots[1]}'\n`), 'complete 386 root assignment');
  assert.match(body, /2> _out\/native-lock\/stderr\.log \| tee _out\/native-lock\/raw\.jsonl/);
  assert.match(body, /node \.github\/scripts\/audit-lock-native\.cjs _out\/native-lock/);
  assert.match(body, /actions\/upload-artifact@v6/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
  assert.match(body, /name: native-lock\n          path: _out\/native-lock\/\n          if-no-files-found: error/);
  assert.doesNotMatch(body, /continue-on-error|max-parallel|qemu|rosetta|PLAN9ASM_CROSS_EXEC|overlay/);
  assert.match(source, /needs: \[ownership, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, vector-value-fault-priority\]/);
  assert.ok(source.includes(priority), 'complete source-body dispatch before module tests');
  assert.equal(source.split('- lane: lock-scalar-dispatch\n').length - 1, 1);
  assert.equal(source.split(`- ${dispatch}\n`).length - 1, 1);
  return body;
}

test('LOCK source dispatch and both real native ISAs are mandatory before full modules', () => {
  requireLockOwners(workflow);
});

test('LOCK priority rejects cropped, duplicate, skipped, emulated or unverified owners', () => {
  const body = requireLockOwners(workflow);
  for (const [before, after] of [
    ['runs-on: ubuntu-latest', 'runs-on: windows-latest'],
    ['gcc-i686-linux-gnu libc6-dev-i386-cross', 'gcc-i686-linux-gnu'],
    ['scripts/install-ci-llvm22-ubuntu.sh', 'scripts/install-ci-llvm23-ubuntu.sh'],
    ['-timeout=5m -json', '-timeout=6m -json'],
    ['go test -p=1 . -list', 'printf missing-inventory'],
    ['node .github/scripts/audit-lock-native.cjs', 'printf missing-audit'],
    ['if-no-files-found: error', 'if-no-files-found: ignore'],
    ['set -euo pipefail', 'set -euo pipefail\n          qemu-i386 true'],
    ...roots.flatMap(root => [[root, root + '/-O0'], [root, root + '|' + root]]),
  ]) {
    assert.ok(body.includes(before));
    assert.throws(() => requireLockOwners(workflow.replace(body, body.replace(before, after))),
      undefined, before);
  }
  for (const changed of [
    workflow + '\n  lock-native-priority:\n',
    workflow.replace(', lock-native-priority,', ','),
    workflow.replace(priority, priority.replace(dispatch, dispatch + '/sample')),
    workflow.replace(priority, priority.replace('timeout: 5m', 'timeout: 6m')),
  ]) assert.throws(() => requireLockOwners(changed));
});

test('LOCK execution auditing requires original owners and all four completed backend children', () => {
  const {auditNativeLock, roots: actual, children} = require('./audit-lock-native.cjs');
  assert.deepEqual(actual, roots);
  assert.deepEqual(children, roots.flatMap(root => ['-O0', '-O2'].map(level => root + '/' + level)));
  const event = (Action, Test) => ({Action, Package: 'github.com/xgo-dev/plan9asm', ...(Test ? {Test} : {})});
  const events = [event('start')];
  for (const root of roots) {
    events.push(event('run', root));
    for (const child of children.filter(name => name.startsWith(root + '/'))) {
      events.push(event('run', child), event('pass', child));
    }
    events.push(event('pass', root));
  }
  events.push(event('pass'));
  const inventory = roots.join('\n') + '\n';
  const encode = values => values.map(JSON.stringify).join('\n') + '\n';
  assert.deepEqual(auditNativeLock(inventory, encode(events)),
    {roots: 2, children: 4, passed: 6, skipped: 0, incomplete: 0});
  for (const child of children) {
    for (const changed of [
      events.filter(value => value.Test !== child),
      events.filter(value => value.Test !== child || value.Action !== 'pass'),
      events.map(value => value.Test === child && value.Action === 'pass' ? {...value, Action: 'skip'} : value),
      events.map(value => value.Test === child ? {...value, Test: child + '-object-only'} : value),
    ]) assert.throws(() => auditNativeLock(inventory, encode(changed)));
  }
  for (const changed of [events.slice(0, -1), [...events, event('pass')],
    [event('start'), {...event('output'), Output: encode(events)}, event('pass')]]) {
    assert.throws(() => auditNativeLock(inventory, encode(changed)));
  }
});
