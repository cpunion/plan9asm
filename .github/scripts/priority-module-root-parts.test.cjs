const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const workflow = fs.readFileSync(
  path.join(__dirname, '../workflows/go-ci.yml'), 'utf8',
);
const script = fs.readFileSync(
  path.join(__dirname, '../../scripts/test-module-ci.sh'), 'utf8',
);

function job(name) {
  const marker = `\n  ${name}:\n`;
  const start = workflow.indexOf(marker);
  assert.notEqual(start, -1, `missing job ${name}`);
  const body = workflow.slice(start + marker.length);
  const end = body.search(/^  [a-zA-Z][a-zA-Z0-9_-]*:\s*$/m);
  return end < 0 ? body : body.slice(0, end);
}

function step(body, name) {
  const marker = `      - name: ${name}\n`;
  const start = body.indexOf(marker);
  assert.notEqual(start, -1, `missing step ${name}`);
  const end = body.indexOf('\n      - name:', start + marker.length);
  return body.slice(start, end < 0 ? body.length : end);
}

const roles = [
  ...Array.from({ length: 8 }, (_, index) => `root-${index}`),
  'corpus',
  'other',
];

test('each module flavor splits only root into eight exhaustive job roles', () => {
  for (const name of ['test', 'race', 'coverage']) {
    const body = job(name);
    const match = body.match(/^        module-role: \[([^\]]+)\]$/m);
    assert.ok(match, `${name}: missing explicit module scheduling roles`);
    assert.deepEqual(match[1].split(',').map(value => value.trim()), roles);

    const mappings = Array.from(body.matchAll(
      /^          - \{module-role: (root-[0-7]|corpus|other), module-owner: ([012]), root-partition: (-?\d+)\}$/gm,
    ), ([, role, owner, part]) => [role, Number(owner), Number(part)]);
    assert.deepEqual(mappings, [
      ...Array.from({ length: 8 }, (_, index) => [`root-${index}`, 0, index]),
      ['corpus', 1, -1],
      ['other', 2, -1],
    ], `${name}: role mapping repeats or loses package/shard ownership`);
    assert.doesNotMatch(body, /^        module-owner: \[/m,
      `${name}: Cartesian ownership repeats corpus/rest`);
    assert.match(body, /^      fail-fast: false$/m);
    assert.doesNotMatch(body, /max-parallel:|continue-on-error:/);
  }
});

test('root job fanout retains original workers, per-command and job budgets', () => {
  assert.match(script, /test\) timeout=45m ;;/);
  assert.match(script, /race\) timeout=60m; flags=\(-race\) ;;/);
  assert.match(script, /coverage\) timeout=45m; flags=\(-coverprofile=coverage-root\.out\) ;;/);
  assert.match(script, /-shards=32 -workers=2/);
  assert.match(script, /GOMAXPROCS=2/);
  for (const [name, minutes] of [['test', 60], ['race', 75], ['coverage', 60]]) {
    assert.match(job(name), new RegExp(`^    timeout-minutes: ${minutes}$`, 'm'));
  }
  assert.match(job('module-completion'), /^    timeout-minutes: 30$/m);
});

test('executor and completion both request the same explicit root partition count', () => {
  assert.match(script, /^root_partitions=8$/m);
  assert.match(script, /-module-root-partitions=\$root_partitions/);
  assert.match(script, /-module-root-partition=\$4/);
  assert.match(script, /\[\[ \$# != 4/);
  for (const [name, flavor, title] of [
    ['test', 'test', 'Go test (Ubuntu)'],
    ['race', 'race', 'Go test (race)'],
    ['coverage', 'coverage', 'Go test with coverage'],
  ]) {
    assert.match(step(job(name), title), new RegExp(
      `bash scripts/test-module-ci\\.sh ${flavor} execute \\$\\{\\{ matrix\\.module-owner \\}\\} \\$\\{\\{ matrix\\.root-partition \\}\\}`,
    ));
  }
  assert.match(step(job('module-completion'), 'Check complete actual module inventory and coverage'),
    /bash scripts\/test-module-ci\.sh \$\{\{ matrix\.flavor \}\} complete/);
});

test('each root part has an exact separate artifact and strict download destination', () => {
  for (const name of ['test', 'race', 'coverage']) {
    const body = job(name);
    assert.match(body, /name: module-(?:test|race|coverage)-.*\$\{\{ matrix\.module-role \}\}/);
    assert.match(body, /if-no-files-found: error/);
  }
  const aggregate = job('module-completion');
  for (let part = 0; part < 8; part++) {
    const download = step(aggregate, `Download root part ${part}`);
    assert.match(download, new RegExp(
      `^          name: module-\\$\\{\\{ matrix\\.flavor \\}\\}-\\$\\{\\{ matrix\\.go-version \\}\\}-root-${part}$`, 'm',
    ));
    assert.match(download, new RegExp(
      `^          path: _out/module-\\$\\{\\{ matrix\\.flavor \\}\\}-owners/owner-0000/part-000${part}$`, 'm',
    ));
  }
  assert.match(step(aggregate, 'Download corpus owner'), /-corpus$/m);
  assert.match(step(aggregate, 'Download remaining-package owner'), /-other$/m);
  assert.doesNotMatch(aggregate, /merge-multiple:|pattern:/);
  assert.match(aggregate, /needs: \[test, race, coverage\]/);
  assert.match(aggregate, /^    if: \$\{\{ !cancelled\(\) \}\}$/m);
});

test('owner-two side gates run once and Windows remains a complete independent suite', () => {
  const body = job('test');
  for (const title of [
    'Required exact downloaded AMD64 numerical fixture',
    'Go test (cmd/plan9asm, Ubuntu)',
    'Go test (cmd/plan9asmll, Ubuntu)',
  ]) {
    assert.match(step(body, title), /matrix\.module-owner == 2/);
  }
  assert.match(body,
    /\{os: windows-latest, go-version: '1\.27\.x', module-role: other, module-owner: 2, root-partition: -1, nested_modules: true\}/);
  assert.match(step(body, 'Go test with coverage (Windows)'), /go test -timeout=30m \.\/\.\.\./);
  assert.match(body,
    /go-version: \['1\.20\.x', '1\.21\.x', '1\.22\.x', '1\.23\.x', '1\.24\.x', '1\.25\.x', '1\.26\.x', '1\.27\.x'\]/);
});

test('selected native CI remains a separate sixteen-owner serial protocol', () => {
  const body = job('cross-runtime');
  assert.match(body, /owner: \[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15\]/);
  assert.match(body, /-ci-owner \$\{\{ matrix.owner \}\} -shards 16 -timeout 20m/);
  assert.doesNotMatch(body, /module-root-partition|module-owner|(?:^|\s)-workers(?:=|\s)/);
});
