const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const roots = [
  'TestCrossLinuxRuntimeMatrixX86X87SignedI64IdentityOriginalGoAndLLVM',
  'TestAMD64X87ConditionalMoveRuntimeConditions',
  'TestX86RawX87ClearExceptionRuntimeSemantics',
  'TestX86RawX87SoftwareClearPreservesAllStatusFields',
];
const children = [];
for (const arch of ['amd64', '386']) {
  const parent = `${roots[0]}/${arch}`;
  children.push(parent);
  for (const api of ['text', 'module', 'Go']) {
    for (const optimized of [false, true]) {
      for (const backend of [0, 2]) children.push(`${parent}/${api}/IR_O2=${optimized}/backend-O${backend}`);
    }
  }
}

function auditNativeX87(inventory, raw) {
  const discovered = inventory.split(/\r?\n/).filter(line => line.startsWith('Test'));
  assert.deepEqual(discovered.slice().sort(), roots.slice().sort(), 'complete fresh native x87 inventory changed');
  const expected = new Set([...roots, ...children]);
  const started = new Set(), passed = new Set();
  let complete = false;
  for (const line of raw.split(/\r?\n/).filter(Boolean)) {
    const event = JSON.parse(line);
    if (event.Action === 'build-output' && typeof event.ImportPath === 'string' && event.ImportPath && !event.Test) continue;
    assert.equal(event.Package, 'github.com/xgo-dev/plan9asm', 'foreign native x87 package');
    assert.ok(!complete, 'event after native package terminal');
    assert.ok(!['fail', 'skip', 'build-fail'].includes(event.Action), 'required native x87 execution failed or skipped');
    if (event.Test) {
      const owner = event.Test.split('/')[0];
      assert.ok(expected.has(event.Test), 'foreign native root or hardware/ABI/backend child');
      if (event.Test !== owner) assert.ok(started.has(owner) && !passed.has(owner), 'child outside its live owner');
      for (const parent of expected) {
        if (event.Test.startsWith(parent + '/')) {
          assert.ok(started.has(parent) && !passed.has(parent), 'child outside a required live architecture parent');
        }
      }
      if (event.Action === 'run') {
        assert.ok(!started.has(event.Test), 'duplicate native start');
        started.add(event.Test);
      } else {
        assert.ok(started.has(event.Test), 'event for an unstarted native owner');
        if (event.Action === 'pass') {
          assert.ok(!passed.has(event.Test), 'duplicate native terminal');
          passed.add(event.Test);
        }
      }
    } else if (event.Action === 'pass') {
      assert.deepEqual([...started].sort(), [...expected].sort(), 'missing native x87 execution axis');
      assert.deepEqual([...passed].sort(), [...expected].sort(), 'unfinished native x87 execution axis');
      complete = true;
    }
  }
  assert.ok(complete, 'native x87 package did not finish successfully');
  return {roots: roots.length, children: children.length, passed: passed.size, skipped: 0, incomplete: 0};
}

module.exports = {auditNativeX87, roots, children};

if (require.main === module) {
  assert.equal(process.argv.length, 3, 'one actual native x87 report directory required');
  const directory = process.argv[2];
  const result = auditNativeX87(fs.readFileSync(path.join(directory, 'inventory.log'), 'utf8'),
    fs.readFileSync(path.join(directory, 'raw.jsonl'), 'utf8'));
  process.stdout.write(JSON.stringify(result, null, 2) + '\n');
}
