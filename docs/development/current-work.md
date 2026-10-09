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
through `5348b5a8`. The SDK literal/numerical join, signed ARM64 carrier reads,
TLS fail-closed grammar, x86 BT family and ARM VFP register effects are integrated.
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
  writes preserve the other physical D-register half. Linux/QEMU passed 255
  original-Go cases and all four IR/backend optimization combinations. This
  proves the checked register effects, not FPSCR, NaN/overflow or rounding modes.
- Native header planning admits six ordinary function directives and omits
  uninstantiated generic property roots, matching actual compiler headers.
  Concrete instances and foreign dependencies still require their complete
  properties. The generic regression drops 32 unnecessary CPU proposals to
  zero; eleven actual object consumers and evidence-mutation checks pass.
- The downloaded fixed-local CLI accepts its Darwin macOS triple, still bound
  to the exact source/target identity. This does not grant arbitrary local
  frames or prove that the three external crypto examples now pass.

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
- `codex/pr40-bt-vfp-integration-verify-20261010`, `5348b5a8`:
  `_out/bt-vfp-integration/` is running the merged BT/TLS/VFP/carrier roots,
  header/scanner gates, official five-architecture check and vet. Do not edit
  this tree or accept partial logs as final success.
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
  callee ABI. No ledger credit or whole-shard completion follows from live logs.
- `codex/pr40-pragma-performance-verify-20261010`, `10e83020`:
  the coroutine repeat remains active after 775-second planning. Read-only
  diagnosis found an independent generic-root fallback, fixed later by
  `d3b0d076`; this running snapshot does not contain that fix. Never change its
  source or retroactively credit it with the improved planning behavior.

The official five-arch classification gate passed again at `a43d9cc2` with zero
unsupported/parse errors. Actual Go 1.20–1.27 before/after source comparisons
isolate exactly eight formerly false-supported aligned SP/BP move forms now
requiring original-frame context, plus the two TLS roles described above.
Encoder inventories/fingerprints are not exhaustive instantiated semantic
matrices. Context is not semantic/ABI/runtime success.
Remaining strict SDK failures involve original frames/entry state, ABIInternal
transport, numerical globals, TLS and alignment; unbound storage is not a fix.

## Next steps

1. Close the merged finite verification gate. Preserve RED/GREEN and actual
   native counterparts; do not equate table names with semantics.
2. Investigate the shared packed local-frame/alignment failure with a first-
   party fixture and exact original frame/caller contracts. Separately audit
   the remaining ARM conversion status/exception/rounding semantics.
3. Reduce repeated full CPP parsing inside one immutable planning operation,
   using measured first-party performance regressions. Retain all source,
   projection, graph, mutation and cancellation checks; no cross-run trust cache.
4. Finish the fixed broad sample, recheck repaired classes on a fresh snapshot,
   and rerun integrated compatibility/race/SDK/native gates before publishing
   the fork batch. Do not combine revisions into a ledger.

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
