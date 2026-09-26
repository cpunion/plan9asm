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
- Develop on `codex/pr40-arm64-raw-20260926`, implementation through
  `3549877`. Inspect status, worktrees and running processes before editing.
- Runners: `codex/pr40-fp16-20260926` (root gates and external replay) and
  `codex/pr40-ecosystem-fixes-20260925` (cross runtime and discovery).
  Inspect each HEAD and logs before advancing; never alter a running snapshot.
- Use Go 1.27 and LLVM 22 only. Root focused compatibility also uses Go 1.20.
  Cross runtime requires checksum-pinned QEMU 10.2.3, not QEMU 8.2.

## Completed gates and recent implementation

- At clean `6e03524`, full root passed in 755 seconds and required complete
  cross-runtime tests passed in 581 seconds. Logs are respectively
  `_out/full-root-pool-contract.log` and
  `_out/cross-runtime-pool-contract.log` in the runners.
- The same snapshot passed the official five-architecture classification gate
  and strict benchmark: 184/184 files, no N/A, 33 target-seconds.
  Observed form classification is not all-encoder runtime-semantic coverage.
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
- These batches passed focused Go 1.20/1.27, CLI, vet, three-OS LLVM 22 object
  checks and Linux/QEMU runtime tests. Logs use `raw-pool-sve-*`,
  `raw-pool-counters-flags-*`, `raw-pool-offset-*`, `raw-pool-vector-all-*`.
  The combined full gates must now be rerun on the final clean snapshot.

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

- `simd@v1.21.1` ARM64 at `422a906` and `e67462c`: 41/47 files on each
  of Darwin/Linux/Windows (123/141). Sorting now passes; six files still fail.
  See `_out/simd-counters-flags.{json,log}` and
  `_out/simd-bounded-pool.{json,log}` in the root runner.
- After `3549877`, a diagnostic proves the exp SVE pool; a clean complete
  external replay is still required before claiming either exp file passes.
- Mazarin's exact failing Linux/arm64 configuration at `6e03524` passed
  14 files with one concrete source-ABI N/A. This is not a whole-module pass.
- Puter activation at `6e03524`: 5/8 files per ARM64 OS; LLVM 22 independently
  rejects the remaining raw words `0x6ea0f16c`, `0x6ebee0e4`, `0x6ea0f001`
  as invalid encodings. Do not reinterpret them as their macro names.
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
3. Keep genuine upstream source defects failed. Gmsm/gmgo contain byte-swapped
   SM4 WORDs; puter includes invalid words and private Apple instructions;
   fiber/ai includes private Apple instructions. No permission was received
   to isolate proven defects as `blocked_external` or patch exact versions.
4. Run all 32 discovery shards with identical frozen provenance, publish
   validated ledger evidence, then batch-push to the allowed fork.
5. Shard 0's old runner lost communication, confirmed by its check annotation.
   It is infrastructure failure, not a pass and not proven OOM. Rerun it.
6. Clean owned generated IR/candidate caches and containers after use; retain
   JSON reports/logs. Do not delete global module caches or unrelated containers.
