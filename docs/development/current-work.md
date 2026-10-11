# Current work: close the PR 40 CI repair batch

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Tests of a component or
an older revision are not current-head CI or assembly-ledger passes.

## Fixed delivery scope

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch before fork CI and review pass.
Never push to, approve, merge or close upstream PR 40.

The integration branch `codex/pr40-ci-closure-20261011` remains its historical
`aedfac9b` snapshot. The published repair source is
`codex/pr40-ci-failure-replay-20261010` at `0c61d468`. Preserve both and the
separate scalar-FMA development branch. Only original CI failure classes,
not speculative instruction-family audits, belong to the next repair batch.

Fork CI 38043255488 at `aedfac9b` finished naturally with failure, as did its
predecessor 38018079008. Policy preflight passed but all six priority shards
failed; 36/42 again reached the unchanged 330-minute deadline. Their six reports
retain 540 versions: 224 passes, 235 failures, 78 source N/A and three explicit
skips, with 2,679 ordinary translations. The remaining 5,527 versions are not
reported. These partial diagnostic outcomes cannot update the assembly ledger.
Artifacts are in the earlier follow-up tree's
`_out/failed-gates/ci-38043255488/`.

The aggregate completed its audit in 4,778 seconds and failed because six
reports are not 64 complete shards. Downstream module/cross completion gates
also correctly failed for missing owner reports after priority failure prevented
those tests from running; these are not newly observed unit-test failures.
Do not bypass their strict completeness checks. The published bounded CPP
index repair specifically addresses the slow portable-report audit.

The five code repairs plus checkpoint through `0c61d468` were published only
to fork Draft PR 4 after the previous run finished naturally. They include the
deadline fixture, low-product family, portable CPP index, broadcast memory result and native
extension result repairs. The full root run at frozen `b5af3a10` now passed;
new broadcast/extension code has its separate focused evidence below.
Fresh CI 38096927749 is running: policy preflight passed; shards 5/29/38/39
are terminal failures, with 167 passes, 168 failures, 57 source N/A and one
invalid-source skip among their 393 versions. Shards 36/42 remain running.
Many failures still stop at the mixed TEXT/DATA, legacy module-origin and
native result effects repaired only in local unpublished commits. Shard 5's
packdb result observer also refuses VPSUBB; repair the complete packed integer
add/subtract consumer family with independent tests. Let the run finish
naturally before the next batch.
Keep the upstream-connected branch unchanged until fresh fork CI and review
pass. A running CI is not a pass or ledger update.

Local commits `2dcf6471` and `dc3e53d8` close the scalar-insertion and GP/X
result-read families. Their tree `codex/pr40-scalar-lane-native-20261011` is
retains the terminal 32-partition/two-worker full root gate under
`_out/root-exhaustive-dc3e53d8/`: 30 shards passed and two failed. Its original
report remains failed. Focused evidence remains in
`_out/scalar-lane-native/` and `_out/gp-x-native/`.

That full run also reproduced a 45-minute timeout in the monolithic Darwin
ARM64 min/max state owner (partition 27). The repair on
`codex/pr40-minmax-runtime-partitions-20261011` gives each of FMAX, FMIN,
FMAXNM and FMINNM its own native root, matching the existing Linux owners.
The disjoint union still checks all 14,600 original states and twelve LLVM
lanes; no bank, sample, API or optimization axis was removed. The independent
owner/union RED and five structural GREEN roots, including actual Go 1.20,
are retained under `_out/minmax-runtime-partitions/`. The frozen native run at
`1d0ce3a0` passed all four families: FMAX (638 seconds), FMIN (702 seconds),
FMAXNM (603 seconds) and FMINNM (553 seconds). Together they retain all 14,600
original states and 175,200 LLVM comparisons. An integration follow-up also
checks the actual runner's hash
partition: the initial names collided in two shards, so the names now use
`TestARM64ScalarMinMaxNativeRuntime*` and occupy four distinct ordinary shards.
Retain `_out/minmax-shard-spread/` RED/GREEN logs. Structural checks alone are
not runtime passes, and frozen native reports retain their original revision.
Keep the old timeout failed. The redundant full-root attempt at `2fbf0011`
was canceled locally under severe host load and is incomplete, not green;
its separate nested-CLI/internal-package checks both finished successfully
(12 internal packages passed; the header helper has no test files).

Byte-shuffle repair `6430b7f8` is committed locally. Development continues in
`codex/pr40-frame-data-20261011`, with mixed DATA evidence in `_out/frame-data/`.
The frozen `dc3e53d8` full-root run has a real failure in partition 20:
`TestAMD64MovSyscallAndCRC32Coverage` admits a bare MOVLQSX symbol. Commit
`7d4603f9` repairs its shared operand class, retaining that failed full-run
evidence. Do not edit the frozen tree or relabel it. The grammar rejects bare names
before creating LLVM storage; independent Go oracles cover all fourteen members
and retain legal SB/absolute controls. All 22 current-Go focused roots and eight
actual Go 1.20 roots pass, as does vet. The original failed full report stays
failed. The current-Go DATA/adjacent suite completed with 59 passing roots and
zero failures/skips; focused actual Go 1.20, CLI extraction and unchanged official
classification also passed.
Independent ordinary MOV native-result and dead-read REDs are retained in
`_out/scalar-move-native/`. The five B/W/L/Q/D spellings now share their native
register-addressed load shape, exact volatile read and lowering-only byte
effect. Complete Go tables, public/source/CFG boundaries and sparse byte/FLAGS/pointer tests
remain independent. Actual Go REX bytes additionally expose high-byte aliases;
the consumer tracks their real low-byte destination and refuses implicit
physical SP/BP writes. Twenty-one current-Go related roots and nine actual
Go 1.20 roots pass. Fourteen original-Go/LLVM live/dead kernels pass all three
APIs at O0/O2: 1,792 numerical and 168 fault/continuation outcomes per lane.
Rosetta remains a diagnostic, not native Linux execution. Eight final roots
pass, including five-target O0/O2 objects; CLI and unchanged five-architecture
classification also pass. All sixteen final named-address/family regression
roots and final vet pass, retaining Go's distinct SB role. The last actual Go
1.20 family rerun also passes all eight roots. Original exact-version replay
still owns candidate closure.
No external library was compiled locally and no ledger status changed.
Seven CI candidates stop at the original-ZIP/go.mod comparison. The repair
branch `codex/pr40-legacy-module-origin-20261011` shares the existing public
SumDB signature/inclusion implementation with the whole-ZIP observer. An
absent root go.mod now has a distinct authenticated metadata role, never a
copied ZIP member or fabricated main package. Actual canonical cache paths,
full ZIP/tree hashes and permanent mutation revocation remain checked.
Source-only exact aez inspection reproduced the old failure and passed the
new observation/authentication/revocation checks without compiling external
code; logs remain under `_out/legacy-module-origin/`. First-party metadata,
source-origin and consumer-role regressions retain their RED/GREEN evidence.
Legacy primary queries now run from independent exact-version modules, using
only the existing cache's file-proxy index. A separate metadata-preparation
phase accepts registered Go module-resolution progress, then the actual
read-only query retains its original dependency role, implicit Go 1.16 language
and exact original package directory. Both frame factories independently
observe the complete actual source/export graph; the query grants no frame or
numerical permission. Actual aez source-only inspection reproduces both the
cache-CWD and GOPROXY=off failures and passes the cache-only query, without
compiling external code. First-party query and source-origin race tests,
actual Go 1.20 checks and vet pass. All 32 broader current-Go original-source,
LLVM 22 object and numerical roots pass without failures/skips (640 seconds
for the profile package). The independent query-repeat check also preserves
compiler exports across fresh temporary boundaries. Rosetta numerical results
are not native Linux execution. Original external compiler/object replay remains
required: no affected library or assembly-ledger status is promoted here.

The active integration branch is `codex/pr40-frame-data-20261011`, including
the metadata/query repairs through `4b739b56` and min/max shard-spread guard.
Commit the checkpoint before its new bounded complete main-module verification;
retain that exact frozen revision in the output directory. Develop subsequent
CI repairs in an existing idle worktree, not the frozen verification tree.
The historical `aedfac9b` tree's old
`_out/ci-closure/publish-after-ci.sh` already completed publication and is not
a reusable publisher. Its replacement old-run watcher is terminal. The current
read-only watcher is in the published tree at
`_out/ci-failure-replay/watch-0c61d468.log`; it has no publication action.
Retain all terminal logs and earlier disk-exhaustion evidence.

The local cleanup consolidated 1,084 inactive clean worktrees and removed 333
merged local branch refs. Unmerged refs and all reports were preserved; dirty,
current development and running verification trees were not removed. Historical
`_out` paths in this checkpoint may now live under the original worktree basename
in the sibling `plan9asm-cleanup-evidence-20261011` directory. Its recovery journals
retain exact commits and refs; the source checkout can be reconstructed with
`git worktree add`. Only authenticated idle Go build caches were cleaned. Do not
mistake removal of redundant checkouts for an updated test result or ledger.

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

- Shards 29/39's mixed TEXT/DATA refusals now have an independent complete
  current-Go DATA/source/object observation for all three AMD64 OS targets,
  preserving TEXT positions, relocations and Windows unwind identity. This
  alone grants no memory permission. The fixed-local consumer separately binds
  bounded private readonly, relocation-free tables to their actual Go bytes,
  native extent/alignment and immutable LLVM symbol resolution. Source/consumer
  mutation, coalesced definitions, out-of-bounds, writes and address escape
  remain rejected. The complete KMOV/VAES physical-boundary observer shares
  decoding without widening raw RIP translation. The 119 registered literal
  operations passed three-target, three-API LLVM 22 O0/O2 objects; eight original
  Go/LLVM kernels passed 2,048 numerical outcomes per API/optimization lane.
  Rosetta is a legacy SSE diagnostic, not native Linux or AVX512 execution.
  The independent whole-object reader and source/listing mutation controls pass.
  Sixteen actual Go 1.20 roots, vet, CLI extraction and unchanged five-target
  classification pass. All 59 broader current-Go roots pass; original exact-
  version remote replay and current-head CI still own library closure.
- Shard 38's ClarkQAQ/uw hex decoder exposed missing PSHUFB/VPSHUFB result
  effects. The complete family now shares the binary grammar across lowering,
  caller/result effects, native memory tuples and per-function features.
  Independent REDs additionally reproduce removable dead reads, stale upper
  aliases, isolated `.Z` LLVM selection failure, WIG refusal and the conflated
  physical/named mode32 mask limits. Original legacy m128 preserves its native
  alignment/fault opcode; E4NF vector reads retain their entire volatile tuple.
  Unknown bytes, masks, physical high banks and virtual legacy storage remain
  constrained. Complete Go rows, five-target three-API objects, 56 portable
  numerical/alias kernels and 20 live/dead fault kernels pass, including actual
  original legacy Go bytes. Native VEX/EVEX execution is not inferred.
  All 77 current-Go family/adjacent roots pass (190.413s), as do fifteen actual
  Go 1.20 roots, CLI extraction, vet and unchanged official classification.
  Exact-version remote replay and current-head CI still own library closure;
  this local repair is not part of the running published CI.
- Shard 5's go-simd/base32 decode counterexample exposed the same omitted
  lowering-only memory consumer in the complete GP/X transfer family. Actual
  Go accepts the independent REDs, while all five targets and three translator
  APIs initially refuse their native numerical results. MOVL/MOVQ/MOVD and
  VMOVD/VMOVQ now reuse the scalar-insertion continuation check with their own
  shared typed widths, zero fill and legacy/VEX/EVEX alias effects. Public
  declarations, pointers/indexes, source joins, writes, raw origins and frame
  contracts stay strict. The 248 target/form functions have original Go and
  LLVM objects, including all four X128 dwords; each module stays below 64
  functions. Ten live/dead kernels pass original legacy-Go/LLVM numerical and
  fault comparisons through three APIs and O0/O2 (1,280 numerical outcomes,
  120 fault/continuation outcomes per lane). Native VEX/EVEX execution is not
  inferred. All 46 current-Go family/adjacent roots pass (84.697s), including
  the shared scalar insertion runtime rerun; thirteen actual Go 1.20 roots,
  CLI extraction, vet and unchanged official classification pass. Retain
  `_out/gp-x-native/`; the original full library remains failed until remote
  exact-version replay, and this local commit is not in running CI.
- Shards 5/42's exact marcelocantos/hash and frozen versions exposed the
  lowering-only observer's missing native scalar insertion read, not missing
  encodings. Independent REDs reproduce it across all five x86 targets and
  three APIs. The existing shared grammar now consumes the paired exact-width
  volatile load for all eight PINSR/VPINSR names and INSERTPS/VINSERTPS; sparse
  lane definitions and upper aliases stay distinct. Public declaration,
  pointer/index, source-CFG, stores and raw/frame-entry boundaries remain strict.
  The 106 target/form functions have actual Go and LLVM objects. Twenty-two
  live/dead/zero-selector kernels pass three-API LLVM O0/O2 numerical and fault
  comparisons against original legacy Go machine code and an independent byte
  reference. Each runtime lane checks 2,816 numerical outcomes and 264 native
  faults/continuations; Rosetta is not native Linux or VEX/EVEX execution.
  The 71-root current-Go family/regression run passed (174.590s), followed by
  ten final roots including the whole-source negatives and adjacent lowering
  fallback/raw-boundary regressions (54.759s). Thirteen actual Go 1.20 roots,
  CLI extraction, vet and unchanged official five-architecture classification
  passed. Exact discovering-version remote replay and current-head CI still
  own library closure; neither original candidate is promoted to a pass.
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
- The published follow-up batch closes shard 5's low-product caller-effect
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
  actual Go 1.20 check pass. Commit `7579e5d2` is included in `0c61d468` CI.
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
- Frozen `b5af3a10` in `codex/pr40-failed-gates-20261010` (terminal pass):
  `_out/root-exhaustive-b5af3a10/` completed all 32 partitions: 5,741 root passes,
  zero failures, 39 explicit host-inapplicable skips and zero incomplete roots;
  130,423 subtests/seeds passed. Native cross-runtime counterparts remain required.
  The separate `_out/failed-gates/cpp-b5af3a10-full-retry.jsonl` is now
  terminal: all 104 CPP-prefix roots passed. These are first-party tests, not
  local index scans or external-library compilation. New broadcast and native
  extension repairs are not in this frozen snapshot. Preserve raw reports;
  neither result is complete current-head cross-runtime or ecosystem coverage.

The older whole-module run at `044bb668` failed: 18/20 packages passed, root
had one obsolete assertion failure and the corpus driver had three fixture
timeouts. Those regressions have newer repairs, not a retroactive pass.

## Ordered next gates

1. Let published `0c61d468` CI 38096927749 finish naturally. Retain all six
   priority reports and compare exact failures before the next fork-only push.
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
