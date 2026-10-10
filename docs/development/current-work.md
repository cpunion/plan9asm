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
through `6b448609`. Packed x86 conversions, x86/ARM scalar comparisons,
Wasm float comparisons/minmax and complete ordinary Go-partner tag scopes
join the preceding repairs. ARM32's expanded runtime matrix, ARM64 conditional
comparisons and the original observer's SDK macro/include roles pass their
component gates. Scalar ARM64 min/max has a separate independent TDD owner.
Current official classification passes; see below.
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
- All 24 x86 scalar float-to-integer names share one grammar for the 32
  independent Go encoder rows, native status/effects and physical 386 WIG
  widths. Native indefinite values, exceptions, rounding/SAE and dead memory
  effects replace pure casts. Original Go and three LLVM APIs agree over 192
  constant and 3,584 dynamic legacy rows with four optimization combinations.
  Five-target objects cover the AVX/EVEX forms, but native AVX512, 386 runtime
  and unmasked traps still require their explicit Linux owners. Rosetta is not
  that hardware gate. The component's 26 family roots, 62 neighboring roots,
  Go 1.20 and vet passed before integration as `afab61cb`.
- Ordinary metadata protocol v2 now requires its own transitive dependency
  source proof: exact actual-Go imports, explicit embed selection, source
  inventories and versioned module/ZIP/go.mod ownership. Test-role permissions
  are not borrowed. Independent readers and final-object guards remain strict.
  Exact Go-only tag queries retain complete same-package ASM registration;
  planning-only queries retain private source guards through cached replay and
  later phases. Producer and consumer ancestor aliases remain bound to their
  actual live origins without weakening source-internal symlink rejection.
  The default-environment merged corpus gate at `8481dae4` passes all 57 roots;
  its complete nested CLI also passes all 251 roots in 983.672 seconds, with
  default temporary paths and vet. The fresh external repeat fails separately.
- All sixteen Go-named ARM64 scalar integer conversions and 84 registered raw
  forms share native FPCR/FPSR effects, W zeroing and precise scalar definedness.
  The complete family includes six SIMD integer-to-float widths, including H/H.
  Current and historical Go each pass 124 affected roots; original Go/QEMU and
  four LLVM optimization axes agree over 87,040 exceptional, 3,200 status and
  1,712 bank/ZR observations. Another 352 rows independently prove actual NEP
  merge behavior, not merely a writable control bit. The bank oracle now uses
  seven exact partitions: the longest LLVM stage drops from 189.04 to 30.78
  seconds, with all rows retained. Whole-container peak includes compilation
  (2.992 GiB); runtime sampling peaks at 1.055 GiB and no OOM occurs. This does
  not certify arranged/fixed-point conversions, all AFP behavior or trap handlers.
- Sealed private frames transport opaque pointer/slice components and canonical
  zero/one bool bytes independently of numerical definedness. Complete SETcc,
  both CMOV alternatives, overlap revocation, byte-alias and eight-residue
  checks remain mandatory. Original ABI0 byte/canary observers and LLVM i1-to-i32
  callers pass; no C bool ABI, pointee extent, caller-buffer, loop or DATA grant
  follows. The merged transport gate at `f3044806` passes all eight roots.
- Partial discovery reports retain the eight-candidate cadence and also publish
  after a completed candidate when a minute has elapsed. In-progress candidates
  never gain credit. Full provenance checks and atomic writes remain mandatory;
  36 focused current-Go, race and historical-Go roots pass.
- Numerical frame reentry tests now assert that their callback actually fires
  and follow the public Verify route through role-specific helpers. The RED
  showed that the old immediate-caller fixture missed the callback; `814e46d2`
  changes only the fixture, not source-lifetime permissions. Focused ordinary
  tests passed in 87.311 seconds; all three lifecycle roots passed race in
  294.570 seconds, including cancellation and copied-owner rejection.
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
- The remaining 38 packed conversion names plus four existing names share one
  native typed emitter, checked against 148 independent Go encoder rows. The
  legacy runtime matrix retains 32,256 original-Go observations, three LLVM
  APIs and four optimization axes. Five-target objects and bounded native
  fixture compilation are not AVX512 execution. The production rewrite removes
  284 net lines instead of adding per-spelling fallback implementations.
- Eight x86 COMI/UCOMI names and twelve Go encoder rows now retain native MXCSR
  effects and exactly two read-only scalar operands. Independent review fixes
  Context-before-malformed-source ordering and the direct helper target guard.
  Raw VEX.L=1 has no portable physical semantics and cannot be canonicalized
  to L=0. All four native Linux owners have their actual fixture programs
  compiled, but host/Rosetta does not certify their AVX512/trap execution.
- ARM64 scalar comparisons retain native FPCR/FPSR/NZCV, including signaling,
  zero/register operands and raw H/S/D widths. The original-Go runtime matrix
  covers 165,888 observations; target preflight retains malformed/dead errors.
- Four named ARM64 FCCMP/E operations and raw H/S/D forms now gate the native
  FP comparison itself, not merely its result flags. The independent RED
  observes an erroneous Invalid exception on a false predicate. Complete Go
  rows, 5,128 actual words, 7,680 raw witnesses, full flag/byte/CFG effects and
  three-API three-target objects pass. Native Darwin and Linux/QEMU each pass
  655,360 original-Go whole-state rows against all three LLVM APIs and four
  optimization axes. No trap-handler claim follows. Current Go and Go 1.20
  each pass 67 combined plus one supplemental root; race, vet, CLI extraction
  and five-arch classification pass. The fifteen merged form/effect/target
  roots pass in 5.485 seconds. The closed Linux run had no OOM, although cold
  compilation reached its three-GiB limit and caused reclaim events.
- All sixteen Wasm F32/F64 comparisons and min/max use a typed stack grammar.
  Comparisons produce i32. Min/max quiet NaNs through an integer-bit result
  check because LLVM22 O2 can fold minimum(x,x) to a signaling input unchanged.
  Original Go and 24 real Wasm object/runtime axes each pass 25,600 rows; the
  complete 68-root Wasm gate passes without skips. No blockaddress is emitted.
- Go's ARM32 CMPF/CMPD and registered raw VCMP/VCMPE retain native FPSCR
  signaling and cumulative exceptions. Named operations capture status into
  modeled APSR; raw comparisons do not update APSR before VMRS. The original
  24-case Go oracle and four LLVM optimization axes pass on Linux/QEMU with
  no OOM. An independent all-path RED caught raw predicates bypassing the CFG
  proof after skipped writers/calls; the shared predicate-input registry now
  closes that gap without expanding the existing raw closed-pair contract.
  The merged 23-root gate passes. The expanded Linux/QEMU oracle now passes
  13,952 original-Go full-bank/state/condition observations and 167,424 LLVM
  comparisons across 240 linked/executed objects: three LLVM APIs, two IR
  prepasses and two backend optimization levels. Its twenty batches release
  each module/context before continuing. Trap-handler delivery is not tested.
  Typed-Go's separate 548-form, four-target object gate passes; it is not a
  fourth executed API. The required Linux runtime owner has source-selection
  and CI-dispatch regressions. All nine merged consumer roots pass in 8.393
  seconds; the component also passes current Go, Go 1.20 and vet. The closed
  validator had no OOM; its 3.198-GB peak includes cold compilation, not just
  the approximately 186-MB sampled runtime. Owned caches are reclaimed.
- Ordinary header planning defers only exact actual-Go source rejections,
  never metadata/tool/owner failures. Final scope closure includes Go-only
  tags and even syntax-invalid Go partners, whose actual original compiler
  refusals must be retained. Independent readers reject deleted refusals or
  objects. The complete component passes 64 current-Go roots, 59 Go 1.20
  roots and 17 race roots twice. The embed fixture now requires nine actual
  LLVM objects, not the former six; no historical report is relabeled.
  The merged `dd55f82f` gate independently passes all 64 roots in 411.441
  seconds, vet and unchanged-source guards.
- Original AMD64 observations retain actual SDK dependency macro registrations
  as guarded compiler-source inputs, not ordinary target CPP permissions.
  A bounded literal-include closure guards sibling headers, shared fallbacks
  and directory shadowing under Go's fixed package-CWD search. Unknown dynamic
  includes, root escape and changed sources fail. No expanded source, private
  frame or caller permission follows. The complete component passes 31 current
  roots in 878.240 seconds, 27 Go 1.20 roots, fourteen race executions and vet.
  The external crypto recheck still requires a fresh frozen report.

The merged `2ddd94da` instruction gate passes 146 roots on both current Go
(170.817 seconds) and Go 1.20 (161.545 seconds), plus CLI table extraction and
vet. Its official check rejected an AMD64 baseline drift: twenty forms move
from synthetic-stack support to Context. Eight-SDK/five-arch before/after
reports isolate exactly those SP/BP address roles. Independent Go accepts all
38 conversion names with six SP/BP base/index shapes; the three APIs now reject
unbound physical addresses with empty IR. Ordinary GP controls compile all
fifteen five-target/API objects. The reviewed baseline changes only these exact
fields. At `b745f9a1`, current Go and actual Go 1.20 five-arch classification,
nested CLI complete-family extraction and root vet all pass. Context is not
translation/runtime success.

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

- `codex/pr40-full-family-verify-20261010`, `044bb668`:
  `_out/full-family/` is running the twenty-package full module on a separate
  frozen tree. An obsolete ARM64 helper assertion still expects pure fcmp;
  `47db8b1b` fixes that fixture after reproducing the failure, and its scalar
  compare/helper gate passes in 77.128 seconds. Do not edit the running tree
  or relabel this already failed historical full report.
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
  five-architecture classification and vet sequence is terminal separately
  under `_out/typed-native-module-4f8cd6bc/`.
  Its root partitions are terminal: 5,544 pass, one historical reentry-fixture
  failure and 39 host-inapplicable runtime roots. `814e46d2` fixes that fixture
  independently; the other nineteen module packages pass. The historical
  whole run remains failed. Its closed 3.2-GB cache is reclaimed, not reports.
- `codex/pr40-source-tls-verify-20261010`:
  `_out/source-tls-verification/` retains the strict SDK result: 184 attempted,
  163 pass, 21 fail, no skip/N/A/unattempted. Four former passes exposed zero-
  initialized TLS, not a new valid implementation regression. Timing budgets
  pass; correctness does not. Reviewed Go 1.20–1.27 before/after reports isolate
  one MOV TLS form per x86 architecture; `a43d9cc2` updates only those baselines.
  The five-architecture classification gate passed after that review.
  After every old consumer closed, the tree advanced to `afab61cb`.
  `_out/native-closure-afab61cb/` is terminal: selected conversion families
  passed in 21.373 seconds, official five-arch classification and root vet
  passed, but the complete nested CLI failed in 889.685 seconds. Five roots
  share `source inventory directory escapes its bound root` under default
  temporary-path aliases; the independent repair reproduces that environment.
  Its one-GB owned build cache is reclaimed; failed reports remain unchanged.
  The later `f3044806` merged gate passes transport and all thirteen CLI-owner
  roots but exposes the standalone producer's ancestor alias. `aa23a36d` closes
  that remaining role. Only after termination was this tree advanced to
  `8481dae4`; `_out/merged-owners-8481dae4/` retains the passing 57-root corpus
  gate and all 251 nested CLI roots, using default temporary paths. Vet and
  clean/source guards pass; its closed one-GB cache is reclaimed.
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
  owner canceled this old diagnostic batch after four hours: its batch barrier
  left one slow TLS candidate blocking idle workers. The independent offline
  reader validates five passes, ten failures and 52 objects; all other
  candidates remain pending. Exit 137 was explicit cancellation, not OOM or
  success. Reports/binaries remain; the closed five-GB cache is reclaimed.
  Three crypto terminal failures share
  native local-frame/alignment requirements; libc exposes an ARM64 indirect
  callee ABI. Gnark reached its unchanged sixty-minute planning deadline on
  this pre-optimization revision. Its failed journal is retained. The separate
  `codex/pr40-refill-sample-20261010` snapshot will repeat the identical fixed
  selection with four refillable workers and the newly integrated repairs.
  No ledger credit or whole-shard completion follows from logs.
- `codex/pr40-pragma-performance-verify-20261010`, `10e83020`:
  the coroutine repeat is terminal: all 112 eligible profile/object scopes
  pass, with no N/A/skip, after 775-second planning and 8,701 seconds total.
  A separate network-free frozen reader reproduces the progress bytes exactly.
  The generic-root optimization is absent from this historical snapshot;
  its success cannot certify a newer source-owner protocol or faster planning.
  Only after all consumers closed was the tree advanced to `b024aec6` for
  `_out/priority-crypto-b024aec6/`: all three fixed CI crypto candidates fail
  the same original-observation SDK macro-role check. No candidate passes.
  A network-free independent reader reproduces progress bytes; its stopped
  container/cache are closed. A first-party three-OS role RED now reproduces
  this without external source execution.
- `codex/pr40-coroutine-performance-verify-20261010`, `dfa0d559`:
  the independent repeat failed after 95.885-second planning with zero objects.
  Generic preflight now reaches a real `durable` metadata consumer whose
  transitive protobuf dependency uses embed; the ordinary source-owner proof
  does not yet close that role. This early failure is not a speedup. A second
  read-only container reproduced the exact progress bytes; reports remain under
  `_out/coroutine-performance/`, and its closed owned cache was reclaimed.
  A later clean `c0fc9e1e` repeat is separately terminal under
  `_out/ordinary-closure-c0fc9e1e/`: 141.896-second planning then a `durable`
  metadata failure, `noncanonical complete package assembly registration`.
  It has zero objects, not a performance pass. A network-free frozen reader
  independently reproduced the progress bytes. Its container/cache are closed.
  The fresh `8481dae4` retry is terminal under
  `_out/ordinary-closure-8481dae4/`, with the same exact version, 2 CPU/3 GiB
  executor and sixty-minute phase bounds: actual Go rejects linux/386 durable
  source with undefined sizeOfG. Header planning aborts with zero objects.
  Independent replay passes; its container/cache are closed. The first-party
  repair must preserve every final tag scope and actual source rejection,
  including tags present only on Go partners; neither is assembly success.
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
2. Recheck the fixed external failures on a fresh frozen revision with the
   closed header/tag denominator and SDK dependency macro/include repairs.
3. Finish the scalar ARM64 min/max family, retaining native NaN/status and
   actual NEP effects. Its independent matrix also exposed virtual/physical
   NZCV transport through MRS; a separate owner must close that complete
   MRS/MSR family before claiming it solved. ARM32 full-state and ARM64
   conditional compares are terminal and passing, not still-running gates.
   Eleven x86 Linux native AVX512 runtime roots remain unexecuted: the current
   local executor and QEMU lack the required hardware. A real compatible
   executor is requested; object success cannot replace those gates.
4. Finish the fixed broad sample, recheck repaired classes on a fresh snapshot,
   and rerun integrated compatibility/race/SDK/native gates before publishing
   the fork batch. Do not combine revisions into a ledger.
5. DATA/caller joins, loops, split prologues and private helpers remain distinct
   requirements for unchanged crypto sources; opaque slice/bool transport does
   not grant them or establish an external crypto pass.

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
