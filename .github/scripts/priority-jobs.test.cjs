const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const { priorityShards, waitForPriorityJobs } = require('./priority-jobs.cjs');

function jobs(status, conclusion = null) {
  return priorityShards.map(shard => ({
    name: `discovered-corpus (shard ${shard}/32)`,
    status,
    conclusion,
    // GitHub can supply a creation timestamp even while the job is queued.
    started_at: '2026-09-28T00:00:00Z',
  }));
}

test('releases other jobs only when every priority job has started', async () => {
  const queued = jobs('queued');
  const partial = jobs('in_progress');
  partial[4].status = 'queued';
  const snapshots = [queued, partial, jobs('in_progress')];
  let polls = 0;
  let waits = 0;

  await waitForPriorityJobs({
    listJobs: async () => snapshots[polls++],
    sleep: async () => { waits++; },
    log: () => {},
    maxPolls: 3,
  });

  assert.equal(polls, 3);
  assert.equal(waits, 2);
});

test('a completed failure counts as started, not as passing', async () => {
  const snapshot = jobs('in_progress');
  snapshot[0].status = 'completed';
  snapshot[0].conclusion = 'failure';

  await waitForPriorityJobs({ listJobs: async () => snapshot, log: () => {} });
  assert.equal(snapshot[0].conclusion, 'failure');
});

test('missing, skipped and cancelled jobs cannot open the startup gate', async () => {
  for (const snapshot of [[], jobs('completed', 'skipped'), jobs('completed', 'cancelled')]) {
    await assert.rejects(waitForPriorityJobs({
      listJobs: async () => snapshot,
      sleep: async () => {},
      log: () => {},
      maxPolls: 1,
    }), /not started/);
  }
});

test('duplicate jobs and API failures cannot produce a successful gate', async () => {
  const snapshot = jobs('in_progress');
  snapshot.push(snapshot[0]);
  await assert.rejects(waitForPriorityJobs({
    listJobs: async () => snapshot,
    log: () => {},
  }), /duplicate/);

  await assert.rejects(waitForPriorityJobs({
    listJobs: async () => { throw new Error('API unavailable'); },
    log: () => {},
  }), /API unavailable/);
});

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
  assert.deepEqual(priority, priorityShards);
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

test('all other entry jobs wait for startup, not priority completion', () => {
  for (const name of [
    'fmt', 'build', 'benchmark', 'reported_library_list', 'discovered_library_corpus',
    'stdlib-corpus', 'cross-runtime', 'test', 'arm-scan', 'race', 'coverage',
  ]) {
    assert.match(jobSection(name), /^    needs: priority_started$/m, name);
  }
  assert.doesNotMatch(jobSection('priority_started'), /^    needs:/m);
  assert.match(jobSection('priority_started'), /listJobsForWorkflowRunAttempt/);
  assert.match(jobSection('priority_started'), /GITHUB_RUN_ATTEMPT/);
});
