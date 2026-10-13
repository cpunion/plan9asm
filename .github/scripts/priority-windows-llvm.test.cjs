const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { test } = require('node:test');

const action = fs.readFileSync('.github/actions/setup-windows-llvm22/action.yml', 'utf8');

test('MSYS2 bootstrap retries are bounded and the final failure remains fatal', () => {
  const bootstrap = action.slice(action.indexOf('  steps:'), action.indexOf('    - name: Install LLVM 22'));
  const attempts = bootstrap.split(/(?=    - name:)/).filter(step => step.includes('uses: msys2/setup-msys2@v2'));
  assert.equal(attempts.length, 3);
  for (const [index, step] of attempts.entries()) {
    assert.match(step, new RegExp(`id: msys2_attempt_${index + 1}`));
    assert.match(step, /msystem: CLANG64/);
    assert.match(step, /path-type: inherit/);
    assert.match(step, /update: true/);
    if (index < 2) {
      assert.match(step, /continue-on-error: true/);
    } else {
      assert.doesNotMatch(step, /continue-on-error:/);
    }
    if (index > 0) {
      for (let prior = 1; prior <= index; prior++) {
        assert.match(step, new RegExp(`steps\\.msys2_attempt_${prior}\\.outcome == 'failure'`));
      }
    }
  }
});

function install(outcomes, version = '22.1.8') {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'plan9asm-llvm-install-'));
  try {
    const installStep = action.slice(action.indexOf('    - name: Install LLVM 22'));
    const run = installStep.slice(installStep.indexOf('      run: |\n') + '      run: |\n'.length)
      .split('\n').map(line => line.startsWith('        ') ? line.slice(8) : line).join('\n');
    const script = `
mock_pacman_calls=0
outcomes=(${outcomes.join(' ')})
pacman() {
  mock_pacman_calls=$((mock_pacman_calls + 1))
  printf 'PACMAN:'
  printf ' %s' "$@"
  printf '\\n'
  return "\${outcomes[mock_pacman_calls - 1]}"
}
sleep() { test "$1" = 5; printf 'RETRY_WAIT\\n'; }
llvm-config() {
  if [[ "$1" == --version ]]; then printf '%s\\n' "$MOCK_LLVM_VERSION";
  else printf 'checked-flags\\n'; fi
}
cygpath() { printf 'checked-clang64-bin\\n'; }
${run}`;
    const result = spawnSync('bash', ['-c', script], {
      encoding: 'utf8', timeout: 5000,
      env: {
        ...process.env,
        LLVM_VERSION: '22.1.8', LLVM_PACKAGE_VERSION: '22.1.8-2',
        LLVM_RUNTIME_PACKAGE_VERSION: '22.1.8-1', MOCK_LLVM_VERSION: version,
        GITHUB_ENV: path.join(directory, 'environment'), GITHUB_PATH: path.join(directory, 'path'),
      },
    });
    assert.ifError(result.error);
    return {
      ...result,
      environment: fs.existsSync(path.join(directory, 'environment'))
        ? fs.readFileSync(path.join(directory, 'environment'), 'utf8') : '',
    };
  } finally {
    fs.rmSync(directory, { recursive: true, force: true });
  }
}

test('actual installer script retries the identical pinned transaction only', () => {
  for (const outcomes of [[0], [67, 0], [67, 68, 0]]) {
    const result = install(outcomes);
    assert.equal(result.status, 0, result.stderr);
    const transactions = result.stdout.split('\n').filter(line => line.startsWith('PACMAN:'));
    assert.equal(transactions.length, outcomes.length);
    assert.equal(new Set(transactions).size, 1);
    assert.match(transactions[0], /llvm-22\.1\.8-2-any\.pkg\.tar\.zst/);
    assert.doesNotMatch(transactions[0], /llvm-23/);
    assert.equal((result.stdout.match(/RETRY_WAIT/g) || []).length, outcomes.length - 1);
    assert.match(result.environment, /CGO_ENABLED=1/);
  }
});

test('exhausted installation and wrong LLVM version publish no build environment', () => {
  const exhausted = install([67, 68, 69]);
  assert.equal(exhausted.status, 69, exhausted.stderr);
  assert.equal((exhausted.stdout.match(/PACMAN:/g) || []).length, 3);
  assert.equal(exhausted.environment, '');
  const wrongVersion = install([0], '23.0.0');
  assert.equal(wrongVersion.status, 1);
  assert.match(wrongVersion.stderr, /expected LLVM 22\.1\.8, got 23\.0\.0/);
  assert.equal(wrongVersion.environment, '');
});
