# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md) before continuing.
Use `git worktree list` to locate the persistent development and frozen
verification trees; do not reconstruct this work in a temporary checkout.

## Authority and promotion

Upstream PR 40 remains Draft. Repair batches use fork Draft PR 4, targeting
`codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to `cpunion`;
do not update the upstream-connected branch until current-head fork CI,
review and exclusion-policy gates pass.

The last inspected fork run, 36662878534 at `3aeb6b7c`, finished with failures
in discovery shard 39 and its aggregate. Later local repairs are not pushed.
Let each new CI run finish, inspect failures, and batch verified fixes.
Keep scan and coverage funnel tables in the PR body, not this checkpoint.

## Frozen inventory and historical evidence

The imported scan has continuous ranges from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. The bounded historical inventory is complete; this does
not cover the later incremental interval or all historical Module Index data.
The standalone inventory and cgo records stay outside this repository.

Source `754cbebb` finished all 64 discovery shards but failed five shards.
Audited evidence `3c32ecc3` records those outcomes: complete accounting is not
verified success. Read its manifest/progress, not a remembered percentage.
Ten exact-version download failures have a checksum-verified read-only proxy
checkpoint. The other failure was a Go compiler/GC crash in spidermonkeywasm2go;
a single successful retry does not establish a fix. Neither class is source N/A.

Keep that source, input ledger, tools and reports unchanged. New source needs
fresh reports; do not copy pass flags or merge reports from different revisions.

## Integrated development

`codex/pr40-guarded-integration-20261001` holds the integrated repairs. Its
`91b8717a` checkpoint passed both corpus/scanner unit suites and vet, and was
scanned against all five official corpora for Go 1.20–1.27 source tables.
These are enumeration/classification checks, not eight toolchain runtime runs.

- ARM64 local-control, raw-pool, caller-SP, SB memory and register-address
  families have independent Go/LLVM runtime regressions. Compile-only probes
  end with a faulting branch rather than inventing an ordinary return contract.
  Unknown register calls/returns must retain context failures.
- `f8cae321` filters zero-size asmdecl warnings only for a selected file's exact
  literal TEXT line and symbol. Other size/offset/width errors remain visible.
  Both current-Go native oracles and focused Go 1.20 unit checks passed.
  The rhnvrm/tools replay still correctly finds a separate nonzero ABI mismatch.
- `91b8717a` adds case-sensitive frontend registration inventory, including
  common/architecture names, aliases, pseudos and explicit sentinels. It does
  not change the existing form denominator or claim LLVM/runtime coverage.
- Official ARM64 classification changed only `BL (R2)`, `CALL (R15)` and
  `JMP (R29)` from supported to context-required. Audit full form differences
  across all versions before updating fingerprints; retain historical exceptions.

## Active independent repairs

Use separate persistent worktrees and commits, then review before integration:

Architecture applicability must not hide tool failures. The Go assembler
probe now rejects infrastructure/unknown failures, bounds output and process
lifetime, and retains each actual source rejection diagnostic. Generated
header constants remain visible until the real package build provides them.
Replay every schema-9 shard after the integrated source is frozen.

1. Ordinary applicability evidence: replay exact source selection, including
   root-directory files, ignored/nested directory boundaries, suffixes and tags.
   Generic empty-config reasons are insufficient. Keep compact deduplicated
   headers/hashes and module identity, not gzip or full-source blobs in ledgers.
   Update strict reports, aggregation and ledger verification together.
2. ARM native continuation: `d13a8001` has Linux kuser ABI/runtime evidence,
   but is awaiting framed-tail/LR and conditional CPSR review. Do not integrate
   it as a full stdlib pass. asyncPreempt requires an actual native entry/stack
   contract; a zero-argument C entry or ordinary return is not equivalent.
3. ARM64 Go frame/ABI: `a3a5e484` adds strict return guards and `c379d328`
   fixes raw source RET widths with actual Go-object/runtime oracles. They are
   not yet integrated: p256 and LZ4 expose missing private-frame alias/ABI
   contracts. A declaration-free register helper must not receive a guessed
   integer return that overwrites R0. Referenced Go ABI0 frame bridging needs
   independent runtime tests and must not conflate explicit ABIInternal entry.

## Completion sequence

1. Finish each bounded repair with true red/green evidence, affected-target
   LLVM 22 objects and required runtime oracles. Commit promptly; batch pushes.
2. Integrate reviewed fixes, replace this checkpoint, commit, and freeze source.
3. Run official/form gates, all stdlib configurations, benchmark, exhaustive
   root and nested-module tests, vet/build, and required cross-runtime gates.
4. Run all 64 discovery shards in bounded parallel batches using verified
   read-only proxies. Each candidate workspace is owned and removed after use;
   never clean the user's shared Go cache or a live writer's cache.
5. Audit current-source reports and publish matching assembly evidence in a
   separate worktree. Every version needs a tested pass or an explained skip;
   failed/pending are not accepted final states. Do not relabel tool failures.
6. Push the batch to fork PR 4, finish CI/review/coverage, then promote to PR 40.

Native-byte-layout/JIT exclusions remain provisional and require review.
Third-party final-link/runtime results belong in the separate compatibility
repository. Its schema-4 vendor proof must be captured before Go oracles and
rechecked after execution. Those local workflows are not deployed yet; old
schema-3 diagnostic results cannot be promoted. The llgo integration also has
a genuine failure against its old pinned plan9asm dependency; a local replace
is development evidence, not passing pinned CI.
