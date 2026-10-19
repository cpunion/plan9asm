'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {jobBody, jobNeeds, requireJobNeeds, requireCompleteWindowsNeeds} = require('./workflow-needs.cjs');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
// Independent complete gate list: adding a gate to the implementation helper
// cannot replace an original success barrier or silently change this oracle.
const gates = [
  'ownership', 'native-x86-oracles', 'sdk-priority', 'arm64-fallback-cross-priority',
  'lock-native-priority', 'unlocked-rmw-native-priority', 'vector-value-fault-priority',
  'packed-conversion-native-complete', 'packed-reduction-native-complete',
];

function replaceNeeds(source, name, declaration) {
  const body = jobBody(source, name);
  const changed = body.replace(/^    needs:.*$/m, '    needs: ' + declaration);
  assert.notEqual(body, changed, 'negative/format control must change its actual job');
  return source.replace(body, changed);
}

test('required Windows gates use their own parsed dependency set, independent of order', () => {
  assert.deepEqual(jobNeeds(workflow, 'complete-windows').slice().sort(), gates.slice().sort());
  requireCompleteWindowsNeeds(workflow);
  const reversed = replaceNeeds(workflow, 'complete-windows', '[' + gates.slice().reverse().join(', ') + ']');
  requireCompleteWindowsNeeds(reversed);
  const quotes = replaceNeeds(workflow, 'complete-windows', JSON.stringify(gates));
  requireCompleteWindowsNeeds(quotes);
  const extended = reversed + '\n  additional-required-gate:\n    runs-on: ubuntu-latest\n';
  requireCompleteWindowsNeeds(replaceNeeds(extended, 'complete-windows',
    '[' + [...gates, 'additional-required-gate'].join(', ') + ']'));
});

test('omitting any original or either new native gate still fails', () => {
  for (const gate of gates) {
    const reduced = '[' + gates.filter(name => name !== gate).join(', ') + ']';
    assert.throws(() => requireCompleteWindowsNeeds(replaceNeeds(workflow, 'complete-windows', reduced)),
      undefined, 'missing gate must fail: ' + gate);
  }
});

test('duplicate, non-concrete, self, absent or borrowed dependency declarations fail', () => {
  for (const declaration of [
    '[' + [...gates, gates[0]].join(', ') + ']',
    '[' + [...gates, 'not-a-real-job'].join(', ') + ']',
    '[' + [...gates, 'complete-windows'].join(', ') + ']',
    '${{ needs.from_another_job }}', '[]', '[ownership,]',
  ]) assert.throws(() => requireCompleteWindowsNeeds(replaceNeeds(workflow, 'complete-windows', declaration)));
  const body = jobBody(workflow, 'complete-windows');
  assert.throws(() => requireCompleteWindowsNeeds(workflow.replace(body,
    body.replace(/^    needs:.*$/m, '    needs: ownership\n    needs: sdk-priority'))));
  assert.throws(() => requireCompleteWindowsNeeds(workflow.replace(body,
    body.replace(/^    needs:.*\n/m, ''))));
  assert.throws(() => requireCompleteWindowsNeeds(workflow + '\n  complete-windows:\n'));
  assert.throws(() => requireCompleteWindowsNeeds(workflow.replace('\n  ownership:\n', '\n  missing-ownership:\n')));
});

test('main and reusable CLI completion retain every actual owner gate', () => {
  for (const [file, name, required] of [
    ['go-ci.yml', 'module-completion', ['test', 'race', 'coverage']],
    ['windows-module.yml', 'complete', ['root-parts', 'other-owners']],
  ]) {
    const source = fs.readFileSync('.github/workflows/' + file, 'utf8');
    requireJobNeeds(source, name, required);
    requireJobNeeds(replaceNeeds(source, name, '[' + required.slice().reverse().join(', ') + ']'), name, required);
    for (const missing of required) {
      const reduced = '[' + required.filter(gate => gate !== missing).join(', ') + ']';
      assert.throws(() => requireJobNeeds(replaceNeeds(source, name, reduced), name, required));
    }
  }
});

test('concrete block and scalar needs retain the same owned success requirements', () => {
  const source = 'jobs:\n  first:\n    runs-on: ubuntu-latest\n' +
    '  second:\n    needs: first\n' +
    '  complete:\n    needs:\n      - first\n      - second\n';
  assert.deepEqual(jobNeeds(source, 'second'), ['first']);
  assert.deepEqual(requireJobNeeds(source, 'complete', ['second', 'first']), ['first', 'second']);
  assert.throws(() => requireJobNeeds(source.replace('      - second\n', ''), 'complete', ['first', 'second']));
});
