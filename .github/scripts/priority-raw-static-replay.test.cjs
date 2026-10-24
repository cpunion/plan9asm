const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const roots = [
  'TestX86RawStaticReplayCapturesDecoderBeforeSymbolRebinding',
  'TestX86RawStaticReplayScalarResultsThreeAPIs',
  'TestX86RawStaticReplayPortableValuesAndMutablePool',
  'TestX86RawReturnMaterializedReadCannotAuthorizeChangedSource',
  'TestX86RawStaticRegisterReturnDoesNotInventSourceBytes',
];
const block = '- lane: raw-static-replay\n' +
  "            module: '.'\n            roots:\n" +
  roots.map(root => `              - ${root}\n`).join('') +
  '            timeout: 5m';

function check(text) {
  assert.equal(roots.length, 5);
  assert.equal(new Set(roots).size, 5);
  assert.ok(text.includes(block), 'complete raw decoder, source, values and rejection owners');
  assert.equal(text.split('- lane: raw-static-replay\n').length - 1, 1);
  for (const root of roots) {
    assert.equal(text.split(`- ${root}\n`).length - 1, 1, root);
  }
}

test('raw static replay retains five complete owners under its original five-minute budget', () => {
  check(workflow);
});

test('raw replay cannot omit, substitute, duplicate, crop or enlarge an owner', () => {
  check(block);
  for (const changed of [
    '',
    block + '\n' + block,
    block.replace('raw-static-replay', 'unrelated-lane'),
    block.replace("module: '.'", 'module: cmd/plan9asmll'),
    block.replace('timeout: 5m', 'timeout: 6m'),
    block.replace('            timeout: 5m',
      '              - TestUnrelatedAdditionalOwner\n            timeout: 5m'),
    ...roots.flatMap(root => [
      block.replace(`              - ${root}\n`, ''),
      block + `\n              - ${root}\n`,
      block.replace(root, 'TestUnrelatedReplacement'),
      block.replace(root, root + '/scalar32'),
    ]),
  ]) {
    assert.throws(() => check(changed));
  }
});
