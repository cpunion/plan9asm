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
through `cde997f3`. The live SDK literal/numerical join, actual CLI consumer,
ordinary imported-header property closure and typed Go directives are integrated.
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
  `_out/family-matrix-module/` runs the complete main-module inventory with
  two root-partition workers and an independently owned cache. Inspect terminal
  reports; it was still running at this checkpoint. Four observed failures are
  fixed in the integration tree: needless DATA resolver calls, native-aligned
  conformance axis/alignment, obsolete dead-wide-vector rejection, and the old
  ARM64 DATA diagnostic. Preserve the failed report. Nested CLIs remain separate.
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

The official five-arch classification gate passed again at `7ecf2dde` with zero
unsupported/parse errors. Actual Go 1.20–1.27 before/after source comparisons
isolate exactly eight formerly false-supported aligned SP/BP move forms now
requiring original-frame context. Context is not semantic/ABI/runtime success.
Remaining strict SDK failures involve original frames/entry state, ABIInternal
transport, numerical globals, TLS and alignment; unbound storage is not a fix.

## Next steps

1. Close typed signed ARM64 byte-carrier wrapper reads against actual Go
   objects. The initial signed int8/int16/int32 counterexamples fail while their
   unsigned controls pass. Do not expand the bounded source/caller role merely
   to admit SDK atomic or arbitrary external pointer entries.
2. Classify actual Go TLS CMP/TEST forms with their source-encoder/native ABI
   requirements. A virtual TLS register initialized to zero is not valid TLS
   semantics and must not restore a false benchmark pass.
3. Expand a fixed independent external sample and recheck published CI failure
   examples using one new frozen source. Collect the complete main-module and
   nested CLI results; rerun integrated compatibility/race/SDK/native gates
   before publishing the fork batch. Do not combine revisions into a ledger.

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
