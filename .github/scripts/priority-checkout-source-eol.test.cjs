'use strict';

const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const test = require('node:test');

const attributes = fs.readFileSync('.gitattributes', 'utf8');
const source = 'package fixture\n\nconst assembly = `TEXT original(SB),4,$0-0\nRET\n`\n' +
  'const oracle = `int main(void) { return 0; }\n`\nconst escaped = "\\r\\n"\n';
const fixtures = {
  'original.go': source,
  'nested/original_test.go': source,
  'script.sh': '#!/bin/sh\nprintf "original\\n"\n',
  'ordinary.txt': 'ordinary text\nsecond line\n',
  'external.s': 'TEXT original(SB),4,$0-0\nRET\n',
};
const digest = bytes => crypto.createHash('sha256').update(bytes).digest('hex');

function checkoutFixture(rules) {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'plan9asm-checkout-eol-'));
  const environment = {...process.env};
  for (const key of Object.keys(environment)) {
    if (key.startsWith('GIT_')) delete environment[key];
  }
  Object.assign(environment, {
    GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: os.devNull, GIT_ATTR_NOSYSTEM: '1',
  });
  try {
    const git = args => {
      const result = spawnSync('git', ['-c', 'core.attributesFile=' + os.devNull,
        '-c', 'core.hooksPath=' + os.devNull, '-c', 'commit.gpgsign=false',
        '-c', 'user.name=Checkout fixture', '-c', 'user.email=fixture@example.invalid',
        '-C', directory, ...args], {
        env: environment, encoding: 'utf8', timeout: 5000,
      });
      assert.ifError(result.error);
      assert.equal(result.status, 0, result.stderr);
      return result.stdout;
    };
    git(['init', '--quiet']);
    git(['config', '--local', 'core.autocrlf', 'true']);
    git(['config', '--local', 'core.safecrlf', 'false']);
    fs.writeFileSync(path.join(directory, '.gitattributes'), rules);
    for (const [name, text] of Object.entries(fixtures)) {
      const target = path.join(directory, name);
      fs.mkdirSync(path.dirname(target), {recursive: true});
      fs.writeFileSync(target, text);
    }
    git(['add', '.']);
    git(['commit', '--quiet', '-m', 'original fixture']);
    for (const name of Object.keys(fixtures)) fs.unlinkSync(path.join(directory, name));
    // This actually restores bytes through Git's worktree conversion, not a
    // handcrafted CRLF replacement or a global user configuration change.
    git(['checkout-index', '--all', '--force']);
    return Object.fromEntries(Object.keys(fixtures).map(name =>
      [name, fs.readFileSync(path.join(directory, name))]));
  } finally {
    fs.rmSync(directory, {recursive: true, force: true});
  }
}

function requireOriginalGoBytes(observed) {
  for (const name of ['original.go', 'nested/original_test.go']) {
    assert.equal(digest(observed[name]), digest(Buffer.from(fixtures[name])),
      'checkout must preserve complete original Go driver/ASM/C bytes: ' + name);
    assert.equal(observed[name].includes(Buffer.from('\r\n')), false, name);
  }
  assert.equal(observed['script.sh'].toString(), fixtures['script.sh']);
  for (const name of ['ordinary.txt', 'external.s']) {
    assert.equal(observed[name].toString(), fixtures[name].replaceAll('\n', '\r\n'),
      'control must prove autocrlf conversion without normalizing external authority: ' + name);
  }
}

test('actual autocrlf checkout preserves complete Go source hashes and existing shell LF policy', () => {
  requireOriginalGoBytes(checkoutFixture(attributes));
});

test('Go checkout policy rejects missing substituted or CRLF source rules using real Git controls', () => {
  const complete = '*.sh text eol=lf\n*.go text eol=lf\n';
  requireOriginalGoBytes(checkoutFixture(complete));
  for (const changed of [
    complete.replace('*.go text eol=lf\n', ''),
    complete.replace('*.go text', '*.notgo text'),
    complete.replace('*.go text eol=lf', '*.go text eol=crlf'),
  ]) {
    const observed = checkoutFixture(changed);
    assert.equal(observed['ordinary.txt'].toString(), fixtures['ordinary.txt'].replaceAll('\n', '\r\n'));
    assert.throws(() => requireOriginalGoBytes(observed));
  }
});
