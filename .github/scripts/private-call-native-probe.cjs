'use strict';

// Fork-only diagnostic. The actual numerical driver and its tests are unchanged.
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const {auditNativeOwners} = require('./audit-native-owners.cjs');
const plan = require('./private-call-native-plan.json');

const sha = data => crypto.createHash('sha256').update(data).digest('hex');
const readJSON = file => JSON.parse(fs.readFileSync(file, 'utf8'));
const writeJSON = (file, data) => fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n');
const execute = (file, args) => cp.execFileSync(file, args, {
  encoding: 'utf8', timeout: 60000, maxBuffer: 8 * 1024 * 1024,
}).trim();
const selector = '^(' + plan.roots.join('|') + ')$';
const commands = {
  inventory: ['test', '-p=1', '.', '-list', selector],
  native: ['test', '-p=1', '.', '-run', selector, '-count=1', '-timeout=15m', '-json'],
};

function sourceState(expectedHead) {
  assert.match(expectedHead, /^[a-f0-9]{40}$/);
  assert.equal(execute('git', ['rev-parse', 'HEAD']), expectedHead);
  assert.equal(execute('git', ['status', '--porcelain']), '', 'source tree changed');
  const sources = {};
  for (const [file, expected] of Object.entries(plan.source_sha256)) {
    sources[file] = sha(fs.readFileSync(file));
    assert.equal(sources[file], expected, 'planned source changed: ' + file);
  }
  return {head: expectedHead, tree: execute('git', ['rev-parse', 'HEAD^{tree}']), sources};
}

function toolState() {
  assert.equal(process.version, 'v22.23.4');
  assert.equal(process.platform, 'linux');
  assert.equal(process.arch, 'x64');
  const environment = JSON.parse(execute('go', ['env', '-json',
    'GOVERSION', 'GOOS', 'GOARCH', 'GOROOT', 'GOTOOLDIR', 'GOFLAGS', 'GOTOOLCHAIN']));
  assert.equal(environment.GOVERSION, 'go1.27.2');
  assert.equal(environment.GOOS, 'linux');
  assert.equal(environment.GOARCH, 'amd64');
  assert.equal(environment.GOFLAGS, '');
  assert.equal(environment.GOTOOLCHAIN, 'local');
  const tools = {};
  for (const name of ['go', 'node', 'llc-22', 'clang-22', 'opt-22', 'llvm-config-22']) {
    const file = fs.realpathSync(execute('which', [name]));
    tools[name] = {path: file, sha256: sha(fs.readFileSync(file))};
  }
  for (const name of ['asm', 'compile', 'link']) {
    const file = fs.realpathSync(path.join(environment.GOTOOLDIR, name));
    tools[name] = {path: file, sha256: sha(fs.readFileSync(file))};
  }
  const llvm = execute('llvm-config-22', ['--version']);
  const llc = execute('llc-22', ['--version']);
  const clang = execute('clang-22', ['--version']);
  const opt = execute('opt-22', ['--version']);
  assert.match(llvm, /^22\./);
  assert.match(llc, /LLVM version 22\./);
  assert.match(clang, /clang version 22\./);
  assert.match(opt, /LLVM version 22\./);
  return {node: process.version, environment, llvm, llc, clang, opt, tools};
}

function runCommand(directory, name, args) {
  writeJSON(path.join(directory, name + '-command.json'), {file: 'go', args});
  const stdout = fs.openSync(path.join(directory, name === 'inventory' ? 'inventory.log' : 'raw.jsonl'), 'w');
  const stderr = fs.openSync(path.join(directory, name + '.stderr'), 'w');
  let result;
  try {
    result = cp.spawnSync('go', args, {stdio: ['ignore', stdout, stderr], timeout: 18 * 60 * 1000});
  } finally {
    fs.closeSync(stdout);
    fs.closeSync(stderr);
  }
  writeJSON(path.join(directory, name + '-exit.json'), {
    status: result.status, signal: result.signal, error: result.error?.message ?? null,
  });
  assert.equal(result.status, 0, name + ' command failed');
}

function auditReceipt(directory, expectedHead) {
  const before = readJSON(path.join(directory, 'before.json'));
  const after = readJSON(path.join(directory, 'after.json'));
  assert.match(expectedHead, /^[a-f0-9]{40}$/);
  assert.equal(before.source.head, expectedHead);
  assert.deepEqual(after, before, 'source or actual tools changed during execution');
  assert.deepEqual(before.source.sources, plan.source_sha256);
  assert.equal(before.tools.node, 'v22.23.4');
  assert.equal(before.tools.environment.GOVERSION, 'go1.27.2');
  assert.equal(before.tools.environment.GOOS, 'linux');
  assert.equal(before.tools.environment.GOARCH, 'amd64');
  assert.equal(before.tools.environment.GOFLAGS, '');
  assert.equal(before.tools.environment.GOTOOLCHAIN, 'local');
  assert.match(before.tools.llvm, /^22\./);
  assert.match(before.tools.llc, /LLVM version 22\./);
  assert.match(before.tools.clang, /clang version 22\./);
  assert.match(before.tools.opt, /LLVM version 22\./);
  for (const name of ['go', 'node', 'asm', 'compile', 'link', 'llc-22', 'clang-22', 'opt-22', 'llvm-config-22']) {
    assert.ok(path.isAbsolute(before.tools.tools[name].path));
    assert.match(before.tools.tools[name].sha256, /^[a-f0-9]{64}$/);
  }
  for (const name of ['inventory', 'native']) {
    assert.deepEqual(readJSON(path.join(directory, name + '-command.json')),
      {file: 'go', args: commands[name]}, 'actual command differs from the complete selector');
    assert.deepEqual(readJSON(path.join(directory, name + '-exit.json')),
      {status: 0, signal: null, error: null}, 'nonzero or incomplete actual command');
    fs.readFileSync(path.join(directory, name + '.stderr'));
  }
  const inventory = fs.readFileSync(path.join(directory, 'inventory.log'), 'utf8');
  const raw = fs.readFileSync(path.join(directory, 'raw.jsonl'), 'utf8');
  const result = auditNativeOwners(inventory, raw, plan.roots, plan.children);
  const output = raw.split(/\r?\n/).filter(Boolean).map(line => JSON.parse(line))
    .filter(event => event.Test === plan.native_root && event.Action === 'output')
    .map(event => event.Output).join('');
  const prefix = 'actual CPU/OS-state and original-opcode preflight: ';
  assert.equal(output.split(prefix).length - 1, 1, 'missing or duplicate actual native preflight');
  assert.match(output, /actual CPU\/OS-state and original-opcode preflight: axis=1 SSE2=1 /);
  return {protocol: 'private_call_native_probe_v1', source: before.source, tools: before.tools,
    ...result, nativeKernels: plan.kernels, inputsPerKernel: plan.inputs_per_kernel,
    inventorySHA256: sha(inventory), rawSHA256: sha(raw)};
}

module.exports = {auditReceipt};
if (require.main === module) {
  const [mode, directory] = process.argv.slice(2);
  assert.equal(process.argv.length, 4);
  assert.ok(['run', 'audit'].includes(mode));
  const expectedHead = process.env.GITHUB_SHA;
  if (mode === 'run') {
    fs.mkdirSync(directory, {recursive: true});
    const before = {source: sourceState(expectedHead), tools: toolState()};
    writeJSON(path.join(directory, 'before.json'), before);
    fs.copyFileSync('/proc/cpuinfo', path.join(directory, 'cpuinfo.txt'));
    try {
      runCommand(directory, 'inventory', commands.inventory);
      const discovered = fs.readFileSync(path.join(directory, 'inventory.log'), 'utf8')
        .split(/\r?\n/).filter(line => line.startsWith('Test')).sort();
      assert.deepEqual(discovered, plan.roots.slice().sort());
      runCommand(directory, 'native', commands.native);
    } finally {
      writeJSON(path.join(directory, 'after.json'), {source: sourceState(expectedHead), tools: toolState()});
    }
  } else {
    sourceState(expectedHead);
    const result = auditReceipt(directory, expectedHead);
    writeJSON(path.join(directory, 'audit.json'), result);
    process.stdout.write(JSON.stringify(result, null, 2) + '\n');
  }
}
