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

function shards(section) {
  const match = section.match(/^        shard: (\[[^\n]+\])$/m);
  assert.ok(match, 'missing shard matrix');
  return JSON.parse(match[1]);
}

test('priority and remaining matrices partition all 32 shards exactly once', () => {
  const priority = shards(jobSection('discovered_library_priority'));
  const remaining = shards(jobSection('discovered_library_corpus'));
  // Actual failed shards from completed fork run 36371763306.
  assert.deepEqual(priority, [11, 15, 16, 19, 30]);
  assert.deepEqual([...priority, ...remaining].sort((a, b) => a - b),
    Array.from({ length: 32 }, (_, index) => index));
});

test('priority and remaining shards use identical test and artifact steps', () => {
  function steps(name) {
    const section = jobSection(name);
    return section.slice(section.indexOf('    steps:\n')).trim();
  }
  assert.equal(steps('discovered_library_priority'), steps('discovered_library_corpus'));
  assert.match(jobSection('discovered_library_corpus_verify'),
    /needs: \[discovered_library_priority, discovered_library_corpus\]/);
});

test('discovery matrices have no explicit concurrency cap', () => {
  for (const name of ['discovered_library_priority', 'discovered_library_corpus']) {
    assert.doesNotMatch(jobSection(name), /^      max-parallel:/m, name);
  }
});

test('discovery matrices verify checksums against direct signed databases', () => {
  for (const name of ['discovered_library_priority', 'discovered_library_corpus']) {
    const section = jobSection(name);
    assert.match(section, /^          GOSUMDB:.*sum\.golang\.google\.cn.*sum\.golang\.org https:\/\/sum\.golang\.org/m, name);
    assert.doesNotMatch(section, /^          GOSUMDB: ['"]?off/m, name);
  }
});
