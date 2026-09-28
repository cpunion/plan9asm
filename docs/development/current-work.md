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
- Full fork run `36333189048` at `a48ebfc2` finished. Shards 0, 9, 14, 28
  and 29, plus aggregate verification, failed. All other jobs passed.
  A green Codecov check alone is not evidence of reviewed patch coverage.
- This batch is integrated on `codex/pr40-fork-repair-20260928`, based on
  the published staging head. It retains the embedded-module repair and
  fork-first runner selection. Push the batch to the existing fork PR head
  without updating the upstream-connected contribution branch.

## Repairs and local evidence

- Go-selected asmdecl checking now attributes ABI diagnostics to the exact
  candidate files. Test-only dependency failures are retried without tests;
  production dependency errors and the candidate's own ABI errors still fail.
- The parser and both command frontends share Go's symbol-offset grammar,
  including whitespace and constant expressions, with red/green regressions.
  Invalid unresolved FP offsets remain rejected.
- Focused clean-source external checks pass Cloudflare SIDH, Apache Arrow
  v15/v18, Query-farm Arrow v18, Milvus Arrow v17, acolita/crypto,
  gitpod-io/golang-crypto and hashicorp/go.net. The local repair branch is
  `codex/pr40-asmdecl-dependency-20260928`; focused reports remain in `_out/`.
- All five ancestor-source diagnostic shards finished. Seven ordinary
  failures (two proxy timeouts and five foreign asmdecl attributions) now
  pass focused checks on the repaired code. Those reports also include four
  native-layout exceptions and one invalid-source exception, never passes.
  These old reports cannot establish current-source canonical coverage.
- The repaired code passed the Go 1.27 full root and both frontend suites,
  Go 1.20/1.27 corpus-command suites, corpus-command race tests, Go 1.20
  focused parser tests and root vet. The frontends require Go 1.24.
  A first 20-minute root run timed out under heavy shard contention; the
  60-minute rerun passed in about 17 minutes.
- The current-Go official observed-form gate passes all five supported
  architectures with zero unsupported forms or parse errors. All 44 Go 1.27
  standard-library target settings compile with LLVM 22. The ARM64 Plan 9
  reference-corpus gate also passes. These are not runtime coverage claims.

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

1. Verify integration matches the tested code, commit the batch, then rebind
   the all-pending assembly ledger with the validated updater. Documentation
   affects source fingerprints; derived assembly-ledger output does not.
2. Push the batch only to the fork staging head. Let one full CI run finish
   before another push; inspect failures and prepare fixes separately.
3. Download all 32 reports and audit their exact source, tools, candidate
   ownership and scan-ledger provenance. Publish passing and failing evidence
   through `scripts/update-assembly-ledger.sh`, never hand-edit pass flags.
4. Resolve all CI failures, review findings, patch-coverage requirements and
   the provisional exception scope before ready/promotion. Keep inventory,
   translation/object compilation and executed runtime claims distinct.

The separate server inventory and cgo records are outside this CI repair.
Remove only owned temporary workspaces/caches after completed runs; preserve
shared caches, frozen evidence and unrelated worktrees.
