const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const workflow = fs.readFileSync(path.join(__dirname, '../workflows/go-ci.yml'), 'utf8');

function jobSection(name) {
  const start = workflow.indexOf(`\n  ${name}:\n`);
  assert.notEqual(start, -1, `missing job ${name}`);
  const rest = workflow.slice(start + 1);
  const next = rest.slice(1).search(/^  [a-zA-Z0-9_-]+:\s*$/m);
  return next < 0 ? rest : rest.slice(0, next + 1);
}

test('translator jobs require policy preflight, not external compatibility', () => {
  for (const name of [
    'fmt', 'build', 'benchmark',
    'stdlib-corpus', 'cross-runtime', 'test', 'arm-scan', 'race', 'coverage',
  ]) {
    const section = jobSection(name);
    assert.match(section, /^    needs: ci_policy$/m, name);
    // Native needs success gating must not be bypassed by always()/!cancelled().
    assert.doesNotMatch(section, /^    if:/m, name);
  }
  assert.doesNotMatch(workflow, /discovered_library|reported_library|testdata\/discovery/);
});

test('all translator matrix failures remain failures', () => {
  for (const name of ['test', 'race', 'coverage', 'cross-runtime']) {
    const job = jobSection(name);
    assert.match(job, /^      fail-fast: false$/m);
    assert.doesNotMatch(job, /continue-on-error:/);
  }
});

test('cheap CI policy checks pass before instruction and official corpus jobs', () => {
  const policy = jobSection('ci_policy');
  assert.match(policy, /bash scripts\/install-ci-llvm22-ubuntu\.sh/);
  assert.match(policy, /node --test \.github\/scripts\/priority-\*\.test\.cjs/);
  assert.match(policy, /go test \. -run '\^TestCI' -count=1/);
  assert.doesNotMatch(policy, /^    (needs|if):/m);
  assert.ok(workflow.indexOf('\n  ci_policy:\n') <
    workflow.indexOf('\n  test:\n'));
});
