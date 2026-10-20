'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const {jobBody} = require('./workflow-needs.cjs');

const source = fs.readFileSync(path.join(__dirname, '../workflows/go-ci.yml'), 'utf8');
const exactSDK = '1.27.2';
const minors = ['1.20.x', '1.21.x', '1.22.x', '1.23.x', '1.24.x',
  '1.25.x', '1.26.x', exactSDK];
const matrixJobs = new Set(['stdlib-corpus', 'test', 'module-completion']);
const jobs = ['fmt', 'build', 'benchmark', 'stdlib-corpus', 'cross-runtime',
  'cross-runtime-verify', 'test', 'arm-scan', 'race', 'coverage', 'module-completion'];

function requireSetups(workflow) {
  for (const name of jobs) {
    const body = jobBody(workflow, name);
    const setups = body.split(/(?=^      - )/m).filter(step =>
      /^        uses: \.\/\.github\/actions\/setup-go$/m.test(step));
    assert.equal(setups.length, name === 'stdlib-corpus' ? 2 : 1,
      name + ': every actual Go installation must be covered');
    const expected = matrixJobs.has(name) ? '${{ matrix.go-version }}' : `'${exactSDK}'`;
    const versions = setups.map(step => {
      assert.match(step, /^        with:$/m);
      assert.doesNotMatch(step, /go-version-file:|check-latest:/);
      const fields = [...step.matchAll(/^          go-version: (.+)$/gm)];
      assert.equal(fields.length, 1, name + ': one explicit SDK selection required');
      return fields[0][1];
    });
    assert.deepEqual(versions, name === 'stdlib-corpus' ? [`'${exactSDK}'`, expected] : [expected],
      name + ': current SDK must use the registered exact producer');
    assert.doesNotMatch(body, /uses: actions\/setup-go@/);
  }
}

function requireMatrices(workflow) {
  const regular = jobBody(workflow, 'test');
  const versions = regular.match(/^        go-version: \[(.+)\]$/m);
  assert.ok(versions, 'complete test SDK matrix required');
  assert.deepEqual([...versions[1].matchAll(/'([^']+)'/g)].map(match => match[1]), minors);
  const nested = [...regular.matchAll(/- \{go-version: '([^']+)', nested_modules: true\}/g)];
  assert.deepEqual(nested.map(match => match[1]), minors.slice(4));

  const corpus = [...jobBody(workflow, 'stdlib-corpus').matchAll(
    /^          - \{os: ([^,]+), .*go-version: '([^']+)', .*targets: '([^']+)'\}$/gm)];
  assert.equal(corpus.length, 10, 'eight Linux SDK lanes and both native hosts remain required');
  assert.deepEqual(corpus.slice(0, 8).map(match => [match[1], match[2]]),
    minors.map(version => ['ubuntu-latest', version]));
  assert.deepEqual(corpus.slice(8).map(match => [match[1], match[2]]),
    [['macos-latest', exactSDK], ['windows-latest', exactSDK]]);

  const complete = [...jobBody(workflow, 'module-completion').matchAll(
    /^          - \{flavor: ([^,]+), go-version: '([^']+)'\}$/gm)];
  assert.deepEqual(complete.map(match => [match[1], match[2]]),
    [...minors.map(version => ['test', version]), ['race', exactSDK], ['coverage', exactSDK]]);
}

function requireCurrentFixture(workflow) {
  const body = jobBody(workflow, 'test');
  const step = body.split(/(?=^      - )/m).find(step =>
    step.startsWith('      - name: Required exact downloaded AMD64 numerical fixture\n'));
  assert.ok(step, 'current-only numerical owner must remain present');
  assert.ok(step.includes("if: runner.os == 'Linux' && matrix.go-version == '" +
    exactSDK + "' && matrix.module-owner == 2\n"),
  'the exact current matrix must still activate its numerical owner');
  assert.match(step, /go test \.\/gotoolprofile -run="\$scope" -count=1 -timeout=10m/);
}

function requireArtifactScope(workflow) {
  for (const flavor of ['race', 'coverage']) {
    assert.ok(jobBody(workflow, flavor).includes(
      'name: module-' + flavor + '-' + exactSDK + '-${{ matrix.module-role }}\n'),
    flavor + ': artifact identity must match the exact completion matrix');
  }
  assert.ok(jobBody(workflow, 'windows-module').includes(
    'name: test (windows-latest, ' + exactSDK + ', ${{ matrix.module }})\n'));
  assert.doesNotMatch(workflow, /1\.27\.x/, 'no current floating SDK selector or stale scope remains');
}

function requireCurrentScope(workflow) {
  requireSetups(workflow);
  requireMatrices(workflow);
  requireCurrentFixture(workflow);
  requireArtifactScope(workflow);
}

// Negative controls start from a complete positive fixture while the actual
// workflow is RED. This does not relax compiler-observation admission.
const fixture = source.replaceAll('1.27.x', exactSDK);

test('formal current Go owners use the audited exact SDK', () => {
  requireSetups(source);
});

test('all eight minor lanes and exact current completion scopes remain present', () => {
  requireMatrices(source);
});

test('the exact current matrix still executes its required numerical fixture', () => {
  requireCurrentFixture(source);
});

test('current artifacts and Windows labels identify the selected exact SDK', () => {
  requireArtifactScope(source);
});

test('every current owner rejects a foreign patch or floating selector', () => {
  requireCurrentScope(fixture);
  for (const name of jobs) {
    const before = jobBody(fixture, name);
    for (const version of ['1.27.3', '1.27.x', '1.27', 'stable']) {
      const after = before.replaceAll(exactSDK, version);
      assert.notEqual(after, before, name + ': mutation must change the actual owner');
      assert.throws(() => requireCurrentScope(fixture.replace(before, after)), name + ': ' + version);
    }
  }
});

test('neither stale current-only conditions nor artifact names can borrow another pin', () => {
  requireCurrentScope(fixture);
  for (const [before, after] of [
    ["matrix.go-version == '1.27.2'", "matrix.go-version == '1.27.x'"],
    ['module-race-1.27.2-', 'module-race-1.27.x-'],
    ['module-coverage-1.27.2-', 'module-coverage-1.27.3-'],
    ["{flavor: coverage, go-version: '1.27.2'}", "{flavor: coverage, go-version: '1.27.3'}"],
  ]) {
    assert.ok(fixture.includes(before), 'mutation must affect the actual workflow');
    assert.throws(() => requireCurrentScope(fixture.replace(before, after)));
  }
});
