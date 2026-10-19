'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const {jobBody} = require('./workflow-needs.cjs');

const workflow = fs.readFileSync(path.join(__dirname, '../workflows/pr40-owned-process-probe.yml'), 'utf8');
const condition = "${{ !cancelled() && steps.native_toolchain.outcome == 'success' && steps.native_go.outcome == 'success' }}";
const owners = [
  'Check actual native flags, ABI transport and guarded memory',
  'Execute original Go and LLVM native MMX streaming stores',
  'Execute LLVM floating comparison results and native FP effects',
  'Execute translated scalar BMI values and flags',
  'Execute complete numerical regression programs',
  'Execute original native duplicate moves and whole-vector effects',
  'Execute original native scalar reciprocal estimates',
  'Execute original native retained LEAW pointer bytes and flags',
  'Execute original complete floating programs with actual EVEX DQ state',
  'Execute original native integer stores align and division',
];

function check(source) {
  const job = jobBody(source, 'native-x86-oracles');
  assert.match(job, /- name: Install LLVM 22 and required native i386 toolchain\n        id: native_toolchain\n/);
  assert.match(job, /- uses: \.\/\.github\/actions\/setup-go\n        id: native_go\n/);
  const steps = job.split(/(?=^      - )/m);
  const execution = steps.filter(step => step.includes('GOTOOLCHAIN: local'));
  assert.equal(execution.length, owners.length);
  for (const name of owners) {
    const matches = execution.filter(step => step.startsWith('      - name: ' + name + '\n'));
    assert.equal(matches.length, 1, name);
    const step = matches[0];
    assert.ok(step.includes('        if: ' + condition + '\n'), name + ' may be hidden by a prior owner failure');
    assert.equal((step.match(/^        if:/gm) || []).length, 1);
    assert.match(step, /set -euo pipefail/);
    assert.match(step, /go test \. -run/);
    assert.match(step, /node \.github\/scripts\/audit-/);
  }
  assert.doesNotMatch(job, /continue-on-error|\|\| true/);
}

test('independent native receipt owners keep running after another owner fails', () => {
  check(workflow);
});

test('native isolation policy rejects missing setup gates and masked failures', () => {
  // Construct a qualifying policy fixture without changing the workflow.
  let qualified = workflow
    .replace('      - name: Install LLVM 22 and required native i386 toolchain\n',
      '      - name: Install LLVM 22 and required native i386 toolchain\n        id: native_toolchain\n')
    .replace('      - uses: ./.github/actions/setup-go\n',
      '      - uses: ./.github/actions/setup-go\n        id: native_go\n');
  for (const name of owners) {
    const title = '      - name: ' + name + '\n';
    qualified = qualified.replace(title, title + '        if: ' + condition + '\n');
  }
  check(qualified);
  for (const name of owners) {
    const title = '      - name: ' + name + '\n';
    for (const replacement of ['', '        if: ${{ success() }}\n', '        if: ${{ always() }}\n']) {
      assert.throws(() => check(qualified.replace(title + '        if: ' + condition + '\n', title + replacement)), name);
    }
  }
  for (const mutation of [
    qualified.replace('id: native_toolchain', 'id: other_toolchain'),
    qualified.replace('id: native_go', 'id: other_go'),
    qualified.replace('set -euo pipefail', 'set -euo pipefail\n          continue-on-error: true'),
  ]) assert.throws(() => check(mutation));
});
