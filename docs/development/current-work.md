# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Inspect actual clean
revisions, frozen tools and complete raw reports before accepting results.
Component tests and namespace classification are not current-head CI.

## Contribution and owners

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected head before fork current-head CI/review
pass. Never push to, approve, merge or close upstream PR 40.

Published `c09eff1e`, completed run `37711752851`, remains failed: four priority
shards failed and two reached deadlines. Downstream skips and failed completion
aggregates are not passes. Batch verified repairs before pushing.

Development is on `codex/pr40-result-effects-followup-20261010`, integrated
through `1b09e349`. Native scalar/NEON ARM conversion status, packed private-
frame proofs and the refillable external-shard queue join the preceding fixes.
Root alone publishes the assembly ledger.
Never edit, rebase or import into running snapshots.

## Integrated systematic repairs

The batch extends earlier native whole128 moves, aligned DATA, complete CLMUL,
packed shifts, bounded numerical DATA reads and pure-GP scratch propagation:

- Full X/Y/Z byte definedness, wider-alias preserve/zero rules and CFG joins.
  Low-X initialization cannot establish unknown wider bytes.
- Complete mode32 static-SB address roles: Go ignores base annotations but
  retains indexes. No frame, bounds or pointer authority follows.
- Twenty non-scalar INSERT/EXTRACT names: all chunk, immediate, alias and mask
  axes. Masked extraction requires independent read and store authority.
- Complete VPTERNLOGD/Q forms and all 256 alias-aware truth tables. Dead
  outputs never erase required memory effects.
- Twelve packed-move result families. Seven unaligned forms preserve full
  volatile or guarded masked-lane reads. Five aligned forms use the original
  side-effecting opcode to preserve native alignment and fault behavior.
- Private conditional availability paired with full block-insertion reads,
  including zero-mask/dead E6NF tuples. Public observers and raw/frame/store
  contracts remain separate.
- Bounded negative header-preflight reuse with repeated source/tool guards,
  mutation invalidation and cancellation. Cached queries are not test credit.
- Complete scalar CMP/TEST grammar, native-width/sign-extension rules and
  observable dead reads; source, conditional effects and lowering agree.
- DATA floating leaves retain unary signs/parentheses and encode their actual
  four/eight-byte Go width. Invalid numeric syntax cannot become zero/address.
- Complete native vector byte boundaries reuse typed family decoders, including
  non-temporal moves. Physical RIP observation is not a source/memory grant;
  unknown/truncated encodings and calls discard the whole observation.
- Independent ordinary header-property replay retains original imported sources
  and complete definitions. The unnecessary-header-query regression drops from
  32 queries to one; actual eligible profiles still need their own objects.
- PCALIGN/PCALIGNMAX, PCDATA and FUNCDATA retain their actual typed operand and
  original-source roles. Invalid or dead directives cannot bypass validation.
- SDK CRC joins the live literal-data owner with the exact numerical source,
  original native PC and CPP consumer. Three APIs and four optimization settings
  compiled twelve objects from unchanged SDK assembly; actual default/explicit
  CLI environments also pass. Both owners remain guarded through emission.
- SDK lifecycle race fixtures keep all callback/API axes but give each original
  role its own existing five-minute stage budget. The old whole-fixture timeout
  remains preserved; all six finite race roots now pass without extending it.
- Signed ARM64 wrapper loads share the bounded byte-carrier contract. Original
  Go and three LLVM APIs at O0/O2 passed Darwin and Linux/QEMU checks over 64
  signed/unsigned cases, including native-tail and neighboring-byte canaries.
- Actual Go TLS CMP/TEST and MOV roles no longer turn an unbound TLS register
  into zero. They fail with context before lowering, including dead paths.
  This closes false passes; it does not implement a native TLS bridge.
- Complete BT/BTC/BTR/BTS W/L/Q grammar supplies ordinary and LOCK effects.
  Every imm8, Go optab row, signed memory index, byte boundary and dead native
  read/write has an independent witness. Dynamic virtual FP remains context.
  The finite gate passed 95 current-Go roots and Go 1.20 compatibility; native
  Linux execution remains a required counterpart to the host/Rosetta gate.
- ARM conversion rows 84–87 now model implicit F15 scratch and named single
  writes preserve the other physical D-register half. Native VCVT preserves
  FPSCR rounding and cumulative exceptions, including dead results and invalid
  conversions that pure LLVM casts made poison. Separate word/float endpoint
  types admit raw S-word/F64 combinations without admitting named S source.
  Explicit native ARM/Thumb targets are required; invalid mixed/dead raw source
  remains an error before target context. Linux/QEMU compares 512 physical
  forms and 15,680 dynamic status rows with original Go over four optimization
  axes, alongside the earlier 255 register-effect cases. Trap-enabled handlers
  and other floating arithmetic are not certified by this finite gate.
- All eight classic ARM NEON integer/float32 conversions retain one native D/Q
  opcode. Their fixed standard controls differ from scalar VFP, while cumulative
  exceptions remain observable even with dead outputs. Original Go/QEMU and
  four LLVM optimization combinations agree over 5,504 full-register/status/
  canary rows, including every D/Q bank and alias, exceptional inputs and
  ambient rounding/FZ/DN changes. Current-Go and Go 1.20 each pass 87 affected
  roots with unchanged official ARM classification fingerprints. This does not
  certify ARM64, fixed-point, FP16 or unrelated floating operators.
- Native header planning admits six ordinary function directives and omits
  uninstantiated generic property roots, matching actual compiler headers.
  Concrete instances and foreign dependencies still require their complete
  properties. The generic regression drops 32 unnecessary CPU proposals to
  zero; eleven actual object consumers and evidence-mutation checks pass.
- The downloaded fixed-local CLI accepts its Darwin macOS triple, still bound
  to the exact source/target identity. This does not grant arbitrary local
  frames or prove that the three external crypto examples now pass.
- CPP planning reuses only complete filename/text parses within one invocation.
  Public projections and complete graph relations are rechecked on every use.
  An independent 8-ASM/4-Go/32-CPU model improves from 1.214 to 0.235 seconds
  and 2.234 to 0.476 GB allocated. Independent plan shells can share immutable
  source graphs; a whole shared mutable plan is not a concurrent API.
- Private numeric/call-free frames reuse all twelve packed and seven legacy
  move grammars, KMOV definedness, BP private aliases and eight mod64 residues.
  Independent byte-alias replay prevents relocation from changing overlaps.
  Masked stores revoke overlapping pointer spills without initializing unknown
  retained bytes. No DATA, caller, slice/bool, loop, helper or split-prologue
  permission follows. All 120 related roots passed; native AVX512 execution
  remains separate from the 340-case compile-only oracle and Rosetta SSE gate.
- External `all` scheduling refills vacant workers instead of waiting for batch
  neighbors. Each shard closes its own cache. Failures remain sticky and the
  final aggregate receives the actual shard count, not an implicit 64. Actual
  shell-entry tests passed current-Go repeated/race, Go 1.20 and Linux gates.

At `6d356caa`, combined family tests passed in 83.066 seconds. Block-native
reads at `b1dbf01d` passed their 13.465-second complete gate and Go 1.20 gate.
Aligned packed moves at `2b280b82` passed 45-root component, 39-root legacy/raw
and Go 1.20 gates before integration. Logs stay in each owner's ignored
`_out/`. These are finite scopes: portable/Rosetta tests and compiled native
oracles do not replace required Linux/amd64 native AVX512 runtime roots.

At `0ada7f5d`, the combined instruction gate passed 80 roots/627 nodes without
skips. DATA/parser/budget/source-binding tests passed 65 roots/607 nodes.
Go 1.20 independently passed 17 roots/223 nodes for CMP/TEST, DATA and the full
vector byte-work boundary. The boundary component passed 85 current-Go roots
and eight Go 1.20 roots, including 2,544 original Go forms, 64 non-temporal
forms, all 1,280 selected immediate witnesses and every instruction-byte cut.

## Frozen verification

- `codex/pr40-family-matrix-verify-20261010`, `71e509f1`:
  `_out/family-matrix-module/` completed the twenty-package inventory: nineteen
  packages passed. Root has 5,460 passed, four failed and 39 host-inapplicable
  runtime roots. Those four historical failures are
  fixed in the integration tree: needless DATA resolver calls, native-aligned
  conformance axis/alignment, obsolete dead-wide-vector rejection, and the old
  ARM64 DATA diagnostic. Six focused roots passed on a newer frozen revision;
  preserve the failed full report rather than relabeling it. Its owned cache
  is closed and reclaimed. Nested CLIs remain separate.
- `codex/pr40-bt-vfp-integration-verify-20261010`:
  the initial `5348b5a8` build failed because BT referenced the renamed TLS
  helper. Preserve `_out/bt-vfp-integration/build.log`; `6f5e7902` fixes that
  one reference. Its separate `_out/bt-vfp-integration-6f5e7902/` gate completed:
  35 instruction roots and 48 header/scanner roots passed without skips,
  official five-architecture classification and vet passed. Only after that
  run ended was the tree advanced to `dfa0d559`: separate
  `_out/planning-index-dfa0d559/` focused, actual-consumer, race and vet gates
  also passed. Historical directories keep their exact source revision.
  The later `2eaaeb5c` Linux queue gate passed six roots and cleaned only its
  own cache. After it closed, the tree advanced to `4f8cd6bc`; the full module,
  five-architecture classification and vet sequence is now running separately
  under `_out/typed-native-module-4f8cd6bc/`. Never advance this active snapshot.
- `codex/pr40-source-tls-verify-20261010`, `600b08c7`:
  `_out/source-tls-verification/` retains the strict SDK result: 184 attempted,
  163 pass, 21 fail, no skip/N/A/unattempted. Four former passes exposed zero-
  initialized TLS, not a new valid implementation regression. Timing budgets
  pass; correctness does not. Reviewed Go 1.20–1.27 before/after reports isolate
  one MOV TLS form per x86 architecture; `a43d9cc2` updates only those baselines.
  The five-architecture classification gate passed after that review.
- `codex/pr40-sdk-numerical-cli-20261010`, `a5d9396b`:
  the complete nested CLI gate passed all 239 roots, without skips. The later
  downloaded Darwin/mirror change passed its four actual object consumers and
  identity-rejection controls independently; no combined current-head claim.
- `codex/pr40-sdk-directive-verify-20261010`, `7ecf2dde`:
  `_out/sdk-directive-verification/` retains passing five-architecture official
  classification and the six-root failed-case/guard recheck. The strict SDK
  benchmark still has seventeen failures: original AMD64 CRC now passes, while
  386 runtime TLS addressing exposes another previously unproved native role.
  Timing budgets pass; correctness does not. The native AVX512 execution gate
  has not been replaced by host object or portable execution tests.
- `codex/pr40-result-effects-verify-20261010`, `0529fa9c`:
  `_out/result-effects-module/reports/module.json` completed: 19/20 packages
  passed, root 31/32 partitions passed. The sole root failure was the obsolete
  MOVAPD assertion fixed by `5a438791`. Preserve that failed historical report.
- `codex/pr40-wide-result-verify-20261010`, `6d356caa`:
  `_out/wide-result-verification-20261010/reports/shard-4.json` proves
  `github.com/klauspost/crc32@v1.3.0` passed eleven-target planning, two
  applicable ASM files and six LLVM objects (three OSes times amd64/arm64).
  It does not prove runtime or other shard candidates. Strict SDK benchmark
  attempted 184 files: 167 pass, 17 fail, zero skip/N/A/unattempted. Time
  budgets passed but correctness did not.
- `codex/pr40-cpp-register-matrix-20261009`, `3e82acad`:
  `_out/broad-external-sample-20261009/selection.json` fixes 21 exact versions
  across 18 shards before outcomes. After separate frozen-source retries, all
  21 candidates are terminal: fourteen passes, two source N/A and five failures,
  with 138 ordinary plus seven instrumentation translations. Original aborted
  sourcegraph and unstarted scopes remain preserved; the separate retry passes.
  Beeketing/snappy passes, while gd fails the native indirect-call ABI contract.
  `audit-resumed.json` checks the exact fixed selection and terminal journals;
  independent frozen readers also validate the retry reports. No whole shard
  is complete. Preserve original failed/aborted journals and final-set reports.
  Math32 includes actual five-architecture objects, not merely planned targets.
- `codex/pr40-header-preflight-reuse-20261010`, `a95b9713`:
  `_out/planning-performance-validation-20261010/` reruns the two planning
  failures with unchanged sixty-minute candidate bounds. Purego completed
  planning but failed actual ARM64 call/native-scratch ABI contracts.
  Coroutine completed all 112 eligible profile/object scopes and passed; the
  independent progress reader validates both terminal candidate reports.
  The candidate timeout currently bounds each phase, not total candidate time.
  Inspect its final owner/report; a retained resource-interruption note
  disqualifies clean timing comparisons.
- `codex/pr40-wide-sample-verify-20261010`, `961968b1`:
  `_out/wide-sample/selection.json` fixes fifty new stratified exact versions
  and three published-CI examples, before outcomes, across 37 shards. The
  isolated bounded batch remains active. Three crypto terminal failures share
  native local-frame/alignment requirements; libc exposes an ARM64 indirect
  callee ABI. Gnark reached its unchanged sixty-minute planning deadline on
  this pre-optimization revision. Its failed journal is retained while later
  shards continue. No ledger credit or whole-shard completion follows from logs.
- `codex/pr40-pragma-performance-verify-20261010`, `10e83020`:
  the coroutine repeat remains active after 775-second planning. Read-only
  diagnosis found an independent generic-root fallback, fixed later by
  `d3b0d076`; this running snapshot does not contain that fix. Never change its
  source or retroactively credit it with the improved planning behavior.
- `codex/pr40-coroutine-performance-verify-20261010`, `dfa0d559`:
  the independent repeat failed after 95.885-second planning with zero objects.
  Generic preflight now reaches a real `durable` metadata consumer whose
  transitive protobuf dependency uses embed; the ordinary source-owner proof
  does not yet close that role. This early failure is not a speedup. A second
  read-only container reproduced the exact progress bytes; reports remain under
  `_out/coroutine-performance/`, and its closed owned cache was reclaimed.
- `codex/pr40-arm-vfp-conversion-status-audit-20261010`, `8d12b43c`:
  `_out/arm-fp-status/handoff.md` indexes original-Go REDs and the final seven
  Linux/QEMU roots, 80 current-Go roots and the same 80 Go 1.20 roots, all with
  no failures/skips. Both actual SDK ARM classification fingerprints remain
  unchanged. Its owned stopped containers/cache are closed and reclaimed.
- `codex/pr40-packed-native-frame-20261010`, `579e2b24`:
  `_out/packed-native-frame/handoff.txt` retains 120-root and eleven-root Go
  1.20 gates, three-OS/three-API 195-function objects, numerical/canary checks,
  byte-alias negatives and native-execution boundaries. Its owned cache is
  reclaimed; this does not close the external crypto composite frame.
- `codex/pr40-arm-neon-conversion-status-20261010`, `e7c66a60`:
  `_out/arm-neon-status/handoff.md` retains the original-Go RED and complete
  Linux ten-root, current-Go 87-root and historical 87-root GREEN gates. The
  original optimized lowering lost dead exceptions and status; the native
  whole-vector repair passes all four optimization combinations. Its owned
  cache and stopped containers are closed. Integration `1b09e349` separately
  passed `_out/neon-merged-1b09e349/test.log` in 7.887 seconds.

The official five-arch classification gate passed again at `a43d9cc2` with zero
unsupported/parse errors. Actual Go 1.20–1.27 before/after source comparisons
isolate exactly eight formerly false-supported aligned SP/BP move forms now
requiring original-frame context, plus the two TLS roles described above.
Encoder inventories/fingerprints are not exhaustive instantiated semantic
matrices. Context is not semantic/ABI/runtime success.
Remaining strict SDK failures involve original frames/entry state, ABIInternal
transport, numerical globals, TLS and alignment; unbound storage is not a fix.

## Next steps

1. Finish the next isolated TDD batches and their merged gates. Preserve actual
   native counterparts; do not equate table names with semantics.
2. Close the ordinary transitive dependency/embed source role with its own
   complete source/ZIP/selection owner, producer and independent consumer.
   Do not borrow test-package permissions or remove the failing embed guard.
3. Finish the x86 scalar float-to-integer family and extend the same systematic
   audit to ARM64 scalar conversion formats and native status. The ARM NEON
   finite family above is complete; do not substitute its standard controls
   for ARM64 FPCR rules. The independent x86 original-Go audit in
   `_out/scalar-float-audit/` already
   finds 32/192 row differences without IR optimization and 184/192 with O2:
   precision flags disappear and poison erases result stores. Preserve this
   RED; object/table coverage alone did not detect it. Exact scope is 48 legacy
   functions, four MXCSR modes and four optimization combinations on Rosetta.
4. Finish the fixed broad sample, recheck repaired classes on a fresh snapshot,
   and rerun integrated compatibility/race/SDK/native gates before publishing
   the fork batch. Do not combine revisions into a ledger.
5. Actual slice/bool transport, DATA/caller joins, loops, split prologues and
   private helpers remain distinct requirements for unchanged crypto sources;
   do not infer them from the newly completed numeric private-frame class.

## Inventory, ledger and resources

Inventory remains on the source-only scanner; compilation/execution uses
separate disposable validators. No local background index scanner. cgo
inventory stays outside this repository.

At `3e82acad`, committed assembly evidence records five candidate passes, one
old CRC failure and 6,061 pending. Neither old passes nor the newer CRC report
may be relabeled current-head evidence. Publish only automatic matching
schema-11 source/scan/tool/file/target/profile evidence; never mix revisions.

Independent staging retains 21,257,856 contiguous index entries, 1,051,454
unique paths, 6,700 assembly versions and 84,225 files after path-plus-major
remerge. This is not imported translation coverage; the committed denominator
remains 6,067. Do not rescan old ranges with obsolete path-only deduplication.
Keep funnel tables in the PR body.

Disk exhaustion interrupted the broad sample. Two historical, closed fixture
caches were reclaimed; their reports, binaries and sources remain intact. The
fixed sample retries and planning-performance batch are now terminal. Clean
only exact owned derived caches after every consumer closes, never shared/live
caches. Compile one test binary for grouped gates and bound cold builds.
