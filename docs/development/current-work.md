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
through `741ccef4`. Independent owners have complete scalar CMP/TEST and SDK
literal/numerical joins in progress. Root alone publishes the assembly ledger.
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

At `6d356caa`, combined family tests passed in 83.066 seconds. Block-native
reads at `b1dbf01d` passed their 13.465-second complete gate and Go 1.20 gate.
Aligned packed moves at `2b280b82` passed 45-root component, 39-root legacy/raw
and Go 1.20 gates before integration. Logs stay in each owner's ignored
`_out/`. These are finite scopes: portable/Rosetta tests and compiled native
oracles do not replace required Linux/amd64 native AVX512 runtime roots.

## Frozen verification

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
  across 18 shards before outcomes. Independent final readback confirms
  12 passes, two source N/A, four failures, 129 ordinary plus seven
  instrumentation translations. Sourcegraph was interrupted by low disk and
  remains incomplete; gd and Beeketing/snappy were not started. No whole shard
  is complete. Preserve original failed/aborted journals and final-set reports.
  Math32 includes actual five-architecture objects, not merely planned targets.
- `codex/pr40-header-preflight-reuse-20261010`, `a95b9713`:
  `_out/planning-performance-validation-20261010/` reruns the two planning
  failures with unchanged sixty-minute candidate bounds. Purego completed
  planning but failed actual ARM64 call/native-scratch ABI contracts.
  Inspect coroutine's final owner/report; a retained resource-interruption
  note disqualifies clean timing comparisons.

The official five-arch classification gate passed at `6d356caa` with zero
unsupported/parse errors. Context forms are not semantic/ABI/runtime passes.
Remaining strict SDK failures involve original frames/entry state, ABIInternal
transport, numerical globals and alignment; unbound storage is not a fix.

## Next steps

1. Finish/review complete CMP/TEST formats, physical byte views, immediate
   semantics and dead-read tests. Correct the independent byte-budget fixture
   for the wider vector domain without increasing production limits.
2. Join existing live SDK read-only DATA initializer ownership to exact
   source/CPP/native numerical reads. Original SDK CRC needs DATA/header/
   namespace and EVEX-boundary closure; bypassing guards cannot provide it.
3. Collect frozen results, resume unfinished fixed samples on a clean bounded
   executor, and rerun integrated, nested CLI, compatibility, race, SDK and
   native/cross-runtime gates before publishing the fork batch.

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

Disk exhaustion interrupted the broad sample. Preserve all reports, selections
and tool/source hashes. Clean only exact owned derived caches after every
consumer closes, never shared/live caches. Compile one test binary for grouped
gates and avoid concurrent large cold builds.
