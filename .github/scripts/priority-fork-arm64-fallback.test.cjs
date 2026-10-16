const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

const workflow = fs.readFileSync('.github/workflows/pr40-owned-process-probe.yml', 'utf8');

function job(source, name) {
  const marker = `\n  ${name}:\n`;
  assert.equal(source.split(marker).length - 1, 1, `one exact ${name} owner required`);
  const rest = source.slice(source.indexOf(marker) + marker.length);
  const end = rest.search(/^  [a-zA-Z0-9_-]+:\s*$/m);
  return end < 0 ? rest : rest.slice(0, end);
}

function requirePriority(source) {
  const body = job(source, 'arm64-fallback-cross-priority');
  assert.match(body, /^    runs-on: ubuntu-latest$/m);
  assert.match(body, /^    timeout-minutes: 20$/m);
  assert.doesNotMatch(body, /^    (?:if|needs):|strategy:|matrix:|continue-on-error|\|\| true| -skip| -run|go test/m);
  for (const line of ["GOFLAGS: ''", "GOENV: 'off'", "GOWORK: 'off'", 'GOTOOLCHAIN: local', "PLAN9ASM_CROSS_EXEC: '1'"]) {
    assert.ok(body.includes(`      ${line}\n`), line);
  }
  assert.match(body, /timeout-minutes: 14\n        run: >-\n          bash scripts\/install-ci-llvm22-ubuntu\.sh\n          gcc-i686-linux-gnu gcc-arm-linux-gnueabihf\n          gcc-aarch64-linux-gnu/);
  assert.match(body, /timeout-minutes: 5\n        run: bash scripts\/install-ci-qemu\.sh/);
  assert.match(body, /uses: \.\/\.github\/actions\/setup-go\n        with:\n          go-version: '1\.27\.2'/);
  assert.match(body, /uses: actions\/setup-node@v6\n        with:\n          node-version: '22'/);
  assert.match(body, /run: \|\n          set -euo pipefail\n          mkdir -p _out\/cross-runtime-owner\n          bash scripts\/test-arm64-fallback-ci\.sh execute/);
  assert.equal((body.match(/bash scripts\/test-arm64-fallback-ci\.sh execute/g) || []).length, 1);
  assert.match(body, /if: \$\{\{ !cancelled\(\) \}\}\n        uses: actions\/upload-artifact@v6\n        with:\n          name: arm64-fallback-cross-priority\n          path: _out\/cross-runtime-owner\/arm64-fallback-cross\/\n          if-no-files-found: error/);
  assert.match(job(source, 'complete-windows'), /^    needs: \[ownership, native-x86-oracles, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, unlocked-rmw-native-priority, vector-value-fault-priority\]$/m);
}

test('complete ARM64 CLI oracle is a required bounded Linux cross priority', () => {
  requirePriority(workflow);
});

test('ARM64 fork priority rejects missing ownership, tools, execution, artifacts or success barrier', () => {
  const marker = '\n  arm64-fallback-cross-priority:\n';
  requirePriority(workflow);
  const body = job(workflow, 'arm64-fallback-cross-priority');
  for (const [before, after] of [
    ['arm64-fallback-cross-priority:', 'removed-arm64-priority:'],
    ['runs-on: ubuntu-latest', 'runs-on: windows-latest'],
    ['timeout-minutes: 20', 'timeout-minutes: 25'],
    ["PLAN9ASM_CROSS_EXEC: '1'", "PLAN9ASM_CROSS_EXEC: '0'"],
    ['gcc-aarch64-linux-gnu', 'removed-aarch64-linker'],
    ['gcc-i686-linux-gnu gcc-arm-linux-gnueabihf', 'gcc-i686-linux-gnu'],
    ['install-ci-qemu.sh', 'install-system-qemu.sh'],
    ["go-version: '1.27.2'", "go-version: '1.26.2'"],
    ["node-version: '22'", "node-version: '20'"],
    ['mkdir -p _out/cross-runtime-owner', 'mkdir -p _out/foreign'],
    ['test-arm64-fallback-ci.sh execute', 'test-arm64-fallback-ci.sh complete'],
    ['if: ${{ !cancelled() }}', 'if: ${{ success() }}'],
    ['path: _out/cross-runtime-owner/arm64-fallback-cross/', 'path: _out/foreign/'],
    ['if-no-files-found: error', 'if-no-files-found: ignore'],
  ]) {
    const changed = before === 'arm64-fallback-cross-priority:' ? workflow.replace(before, after) :
      workflow.replace(body, body.replace(before, after));
    assert.notEqual(changed, workflow, `mutation must change its own job: ${before}`);
    assert.throws(() => requirePriority(changed), undefined, `must reject mutation: ${before}`);
  }
  assert.throws(() => requirePriority(workflow.replace(
    'needs: [ownership, native-x86-oracles, sdk-priority, arm64-fallback-cross-priority, lock-native-priority, unlocked-rmw-native-priority, vector-value-fault-priority]', 'needs: [ownership, sdk-priority]')));
  assert.throws(() => requirePriority(workflow.replace(marker, marker + '    continue-on-error: true\n')));
  assert.throws(() => requirePriority(workflow.replace(marker, marker + '    if: ${{ always() }}\n')));
  assert.throws(() => requirePriority(workflow + marker));
});

test('fork priority consumes the unchanged complete strict main driver and its four mandatory children', () => {
  const driver = fs.readFileSync('scripts/test-arm64-fallback-ci.sh', 'utf8');
  assert.match(driver, /root=TestFallbackARM64WideFrameCompilesWithLLVM22\n/);
  assert.match(driver, /-dir cmd\/plan9asmll/);
  assert.match(driver, /-ci-selector "\^\$\{root\}\$" -ci-isolate-root "\$root"/);
  assert.match(driver, /-shards 2 -timeout 5m/);
  assert.match(driver, /for owner in 0 1; do/);
  assert.match(driver, /-verify-ci-partitions "\$base\/owners"/);
  assert.match(driver, /node \.github\/scripts\/audit-arm64-fallback\.cjs/);
  const auditor = fs.readFileSync('.github/scripts/audit-arm64-fallback.cjs', 'utf8');
  assert.match(auditor, /\['darwin', 'linux', 'windows', 'native_complete-argument_oracle'\]/);
  assert.match(auditor, /roots: 1, children: 4, passed: 5, skipped: 0, incomplete: 0/);
  const qemu = fs.readFileSync('scripts/install-ci-qemu.sh', 'utf8');
  assert.match(qemu, /^version=10\.2\.3$/m);
  assert.match(qemu, /sha256sum --check --status/);
});
