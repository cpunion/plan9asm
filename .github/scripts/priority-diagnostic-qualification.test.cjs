'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const test = require('node:test');
const {jobBody, jobNeeds, windowsRequiredNeeds} = require('./workflow-needs.cjs');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const allNeeds = [...windowsRequiredNeeds, 'complete-windows'];
const condition = '    if: ${{ !cancelled() }}';

function requireSchedule(source) {
  assert.deepEqual(jobNeeds(source, 'complete-windows').slice().sort(),
    windowsRequiredNeeds.slice().sort());
  const windows = jobBody(source, 'complete-windows');
  assert.deepEqual(windows.split('\n').filter(line => /^    if:/.test(line)), [condition],
    'Windows execution must wait for terminal dependencies, not require their success');
  assert.match(windows, /^    uses: \.\/\.github\/workflows\/windows-module\.yml$/m);
}

function requireQualification(source) {
  assert.deepEqual(jobNeeds(source, 'diagnostic-qualification').slice().sort(), allNeeds.slice().sort(),
    'final qualification must own all original gates and actual Windows completion');
  const final = jobBody(source, 'diagnostic-qualification');
  assert.deepEqual(final.split('\n').filter(line => /^    if:/.test(line)), [condition]);
  assert.match(final, /      - name: Require all original gates and complete Windows\n        env:\n          DIAGNOSTIC_NEEDS_JSON: \$\{\{ toJSON\(needs\) \}\}\n        run: node \.github\/scripts\/diagnostic-qualification\.cjs(?:\n|$)/);
  assert.doesNotMatch(final, /continue-on-error|\|\|\s*true/);
}

function successfulNeeds() {
  return Object.fromEntries(allNeeds.map(name => [name, {result: 'success', outputs: {}}]));
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

test('Windows execution retains all nine predecessors but is not hidden by their failure', () => {
  assert.equal(windowsRequiredNeeds.length, 9);
  requireSchedule(workflow);
});

test('an independent final diagnostic qualification still requires every actual success', () => {
  requireQualification(workflow);
});

test('qualification policy rejects removed gates, success-only scheduling and unbound observations', () => {
  requireSchedule(workflow);
  requireQualification(workflow);
  for (const name of ['complete-windows', 'diagnostic-qualification']) {
    const body = jobBody(workflow, name);
    for (const replacement of ['', '    if: ${{ success() }}', '    if: ${{ always() }}']) {
      const changed = workflow.replace(body, body.replace(condition, replacement));
      assert.throws(() => name === 'complete-windows' ? requireSchedule(changed) : requireQualification(changed));
    }
    const expected = name === 'complete-windows' ? windowsRequiredNeeds : allNeeds;
    for (const missing of expected) {
      const changed = workflow.replace(body, body.replace(/^    needs:.*$/m,
        '    needs: [' + expected.filter(value => value !== missing).join(', ') + ']'));
      assert.throws(() => name === 'complete-windows' ? requireSchedule(changed) : requireQualification(changed),
        undefined, 'omitted required gate: ' + missing);
    }
    const duplicate = workflow.replace(body, body.replace(/^    needs:.*$/m,
      '    needs: [' + [...expected, expected[0]].join(', ') + ']'));
    assert.throws(() => name === 'complete-windows' ? requireSchedule(duplicate) : requireQualification(duplicate));
  }
  for (const changed of [
    workflow.replace('DIAGNOSTIC_NEEDS_JSON: ${{ toJSON(needs) }}', 'DIAGNOSTIC_NEEDS_JSON: \'{}\''),
    workflow.replace('run: node .github/scripts/diagnostic-qualification.cjs', 'run: echo qualification passed'),
    workflow.replace('run: node .github/scripts/diagnostic-qualification.cjs',
      'run: node .github/scripts/diagnostic-qualification.cjs || true'),
  ]) assert.throws(() => requireQualification(changed));
});

test('only all ten genuine success results qualify; Windows cannot replace an original gate', () => {
  const {qualifyDiagnosticNeeds} = require('./diagnostic-qualification.cjs');
  assert.deepEqual(qualifyDiagnosticNeeds(successfulNeeds()), {schema: 1, successfulJobs: allNeeds});
  for (const name of allNeeds) {
    for (const result of ['failure', 'skipped', 'cancelled', 'queued', 'in_progress', '', true, null]) {
      const needs = successfulNeeds();
      needs[name].result = result;
      assert.throws(() => qualifyDiagnosticNeeds(needs), undefined, name + ': ' + result);
    }
    const missing = successfulNeeds();
    delete missing[name];
    assert.throws(() => qualifyDiagnosticNeeds(missing), undefined, 'missing ' + name);
    for (const malformed of [null, [], 'success', {outputs: {result: 'success'}}]) {
      const needs = successfulNeeds();
      needs[name] = malformed;
      assert.throws(() => qualifyDiagnosticNeeds(needs), undefined, 'malformed ' + name);
    }
  }
  for (const needs of [null, [], true, 'success', {},
    {...successfulNeeds(), extra: {result: 'success', outputs: {}}}]) {
    assert.throws(() => qualifyDiagnosticNeeds(needs));
  }
  const forged = clone(successfulNeeds());
  forged['native-x86-oracles'] = {result: 'failure', outputs: {result: 'success'}};
  forged['complete-windows'].outputs = {qualification: 'success'};
  assert.throws(() => qualifyDiagnosticNeeds(forged));
});

test('the actual qualification command exits nonzero for failed, skipped, absent or malformed needs', () => {
  const script = path.join(__dirname, 'diagnostic-qualification.cjs');
  const execute = input => spawnSync(process.execPath, [script], {
    encoding: 'utf8', timeout: 5000,
    env: {...process.env, DIAGNOSTIC_NEEDS_JSON: input},
  });
  const success = execute(JSON.stringify(successfulNeeds()));
  assert.equal(success.error, undefined);
  assert.equal(success.status, 0, success.stderr);
  assert.deepEqual(JSON.parse(success.stdout), {schema: 1, successfulJobs: allNeeds});
  for (const name of allNeeds) {
    for (const result of ['failure', 'skipped', 'cancelled']) {
      const needs = successfulNeeds();
      needs[name].result = result;
      const failed = execute(JSON.stringify(needs));
      assert.equal(failed.error, undefined);
      assert.notEqual(failed.status, 0, name + ': ' + result);
      assert.equal(failed.stdout, '');
    }
    const missing = successfulNeeds();
    delete missing[name];
    assert.notEqual(execute(JSON.stringify(missing)).status, 0, 'missing ' + name);
  }
  for (const input of ['', '{', 'null', '[]', '{}']) {
    assert.notEqual(execute(input).status, 0, input);
  }
});
