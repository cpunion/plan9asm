# Current work: concurrent inventory and assembly coverage

Continue remote inventory and local corpus coverage concurrently. Do not wait
for history backfill to finish before testing already-discovered assembly.
Read [validation](validation.md) and [discovery verification](discovery-verification.md).

## Authority and evidence

- Upstream xgo-dev PR 40 remains Draft. Fork cpunion PR 3 was closed after
  its CI passed; do not reopen it just to test subsequent changes. This batch
  is staged in fork Draft PR 4 against the fork's
  `codex/expand-ecosystem-corpus-20260913` branch. Do not update that
  upstream-connected branch until current-head CI and exclusion review pass.
- Fork [run 36524259971, attempt 2](https://github.com/cpunion/plan9asm/actions/runs/36524259971)
  passed all 128 jobs and all 64 discovery shards at source `fe028700`.
  All raw reports were independently audited; local evidence commit
  `29d624c1` records its complete, verified assembly ledger.
- That checkpoint proves only its frozen source and inventory. Importing new
  records or changing source requires fresh reports, not copied pass flags.
  Derive current counts from the ledgers and keep funnel tables in the PR body.

## Two independent lanes

1. The standalone server only inventories source. Its bounded history target
   is `2025-10-01T00:00:00Z`, inclusive, with the existing head cursor preserved.
   It publishes a locked snapshot at the cutoff and exits. Direct cgo inventory
   and deployment settings stay outside this repository.
2. Local coverage imports a validated assembly-only checkpoint into a separate
   persistent worktree, commits it, then freezes source, tools and input ledger.
   Prioritize shards with the most exact versions absent from the last verified
   checkpoint. Every candidate and all 64 shards remain required; priority is
   scheduling, not a coverage exemption.

Queue newer scan exports outside the runner worktree. Do not merge them while
tests run. The cutoff watcher must collect/export only during an active earlier
coverage batch, not start a competing import or duplicate full corpus run.

## Current repairs

- Issue [44](https://github.com/xgo-dev/plan9asm/issues/44): public Go binding
  now retains x86 ABI0 frames for declared callees, including result-only
  calls and nested aggregates. Framed amd64 BP references share SP storage.
  Indirect callback signatures require a complete, straight-line typed stack
  forwarding proof; names alone never select this path. The focused
  `TestIssue44`, `TestX86ABI0Forward` and `TestGoABI0Nested` tests check LLVM 22
  objects and values reaching direct/callback callees. Keep full external
  project runtime claims separate from these independent oracles.
- Discovery exposed raw Intel VMCALL in a TamaGo fork. Validate both KVM
  hypercall encodings, VMCALL and VMMCALL, across all five x86 platform
  targets. Privileged instructions have decode/ABI/object tests, not ordinary
  user-space execution tests. Replay the discovering shard after integration.
- Intel ixl-go exposed ENQCMD; ENQCMD/ENQCMDS now share one raw address
  grammar and lowerer. Tests cover 32/64-bit modes, SIB/displacements,
  extended registers, 16/32-bit address overrides, explicit FS/GS prefixes,
  64-byte memory inputs and retry/status flags. Raw PC-relative command
  sources still fail closed when source-layout context is unavailable.
  Run `go test . -run '^TestX86(RawEnqueue|Enqueue)' -count=1` and the command's
  supported-op extraction regression, then replay discovery shard 39.
  LLVM inline-asm memory constraints discarded FS/GS pointer address spaces
  in the regression; explicit segment prefixes are required. Audit the older
  cache/descriptor inline-asm lowerers for the same issue separately.
- A complete llgo integration probe also exposed a separate C ABI pass bug:
  sizing an intrinsic's metadata parameter before excluding LLVM intrinsics.
  The isolated fix and child-process regression are in
  merged [xgo-dev/llgo PR 2709](https://github.com/xgo-dev/llgo/pull/2709).
  End-to-end coverage must test the actual llgo revision, not a local overlay.
- The cutoff inventory exposed raw ENDBR64 in ethereum-vanity-address.
  The complete ENDBR32/64 byte family now has instruction-boundary, Go-form,
  five-target LLVM 22 compilation and GP/flags-preservation tests. ENDBR32
  remains raw-only because Go does not name it. These tests do not establish
  operating-system CET enforcement. Replay the discovering external shard.
- The assembly ledger must match all discovered assembly candidates and the
  freshly verified shard outcomes. Passes require positive compiled counts;
  every source or explicit skip retains scoped reasons/evidence. Pending and
  failed are working states, not accepted final outcomes. Publish the full
  ledger before promotion; CI rejects an incomplete or mismatching snapshot.
- Keep all discovered instruction, signature and LLVM-object coverage here.
  Third-party llgo final-link and runtime coverage belongs in a separate
  `llgo-compat` repository. Its module/target results must be independently
  evidenced; neither an object compilation nor `ld -r` is a final-link pass.
  Cgo compatibility will use a separate repository and the standalone cgo
  inventory, not this assembly ledger.
- Package checks are batched per target with precise per-package fallback;
  independent LLVM translator processes are bounded to two. All candidates,
  files and target outcomes remain accounted for across 64 shards.
- Cache-writer regressions reproduce descendants surviving success, failure
  and cancellation, including inherited output-pipe hangs. Captured Unix
  commands now own and terminate their process groups; Git automatic
  maintenance stays foreground. Bounded directory-not-empty retries still
  report persistent cleanup failures. Other filesystem errors are not retried.
- Priority shards are 5, 29, 36, 38 and 42. All must pass before the remaining
  jobs run; all 64 shards and their strict aggregate remain mandatory.
- Go 1.27.1 and LLVM 22 are pinned for external corpus evidence. Build and
  test with Go while llgo support is incomplete; do not add `!llgo` tags.

## Provisional native-layout proposal

GopherJRE, GoJIT and Sharkie use native object byte layouts or private JIT
continuations that ordinary semantic LLVM translation does not preserve.
Four exact-version exceptions in `testdata/corpus/native-layout.json` are
provisional, with source hashes and Go object witnesses. Other files still
compile; exceptions count separately from passes. Keep both PRs Draft until
review accepts this policy or a compatible implementation replaces it.

## Next actions

1. Run four local shards concurrently with distinct reports and an owned shared
   build cache. Remove candidate sources/output after processing. Keep every
   applicable file and target, including non-host architectures.
2. Query audited partial progress with `scripts/discovery-status.sh`. Publish
   matching results with `scripts/update-assembly-ledger.sh` in a separate
   evidence worktree, leaving the runner revision and inputs untouched.
3. Reproduce failures in a separate development worktree. Follow the complete
   instruction-family/operand-format TDD procedure; do not recategorize missing
   support as source N/A. Changed source invalidates earlier reports.
4. Audit all 64 reports and strict aggregation before the batch push. Update the
   PR funnel, then let fork CI finish. Resolve review, patch coverage and the
   provisional exclusions before Ready or promotion. Object compilation is not
   every external project's runtime test.

An incremental scan with an empty output ledger starts at now; `-seen-report`
only deduplicates and does not inherit its cursor. Seed a separate output with
the converter before continuing the saved interval. Do not merge an empty
now-to-now scan as if it covered the committed incremental gap.

Clean only owned temporary files. Never commit caches, binaries, compressed
records, personal paths or diagnostic-only reports.
