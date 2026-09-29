# Current work: PR 40 fork validation

Prioritize fork CI and compilation of every already-scanned assembly candidate
before resuming module-index history scanning. Read [validation](validation.md)
and [discovery verification](discovery-verification.md) before changes.

## Authority and evidence

- Upstream xgo-dev PR 40 and fork cpunion PR 3 remain Draft. Push repairs only
  to the fork PR 3 head, `codex/pr40-fork-ci-20260927`. Do not update its
  upstream-connected base until current-head CI and exclusion review pass.
- Fork run [36509240233](https://github.com/cpunion/plan9asm/actions/runs/36509240233)
  checks source `5a0699c0`. Let it finish; audit all 64 reports together.
  Shards 38 and 42 failed while removing Hysteria's private Git cache, with
  `directory not empty`. This is a cleanup failure, not a missing instruction.
- Older run 36458192352 attempt 2 passed all 126 jobs at source `5a894a20`:
  3,944 passed candidates, 826 source N/A and 13 explicit exceptions, with
  no failed or pending candidates. That evidence cannot certify new source.
- The committed inventory remains 12,665,430 index entries, 811,704 unique
  module paths, 802,019 scanned exact versions, 4,783 assembly candidates and
  13,631 retained scan failures. History starts in March 2026, not 2019.
  Keep funnel tables in the PR body and preserve input records during tests.

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

1. Finish and audit the current run. Preserve its reports under ignored
   `_out/`; do not mix them with reports from a repaired source.
2. Validate cleanup regressions, corpus-tool tests, race tests and workflow
   partition checks. Commit the repair and rebind the pending assembly ledger
   from an empty ignored reports directory before pushing to the fork.
3. Run the complete repaired-source CI. Audit all 64 reports against the
   frozen source, tool bytes, scan ledger and candidate ownership; publish
   actual progress through `scripts/update-assembly-ledger.sh`, including
   failures or pending candidates rather than hand-upgrading statuses.
4. Update the PR funnel and performance comparison from verified reports.
   Resolve review, patch coverage and the provisional exclusions before
   promotion. Object compilation is not every external project's runtime test.

An incremental scan with an empty output ledger starts at now; `-seen-report`
only deduplicates and does not inherit its cursor. Seed a separate output with
the converter before continuing the saved interval. Do not merge an empty
now-to-now scan as if it covered the committed incremental gap.

Clean only owned temporary files. Never commit caches, binaries, compressed
records, personal paths or diagnostic-only reports.
