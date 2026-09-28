# Current work: PR 40 fork validation

Prioritize CI and the discovered third-party assembly corpus before resuming
module-index history scanning. Read [validation](validation.md) and
[discovery verification](discovery-verification.md) before changing source or
ledger evidence.

## Authority and frozen evidence

- Upstream xgo-dev PR 40 remains OPEN/DRAFT. Push repairs only to the allowed
  `cpunion` fork Draft PR 3, never directly to upstream. Do not merge, approve
  or close upstream PRs.
- Fork run [36458192352](https://github.com/cpunion/plan9asm/actions/runs/36458192352),
  attempt 2, completed successfully at source `5a894a20`: 126/126 jobs,
  including all 64 discovery shards and the strict aggregate.
- Independent local report verification at that frozen source found 4,783
  selected assembly-bearing module versions: 3,944 passed translation and
  LLVM 22 object compilation, 826 were evidence-backed source N/A, six were
  pinned invalid-source skips, two superseded, one private extension and four
  provisional native-layout skips; zero failed or remained pending.
- Local evidence commit `661acba4` records that exact source's passing ledger
  (`complete=true`, `verified=true`). Do not reuse it after any source or
  ledger change. The staging PR's current source is not yet this evidence
  commit; its pending ledger must be rebound before the next push.
- The inventory remains incomplete historically: 12,665,430 index entries,
  811,704 unique module paths, 802,019 scanned exact versions, 4,783 matched
  versions and 13,631 retained scan failures. Its recorded range starts in
  March 2026; do not claim complete coverage of Go modules or back to 2019.
  Keep the funnel table in the PR body, not this document.

## Staged performance repair

- The earlier 32-shard run had a 129-minute job. Rehashing to 64 shards cut
  the largest shard from 173 to 94 candidates, but the latest successful run
  still had a roughly 112-minute shard. Its `reflectx@v1.8.2` candidate alone
  took 59.2 minutes and covers eight packages across six targets.
- A separate local branch, `codex/pr40-corpus-batch-20260929`, contains three
  unpushed commits: `b57b6355` batches current-Go build/asmdecl checks per
  target with per-package fallback; `be711e25` tests real ABI isolation;
  `904b77df` bounds independent translator processes to two and aggregates
  their reports deterministically after both finish.
- A diagnostic replay of `googlesqlwasm2go@v0.2.0` on `linux/amd64` compiled
  all 11 selected packages in 2m30s with the staged repair, versus 4m00s
  serially (about 37.6% faster). No source or LLVM failures were reclassified.
  Doubling LLVM chunk thresholds was slower and was discarded. Temporary
  diagnostic source and binaries were removed; small logs remain ignored.
- The staged branch passed Go 1.27.1/LLVM 22 `go test ./... -count=1
  -timeout=20m`, both nested command-module suites, full `go vet ./...`, the
  corpus test package and focused race tests. These are local results, not
  current-head CI or full-corpus evidence.

## Provisional native-layout proposal

GopherJRE, GoJIT and Sharkie use native object byte layouts or private JIT
continuations that ordinary semantic LLVM translation does not preserve.
The Draft proposes `skipped_native_layout` for four exact module versions,
pinning source hashes and Go object witnesses. Other files still compile,
and these four are counted separately from passes. This policy is **not
accepted**; keep both PRs Draft until review resolves it. If rejected,
implement a compatible native-object mechanism instead of relaxing checks.

## Next actions

1. Confirm the fork run and strict 64-report audit above remain available.
   Inspect destination remote before any push.
2. In the staged performance branch, update this checkpoint, then rebind the
   pending assembly ledger to its new semantic source fingerprint using an
   empty ignored reports directory. Do not copy the old verified outcomes.
3. Commit the checkpoint/ledger, push only the `cpunion` PR 3 head, cancel
   the automatic duplicate run after the push, wait until it is terminal,
   then rerun the workflow. Keep the owner-based runner switch and every
   coverage gate intact.
4. Watch the complete new-source CI. Audit all 64 current reports and update
   the passing ledger and PR body only when source, tool and ledger provenance
   match. Compare longest job and runner-minutes with the 112-minute baseline.
5. Resolve review, patch coverage and the provisional skip policy before
   ready status or promotion to the upstream-connected fork branch. Keep
   inventory, object compilation and executed runtime claims separate.

Clean only owned temporary files. Never commit caches, generated binaries,
compressed scan results, personal paths or diagnostic-only reports.
