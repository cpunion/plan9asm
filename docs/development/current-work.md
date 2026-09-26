# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). CI repair has priority:
do not start new module-index inventory scans while failures remain.

## Contribution and frozen worktrees

- PR 40 is OPEN/DRAFT at remote head `4cf5ade`: last live inspection found
  74 successful checks, 19 failed discovery shards and a failed aggregate.
  Run `gh pr view` again for current state. No later local repair was pushed.
- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`, never upstream.
  The user requests complete repairs before a batch push. Keep the PR draft.
- Upstream main `7cc8c0f` was fetched and is already an ancestor.
- Develop on `codex/pr40-arm64-raw-20260926`. Inspect status, worktrees and
  running processes before editing.
- Runners: `codex/pr40-fp16-20260926` (root gates and external replay) and
  `codex/pr40-ecosystem-fixes-20260925` (cross runtime and discovery).
  Inspect each HEAD and logs before advancing; never alter a running snapshot.
  Both are frozen at `692fad1`; the root runner currently also hosts the full
  cross-runtime gate. Discovery shards 10 and 17 are running in the ecosystem
  runner with reports under `_out/ci-repair-692fad1-shard17/`.
- Use Go 1.27 and LLVM 22 only. Root focused compatibility also uses Go 1.20.
  Cross runtime requires checksum-pinned QEMU 10.2.3, not QEMU 8.2.

## Completed gates and recent implementation

- At clean `a48a4fd`, full root passed in 642 seconds and required complete
  cross-runtime tests passed in 479 seconds. Logs are respectively
  `_out/full-root-bounded-vector-pool.log` and
  `_out/cross-runtime-bounded-vector-pool.log` in the runners.
- Clean `692fad1` passed the official five-architecture classification gate
  and strict benchmark: 184/184 files, no N/A, 43 target-seconds.
  Observed form classification is not all-encoder runtime-semantic coverage.
  Full root and cross runtime also passed at that snapshot; logs are
  `_out/full-root-index-pool.log` and `_out/cross-runtime-index-pool.log`.
- Earlier batches cover ordinary SVE loads, floating divide/scale, all
  floating-immediate values, and saturating arithmetic including signed
  byte/halfword operations with unsigned large immediates. Consult Git/tests.
- `6d5bf16`: typed SVE effects, including compact/unary, unrelated memory and
  scalar operands, no longer prevent proving a constant-pool address.
- `422a906`: predicate count and ADD/SUB/carry/negation families model actual
  GP overwrites without treating their address-dependent inputs as safe.
- `186ca03`: in-place 64-bit immediate ADD/SUB tracks a bounded pool offset.
  Scalar, pair, structure and SVE replicate loads check actual byte footprints.
  Differing-offset joins, changing-offset loops, out-of-bounds reads,
  truncation, flag exposure and unknown indexed accesses remain rejected.
- `3549877`: vector permutations, arithmetic, shifts, floating operations,
  copies and counts reuse existing typed decoders in pointer-flow analysis.
- `95f9251`, `f19a49d`, `ed613ca`: MOVPRFX, predicate logical/select, unrelated
  scalable spills, RDVL and ADDVL/ADDPL effects reuse validated typed grammars.
- `e2de140`: bounded indexed loads use a backward reaching-definition proof
  across every reachable CFG edge. MOV, masks, unsigned loads, LSR and CSEL
  establish upper bounds. Unknown definitions, value-changing cycles, skipped
  masks, negative signed indexes and overflow remain rejected. Cover all
  integer/FP/vector load widths, index extensions and natural shifts.
- `692fad1`: signed/unsigned vector lane extraction kills and SVE immediate
  duplication effects, including every legal lane/width and shifted forms.
- These batches passed focused Go 1.20/1.27, CLI, vet, three-OS LLVM 22 object
  checks and Linux/QEMU runtime tests. Logs use `raw-pool-sve-*`,
  `raw-pool-counters-flags-*`, `raw-pool-offset-*`, `raw-pool-vector-all-*`,
  `raw-pool-index-*` and `raw-pool-vector-transfer-*`.

## Pool safety boundary

Only R0–R17 can be terminal scratch kills, with explicit parameter/result
contracts. Exclude normal ABI results and every frame-result fallback register.
Unknown signatures and custom ArgRegs stay strict. Actual independent register
overwrites can kill other pointers. Reject calls, indirect exits, pointer
copies/escapes, exclusive-monitor effects and unknown instructions.
No arbitrary pointer arithmetic or unknown WORDs may be permitted for CI.

The bounded proof rejects multiple offsets at the same instruction rather than
discarding a path. Unknown memory footprints fail closed. The current effect
table deliberately permits only decoded instruction families. Diagnostic
overlays under `_out/` are not evidence; keep them synchronized before use.

## External results and ledger

- `simd@v1.21.1` ARM64 at `692fad1`: 45/47 files on each of
  Darwin/Linux/Windows (135/141), up from 41/47. All four math files now pass.
  Both JSON-quoting routines also advance, but the same two byte-processing
  files fail later at `parseIntsNEON`/`parseIntsSVE2`. See the root runner's
  `_out/simd-vector-transfer.{json,log}`; these are object, not runtime results.
- `parseInts` derives a second pointer (`ADD X12,X9,X13,LSL #3`) and reads
  `[X12,#-8]`. This needs alias/offset range and loop/branch reasoning; do not
  simply permit pointer copies. Diagnostic `_out/raw-pool-parseints-detail.log`
  in the development tree shows the instruction sequence, not proof of safety.
- At `692fad1`, outfix passed all three AMD64 OS targets (3/3). Go-highway dev9
  passed its selected ARM64 files across three OS targets (73/73). These exact
  replays do not establish complete discovery candidate or shard passes.
- Mazarin's exact failing Linux/arm64 configuration at `6e03524` passed
  14 files with one concrete source-ABI N/A. This is not a whole-module pass.
- Puter activation at `6e03524`: 5/8 files per ARM64 OS; LLVM 22 independently
  rejects the remaining raw words `0x6ea0f16c`, `0x6ebee0e4`, `0x6ea0f001`
  as invalid encodings. Do not reinterpret them as their macro names.
- Exact-version invalid-source skips now have pinned SHA-256 files, evaluated
  raw WORD expressions, LLVM 22 rejection checks, and distinct report/ledger
  status. Local proof passed for puter v1.2.3, gmsm v0.15.6 and two distinct
  gmgo module paths at v0.1.1.
  A skip is not a translated library. Rerun their shards on a clean snapshot.
- Large multi-package candidates now compile one package per translator
  process, releasing its LLVM objects and output before the next package.
  Rerun `wasm2go` candidates to confirm this removes `signal: killed`; do not
  recast a resource failure as source N/A or an invalid-source skip.
- Complete shard 17 at `6e03524`: 159 selected = 126 passed + 32 source N/A
  + one failed simd. Report is retained in the ecosystem runner under
  `_out/ci-repair-6e03524-shard17/shard-17.json`.
- Validated updater published that evidence via `5cec520`, merged at
  `e67462c`: 4,783 candidates, 4,624 pending, 126 passed, 32 N/A, one failed;
  incomplete and unverified. Evidence is stale against later implementation.
  Do not combine reports from different source/tool/ledger snapshots.
- Historical shard 25 evidence remains on
  `codex/pr40-shard25-evidence-preserved-20260926`; do not mix it into this run.

## Remaining work

1. Run combined gates and simd replay on a clean fixed snapshot. Remaining
   atan2/json quoting pools require real indexed-load/range proofs.
2. Resolve other actual CI failures: Knoxdb/forks need cross-TEXT ABI0
   shared-frame/register semantics; JIT libraries need native code layout and
   address/entry contracts; go-highway retains absent RIP constant pools.
3. Keep unproven upstream defects failed. Only the four confirmed exact
   versions above qualify for `skipped_invalid_source`; private Apple
   instructions alone are not proof of invalid source. Continue other
   applicable forms and candidates normally.
4. Run all 32 discovery shards with identical frozen provenance, publish
   validated ledger evidence, then batch-push to the allowed fork.
5. Shard 0's old runner lost communication, confirmed by its check annotation.
   It is infrastructure failure, not a pass and not proven OOM. Rerun it.
6. Clean owned generated IR/candidate caches and containers after use; retain
   JSON reports/logs. Do not delete global module caches or unrelated containers.
