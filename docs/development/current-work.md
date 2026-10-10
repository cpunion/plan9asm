# Current work: close the PR 40 CI repair batch

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Tests of a component or
an older revision are not current-head CI or assembly-ledger passes.

## Fixed delivery scope

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch before fork CI and review pass.
Never push to, approve, merge or close upstream PR 40.

The integration branch is `codex/pr40-ci-closure-20261011`. Stop opening new
instruction-family audits until this repair batch has run in CI. Preserve the
separate scalar-FMA development branch; it is not part of this batch.

Published `cb8b2ad3` is running fork CI 38018079008. Policy preflight passed;
shards 29/39 failed and four priority shards are still running. Let it finish before pushing
the next repair batch. No current-head corpus or full CI pass is established.

Earlier `c09eff1e` failed run 37711752851: priority shards 5/29/38/39 failed,
36/42 were canceled, and downstream gates did not complete. The four failed
shard logs contain 110 failed candidates (31/23/25/31 respectively). Initial
diagnostic grouping is eight rule classes, not 110 missing instructions:
profile bounds; downloaded-source identity; instruction effects; asmdecl/source
attribution; original profile roles; native entry/call/return; memory/frame
contracts; and remaining instrumentation diagnostics. These are a triage list,
not a claim that the new batch fixes every candidate.

Measure delivery by original failing candidates becoming verified passes and
CI jobs becoming green. Do not substitute growing test-matrix row counts.
New counterexamples join the root-cause list; do not interrupt every batch
with unrelated speculative implementation. Keep correctness gates unchanged.

## Included repairs and retained evidence

- Downloaded-source binding, per-target feature-profile bounds and x86 typed
  effects are already integrated, together with earlier ARM/VFP/NEON and Wasm
  repairs. Match original CI examples to their independent regression tests
  and rerun them; commit presence alone is not evidence of closure.
- SDK observation retains direct members of actual imported SDK packages,
  selected descendants, literal includes and shadow guards. Modules remain
  recursive. Retention errors now preserve their actual cause.
- Exact x/crypto v0.57.0 replay measured 54,155,434 retained source/export bytes
  plus a 17,142,188-byte Go driver, exceeding the former 64-MiB shared bound.
  The shared streaming inventory now has a finite 128-MiB bound, with source/
  tool boundary and mutation regressions. Current Go, Go 1.20, race and Linux
  component gates passed. The original three crypto versions completed with
  zero passes and three failures: runZeroInc/excrypto at actual source/CPU
  routing, tdv/crypto and x/crypto at exact package-versus-module-root binding.
  These are still failures, not approved skips or library passes.
- Downloaded fixed-local capture now distinguishes a selected subpackage from
  its original module root, deriving its declared import path from the guarded
  ZIP/go.mod. The independent first-party fixture reproduced the old refusal
  on Linux, Darwin and Windows; all three translation APIs now produce LLVM 22
  O0/O2 objects for each target. Ancestor metadata changes permanently revoke
  the capability. Current Go and focused Go 1.20 gates passed. Imported SDK
  and exact foreign modules now have independent source/export/ZIP handoffs,
  with actual four-role RED/GREEN controls, three-API O0/O2 objects and
  metadata/source mutation negatives. CPU-profile and complete native frame
  contracts remain separate; this is not crypto replay closure. At `fd57caf0`,
  the Linux original-import/nested-object roots passed (337.394s), as did four
  shared SDK/frame regression roots. Original x/crypto replay is terminal:
  zero passes, one failure at the unrelated local `XOF` interface declaration.
  The former root/import failures are absent, but this is not library closure.
- The AMD64 fixed-local observer's v1-only gate has an independent RED for all
  nine Go-accepted v2/v3/v4 combinations across Linux, Darwin and Windows.
  All twelve target/profile combinations now pass actual source/export and
  three-API LLVM object controls (72 objects), with profile-revocation and
  conditional-CPP negatives. Actual Go 1.20 all-target profiles and existing
  numerical/frame joins passed too. A test-root-owned build cache reduced this
  focused current-Go suite from 140.877s to 62.199s without removing cases;
  this is not a corpus-wide performance or external-library pass.
- The XOF counterexample exposed an unrelated-Go-type/assembly-transport
  boundary error, not a missing mnemonic. Eight independent interface,
  callback, map, channel, floating, recursive, generic and alias cases reproduce
  the refusal. Complete original Go typechecking and declaration roles remain
  mandatory; restricted transport/layout comparison now applies to every
  actual assembly signature's reachable type tree. Seventeen identity
  mutations and fifteen unsupported signature roles still reject. Focused
  current-Go and Go 1.20 gates and three-target, three-API LLVM objects pass.
  At `814de5ce`, the Linux first-party roots passed (533.700s). The original
  exact crypto replay in `_out/ci-closure/types-replay/` is terminal with zero
  passes and one failure: mixed TEXT/DATA is outside the frame observer's
  closed contract. Its package/type gates are not external-library closure.
- CI's new archive/member rejections exposed a second, distinct bound: whole
  module ZIPs and unrelated payloads had inherited the 16-MiB source-buffer
  limit. Origin inventory now streams complete artifacts under Go's 500-MiB
  ZIP/tree checks and one finite graph budget. Ordinary source/export buffers,
  tool roles, exact ZIP/tree hashes and permanent mutation revocation remain
  strict. Independent large-archive/member REDs, race and actual Go 1.20
  component gates pass. Actual original Go downloads of both large fixtures
  pass three targets, three translation APIs and LLVM 22 O0/O2 objects. This
  does not enlarge the separate selected-package source/parser budget or prove
  an external candidate passed; original-library replay remains required.
- Two FP-alias consumers omitted existing complete typed grammars: CMP/TEST
  and scalar/block BROADCAST. They now reuse those grammars for read-only byte
  widths; no suffix guessing, parameter writes or result initialization is
  introduced. Independent original-Go REDs cover both architectures, all
  widths/directions and vector/mask formats. Current Go and Go 1.20 consumer
  gates, real asmdecl aliases and five-target object regressions pass. The old
  Go maskless `.Z` encoder defect has an explicit reserved-byte observation,
  never an accepted translation or a skipped test. The five-architecture
  official classification gate remains unchanged and passes. Original failing
  libraries remain failures until their complete frozen replay succeeds.
- ARM64 FPCR/FPSR, NZCV, scalar binary operations and RNDR named/raw effects
  have independent original-Go and LLVM runtime evidence. Min/max's original
  full Linux run failed at the 45-minute deadline, without an observed value
  mismatch. Keep that failed log. Four disjoint FMAX/FMIN/FMAXNM/FMINNM roots
  now retain the full denominator and the same per-root deadline. All four
  completed their Go/LLVM comparisons at `e693532a`; FMAX/FMIN passed, while
  NM roots failed the final assertion because its AH signed-zero counter is
  only populated by non-NM instructions. The scoped witness repair keeps all
  comparisons and requires actual NEP for every family. Its independent RED/
  GREEN passes. Both fresh full Linux NM roots passed at published `cb8b2ad3`
  (678.94s/679.10s), retaining every original comparison.
- CI policy's old shared-batch-cache fixture is updated for the refill queue:
  caller cache isolation, exhaustive priority order, sticky shard failures and
  a mandatory final aggregate. All 36 Node tests and the three actual-child
  Go queue tests pass; the original three policy failures remain recorded.
- The known ARM64 WORD truncation counterexample now uses one shared Go
  `C_LCON` check before native preflight, decoding and lowering. Rejected
  64-bit constants can no longer become valid low-word opcodes, including
  dead source and missing/foreign targets. Legal wide MOV/bitmask classes
  remain accepted. The independent Go-oracle RED and nine focused GREEN
  roots are retained under `_out/ci-closure/word-class-*`, including actual
  Go 1.20/1.27 constant oracles and three-target LLVM objects. The current
  five-architecture official corpus gate passed without changing its baseline.
  The separate unfinished scalar-FMA branch is still not integrated.

## Active frozen validation

Do not edit tracked source, tools or ledger in these trees while consumers run.
Paths below are relative to their named worktree. Preserve terminal reports
and clean only exact owned caches after every consumer has closed.

- `codex/pr40-result-effects-followup-20261010`, `e693532a` (terminal):
  `_out/minmax-partitions/` is terminal with two passes and two assertion
  failures above. The integration tree's `_out/ci-closure/` holds the scoped
  regression and successful required Linux NM reruns; do not relabel old logs.
- `codex/pr40-original-graph-budget-20261011`, `e8d58a41` (terminal):
  `_out/graph-budget/` retains failures for runZeroInc/excrypto v0.43.3, tdv/crypto
  v0.0.0-20260609182332-5f2de1a9f1e2 and x/crypto v0.57.0, using stamped tools,
  fresh reports/events and one disposable validator. Earlier failed diagnostic
  runs remain in their original worktrees.
- `codex/pr40-ci-closure-20261011`, `fd57caf0` (terminal):
  `_out/ci-closure/imports-replay/` retains the original exact x/crypto v0.57.0
  failure at local `XOF`, after independent Linux source/import gate passes.
  Its candidate cache was cleaned after every consumer closed.
- Detached `191f2cab` (active; archive-bound and CMP/TEST repairs):
  `_out/origin-archive/linux-replay/` retains the Linux first-party passes
  and exact cbysal/go-ethereum v1.17.5 and enceve/crypto replay. Its source,
  tools and ledger are frozen; inspect terminal reports before crediting any
  candidate. The later BROADCAST repair is not part of this snapshot.
- `codex/pr40-refill-sample-20261010`, `02c9ebbb`:
  `_out/refill-sample/` continues the fixed 53-candidate, 37-shard sample with
  four workers. This older revision cannot update the current-head ledger.

The older whole-module run at `044bb668` failed: 18/20 packages passed, root
had one obsolete assertion failure and the corpus driver had three fixture
timeouts. Those regressions have newer repairs, not a retroactive pass.

## Ordered next gates

1. Let published `cb8b2ad3` finish its priority matrix; do not interrupt it.
2. Finish the package/source binding repairs, retaining the original examples
   as failed until each complete replay succeeds. Commit locally in batches.
3. Compare new CI failures with the original eight classes. Fix each class with
   independent RED/GREEN tests and original-example replay; batch the next push.
4. After priority success, let all remaining current-head gates run. Publish
   only matching, complete reports through the automatic ledger writer.

Known later gates remain real failures/blockers: the latest strict SDK report
has 163/184 passes, with frame/entry/ABIInternal/TLS/global/alignment gaps;
eleven required Linux AMD64 AVX512 runtime roots need compatible hardware.
Neither portable execution nor LLVM object success replaces native execution.
Keep the PR draft; there is no current-head completion or ready claim.

## Inventory and ledger boundaries

Inventory runs on the designated scanner, never as a local background process.
The scanner does not compile external code; disposable validators do. cgo
inventory stays outside this repository. Do not rescan committed ranges.

Committed inventory remains 18,758,534 index entries, 965,347 paths and 6,067
assembly versions. Independent staging has 21,257,856 contiguous entries,
1,051,454 paths and 6,700 assembly versions, not imported/tested coverage.
Historical `3e82acad` assembly evidence has five passes, one CRC failure and
6,061 pending; newer mixed-revision samples cannot replace that provenance.

Root alone publishes the ledger using the automatic writer and matching
schema-11 reports, progress v3, assembly-ledger v3, and complete file/target/
profile/source/tool evidence. Never hand-edit passes, mix revisions, load all
records just for status, or commit obsolete compressed/per-run ledgers.
Keep current scan/coverage funnel tables in the PR body.
