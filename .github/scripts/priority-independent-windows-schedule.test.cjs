'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const policy = require('./workflow-needs.cjs');
const {qualifyDiagnosticNeeds} = require('./diagnostic-qualification.cjs');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
// Independent original qualification inventory. Earlier scheduling must not
// substitute a Windows success for any SDK, native or hardware owner.
const scheduling = ['ownership'];
const qualification = [
  'ownership', 'native-x86-oracles', 'sdk-priority', 'arm64-fallback-cross-priority',
  'lock-native-priority', 'unlocked-rmw-native-priority', 'vector-value-fault-priority',
  'packed-conversion-native-complete', 'packed-reduction-native-complete', 'complete-windows',
];

function replaceNeeds(source, name, needs) {
  const body = policy.jobBody(source, name);
  return source.replace(body, body.replace(/^    needs:.*$/m, '    needs: ' + JSON.stringify(needs)));
}

test('complete Windows starts after ownership without waiting for final qualification owners', () => {
  assert.deepEqual(policy.jobNeeds(workflow, 'complete-windows'), scheduling,
    'the diagnostic matrix is still serialized behind final success owners');
  const body = policy.jobBody(workflow, 'complete-windows');
  assert.deepEqual(body.split('\n').filter(line => /^    if:/.test(line)), ['    if: ${{ !cancelled() }}']);
  assert.match(body, /^    uses: \.\/\.github\/workflows\/windows-module\.yml$/m);
});

test('the shared dependency guard accepts early scheduling but still rejects every missing final gate', () => {
  const early = replaceNeeds(workflow, 'complete-windows', scheduling);
  assert.doesNotThrow(() => policy.requireCompleteWindowsNeeds(early),
    'one shared predecessor set still binds scheduling to final qualification');
  for (const missing of qualification) {
    const changed = replaceNeeds(early, 'diagnostic-qualification', qualification.filter(name => name !== missing));
    assert.throws(() => policy.requireCompleteWindowsNeeds(changed), undefined, missing);
  }
  for (const delayed of qualification.filter(name => name !== 'ownership')) {
    assert.throws(() => policy.requireCompleteWindowsNeeds(replaceNeeds(early, 'complete-windows', ['ownership', delayed])),
      undefined, 'a terminal qualification owner must not become a scheduling barrier: ' + delayed);
  }
});

test('scheduling and final policy inventories are separate exact roles', () => {
  assert.deepEqual(policy.windowsSchedulingNeeds, scheduling);
  assert.deepEqual(policy.finalRequiredNeeds, qualification);
  assert.notStrictEqual(policy.windowsSchedulingNeeds, policy.finalRequiredNeeds);
  assert.deepEqual(policy.jobNeeds(workflow, 'diagnostic-qualification'), qualification);
});

test('early Windows completion cannot qualify a pending failed or absent original owner', () => {
  const successful = () => Object.fromEntries(qualification.map(name => [name, {result: 'success', outputs: {}}]));
  assert.deepEqual(qualifyDiagnosticNeeds(successful()), {schema: 1, successfulJobs: qualification});
  for (const name of qualification) {
    for (const result of ['failure', 'cancelled', 'skipped', 'queued', 'in_progress']) {
      const changed = successful();
      changed[name] = {result, outputs: {result: 'success', qualification: 'success'}};
      assert.throws(() => qualifyDiagnosticNeeds(changed), undefined, name + '/' + result);
    }
    const changed = successful();
    delete changed[name];
    assert.throws(() => qualifyDiagnosticNeeds(changed), undefined, 'absent ' + name);
  }
});
