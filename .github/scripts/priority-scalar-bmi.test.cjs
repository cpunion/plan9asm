'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {auditScalarBMI, roots} = require('./audit-scalar-bmi.cjs');

test('BLS and BEXTR runtime repairs have a required native Linux execution owner', () => {
  const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
  const job = workflow.slice(workflow.indexOf('  native-x86-oracles:'), workflow.indexOf('\n  complete-windows:'));
  assert.match(job, /runs-on: ubuntu-latest/);
  const start = job.indexOf('      - name: Execute translated scalar BMI values and flags');
  assert.ok(start >= 0);
  const block = job.slice(start, job.indexOf('      - name:', start + 1));
  for (const root of roots) assert.ok(block.includes(root));
  assert.ok(block.includes('go test . -list "^($roots)$"'));
  assert.ok(block.includes('go test . -run "^($roots)$" -count=1 -timeout=5m -json'));
  assert.ok(block.includes('node .github/scripts/audit-scalar-bmi.cjs "$directory"'));
  assert.doesNotMatch(block, /continue-on-error|\|\| true/);
});

test('native scalar BMI execution cannot be replaced by skips, output text or an incomplete owner', () => {
  const event = (Action, Test) => ({Action, Package:'github.com/xgo-dev/plan9asm', ...(Test ? {Test} : {})});
  const events = [event('start'), ...roots.flatMap(root => [event('run', root), event('pass', root)]), event('pass')];
  const raw = values => values.map(JSON.stringify).join('\n') + '\n';
  const inventory = roots.join('\n') + '\n';
  assert.deepEqual(auditScalarBMI(inventory, raw(events)),
    {roots:2, children:0, passed:2, skipped:0, incomplete:0});
  for (const root of roots) {
    assert.throws(() => auditScalarBMI(inventory, raw(events.filter(event => event.Test !== root))));
    assert.throws(() => auditScalarBMI(inventory, raw(events.map(event =>
      event.Test === root && event.Action === 'pass' ? {...event, Action:'skip'} : event))));
  }
  assert.throws(() => auditScalarBMI(inventory, raw(events.slice(0, -1))));
  assert.throws(() => auditScalarBMI(inventory,
    raw([event('start'), {...event('output'), Output:roots.join('\nPASS ')}, event('pass')])));
});
