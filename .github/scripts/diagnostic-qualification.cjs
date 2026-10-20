'use strict';

const assert = require('node:assert/strict');
const {windowsRequiredNeeds} = require('./workflow-needs.cjs');

const required = [...windowsRequiredNeeds, 'complete-windows'];

// Scheduling Windows after a failed predecessor discovers its real failures;
// it cannot supply success for that predecessor. GitHub's current-run needs
// observations remain the only input, never a job's self-reported outputs.
function qualifyDiagnosticNeeds(needs) {
  assert.ok(needs && typeof needs === 'object' && !Array.isArray(needs),
    'missing actual diagnostic dependency observations');
  assert.deepEqual(Object.keys(needs).sort(), required.slice().sort(),
    'diagnostic qualification requires every original gate and complete Windows');
  for (const name of required) {
    const observation = needs[name];
    assert.ok(observation && typeof observation === 'object' && !Array.isArray(observation),
      'malformed actual dependency result: ' + name);
    assert.ok(Object.hasOwn(observation, 'result'), 'absent actual dependency result: ' + name);
    assert.equal(observation.result, 'success', 'required diagnostic gate did not succeed: ' + name);
  }
  return {schema: 1, successfulJobs: required.slice()};
}

module.exports = {qualifyDiagnosticNeeds};
if (require.main === module) {
  const input = process.env.DIAGNOSTIC_NEEDS_JSON;
  assert.equal(typeof input, 'string', 'actual GitHub needs context is required');
  const result = qualifyDiagnosticNeeds(JSON.parse(input));
  process.stdout.write(JSON.stringify(result) + '\n');
}
