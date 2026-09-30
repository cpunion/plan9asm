const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
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
  const match = section.match(/^        shard:\n((?:          - \d+\n)+)/m);
  assert.ok(match, 'missing shard matrix');
  return Array.from(match[1].matchAll(/^          - (\d+)$/gm),
    ([, value]) => Number(value));
}

test('priority and remaining matrices partition all 64 shards exactly once', () => {
  const priority = shards(jobSection('discovered_library_priority'));
  const remaining = shards(jobSection('discovered_library_corpus'));
  // Retain instruction/ABI regressions and the cache cleanup failure from
  // fork run 36509240233.
  assert.deepEqual(priority, [5, 29, 36, 38, 39, 42]);
  assert.deepEqual([...priority, ...remaining].sort((a, b) => a - b),
    Array.from({ length: 64 }, (_, index) => index));
  for (const name of ['discovered_library_priority', 'discovered_library_corpus']) {
    const section = jobSection(name);
    assert.match(section, /discovered-corpus \(shard \$\{\{ matrix\.shard \}\}\/64\)/);
    assert.match(section, /check-discovered-library-corpus\.sh "\$\{\{ matrix\.shard \}\}" 64/);
  }
});

test('failed exact versions rehash into priority shards', () => {
  const priority = shards(jobSection('discovered_library_priority'));
  for (const exactVersion of [
    'github.com/Qitmeer/go-ethereum@v1.10.12',
    'github.com/RookieCoderrr/neo3fura-ctrverification@v0.0.0-20221201045318-9878de6dbeed',
    'github.com/SysVerification/gokv@v0.0.0-20250508184610-d007325b6ee8',
    'github.com/reddit/milvus/pkg/v3@v3.0.0-20260702082229-182134e29ebf',
    'github.com/apernet/hysteria/app/v2@v2.12.3',
    'github.com/HyNetwork/hysteria/app@v1.3.5',
    'github.com/intel/ixl-go@v0.0.0-20260223204051-3e809238f558',
  ]) {
    const hash = createHash('sha256').update(exactVersion).digest();
    const shard = Number(hash.readBigUInt64BE(0) % 64n);
    assert.ok(priority.includes(shard), `${exactVersion} belongs to shard ${shard}`);
  }
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

test('local discovery helpers use the CI shard count by default', () => {
  for (const name of [
    'check-discovered-library-corpus.sh',
    'discovery-status.sh',
    'update-assembly-ledger.sh',
  ]) {
    const script = fs.readFileSync(path.join(__dirname, '../../scripts', name), 'utf8');
    assert.match(script, /shard_count=\$\{2:-64\}/, name);
  }
});

test('discovery matrices verify checksums against direct signed databases', () => {
  for (const name of ['discovered_library_priority', 'discovered_library_corpus']) {
    const section = jobSection(name);
    assert.match(section, /^          GOSUMDB:.*sum\.golang\.google\.cn.*sum\.golang\.org https:\/\/sum\.golang\.org/m, name);
    assert.doesNotMatch(section, /^          GOSUMDB: ['"]?off/m, name);
  }
});
