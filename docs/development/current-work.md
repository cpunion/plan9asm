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
- Full fork run `36371763306` at `ded20b17` finished without cancellation:
  88 jobs passed; discovery shards 11, 15, 16, 19 and 30, plus the aggregate,
  failed. The previous five failing shards all passed. A green Codecov check
  alone is not evidence of reviewed patch coverage.
- All 32 reports passed the provenance/inventory audit. The historical
  evidence branch is `codex/pr40-ci-evidence-ded20-20260928`; its ledger is
  complete but not verified because seven exact versions failed. It proves
  only the frozen `ded20b17` source, not the next repair head.
- The user now requires priority **success**, not just startup. The next
  workflow uses native dependencies: shards 11, 15, 16, 19 and 30 must all
  pass before other jobs execute. Both matrices still partition all 32 shards
  and feed the strict aggregate; no coverage gate is removed.

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

1. Finish focused Go 1.20/current-Go checks, commit the repair and scheduling
   changes, then rebind the pending assembly ledger with the validated updater.
   Documentation affects source fingerprints; derived assembly output does not.
2. Freeze that source and replay all seven failed exact versions plus relevant
   root/frontend tests. Keep diagnostic replays separate from canonical 32-shard
   evidence. Fix any additional failure before pushing the batch.
3. Push only the fork staging head. As requested, cancel the new automatic run,
   wait for cancellation, then restart the workflow. Its five priority shards
   run first; all other jobs remain blocked until they pass. Watch the full run
   and audit/publish current-source reports with the ledger updater.
4. Resolve all CI failures, review findings, patch-coverage requirements and
   the provisional exception scope before ready/promotion. Keep inventory,
   translation/object compilation and executed runtime claims distinct.

The separate server inventory and cgo records are outside this CI repair.
Remove only owned temporary workspaces/caches after completed runs; preserve
shared caches, frozen evidence and unrelated worktrees.
