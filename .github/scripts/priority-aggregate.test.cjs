const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const test = require('node:test');

const root = path.resolve(__dirname, '../..');
const source = fs.readFileSync(path.join(root, 'scripts/verify-discovered-library-corpus.sh'), 'utf8');

test('aggregate audits reports once for progress, strict completion and ledger comparison', () => {
  const workflow = fs.readFileSync(path.join(root, '.github/workflows/go-ci.yml'), 'utf8');
  const aggregate = workflow.split('\n  discovered_library_corpus_verify:\n')[1].split('\n  stdlib-corpus:\n')[0];
  assert.equal((aggregate.match(/run: scripts\/verify-discovered-library-corpus.sh --progress/g) || []).length, 1);
  assert.doesNotMatch(aggregate, /scripts\/discovery-status.sh/);
  assert.match(aggregate, /Upload discovery progress\n        if: \$\{\{ !cancelled\(\) \}\}/);

  const main = fs.readFileSync(path.join(root, 'cmd/plan9asmcorpus/main.go'), 'utf8');
  const verification = main.split('\n\tif *verifyDiscoveryReports != "" {')[1].split('\n\tif *assemblyLedgerStatusPath != "" {')[0];
  assert.equal((verification.match(/collectDiscoveryProgress\(/g) || []).length, 1);
  assert.doesNotMatch(verification, /verifyDiscoveryCorpusReports\(/);
});

function verifyFixture(t, mode) {
  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'plan9asm-aggregate-test-'));
  t.after(() => fs.rmSync(temporary, { recursive: true, force: true }));
  const scripts = path.join(temporary, 'scripts');
  const bin = path.join(temporary, 'bin');
  const reports = path.join(temporary, 'reports');
  const calls = path.join(temporary, 'calls.jsonl');
  fs.mkdirSync(scripts);
  fs.mkdirSync(bin);
  fs.writeFileSync(path.join(scripts, 'verify-discovered-library-corpus.sh'), source, { mode: 0o755 });
  fs.writeFileSync(path.join(bin, 'go'), `#!${process.execPath}
const fs = require('node:fs');
const args = process.argv.slice(2);
fs.appendFileSync(process.env.FIXTURE_CALLS, JSON.stringify(args) + '\\n');
if (args.includes('./cmd/plan9asmdiscover')) {
  process.stdout.write('{"index_entries":12}\\n');
} else if (args.includes('-discovery-progress-output')) {
  if (process.env.FIXTURE_MODE !== 'invalid') {
    const output = args[args.indexOf('-discovery-progress-output') + 1];
    fs.writeFileSync(output, JSON.stringify({complete: process.env.FIXTURE_MODE === 'pass', verified: process.env.FIXTURE_MODE === 'pass'}) + '\\n', {flag: 'wx'});
  }
  if (process.env.FIXTURE_MODE !== 'pass') {
    process.stderr.write('strict reader remains failed\\n');
    process.exit(1);
  }
} else {
  process.stderr.write('expected one combined strict/progress invocation\\n');
  process.exit(9);
}
`, { mode: 0o755 });
  const result = spawnSync('bash', [path.join(scripts, 'verify-discovered-library-corpus.sh'), '--progress', reports], {
    encoding: 'utf8', timeout: 10000,
    env: { ...process.env, PATH: bin + path.delimiter + process.env.PATH, FIXTURE_CALLS: calls, FIXTURE_MODE: mode },
  });
  assert.ifError(result.error);
  const invocations = fs.existsSync(calls) ? fs.readFileSync(calls, 'utf8').trim().split('\n').map(JSON.parse) : [];
  return { result, invocations, reports };
}

for (const mode of ['pass', 'partial', 'invalid']) {
  test(`combined aggregate ${mode} preserves the strict exit and only audited progress`, (t) => {
    const { result, invocations, reports } = verifyFixture(t, mode);
    assert.equal(result.status, mode === 'pass' ? 0 : 1, result.stderr);
    assert.equal(invocations.length, 2);
    assert.ok(invocations[1].includes('-discovery-progress-output'));
    assert.ok(invocations[1].includes('-verify-discovery-reports'));
    assert.ok(invocations[1].includes('-compare-assembly-ledger'));
    assert.deepEqual(JSON.parse(fs.readFileSync(path.join(reports, 'scan-status.json'), 'utf8')), { index_entries: 12 });
    const progress = path.join(reports, 'assembly-progress.json');
    if (mode === 'invalid') {
      assert.equal(fs.existsSync(progress), false);
    } else {
      assert.deepEqual(JSON.parse(fs.readFileSync(progress, 'utf8')), { complete: mode === 'pass', verified: mode === 'pass' });
    }
    assert.deepEqual(fs.readdirSync(reports).filter(name => name.startsWith('.status-')), []);
  });
}
