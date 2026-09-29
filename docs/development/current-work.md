# Current work: concurrent inventory and assembly coverage

Continue remote inventory and local corpus coverage concurrently. Do not wait
for history backfill to finish before testing already-discovered assembly.
Read [validation](validation.md) and [discovery verification](discovery-verification.md).

## Authority and evidence

- Upstream xgo-dev PR 40 and fork cpunion PR 3 remain Draft. Push repairs only
  to the fork PR 3 head, `codex/pr40-fork-ci-20260927`. Do not update its
  upstream-connected base until current-head CI and exclusion review pass.
- Fork [run 36524259971, attempt 2](https://github.com/cpunion/plan9asm/actions/runs/36524259971)
  passed all 126 jobs and all 64 discovery shards at source `fe028700`.
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
