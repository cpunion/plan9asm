# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). Repair CI before starting
another module-index inventory scan.

## Contribution and worktrees

- PR 40 remains OPEN/DRAFT. Latest live inspection: remote head `4cf5ade`,
  base `7cc8c0f`, 74 successful checks and 20 failures (19 discovery shards
  plus their aggregate). Local repairs have not been pushed.
- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`. Never push to
  `origin` or `xgo-dev`. Batch the repairs before pushing; keep the PR draft.
- Develop on `codex/pr40-arm64-raw-20260926`. Upstream main `7cc8c0f` is
  already an ancestor. Inspect status, worktrees and processes first.
- `codex/pr40-shard25-replay-20260926` is frozen at `8134b3e`. Shard 29
  is still running there. Do not change that checkout or ledger.
- A separate persistent verification worktree is detached at `f151495`.
  Its root `go test ./... -count=1 -timeout=20m` passed in 765 seconds. Official
  five-architecture classification and strict benchmark passed there:
  184/184 files, no N/A, 28 target-seconds plus four seconds for driver build.
  This predates the network-retry and guarded-index commits; it is not
  current-head full-suite evidence.
- Use Go 1.27.1 and LLVM 22 for external modules. Put the actual Go binary in
  PATH; setting only GOTOOLCHAIN can select a different child compiler.
  Required cross execution uses checksum-pinned QEMU 10.2.3.

## Latest committed repairs

- `cc307d5`: reject Go command/compiler/assembler version mismatches as
  infrastructure failures, never source N/A. The shard script pins its child
  Go binary to the recorded GOROOT and checks all three versions.
- `ba157c3`: pin discovery and aggregate CI jobs to Go 1.27.1.
- `f151495`: track every bounded ARM64 constant-pool pointer alias separately.
  MOV and 64-bit ADD/SUB immediate, shifted and extended-register forms
  propagate inclusive offset ranges. Killing an original does not kill its
  copies. Loads must fit at both range endpoints. Escapes, truncation,
  address-dependent flags, unknown indexes and changing-offset joins fail.
- `0f08122`: retry transient module-download failures as well as dependency
  build failures. At most three attempts share the original deadline. A
  synthetic local proxy establishes red/green recovery, persistent failure,
  no retry on 404, cancellation and workspace cleanup. Checksum, compiler,
  resource and toolchain failures are not retried.
- `596b147`: follow actual CFG edges when proving integer index bounds.
  An adjacent CMP must dominate B.cond; CBZ/CBNZ zero edges are also modeled.
  A W comparison does not bound an X register. Bypassed guards, changed flags,
  signed-negative possibilities and converging unconstrained edges fail.

The pointer/guard batches passed all focused pool tests on Go 1.27.1,
focused Go 1.20 compatibility, LLVM 22 ARM64 objects for Darwin/Linux/Windows,
Darwin native runtime and the required Linux/QEMU runtime counterparts.
The existing cross container uses Go 1.27.0 for these root runtime tests.
Both nested CLI test suites and vet passed again after the guarded-index
commit. Corpus unit tests and focused race tests passed for download retry.
Logs in the development worktree use `raw-pool-alias*`,
`raw-pool-guard*` and `discovery-download-retry*`.

## External evidence and remaining real failures

A passing fixture or one target replay is not a passing module or shard.
The latest direct SIMD replay (`_out/simd-guard.json`) still fails both
`parseIntsNEON` and `parseIntsSVE2`; do not upgrade its assembly ledger.

- **SIMD**: `github.com/sebishogun/simd@v1.21.1` and its mirror have 45/47
  applicable ARM64 files passing in earlier three-OS diagnostics. The two
  remaining bytes files need relational index bounds and loop reasoning,
  including derived aliases and post-indexed pair loads. The apparent branch
  word `0x540be400` is numeric pool data, not invalid source. Implement in
  `arm64_raw_pool_*.go`; do not relax the proof merely to relocate the pool.
- **KnoxDB**: `blockwatch.cc/knoxdb`, `github.com/blockwatch-cc/knoxdb` and
  `github.com/os2357/knx` all still resolve latest to v0.2.9. Four Uint8/16/32/64
  AVX2 files require cross-TEXT ABI0 register and frame preservation. The
  root's indirect dispatch enters helper TEXTs sharing DI/SI/BX/R14/R15 and
  the root's result at +48(FP). Widening the exit helper's FP slot alone is
  incorrect. These helpers actually have Go declarations of the form
  `func helper()`, so declaration-free tail-signature inference does not
  apply. Preserve the distinction between their Go entry ABI and assembly
  continuation state. Ordinary AVX2 and Uint64 AVX512 files already pass.
- **Native-layout/JIT**: GopherJRE, GoJIT and both case-distinct Sharkie module
  paths still fail. Their source observes exact code offsets or transfers
  registers/stack through generated machine code. Widening raw branch
  boundaries alone is insufficient. There is no authorized generic
  native-layout skip. A newer GoJIT version was downloaded diagnostically but
  has not been imported or verified as a replacement.
- Exact invalid-source skips (including both go-highway versions), historical
  gVisor/Skywire superseded skips and the pinned fiber/ai private AMX exception
  remain distinct from passes. Their manifests and executable proof tests are
  authoritative. Do not extend their scope without evidence/authorization.

The frozen `8134b3e` diagnostics are stored under
`_out/ci-repair-8134-shardN/shard-N.json`. Completed shards 1, 13, 20, 22,
23, 24 and 25 have zero failures and individually audited reports. Shards
0, 6, 9, 14, 15, 16, 17, 26 and 28 retain the real failure classes above
(shard 16 also predates the dev9 invalid-source fix). Shard 4 completed
168 candidates = 123 passed + 35 source N/A + one invalid-source skip +
nine download TLS-timeout failures; its integrity audit passed, not its
coverage gate. Shard 10 completed 173 candidates with 133 passed, 29 N/A and
11 proxy TLS-timeout failures; its large spanneranalyzer p0/p6 candidates
passed. Shard 29's large reflectx candidate passed 48 object translations.

A separate clean `cc307d5` shard 16 report has 132 passed + 30 N/A + one
go-highway invalid-source skip + one KnoxDB failure. It is not compatible
with the older frozen reports. Keep all of these as diagnostic evidence.

## Next actions and completion gates

1. Finish/audit active tests and shards. Retry network failures with current
   bounded retry support; never reclassify them as source N/A.
2. Continue real semantic fixes above with red/green and runtime tests.
   The address-proof work does not yet resolve the two complete SIMD files.
3. Run full root, both nested CLIs, vet/build, official five-architecture
   coverage, ARM64/stdlib corpus, cross runtime and strict benchmark after
   the final implementation. Preserve evidence by exact source snapshot.
4. Freeze source/tools/ledger, run all 32 shards on identical provenance,
   then use the validated updater to refresh the assembly ledger. Even doc
   commits change provenance; do not combine old reports into the final run.
5. Batch-push the allowed fork, update PR body with validated scan/coverage
   funnels, and inspect current-head CI/review/coverage before making ready.

The committed assembly evidence remains stale and incomplete: 4,783 candidates,
4,624 pending, 126 passed, 32 N/A and one failed at its older revision. Do not
hand-edit it to match diagnostic counts. Reports/binaries belong in ignored
`_out/`; clean only owned generated files and candidate caches, never shared
module caches or unrelated containers. Old progress narratives remain in Git
history instead of accumulating in this checkpoint.
