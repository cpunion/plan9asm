'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const workflow = fs.readFileSync(path.join(__dirname, '../workflows/cmp-test-bt-native-diagnostic.yml'), 'utf8');
const product = 'f0d322e5193a557b8a1b4ac29a4ea06242f49cb3';
const roots = [
  'TestCrossLinuxRuntimeMatrixX86ScalarCompareFlags',
  'TestCrossLinuxRuntimeMatrixX86ScalarCompareDeadReadFaults',
  'TestCrossLinuxRuntimeMatrixX86TESTIndependentCarryFlags',
  'TestCrossLinuxRuntimeMatrixX86BitTestNumericalAndAliases',
  'TestCrossLinuxRuntimeMatrixX86BitTestDeadMemoryFaults',
  'TestCrossLinuxRuntimeMatrixX86BitTestFPPartialWriteback',
];

function requireControl(source) {
  assert.match(source, /branches: \[codex\/cmp-test-bt-native-f0d32-20261109\]/);
  assert.match(source, /if: \$\{\{ github.repository == 'cpunion\/plan9asm' \}\}/);
  assert.match(source, /permissions:\n  contents: read\n/);
  assert.match(source, /cancel-in-progress: false/);
  assert.match(source, /runs-on: ubuntu-24\.04/);
  assert.match(source, /go: \['1\.27\.3', '1\.20\.14'\]/);
  assert.ok(source.includes('      PRODUCT_SHA: ' + product + '\n'));
  assert.match(source, /ref: \$\{\{ env.PRODUCT_SHA \}\}/);
  assert.match(source, /persist-credentials: false/);
  const selected = source.match(/NATIVE_OWNERS: >-\n((?:        Test[^\n]+\n)+)/);
  assert.ok(selected, 'fixed original native owners required');
  assert.deepEqual(selected[1].trim().split(/\s+/), roots);
  assert.match(source, /bash scripts\/install-ci-llvm22-ubuntu\.sh/);
  assert.match(source, /uses: \.\/\.github\/actions\/setup-node22/);
  assert.match(source, /test "\$\(uname -m\)" = x86_64/);
  assert.match(source, /test "\$\(go env GOOS\)\/\$\(go env GOARCH\)" = linux\/amd64/);
  assert.match(source, /test "\$\(go env GOVERSION\)" = "go\$SDK_VERSION"/);
  assert.equal((source.match(/test "\$\(git rev-parse HEAD\)" = "\$PRODUCT_SHA"/g) || []).length, 2);
  assert.match(source, /-list "\$selector"/);
  assert.match(source, /-run "\$selector" -count=1 -timeout=30m -json/);
  assert.match(source, /sha256sum -c "\$EVIDENCE\/source-inputs.sha256"/);
  assert.match(source, /sha256sum -c "\$EVIDENCE\/tool-inputs.sha256"/);
  assert.match(source, /auditNativeOwners\(inventory, raw, roots, children\)/);
  assert.match(source, /if: always\(\)[\s\S]*uses: actions\/upload-artifact@v7/);
  assert.match(source, /if-no-files-found: error/);
  assert.doesNotMatch(source, /continue-on-error:|\|\| true|workflow_dispatch:|pull_request:|secrets:/);
}

test('fork native control binds exact product, all six owners and two SDKs', () => {
  requireControl(workflow);
  for (const [before, after] of [
    [product, '0'.repeat(40)],
    ["'1.20.14'", "'1.20.x'"],
    [roots[0], 'TestUnrelatedOwner'],
    ['ref: ${{ env.PRODUCT_SHA }}', 'ref: main'],
    ['-run "$selector"', '-run TestSample'],
    ['if-no-files-found: error', 'if-no-files-found: warn'],
  ]) {
    assert.notEqual(workflow.replace(before, after), workflow);
    assert.throws(() => requireControl(workflow.replace(before, after)));
  }
});

// Independent exact child inventory from the unchanged six runtime owners.
const children = [];
for (const root of roots.slice(0, 5)) {
  for (const api of ['text', 'module', 'Go']) {
    children.push(root + '/' + api, root + '/' + api + '/backend_O0',
      root + '/' + api + '/IR_O2/backend_O2');
  }
}
children.push(...['text', 'module', 'Go'].map(api => roots[5] + '/' + api));
const sqrt = roots[1] + '/scalar-sqrt';
children.push(sqrt);
for (const api of ['text', 'module', 'Go']) {
  children.push(sqrt + '/' + api);
  for (const optimized of ['false', 'true']) {
    const parent = sqrt + '/' + api + '/IR-O2=' + optimized;
    children.push(parent, parent + '/-O0', parent + '/-O2');
  }
}

function audit(events) {
  const selected = workflow.match(/          node <<'NODE'\n([\s\S]*?)^          NODE$/m);
  assert.ok(selected, 'execute the actual workflow audit body');
  const script = selected[1].split('\n').map(line => line.slice(10)).join('\n');
  const files = new Map([
    ['native-inventory.txt', roots.join('\n') + '\n'],
    ['native.jsonl', events.map(JSON.stringify).join('\n') + '\n'],
  ]);
  const io = {
    readFileSync(name) { assert.ok(files.has(path.basename(name))); return files.get(path.basename(name)); },
    writeFileSync(name, value) { files.set(path.basename(name), value); },
  };
  const execute = new Function('require', 'process', script);
  execute(
    name => {
      if (name === 'node:fs') return io;
      assert.equal(name, './.github/scripts/audit-native-owners.cjs');
      return require('./audit-native-owners.cjs');
    },
    {env: {EVIDENCE: '.', NATIVE_OWNERS: roots.join(' ')}},
  );
  return JSON.parse(files.get('native-audit.json'));
}

function fixture() {
  const event = (Action, Test) => ({Action, Package: 'github.com/xgo-dev/plan9asm', ...(Test ? {Test} : {})});
  const names = [...roots, ...children];
  const parent = name => names.filter(candidate => name.startsWith(candidate + '/'))
    .sort((a, b) => b.length - a.length)[0];
  const visit = name => [event('run', name),
    ...children.filter(child => parent(child) === name).flatMap(visit), event('pass', name)];
  return [...roots.flatMap(visit), event('pass')];
}

test('actual inline native audit requires all 76 original root and child terminals', () => {
  assert.deepEqual(audit(fixture()), {roots: 6, children: 70, passed: 76, skipped: 0, incomplete: 0});
  for (const name of [...roots, ...children]) {
    const events = fixture();
    assert.throws(() => audit(events.filter(event => event.Test !== name)));
    for (const Action of ['skip', 'fail']) {
      assert.throws(() => audit(events.map(event => event.Action === 'pass' && event.Test === name
        ? {...event, Action} : event)));
    }
  }
  const events = fixture();
  assert.throws(() => audit(events.slice(0, -1)));
  assert.throws(() => audit([...events, events[events.length - 1]]));
});
