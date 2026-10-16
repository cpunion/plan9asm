const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {auditNativeUnlocked, roots, children} = require('./audit-unlocked-native.cjs');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');

function requireUnlockedOwner(source) {
  const marker = '\n  unlocked-rmw-native-priority:\n';
  assert.equal(source.split(marker).length - 1, 1, 'one mandatory unlocked native job');
  const rest = source.slice(source.indexOf(marker) + marker.length);
  const end = rest.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  const body = end < 0 ? rest : rest.slice(0, end);
  assert.match(body, /^    runs-on: ubuntu-latest$/m);
  assert.match(body, /scripts\/install-ci-llvm22-ubuntu\.sh gcc-i686-linux-gnu libc6-dev-i386-cross/);
  assert.match(body, /go-version: '1\.27\.2'/);
  assert.match(body, /set -euo pipefail/);
  assert.match(body, /go test -p=1 \. -list "\^\(\$roots\)\$"/);
  assert.match(body, /grep -Fx "\$root" _out\/native-unlocked\/inventory\.log/);
  assert.match(body, /go test -p=1 \. -run "\^\(\$roots\)\$" -count=1 -timeout=5m -json/);
  for (const root of roots) assert.equal(body.split(root).length - 1, 1, root);
  assert.match(body, /2> _out\/native-unlocked\/stderr\.log \| tee _out\/native-unlocked\/raw\.jsonl/);
  assert.match(body, /node \.github\/scripts\/audit-unlocked-native\.cjs _out\/native-unlocked/);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}/);
  assert.match(body, /name: native-unlocked\n          path: _out\/native-unlocked\/\n          if-no-files-found: error/);
  assert.doesNotMatch(body, /continue-on-error|max-parallel|qemu|rosetta|PLAN9ASM_CROSS_EXEC|overlay/);
  assert.match(source, /needs: \[ownership, native-x86-oracles, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, unlocked-rmw-native-priority, vector-value-fault-priority\]/);
  return body;
}

test('both unlocked native ISAs have a distinct required owner before full modules', () => {
  requireUnlockedOwner(workflow);
  const native = fs.readFileSync('amd64_scalar_memory_rmw_linux_amd64_test.go', 'utf8');
  assert.equal((native.match(/x87IA32RequireNativeHost\(t\)/g) || []).length, 4);
  assert.equal((native.match(/\[\]string\{"-O0", "-O2"\}/g) || []).length, 2);
  assert.doesNotMatch(native, /\.Skip|qemu|rosetta|PLAN9ASM_CROSS_EXEC/);
});

test('unlocked native ownership rejects omissions, duplicate roots, emulation and larger budgets', () => {
  const body = requireUnlockedOwner(workflow);
  for (const [before, after] of [
    ['runs-on: ubuntu-latest', 'runs-on: windows-latest'],
    ['scripts/install-ci-llvm22-ubuntu.sh', 'scripts/install-ci-llvm23-ubuntu.sh'],
    ['libc6-dev-i386-cross', 'missing-i386-headers'],
    ['go test -p=1 . -list', 'printf missing-discovery'],
    ['grep -Fx "$root"', 'printf missing-membership'],
    ['-timeout=5m -json', '-timeout=6m -json'],
    ['audit-unlocked-native.cjs', 'missing-auditor.cjs'],
    ['if-no-files-found: error', 'if-no-files-found: ignore'],
    ['set -euo pipefail', 'set -euo pipefail\n          qemu-i386 true'],
    ...roots.map(root => [root, root + '|' + root]),
  ]) {
    assert.notEqual(body.replace(before, after), body);
    assert.throws(() => requireUnlockedOwner(workflow.replace(body, body.replace(before, after))));
  }
  assert.throws(() => requireUnlockedOwner(workflow.replace(', unlocked-rmw-native-priority,', ',')));
  assert.throws(() => requireUnlockedOwner(workflow.replace(', native-x86-oracles,', ',')));
});

test('unlocked audit requires both executed ISAs and all three public API children', () => {
  assert.equal(roots.length, 2);
  assert.equal(children.length, 6);
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
  const encode = rows => rows.map(JSON.stringify).join('\n') + '\n';
  assert.deepEqual(auditNativeUnlocked(inventory, encode(events)),
    {roots: 2, children: 6, passed: 8, skipped: 0, incomplete: 0});
  for (const child of children) {
    for (const altered of [
      events.filter(row => row.Test !== child),
      events.map(row => row.Test === child && row.Action === 'pass' ? {...row, Action: 'skip'} : row),
      events.map(row => row.Test === child ? {...row, Test: child + '-object-only'} : row),
    ]) assert.throws(() => auditNativeUnlocked(inventory, encode(altered)));
  }
  assert.throws(() => auditNativeUnlocked(inventory,
    encode([event('start'), {...event('output'), Output: encode(events)}, event('pass')])));
});
