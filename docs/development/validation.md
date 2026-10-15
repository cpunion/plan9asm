# Validation and contribution completion

## Toolchain

Develop with current Go 1.27 and LLVM 22 exactly. The official compatibility
matrix retains Go 1.20–1.27. Put the selected Go binary on PATH: subprocesses
deliberately use `GOTOOLCHAIN=local`.

```sh
go127_root=$(GOTOOLCHAIN=go1.27.2 go env GOROOT)
export PATH="$go127_root/bin:$PATH"
export GOTOOLCHAIN=local
source scripts/llvm22.sh
llc_cmd=$(find_llvm22_llc)
"$llc_cmd" --version
```

The helper accepts `llc-22`, an unversioned `llc` reporting version 22, or
`LLVM_CONFIG` pointing to LLVM 22. On Homebrew hosts:

```sh
llvm22_root=$(brew --prefix llvm@22)
export PATH="$llvm22_root/bin:$PATH"
export LLVM_CONFIG="$llvm22_root/bin/llvm-config"
```

Missing required LLVM tools, linkers or backends fail, never skip.
Host-inapplicable execution requires its cross-runtime CI counterpart.

The rich CLI's complete ARM64 fallback-frame oracle runs natively on supported
ARM64 hosts and under pinned QEMU in the required Linux cross job. One outer
owner runs its exact nested-module root; a fresh strict reader verifies both
logical partitions and all four required children. Cross scopes bind the actual
Node auditor through the shared tool observer. Object-only Windows skips do not
establish execution of its 512 input rows.

Darwin native C oracles query the actual selected macOS SDK with bounded, owned
`xcrun --sdk macosx --show-sdk-path`, require one complete absolute existing
directory, and pass `--no-default-config -isysroot` to the independently checked
LLVM 22 clang. `DEVELOPER_DIR` selects the actual developer tools; `SDKROOT` and
CGO flags alone cannot override a hard-coded clang default configuration. Do not
edit global clang configurations, guess an SDK version or substitute another
LLVM. Missing SDK selection fails. The shared `internal/nativeoracle` helper is
test-driver configuration, not source/ABI authority; exact downloaded-source
oracles retain their additional live SDK/tool/file identity observations.

## Local and CI gates

Run focused RED/GREEN tests first; preserve failures under ignored `_out/`.
Then run the relevant complete gates:

```sh
go test ./... -count=1 -timeout=45m
(cd cmd/plan9asm && go test ./... -count=1 -timeout=30m)
(cd cmd/plan9asmll && go test ./... -count=1 -timeout=30m)
go vet ./...
go build ./...
test -z "$(git ls-files -z -- '*.go' | xargs -0 gofmt -l)"
git diff --check
node --test .github/scripts/priority-*.test.cjs

scripts/check-go-asm-coverage.sh
scripts/check-arm64-plan9-corpus.sh
scripts/check-stdlib-corpus.sh
scripts/benchmark-compile.sh
```

The [exhaustive runner](../../cmd/plan9asmtest/README.md) can partition expensive
first-party tests without sampling. Its actual inventory, disjoint ownership,
raw outcomes and strict complete-module aggregate remain required. Owner 0 owns
the root package, owner 1 `gotoolprofile`, owner 2 all remaining packages.
Ordinary suites may use `-workers 2`; one collector owns report/coverage writes.
Root scheduling parts do not change the test denominator. Windows keeps all
three modules independent: sixteen root parts for the main module over the same
32 hash shards (two shards per part), four parts per CLI, and one profile/rest
owner each. Every command retains the thirty-minute budget;
only complete cross-owner verification publishes the module's coverage. The
Windows driver binds inherited and explicit `-tags=byollvm` without fallback.
It selects Go, LLVM22, Node, Python and jq before freezing a minimal tool PATH
for both execution and aggregation. Ambient hosted Git runtime-bank and WSL
entries are not inherited. Complete environment equality remains strict; never
remove PATH from the verifier to accept unrelated execution environments.
The readonly DATA consumer matrix has three complete target roots, placed in
different Windows scheduling owners. Each retains all119 operations, all three
APIs and O0/O2 objects; root splitting is not a reduced test denominator.
The two current-Go bytealg CLI object owners likewise keep complete Count and
IndexByte files in separate scheduling owners, each with its own original
five-minute context. A previous root shared that context across both files;
the second file's cancellation was a real failure, not source rejection.

Repeated ARM64 byte-carrier matrices share only a parent-test-owned `GOCACHE`.
Recent Go distributions build builtin observers such as `objdump` there;
recreating an empty cache for each row repeats that build and can exhaust the
unchanged one-minute command limit. Each case still captures independent
original sources, outputs, leases and revocable owners. Do not share those
mutable observations or increase command budgets to hide cache setup costs.

```sh
go run ./cmd/plan9asmtest -module -out _out/module-suite \
  -shards 32 -workers 2 -timeout 45m
```

Run race tests for changed concurrent code and actual Go 1.20 checks for changed
root-module code. Current-SDK producer fixtures use explicit `go1.27` build
roles; historical lanes still test their portable framework. Do not add runtime
skips or `!llgo` tags to hide incomplete compiler support.
Run the formatting check with the selected current Go toolchain. Limit it to
tracked sources: ignored `_out/` holds deliberately unchanged historical failures
and generated observation inputs, not repository formatting targets.

## Native and cross-runtime semantics

```sh
go test . -run '^(TestAMD64ConformanceNativeGo|TestAMD64ConformanceLLVMRuntime)$' -count=1 -v
PLAN9ASM_CROSS_EXEC=1 go test . \
  -run '^(TestCrossLinuxRuntimeMatrix|TestARM64Conformance|TestARMIntegerMemoryRuntimeCross)' \
  -count=1 -v
```

The cross driver runs on Linux/amd64 with matching cross compilers. Install
checksum-pinned QEMU 10.2.3 using `bash scripts/install-ci-qemu.sh`, then prepend
`$PWD/_out/qemu-10.2.3-linux-amd64` to PATH. Ubuntu 24.04's QEMU 8.2 has an SVE
indexed dot-product bug. For i386 also install `gcc-i686-linux-gnu` and
`libc6-dev-i386-cross`. Missing tools must fail. Containers need executable
temporary storage for Go test binaries.

## Official forms and benchmark

An ordinary Go assembler rejection requires actual exit 1 plus a concrete,
fixture-bound diagnostic. Panic, timeout, signal, unknown output or cleanup
failure is not rejection evidence. `TestGoAssemblerKnownMalformedInputDefects`
separately observes known Go assembler bugs without treating them as successful
form coverage or external-library applicability evidence.

The official gate checks classifications, parse failures and versioned form
fingerprints. Inspect differences before changing
`testdata/coverage/go-asm-baseline.json`. Opcode names, observed corpus forms,
encoder rules, LLVM objects and runtime semantics are distinct coverage claims.

The strict benchmark compiles standard-library assembly for
`linux/{386,amd64,arm,arm64}` and `js/wasm`. Failures, skips and N/A fail the
job. Schema-2 summaries retain failed and unattempted counts; missing reports
remain unknown. Reports under `_out/compile-benchmark/` are never overwritten.
Default budgets are 300 seconds per target and 900 seconds total. Diagnostic
overrides use `PLAN9ASM_BENCHMARK_TARGETS`,
`PLAN9ASM_BENCHMARK_MAX_TARGET_SECONDS`,
`PLAN9ASM_BENCHMARK_MAX_TOTAL_SECONDS` and `PLAN9ASM_BENCHMARK_REPORT_DIR`.
Measure before claiming a speedup or changing budgets.

## Frozen verification and PRs

Commit first and keep a tested snapshot unchanged, including documentation.
Develop in another existing idle worktree while a frozen suite runs. A partial
suite or older-source pass cannot certify current-head success.

Inspect remotes and upstream divergence before publishing. Push only to the
allowed fork, create/update the contribution PR, and never directly push,
merge, approve or close upstream PRs. Batch updates so active frozen runs can
finish. Keep PRs draft until current-head CI, review and coverage pass.

External scanning, ledgers, exception policies and third-party compatibility
CI are now owned by [llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
Do not restore them here or gate instruction tests on their historical pending
inventory. Independent instruction regressions and required first-party source
proof tests remain in this repository; moving a workload is not a test pass.
