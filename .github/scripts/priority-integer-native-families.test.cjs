'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {auditIntegerNativeFamilies} = require('./audit-integer-native-families.cjs');

// Independent owner/axis inventory: do not import the auditor's denominator.
const owners = [
  'TestCrossLinuxRuntimeMatrixX86MOVNTIOriginalFlagsAndFaults',
  'TestCrossLinuxRuntimeMatrixX86PackedAlignLegacyNativeFaults',
  'TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNativeAMD64',
  'TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNative386',
];
const axes = ['backend_O0', 'backend_O2', 'IR_O2/backend_O2'];
const packageName = 'github.com/xgo-dev/plan9asm';
const hash = 'a'.repeat(64);
const imageHash = 'b'.repeat(64);
const inventory = owners.join('\n') + '\nok\t' + packageName + '\n';
const host = `IA32 direct host: linux/amd64 kernel=x86_64 CPUsha=${hash} mapsha=${hash}\n`;
const staticImage = `native division static IA32 ELF bytes=4096 sha256=${imageHash}\n`;
const directImage = args => `IA32 direct ELF bytes=4096 sha256=${imageHash} argv=${args}\n`;
const sdk = args => `actual executable=/sdk/bin/go sha256=${hash} argv=${args}\n`;
const originalArgs = '["-test.run=^TestOriginal$" "-test.count=1" "-test.v"]';
const compileArgs = '["test" "-p=1" "-vet=off" "-c" "-o" "/tmp/owned/original.test" "."]';
const event = (Action, Test, Output) => ({Action, Package: packageName,
  ...(Test ? {Test} : {}), ...(Output !== undefined ? {Output} : {})});
const serialize = events => events.map(value => JSON.stringify(value)).join('\n');

function completeEvents() {
  const events = [event('start')];
  for (const owner of owners) {
    const divide = owner.includes('ImplicitDivide');
    const ia32 = owner.endsWith('386');
    events.push(event('run', owner));
    if (divide) events.push(event('output', owner, host));
    for (const axis of axes) {
      const child = owner + '/' + axis;
      events.push(event('run', child));
      if (ia32) {
        events.push(event('output', child, staticImage + directImage('[]') + staticImage));
      }
      events.push(event('pass', child));
    }
    if (divide) {
      events.push(event('output', owner, sdk('["version"]') + sdk(compileArgs)));
      if (ia32) events.push(event('output', owner, staticImage + directImage(originalArgs)));
      events.push(event('output', owner, host));
    }
    events.push(event('pass', owner));
  }
  events.push(event('pass'));
  return events;
}

test('native integer families require all four roots and twelve complete execution axes', () => {
  assert.deepEqual(auditIntegerNativeFamilies(inventory, serialize(completeEvents())),
    {roots: 4, children: 12, passed: 16, skipped: 0, incomplete: 0,
      nativeHosts: 4, sdkBuilds: 2, staticIA32Images: 7, directIA32Executions: 4});
});

test('native integer families reject each missing, failed, skipped or duplicate terminal', () => {
  const names = owners.flatMap(owner => [owner, ...axes.map(axis => owner + '/' + axis)]);
  for (const name of names) {
    const original = completeEvents();
    for (const Action of ['fail', 'skip']) {
      assert.throws(() => auditIntegerNativeFamilies(inventory,
        serialize(original.map(value => value.Action === 'pass' && value.Test === name
          ? {...value, Action} : value))), name + '/' + Action);
    }
    for (const Action of ['run', 'pass']) {
      const at = original.findIndex(value => value.Action === Action && value.Test === name);
      assert.throws(() => auditIntegerNativeFamilies(inventory,
        serialize(original.filter((_, index) => index !== at))), name + '/missing ' + Action);
      const duplicate = original.slice();
      duplicate.splice(at, 0, original[at]);
      assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(duplicate)),
        name + '/duplicate ' + Action);
    }
  }
});

test('native integer families reject malformed packages, inventory changes and printed success', () => {
  for (const changed of ['', inventory + owners[0] + '\n', inventory.replace(owners[0] + '\n', '')]) {
    assert.throws(() => auditIntegerNativeFamilies(changed, serialize(completeEvents())));
  }
  for (const changed of [
    [event('output', undefined, '16 passed')],
    completeEvents().slice(1),
    completeEvents().slice(0, -1),
    [event('start'), ...completeEvents()],
    [...completeEvents(), event('pass')],
    [...completeEvents(), {Action: 'build-output', ImportPath: 'runtime', Output: 'late'}],
    [event('run', owners[0] + '/objects-only'), ...completeEvents()],
    completeEvents().map(value => ({...value, Package: 'foreign/package'})),
  ]) assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)));
  for (const value of [event('unknown'), event('output', owners[0], 'late output')]) {
    const changed = completeEvents();
    changed.splice(-1, 0, value);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)));
  }
  assert.throws(() => auditIntegerNativeFamilies(inventory, '{not JSON}'));
});

test('division native hosts and actual Go SDK builds bracket all three axes', () => {
  for (const owner of owners.slice(2)) {
    for (const replacement of ['', host.replace('linux/amd64', 'linux/arm64'),
      host.replace('kernel=x86_64', 'kernel=aarch64'), host.replace(hash, 'not-a-digest')]) {
      const changed = completeEvents().map(value => value.Test === owner && value.Output === host
        ? {...value, Output: replacement} : value);
      assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)), owner);
    }
    for (const needle of ['["version"]', compileArgs]) {
      const changed = completeEvents().map(value => value.Test === owner && value.Output
        ? {...value, Output: value.Output.replace(sdk(needle), '')} : value);
      assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)), needle);
    }
    const changed = completeEvents();
    const before = changed.findIndex(value => value.Test === owner && value.Output === host);
    const [late] = changed.splice(before, 1);
    changed.splice(before + 2, 0, late);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)), 'late preflight');
    const wrongTool = completeEvents().map(value => value.Test === owner && value.Output
      ? {...value, Output: value.Output.replaceAll('/sdk/bin/go', '/other/bin/go')} : value);
    // A different actual SDK path is valid if both commands bind the same tool.
    assert.doesNotThrow(() => auditIntegerNativeFamilies(inventory, serialize(wrongTool)));
    const splitTool = completeEvents().map(value => value.Test === owner && value.Output
      ? {...value, Output: value.Output.replace(sdk(compileArgs), sdk(compileArgs).replace(hash, imageHash))}
      : value);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(splitTool)), 'changed SDK');
    const liveHostHashes = completeEvents();
    const lastHost = liveHostHashes.findLastIndex(value => value.Test === owner && value.Output === host);
    liveHostHashes[lastHost] = {...liveHostHashes[lastHost], Output: host.replaceAll(hash, imageHash)};
    assert.doesNotThrow(() => auditIntegerNativeFamilies(inventory, serialize(liveHostHashes)));
    const unboundArgs = completeEvents().map(value => value.Test === owner && value.Output
      ? {...value, Output: value.Output.replace('"-p=1" "-vet=off"', '"-p=8"')} : value);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(unboundArgs)), 'changed build');
    const missingPost = completeEvents();
    const post = missingPost.findLastIndex(value => value.Test === owner && value.Output === host);
    missingPost.splice(post, 1);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(missingPost)), 'missing final host');
    const earlyGo = completeEvents();
    const goAt = earlyGo.findIndex(value => value.Test === owner && value.Output?.includes('actual executable='));
    const [originalGo] = earlyGo.splice(goAt, 1);
    const finalChild = earlyGo.findIndex(value => value.Test === owner + '/' + axes[2] && value.Action === 'pass');
    earlyGo.splice(finalChild, 0, originalGo);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(earlyGo)), 'Go before final axis');
  }
});

test('division IA32 requires actual static and direct images in every axis and original Go', () => {
  const owner = owners[3];
  const required = [owner, ...axes.map(axis => owner + '/' + axis)];
  for (const name of required) {
    for (const needle of [staticImage, directImage(name === owner ? originalArgs : '[]')]) {
      const changed = completeEvents().map(value => value.Test === name && value.Output
        ? {...value, Output: value.Output.replace(needle, '')} : value);
      assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)), name + needle);
    }
  }
  for (const replacement of ['bytes=0', 'bytes=-1', 'bytes=guest']) {
    const changed = completeEvents().map(value => value.Output
      ? {...value, Output: value.Output.replaceAll('bytes=4096', replacement)} : value);
    assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(changed)), replacement);
  }
  const forged = completeEvents().map(value => value.Output
    ? {...value, Output: value.Output.replaceAll('IA32 direct ELF', 'QEMU translated ELF')} : value);
  assert.throws(() => auditIntegerNativeFamilies(inventory, serialize(forged)));
});

test('native integer CI is a required exact Linux owner with retained raw observations', () => {
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  const job = workflow.slice(workflow.indexOf('  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  assert.match(job, /runs-on: ubuntu-latest/);
  assert.match(job, /install-ci-llvm22-ubuntu\.sh gcc-i686-linux-gnu libc6-dev-i386-cross/);
  const start = job.indexOf('      - name: Execute original native integer stores align and division');
  assert.ok(start >= 0, 'required native families need a dedicated complete owner');
  const block = job.slice(start, job.indexOf('      - name:', start + 1));
  for (const owner of owners) assert.equal(block.split(owner).length - 1, 1, owner);
  assert.ok(block.includes('directory=_out/native-x86-oracles/integer-families'));
  assert.ok(block.includes('set -euo pipefail'));
  assert.ok(block.includes('go test . -list "^($roots)$" > "$directory/inventory.log"'));
  assert.ok(block.includes('grep -Fx "$root" "$directory/inventory.log"'));
  assert.ok(block.includes('go test . -run "^($roots)$" -count=1 -timeout=5m -json > "$directory/raw.jsonl" 2> "$directory/stderr.log"'));
  assert.ok(block.includes('node .github/scripts/audit-integer-native-families.cjs "$directory"'));
  assert.doesNotMatch(block, /continue-on-error|\|\| true|qemu|objects-only/);
  assert.match(workflow, /needs: \[ownership, native-x86-oracles, sdk-priority,/);
  assert.match(job, /path: _out\/native-x86-oracles\//);
});
