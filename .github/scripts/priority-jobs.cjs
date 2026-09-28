const { setTimeout: delay } = require('node:timers/promises');

// Failed discovery shards from the last complete fork run. The workflow
// regression test keeps this list and the two exhaustive matrices in sync.
const priorityShards = [0, 9, 14, 28, 29];

async function waitForPriorityJobs({
  listJobs,
  log = console.log,
  sleep = delay,
  maxPolls = 120,
}) {
  const names = priorityShards.map(shard => `discovered-corpus (shard ${shard}/32)`);
  let pending = names;

  for (let poll = 0; poll < maxPolls; poll++) {
    const jobs = await listJobs();
    pending = names.filter(name => {
      const matches = jobs.filter(job => job.name === name);
      if (matches.length > 1) {
        throw new Error(`duplicate priority job in current attempt: ${name}`);
      }

      const job = matches[0];
      const started = job && (job.status === 'in_progress' ||
        (job.status === 'completed' &&
          ['success', 'failure', 'timed_out'].includes(job.conclusion)));
      return !started;
    });

    if (pending.length === 0) {
      log('All five priority shards have started; releasing the remaining CI jobs.');
      return;
    }

    log(`Waiting for priority jobs to start: ${pending.join(', ')}`);
    if (poll + 1 < maxPolls) {
      await sleep(10000);
    }
  }

  throw new Error(`Priority jobs have not started: ${pending.join(', ')}`);
}

module.exports = { priorityShards, waitForPriorityJobs };
