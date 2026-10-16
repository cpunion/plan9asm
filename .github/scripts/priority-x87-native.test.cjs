const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const roots = [
  'TestCrossLinuxRuntimeMatrixX86X87SignedI64IdentityOriginalGoAndLLVM',
  'TestAMD64X87ConditionalMoveRuntimeConditions',
  'TestX86RawX87ClearExceptionRuntimeSemantics',
  'TestX86RawX87SoftwareClearPreservesAllStatusFields',
];

function requireX87NativeJob(source) {
  const marker = '\n  x87-native-priority:\n';
  const start = source.indexOf(marker);
  assert.ok(start >= 0, 'missing independent native x87 execution owner');
  let body = source.slice(start + marker.length);
  const end = body.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  if (end >= 0) body = body.slice(0, end);
  assert.match(body, /runs-on: ubuntu-latest/);
  assert.match(body, /gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.match(body, /go-version: '1\.27\.2'/);
  assert.match(body, /set -euo pipefail/);
  assert.match(body, /go test -p=1 \. -list "\^\(\$roots\)\$"/);
  assert.match(body, /go test -p=1 \. -run "\^\(\$roots\)\$" -count=1 -timeout=5m -json/);
  for (const root of roots) assert.equal(body.split(root).length - 1, 1, 'exact complete native owner: ' + root);
  assert.match(body, /2> _out\/native-x87\/stderr\.log \| tee _out\/native-x87\/raw\.jsonl/);
  assert.match(body, /node \.github\/scripts\/audit-x87-native\.cjs _out\/native-x87/);
  assert.match(body, /actions\/upload-artifact@v6/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
  assert.doesNotMatch(body, /continue-on-error|max-parallel|qemu|rosetta|PLAN9ASM_CROSS_EXEC|overlay/);
  return body;
}

test('x87 complete native hardware and original-Go owners cannot be replaced by host skips or objects', () => {
  requireX87NativeJob(workflow);
});

module.exports = {requireX87NativeJob, roots};

test('native x87 scheduling rejects removed, substituted, duplicated, emulated or unverified owners', () => {
  const body = requireX87NativeJob(workflow);
  for (const root of roots) {
    assert.throws(() => requireX87NativeJob(workflow.replace(root, 'TestUnrelatedX87Object')));
    assert.throws(() => requireX87NativeJob(workflow.replace(root, root + '|' + root)));
  }
  for (const [before, after] of [
    ['-timeout=5m -json', '-timeout=6m -json'],
    ['go test -p=1 . -list', 'printf missing-inventory'],
    ['node .github/scripts/audit-x87-native.cjs', 'printf omitted-audit'],
    ['2> _out/native-x87/stderr.log', '2>&1'],
    ['set -euo pipefail', 'set -euo pipefail\n          qemu-i386 true'],
    ['runs-on: ubuntu-latest', 'runs-on: windows-latest'],
  ]) {
    assert.ok(body.includes(before));
    assert.throws(() => requireX87NativeJob(workflow.replace(body, body.replace(before, after))));
  }
});

test('native x87 auditor requires actual complete ordered terminals, not printed PASS text', () => {
  const {auditNativeX87, roots: actualRoots, children} = require('./audit-x87-native.cjs');
  assert.deepEqual(actualRoots, roots);
  const event = (Action, Test) => ({Action, Package: 'github.com/xgo-dev/plan9asm', ...(Test ? {Test} : {})});
  const events = [event('start')];
  for (const root of roots) {
    events.push(event('run', root));
    for (const parent of children.filter(name => name.startsWith(root + '/') && name.split('/').length === 2)) {
      events.push(event('run', parent));
      for (const child of children.filter(name => name.startsWith(parent + '/'))) {
        events.push(event('run', child), event('pass', child));
      }
      events.push(event('pass', parent));
    }
    events.push(event('pass', root));
  }
  events.push(event('pass'));
  const inventory = roots.join('\n') + '\nok\tgithub.com/xgo-dev/plan9asm\n';
  const encode = values => values.map(value => JSON.stringify(value)).join('\n') + '\n';
  assert.deepEqual(auditNativeX87(inventory, encode(events)), {roots: 4, children: 26, passed: 30, skipped: 0, incomplete: 0});
  const child = children[1];
  const earlyParent = events.filter(value => value.Test !== children[0] || value.Action !== 'pass');
  earlyParent.splice(earlyParent.findIndex(value => value.Test === child), 0, event('pass', children[0]));
  for (const changed of [
    earlyParent,
    events.filter(value => value.Test !== child),
    events.filter(value => value.Test !== child || value.Action !== 'pass'),
    events.map(value => value.Test === child && value.Action === 'pass' ? {...value, Action: 'skip'} : value),
    events.map(value => value.Test === child && value.Action === 'pass' ? {...value, Action: 'fail'} : value),
    events.map(value => value.Test === child ? {...value, Test: child + '-object-only'} : value),
    events.slice(0, -1),
    [...events, event('pass')],
    [event('start'), {...event('output'), Output: events.map(value => 'PASS ' + (value.Test || 'package')).join('\n')}, event('pass')],
  ]) assert.throws(() => auditNativeX87(inventory, encode(changed)));
  assert.throws(() => auditNativeX87(inventory.replace(roots[0], 'TestObjectOnly'), encode(events)));
});
