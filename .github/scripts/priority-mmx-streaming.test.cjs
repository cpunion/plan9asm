'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');
const {auditMMXStreaming, roots, children} = require('./audit-mmx-streaming.cjs');
const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');
const block = `      - name: Execute original Go and LLVM native MMX streaming stores
        env:
          GOTOOLCHAIN: local
        run: |
          set -euo pipefail
          directory=_out/native-x86-oracles/mmx-streaming
          mkdir -p "$directory"
          root='${roots[0]}'
          go test . -list "^($root)$" > "$directory/inventory.log"
          grep -Fx "$root" "$directory/inventory.log"
          go test . -run "^($root)$" -count=1 -timeout=5m -json > "$directory/raw.jsonl" 2> "$directory/stderr.log"
          node .github/scripts/audit-mmx-streaming.cjs "$directory"
`;

test('MMX streaming executes original Go and every native LLVM backend axis before modules', () => {
  assert.ok(workflow.includes(block));
  assert.equal(workflow.split(block).length, 2);
  const job = workflow.slice(workflow.indexOf('  native-x86-oracles:'), workflow.indexOf('\n  ownership:'));
  assert.ok(job.includes(block));
  assert.match(job, /runs-on: ubuntu-latest/);
});

test('MMX streaming audit rejects omitted, skipped or substituted original/backend execution', () => {
  const event = (Action, Test) => ({Action, Package: 'github.com/xgo-dev/plan9asm', ...(Test ? {Test} : {})});
  const events = [event('start'), event('run', roots[0])];
  for (const child of children) events.push(event('run', child), event('pass', child));
  events.push(event('pass', roots[0]), event('pass'));
  const encode = values => values.map(JSON.stringify).join('\n') + '\n';
  assert.deepEqual(auditMMXStreaming(roots[0] + '\n', encode(events)),
    {roots:1, children:5, passed:6, skipped:0, incomplete:0});
  for (const child of children) {
    for (const changed of [
      events.filter(value => value.Test !== child),
      events.map(value => value.Test === child && value.Action === 'pass' ? {...value, Action:'skip'} : value),
      events.map(value => value.Test === child ? {...value, Test:child+'-object-only'} : value),
    ]) assert.throws(() => auditMMXStreaming(roots[0]+'\n', encode(changed)));
  }
  assert.throws(() => auditMMXStreaming('TestOtherOwner\n', encode(events)));
  assert.throws(() => auditMMXStreaming(roots[0]+'\n', encode([
    event('start'), {...event('output'), Output:'PASS '+children.join('\nPASS ')}, event('pass'),
  ])));
});
