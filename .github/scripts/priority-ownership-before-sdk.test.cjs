const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {test} = require('node:test');

const workflow = fs.readFileSync(path.join(__dirname, '../workflows/pr40-owned-process-probe.yml'), 'utf8');
const sdk = workflow.slice(workflow.indexOf('\n  sdk-priority:\n'), workflow.indexOf('\n  ownership:\n'));

function requireOwnershipGate(job) {
  assert.equal((job.match(/\n    needs:/g) || []).length, 1);
  assert.match(job, /\n    needs: ownership\n/);
}

test('SDK matrix waits for both previously failing ownership versions', () => {
  requireOwnershipGate(sdk);
});

test('SDK ownership gate rejects absent, substituted and duplicate dependencies', () => {
  for (const changed of [
    sdk.replace(/\n    needs: ownership\n/, '\n'),
    sdk.replace('needs: ownership', 'needs: native-x86-oracles'),
    sdk.replace('needs: ownership', 'needs: ownership\n    needs: ownership'),
  ]) {
    assert.throws(() => requireOwnershipGate(changed));
  }
});
