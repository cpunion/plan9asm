'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {requireCompleteWindowsNeeds} = require('./workflow-needs.cjs');
const {auditProductCMOV, roots, children} = require('./audit-product-cmov.cjs');

test('complete numerical regression programs execute before the exhaustive Windows modules', () => {
  assert.deepEqual(roots, [
    'TestCrossLinuxRuntimeMatrixX86DwordProductIndependentWholeBytes',
    'TestCrossLinuxRuntimeMatrixX86CMOVIndependentWholeBytes',
    'TestCrossLinuxRuntimeMatrixX86AESFamilyWholeByteAliases',
    'TestAMD64VectorTestRuntimeSemantics',
    'TestAMD64IFMARuntimeSemantics',
    'TestAMD64QwordBitLookupRuntimeSemantics',
    'TestCrossLinuxRuntimeMatrixX86CMOVFPPartialCrossSlotWholeBytes',
  ]);
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  const job = workflow.slice(workflow.indexOf('  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  assert.match(job, /runs-on: ubuntu-latest/);
  const start = job.indexOf('      - name: Execute complete numerical regression programs');
  assert.ok(start >= 0, 'original discovering programs need their own native execution step');
  const block = job.slice(start, job.indexOf('      - name:', start + 1));
  for (const root of roots) assert.ok(block.includes(root));
  assert.ok(block.includes('go test . -list "^($roots)$"'));
  assert.ok(block.includes('go test . -run "^($roots)$" -count=1 -timeout=5m -json'));
  assert.ok(block.includes('node .github/scripts/audit-product-cmov.cjs "$directory"'));
  assert.doesNotMatch(block, /continue-on-error|\|\| true/);
  requireCompleteWindowsNeeds(workflow);
});

test('product and CMOV execution retains every root and optimization child without skips', () => {
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
  const raw = values => values.map(JSON.stringify).join('\n') + '\n';
  const inventory = roots.join('\n') + '\n';
  assert.deepEqual(auditProductCMOV(inventory, raw(events)),
    {roots: 7, children: 24, passed: 31, skipped: 0, incomplete: 0});
  for (const name of [...roots, ...children]) {
    assert.throws(() => auditProductCMOV(inventory, raw(events.filter(event => event.Test !== name))));
    for (const Action of ['skip', 'fail']) {
      assert.throws(() => auditProductCMOV(inventory, raw(events.map(event =>
        event.Test === name && event.Action === 'pass' ? {...event, Action} : event))));
    }
  }
  assert.throws(() => auditProductCMOV(inventory, raw(events.slice(0, -1))));
  assert.throws(() => auditProductCMOV(inventory, raw([event('start'), event('pass')])));
});
