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

test('all other entry jobs require successful priority completion', () => {
  for (const name of [
    'fmt', 'build', 'benchmark', 'reported_library_list', 'discovered_library_corpus',
    'stdlib-corpus', 'cross-runtime', 'test', 'arm-scan', 'race', 'coverage',
  ]) {
    const section = jobSection(name);
    assert.match(section, /^    needs: discovered_library_priority$/m, name);
    // Native needs success gating must not be bypassed by always()/!cancelled().
    assert.doesNotMatch(section, /^    if:/m, name);
  }
  assert.doesNotMatch(workflow, /^  priority_started:/m);
});

test('all failed shards run and no priority failure can become success', () => {
  const priority = jobSection('discovered_library_priority');
  assert.match(priority, /^      fail-fast: false$/m);
  assert.doesNotMatch(priority, /continue-on-error:/);
  assert.doesNotMatch(priority, /^    (needs|if):/m);
});
