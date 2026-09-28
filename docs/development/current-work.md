# Current work: PR 40 CI repair

Fix CI before resuming the module-index scan. Follow [validation](validation.md)
and [discovery verification](discovery-verification.md). Do not mix reports
from different source fingerprints or edit a frozen corpus runner.

## Authority and CI

- Upstream xgo-dev PR 40 remains OPEN/DRAFT. Push only to the allowed
  `cpunion` fork; never push to or merge/approve/close upstream PRs.
- Fork staging [PR 3](https://github.com/cpunion/plan9asm/pull/3) is Draft.
  Its remote head is `codex/pr40-fork-ci-20260927`; its base is the existing
  upstream contribution branch `codex/expand-ecosystem-corpus-20260913`.
  Do not update that base during experimental CI.
- Full fork run `36391542382` attempt 2 at `be561a20` finished without
  cancellation: 89 jobs passed; discovery shards 4, 5 and 29 plus the
  aggregate failed. A green Codecov check alone is not evidence of reviewed
  patch coverage.
- All 32 reports passed the provenance/inventory audit. The historical
  evidence branch `codex/pr40-ci-evidence-be561a20` records 4,783 candidates:
  3,941 passed, 825 source N/A, 13 explicitly skipped, and four failed.
  It is complete but not verified, and proves only frozen `be561a20`.
- The current fork workflow uses native priority **success** dependencies.
  Failed versions rehash to 64-shard priorities 5, 29 and 36. The two matrices
  partition all 64 shards, without an explicit parallelism cap, and feed the
  strict aggregate. No coverage gate is removed.
- Fork run `36432163029` attempt 2 is in progress at remote head `7e9bebca`.
  All three priority shards passed. Eleven non-corpus jobs failed the same
  stale CI policy test, which rejected any `GOSUMDB` setting even when it
  selects the official signed checksum database. A focused red/green fix is
  committed locally but will not be pushed before this full run finishes.

## Repairs and local evidence

- Go-selected asmdecl checking now attributes ABI diagnostics to the exact
  candidate files. Test-only dependency failures are retried without tests;
  production dependency errors and the candidate's own ABI errors still fail.
- The parser and both command frontends share Go's symbol-offset grammar,
  including whitespace and constant expressions, with red/green regressions.
  Invalid unresolved FP offsets remain rejected.
- Those earlier repairs now pass the full fork CI outside the five newly
  failing discovery shards, including platform, runtime, official corpus,
  benchmark and coverage jobs. These are separate evidence categories, not
  execution of every external library's own tests.
- All seven failures are gVisor-derived versions whose ring0 raw bytes decode
  to unsuffixed `RDGSBASE`. The complete typed named FSGSBASE lowerer already
  exists; raw-to-Go normalization was missing for its eight L/Q spellings.
- Repair branch `codex/pr40-raw-segment-base-20260928` first reproduces the
  failure, then reuses that named grammar. Width comes from the decoded GPR,
  not DataSize (66 may be ignored). Invalid memory operands, LOCK, extra
  operands and non-64-bit raw encodings fail. Named Go 386 L compatibility
  remains independently tested; it is not a claim of CPU runtime validity.
- Red/green tests cover all 16 GPRs, both widths, ignored prefixes and invalid
  forms. The full reported SWAPGS/RDGSBASE sequence and all siblings produce
  the same semantic IR as named instructions, compiled with LLVM 22 on all
  three amd64 OS targets. The scheduling regression is also red/green.
- The latest failures were module-service transport errors, not missing
  instructions: proxy checksum tiles returned 404 and HTTP/2 module or signed
  checksum streams reset. CI now uses the direct signed checksum database;
  bounded retries recognize HTTP/2 stream resets only on HTTP reads. The
  original failed versions and other locally observed network failures were
  replayed with Go 1.27.1 and LLVM 22: eight passed and one was source N/A;
  none failed. This is targeted evidence, not a replacement for frozen CI.
- A 32-shard job took 129 minutes. Rehashing all 4,783 candidates into 64
  shards reduces the largest shard from 173 to 94 candidates, while retaining
  deterministic exact-version ownership and aggregate verification. The
  current run's three priority shards passed in about 19, 33 and 41 minutes;
  measure the full run before deciding whether further splitting is needed.
- At the local repair head, `go test ./... -count=1 -timeout=20m` and
  `go vet ./...` pass with Go 1.27.1 and LLVM 22. The workflow policy test
  also passed focused checks on Go 1.20 and under Go 1.27 race/coverage.
- The local workflow now runs a cheap `ci_policy` job before the expensive
  priority shards. It executes scheduling regressions and the `TestCI` root
  tests, which would have caught the current 11-job failure before any corpus
  runner started. The gate has red/green scheduling tests and passes
  `actionlint`; it is not yet validated by a fork run.

## Provisional native-layout proposal

GopherJRE exchanges a raw continuation with JIT code, GoJIT finds a native
byte sentinel and enters mid-TEXT, and Sharkie fabricates a code-relative
return address while switching stacks. Ordinary semantic translation does
not preserve these byte-exact native layouts and private calling contracts.

The Draft proposes `skipped_native_layout` for four exact module versions
(three projects, including both Sharkie path spellings). The manifest pins
source hashes, all selected amd64 targets, symbols and independently assembled
Go object byte witnesses. Every other applicable file still compiles. A stale
witness, unpinned target or other translation failure fails the candidate.

This exception policy is **not accepted yet**. Fork Draft CI can evaluate the
proposal under the user's existing staging authorization; it must be explicit
in the PR body and counted separately from passed candidates. Do not claim
full assembly coverage, mark ready or promote it to the upstream contribution
branch until review resolves whether these exclusions are acceptable. If not,
implement a compatible native-object mechanism instead of relaxing checks.

## Next actions

1. Let fork run `36432163029` attempt 2 finish. Inspect all completed job
   logs and all 64 shard reports; do not assume the locally fixed policy test
   is the only failure until the run is terminal. Record the longest shard
   duration and investigate any remaining slow candidates.
2. Batch any further red/green fixes on the separate local repair branch.
   Rebind the pending assembly ledger after every source change and run the
   relevant local gates. Do not mix reports from this remote head with a new
   source fingerprint.
3. Push only the `cpunion` fork staging head. Cancel the superseded automatic
   run, wait for cancellation, then rerun the workflow. Audit all 64 reports
   and update the ledger and PR body from that frozen evidence.
4. Resolve all CI failures, review findings, patch-coverage requirements and
   the provisional exception scope before ready/promotion. Keep inventory,
   translation/object compilation and executed runtime claims distinct.

The separate server inventory and cgo records are outside this CI repair.
Remove only owned temporary workspaces/caches after completed runs; preserve
shared caches, frozen evidence and unrelated worktrees.
