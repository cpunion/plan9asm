const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const test = require('node:test');

const source = fs.readFileSync(
  path.resolve(__dirname, '../../scripts/check-discovered-library-corpus.sh'),
  'utf8',
);

function runCorpusFixture(t, failedShard = '', priorityShards = '') {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'plan9asm-cache-test-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const scripts = path.join(root, 'scripts');
  const bin = path.join(root, 'bin');
  fs.mkdirSync(scripts);
  fs.mkdirSync(bin);
  const calls = path.join(root, 'calls');
  const verified = path.join(root, 'verified');
  const caches = path.join(root, 'caches');
  fs.mkdirSync(caches);
  const runner = path.join(scripts, 'check-discovered-library-corpus.sh');
  fs.writeFileSync(runner, source, { mode: 0o755 });
  fs.writeFileSync(path.join(bin, 'xargs'), `#!/usr/bin/env bash
set -euo pipefail
status=0
while IFS= read -r shard; do
  test -d "$PLAN9ASM_DISCOVERY_BUILD_CACHE"
  printf '%s\t%s\n' "$shard" "$PLAN9ASM_DISCOVERY_BUILD_CACHE" >> "$FIXTURE_CALLS"
  if [[ "$shard" == "$FIXTURE_FAILED_SHARD" ]]; then
    status=123
  fi
done
exit "$status"
`, { mode: 0o755 });
  fs.writeFileSync(path.join(scripts, 'verify-discovered-library-corpus.sh'), `#!/usr/bin/env bash
set -euo pipefail
printf 'verified\n' >> "$FIXTURE_VERIFIED"
`, { mode: 0o755 });
  const result = spawnSync('bash', [runner, 'all', '5'], {
    encoding: 'utf8',
    timeout: 10000,
    env: {
      ...process.env,
      PATH: `${bin}${path.delimiter}${process.env.PATH}`,
      TMPDIR: caches,
      PLAN9ASM_DISCOVERY_PARALLELISM: '2',
      PLAN9ASM_DISCOVERY_PRIORITY_SHARDS: priorityShards,
      FIXTURE_CALLS: calls,
      FIXTURE_VERIFIED: verified,
      FIXTURE_FAILED_SHARD: failedShard,
    },
  });
  assert.ifError(result.error);
  const records = fs.existsSync(calls)
    ? fs.readFileSync(calls, 'utf8').trim().split('\n')
      .map((line) => line.split('\t'))
    : [];
  return { result, records, caches, verified };
}

test('discovery shares a cache only within each bounded parallel batch', (t) => {
  const { result, records, caches, verified } = runCorpusFixture(t);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(records.map(([shard]) => shard), ['0', '1', '2', '3', '4']);
  assert.equal(records[0][1], records[1][1]);
  assert.equal(records[2][1], records[3][1]);
  assert.equal(new Set(records.map(([, cache]) => cache)).size, 3);
  assert.deepEqual(fs.readdirSync(caches), []);
  assert.equal(fs.readFileSync(verified, 'utf8'), 'verified\n');
});

test('failed discovery batches keep failures, clean caches and run the final gate', (t) => {
  const { result, records, caches, verified } = runCorpusFixture(t, '2');
  assert.notEqual(result.status, 0);
  assert.deepEqual(records.map(([shard]) => shard), ['0', '1', '2', '3', '4']);
  assert.deepEqual(fs.readdirSync(caches), []);
  assert.equal(fs.readFileSync(verified, 'utf8'), 'verified\n');
});

test('priority shards run first without omitting or repeating other shards', (t) => {
  const { result, records, caches, verified } = runCorpusFixture(t, '', '4,2');
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(records.map(([shard]) => shard), ['4', '2', '0', '1', '3']);
  assert.equal(records[0][1], records[1][1]);
  assert.equal(records[2][1], records[3][1]);
  assert.equal(new Set(records.map(([, cache]) => cache)).size, 3);
  assert.deepEqual(fs.readdirSync(caches), []);
  assert.equal(fs.readFileSync(verified, 'utf8'), 'verified\n');
});

test('invalid priority lists fail before starting corpus work', (t) => {
  for (const priority of ['5', '2,2', '-1', '1,,2', ',1', '1,', '01', '1 x',
    '18446744073709551616']) {
    const { result, records, caches, verified } = runCorpusFixture(t, '', priority);
    assert.notEqual(result.status, 0, `accepted invalid priority ${priority}`);
    assert.deepEqual(records, []);
    assert.deepEqual(fs.readdirSync(caches), []);
    assert.equal(fs.existsSync(verified), false);
  }
});
