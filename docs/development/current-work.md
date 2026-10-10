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

Old fork CI 38018079008 finished naturally with failure. All six priority
shards failed; 36/42 hit the 330-minute deadline and retained partial reports.
The guarded publisher then fast-forwarded only fork PR 4 to `aedfac9b`, a
35-commit repair batch. New fork CI 38043255488 has passed policy preflight
and started all six priority shards. Let that run finish. Its head and the
integration worktree remain frozen; the upstream-connected branch is unchanged.
No current-head corpus or full CI pass is established. Further observed-failure
repairs now live on `codex/pr40-ci-failure-replay-20261010`, based on `b5af3a10`.
The earlier `codex/pr40-failed-gates-20261010` tree is frozen for broad local
verification; do not edit it while those first-party tests are running.

Current CI shards 5/29/39 finished with 131 passes, 128 failures, 39 source
N/A and one invalid-source skip across 299 versions. Their candidate statuses
are unchanged from the previous run. Shards 36/38/42 are still running. Fresh
artifacts are retained in the earlier follow-up tree under
`_out/failed-gates/ci-38043255488/`; these three reports cannot update the ledger.
The prior watcher died during disk exhaustion. Its publisher had already
finished; do not restart that publisher. A read-only `gh run watch` now writes
the integration tree's `_out/ci-closure/watch-aedfac9b-after-cleanup.log`.

A guarded local publisher is prepared under ignored `_out/ci-closure/` as
`publish-after-ci.sh`. Before launching another watcher or changing the
integration tree, inspect its `publish-after-ci.log`, owned process and
`published-ci.json` if present. It pins one complete source revision, the prior
run/head, fork owner, Draft PR 4 and both branch names. It waits for the prior
run's terminal failure, then rechecks clean source and unchanged destinations
before a normal fast-forward push. Any mismatch stops publication. It never
cancels, force-pushes, merges/closes a PR or writes upstream. Keep the integration
tree frozen while it waits; use a separate tree for further development. A new
CI run is only a started gate, never a pass or ledger update.

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

- Shards 5/29/39 exposed ordinary pointer `MOVBQZX` reads missing from the
  lowering-only result observer. Independent REDs cover the entire fourteen-name
  extension family; a second RED finds dead native loads being removable.
  The existing typed grammar now pairs exact-width volatile native reads with
  conditional numerical destination bytes. No pointer validity, pointee extent,
  caller, frame or standalone declaration permission follows. Complete Go encoder
  attributes, five-target three-API objects, unknown/address/raw negatives and
  whole-register byte effects pass. Twenty-eight live/dead kernels compare actual
  original Go machine code and LLVM against an independent byte reference and
  NULL/guard-page fault observations through three APIs at O0/O2. Local AMD64
  execution uses Rosetta, not native Linux evidence. Sixty-four current-Go roots,
  nine actual Go 1.20 roots, CLI extraction, vet and unchanged five-architecture
  official classification pass. Retain `_out/ci-failure-replay/extension-native-*`.
  Original exact-version replay and current-head CI remain required; no library
  pass or ledger status is inferred from these first-party checks.
- Shard 5's exact packdb version exposed a missing broadcast memory consumer,
  not a new opcode: the result observer accepted only register inputs despite
  the complete shared grammar. Independent REDs reproduce the failure without
  the library. All eighteen memory broadcast names now consume existing exact
  initialized FP/AUTO bytes or lowering-owned same-file DATA extents. Masks
  cannot hide missing inputs; native pointers, unknown masks/merges, negative
  offsets and address overrides still reject. The 104 memory forms have actual
  Go objects and three-API LLVM 22 objects on five x86 targets, in batches of at
  most 64 functions. Portable execution compares 624 dynamic-mask calls per API
  and optimization level against an independent byte reference; it does not
  claim original AVX512 execution. Forty current-Go family/regression roots pass
  (66.031s). Eight actual Go 1.20 roots pass after repairing an existing test's
  maskless `.Z` assumption: old Go's exact reserved bytes remain rejected,
  never supported or skipped. Two current-Go reserved-encoding roots, CLI
  extraction, vet and the unchanged five-architecture official gate pass.
  Evidence is under `_out/ci-failure-replay/` in the new development tree.
  Original exact-version replay and current-head CI remain pending; no ledger
  status was promoted. The 58 mixed-TEXT/frame refusals in the three completed
  shards are a separate, still-unrepaired boundary, not broadcast closure.
- Follow-up CPP planning repairs are under `_out/failed-gates/`: offline JSON
  lost private resource indexes and repeatedly lexed complete originals. The
  independent allocation RED and original shard 38 CPU profile reproduce it.
  A bounded per-invocation index now serves registration, active replay and
  condition-free queries; immutable text shares bytes, never mutable pointers.
  The same 94-candidate accounting audit improved from a five-minute timeout
  to 86 seconds, without dropping source hashes or graph/scope checks. Final
  tracked-file checks pass: 27 current-Go roots with actual five-architecture
  Go objects, 26 Go 1.20 roots, race and vet. Preserve the earlier overlay-only
  old-vet virtual-file failure as diagnostic evidence. The broader CPP attempt timed out at
  12 minutes after 52 root passes; it is not a complete suite pass. After the
  user's disk cleanup, the frozen `b5af3a10` retry passed all 104 roots in
  1,233.229s under the normal 45-minute budget, with no failures or skips.
  Evidence is `cpp-b5af3a10-full-retry.jsonl`. Preserve the older timeout and
  disk failures; this fresh result does not relabel those failed attempts.
  This is a local audit optimization, not a current CI or external-library pass.
- The next, not-yet-published batch closes shard 5's low-product caller-effect
  gap. Independent original-Go W/D/Q controls reproduced unregistered D/Q
  effects; replacing only the old lowerers also reproduced incorrect upper
  aliases at LLVM O0/O2. D/Q now share typed shape, lowering, caller/result
  effects and memory tuples with complete Go-table attributes. E4 masked reads,
  legacy native storage, VEX.WIG and physical-versus-named mode32 masks remain
  distinct. The word family is reused, not replaced. All 68 related current-Go
  roots plus the independent malformed-axis root pass; 15 focused Go 1.20
  roots, CLI extraction, vet and the unchanged five-architecture official
  classification gate pass too. Three-API/five-target objects and 110 portable
  runtime kernels cover full aliases, masks, FLAGS, canaries and inactive NULL
  inputs. This is not original AVX512 execution or external-library closure.
  Keep the original discovering candidates failed until frozen remote replay.
- The new full root run exposed a deadline-fixture startup race: the two-second
  stage allowance could expire before two nested LLVM-linked test binaries
  published a child PID. The failure-only shim now starts a lightweight real
  pipe-holding child while retaining the same deadline, error/output witnesses
  and owned-process cleanup assertion. Three repeated lifecycle runs and the
  actual Go 1.20 check pass. Commit `7579e5d2` is separate from published CI.
  Evidence for both repairs is under `_out/failed-gates/` in the follow-up tree.
  Preserve the failed link attempt caused by disk exhaustion. Nine confirmed
  inactive, task-owned compilation archives were cleaned; reports and source
  were retained. Do not interpret resource failure or a focused rerun as a full
  suite pass.

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
- CI's UTF-16 example exposed another consumer gap: exact literal/CPP result
  comparison refused bound integer constant expansion. Ten actual Go cases
  across both x86 architectures independently reproduced it. The consumer now
  accepts the closed integer value from the original CPP replay, preserving
  every other operand, opcode, label, address and result-byte obligation. Seven
  actual negative source/result cases and operand metadata mutations remain
  rejected. Current-Go full alias regressions and focused Go 1.20 pass; this
  repairs diagnostic consumption, not an external-library or runtime pass.
- CI's historical crypto mirror exposed an implicit-module-language gap.
  Actual Go accepts an original go.mod without a go line: its main query uses
  1.16, while dependency metadata keeps an empty field. Independent primary/
  dependency and three-target REDs reproduce the old blanket rejection. The
  shared role comparison now retains that distinction, without changing source
  files, declaration/ZIP hashes or revocation. Eighteen three-API LLVM 22 O0/O2
  objects, current-Go source regressions and actual Go 1.20 checks pass. This
  source-role repair is not an original-library pass. Frozen `3145792e` Linux
  first-party controls pass, but the original mirror still fails complete
  assembly TEXT/frame observation. Its terminal report and clean tool stamps
  are retained; no ledger status changed.
- Shard 5 exposed an inert raw-NOP field being sent to the physical memory
  handoff. Independent original-Go REDs cover both x86 modes, the complete
  ModRM/SIB axis, prefixes, five targets and all three APIs. NOP now retains
  its actual byte length without acquiring a memory or PC-address effect.
  The pinned decoder's absent REX.W NOP row has a bounded opcode-specific
  adapter. Invalid prefixes, reserved extensions, interior branch targets and
  real-memory handoff mismatches remain rejected. The 614 encoding witnesses,
  30 LLVM 22 objects, portable state oracle and related regressions pass:
  33 current-Go roots, seven Go 1.20 roots, vet and the unchanged five-target
  official classification gate. Rosetta's original-register
  NOP SIGILL is retained as a failed diagnostic; required native Linux
  AMD64/386 roots, not Rosetta, own the original execution gate. Original
  external replay and current-head CI remain required.
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
- Shard 5's BloomFilter result proof exposed missing scalar bit-count effects,
  not a missing LLVM instruction. Independent original-Go REDs cover all nine
  POPCNT/LZCNT/TZCNT W/L/Q forms and their defined FLAGS. Shape, lowering and
  byte effects now share one typed family, reducing production code by 25
  lines. Unknown numerical/pointer/AUTO/native inputs and undefined FLAGS
  remain rejected. Forty-five current-Go result/family regression roots and
  fifteen Go 1.20 roots pass, including 30 five-target three-API objects and
  existing runtime checks. CLI opcode extraction also passes. The discovering
  external candidate remains failed until its complete current-source replay.
- CI policy's old shared-batch-cache fixture is updated for the refill queue:
  caller cache isolation, exhaustive priority order, sticky shard failures and
  a mandatory final aggregate. All 36 Node tests and the three actual-child
  Go queue tests pass; the original three policy failures remain recorded.
- Live aggregate logs did not distinguish tool startup, scan-ledger validation
  and report auditing. The script now logs each phase's start, finish, elapsed
  seconds and actual exit on stderr; canonical paths and JSON stdout remain
  untouched. Independent RED/GREEN fixtures retain failed scans, partial/invalid
  reports, refusal to overwrite and exact cleanup/exit behavior. All 41 Node
  scheduling tests, root CI policy, strict entrypoint tests and shell syntax pass.
- The frozen full unit run found seven failed roots: two stale RNG assertions,
  the earlier WORD preflight's unresolved-constant diagnostic, and four raw-pool
  object fixtures. RNG tests now account separately for the two side-effecting
  intrinsics, while unknown GP inputs still reject. WORD retains the original
  unresolved-symbol diagnostic before decoding. Native preflight now consumes
  only the existing bounded appended-pool proof, not arbitrary dead WORDs as
  data, using a private source view. Independent REDs retain code/data and
  six-family target boundaries, original Go acceptance, four-API three-target
  objects and mutation negatives. All 76 related current-Go roots pass
  (88.734s), as do ten Go 1.20 roots (11.219s). The failed broad attempt that
  treated dead source as data remains recorded and was corrected. These
  focused passes do not relabel the still-running frozen full suite.
- Shard 38's phase journal reports 30–41-minute ordinary-planning phases and
  a 63-minute capture timeout. A first-party profile identifies repeated pure
  condition-free lexical scans outside the existing parser index. The independent
  RED measured 2,460 warm-query allocations; the bounded, complete-text/filename
  lexical index reduces this to 10 without caching source/graph validity. All
  CPU/source pairs, mutation checks, generated-header roles and cancellations
  remain. Twenty-two related roots pass, along with focused race and actual
  Go 1.20 checks. A single-iteration local 8-ASM/4-Go benchmark changed from
  217 ms / 476 MB allocated to 43 ms / 67 MB. This is not a measured CI or
  original-library improvement; the next frozen run must establish that.
- Shard 38 is terminal with 35 passes, 41 failures and 18 source N/A outcomes
  out of 94 versions, at the still-published old revision. Its hex decoder
  exposed the six-name move-mask family's missing result effects. The complete
  Go rows, 672 actual format witnesses, numerical byte/FLAGS negatives and
  five-target three-API LLVM 22 O0/O2 objects now pass. Isolated cases also
  reproduced legacy 386 high-X overacceptance and an LLVM 22 abort caused by
  VPMOVMSKB borrowing SSE2 from sibling instructions. Shape, lowering, feature
  inference and result observation now share the grammar. Forty-five related
  current-Go roots pass (36.612s), with a separate instruction-local feature
  and helper gate; thirteen actual Go 1.20 roots pass (10.591s). CLI extraction,
  vet and unchanged official classification pass. Original external replay and
  current-head CI still own closure, not these component checks.
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
- Detached `191f2cab` (terminal; archive-bound and CMP/TEST repairs):
  `_out/origin-archive/linux-replay/` retains the Linux first-party passes
  and two original failures. cbysal/go-ethereum v1.17.5 passed its former ZIP
  bound but failed mixed TEXT/DATA frame observation. enceve/crypto passed its
  former CMP FP alias refusal but failed PADDQ's native SP storage contract.
  Neither is a pass. All consumers closed and owned candidate caches were
  cleaned. The later BROADCAST and CPP constant repairs are not in this snapshot.
- Detached `2ec32e65` (original-library replays terminal):
  `_out/cpp-constant-replay/` has two failures: GoETW passed literal/CPP constant
  comparison but fails a native MOVOU input read; parquet passed the broadcast
  width gate but fails a native FP span including padding. The separate
  `_out/compare-library-replay/` BLAS failure is now at floating result proof,
  not CMP/TEST alias widths. No passes were promoted; owned caches were cleaned.
  `_out/driver-exhaustive-2ec/` is terminal: the two-worker complete 704-root
  unit rerun passed every root and all 2,826 subtests/seeds, with zero skips.
  This proves frozen `2ec32e65`, not later source changes. The integration tree
  retains its prior 45-minute full-package timeout
  in `_out/ci-closure/cpp-constants/full-driver.log`; it was not a pass.
- `codex/pr40-refill-sample-20261010`, `02c9ebbb` (terminal):
  `_out/refill-sample/` retains the complete fixed 53-candidate, 37-shard sample.
  Every consumer closed and owned caches were cleaned. Its historical partial
  shard reports cannot update the current-head ledger. This idle worktree may
  now be reused only after preserving its artifacts and checking the new owner.
- Detached `3145792e` in the reused sample tree (terminal):
  `_out/default-language-replay/` retains passing Linux source/object controls
  and one failed `carte7000/crypto` replay. The missing go-directive refusal
  is gone, but complete original TEXT/frame observation still fails. The
  disposable validator closed and its owned cache was cleaned.
- Detached `4c90703f` in that reused tree (terminal):
  `_out/module-exhaustive-4c90703f/` is a complete main-module unit run with
  actual package/root inventories, 32 root partitions and two workers. Do not
  reinterpret its failed source snapshot. Host-inapplicable runtime roots retain
  explicit skip counters and still require their cross-runtime counterparts;
  this is not a native Linux, nested CLI, external corpus or current-head pass.
  The later scalar bit-count change is not part of this frozen run.
  Its root package is now terminal: 5,706 passed, seven failed, 39 skipped and
  zero incomplete out of 5,752 roots. All seven failed roots pass the current
  focused rerun; the original report remains failed. The full module run is
  terminal with 19/20 packages passed and only those seven root failures.
  Both nested CLI suites separately passed in the integration tree.
- Frozen `aedfac9b` in the integration tree (terminal failure):
  `_out/root-exhaustive-aedfac9b/` attempted a 32-partition/two-worker root run.
  Preserve that tree for published CI. Partition 12 failed the fixture above;
  partitions 15/16 failed during disk exhaustion, including the compressed
  disp8 source oracle and two fixed-local frame roots. Those three roots now
  separately pass (77.226s). Another disk exhaustion interrupted partition 26
  and truncated the aggregate report; partitions 27–32 were not reached. Its
  source-stack scalar-move and missing architecture-group terminals separately
  pass (1.446s). Neither focused retry repairs the failed full report. Twenty-four
  additional confirmed inactive integration-build archives were cleaned, not
  source or reports. Subsequent user cleanup restored ample headroom. Do not
  clean unrelated caches or interfere with other builds.
- Frozen `b5af3a10` in `codex/pr40-failed-gates-20261010` (running):
  `_out/root-exhaustive-b5af3a10/` is the fresh 32-partition/two-worker root
  rerun. The separate `_out/failed-gates/cpp-b5af3a10-full-retry.jsonl` is now
  terminal: all 104 CPP-prefix roots passed. These are first-party tests, not
  local index scans or external-library compilation. New broadcast and native
  extension repairs are not in this frozen snapshot. Preserve raw reports and
  inspect the root runner's actual terminal outcome before claiming completion.

The older whole-module run at `044bb668` failed: 18/20 packages passed, root
had one obsolete assertion failure and the corpus driver had three fixture
timeouts. Those regressions have newer repairs, not a retroactive pass.

## Ordered next gates

1. Let published `aedfac9b` finish its priority matrix; do not interrupt it.
2. Finish the package/source binding repairs, retaining the original examples
   as failed until each complete replay succeeds. Commit locally in batches.
3. Compare new CI failures with the original eight classes. Fix each class with
   independent RED/GREEN tests and original-example replay; batch the next push.
4. After priority success, let all remaining current-head gates run. Publish
   only matching, complete reports through the automatic ledger writer.

Known later gates remain real failures/blockers: the fresh strict SDK report at
`9f9c1225` still has 163/184 passes, with frame/entry/ABIInternal/TLS/global/alignment gaps;
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
