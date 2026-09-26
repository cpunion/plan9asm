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
- Develop on `codex/pr40-arm64-raw-20260926`. Inspect
  status, worktrees and running processes before editing. Do not push until
  the failing CI classes are fixed and the assembly ledger is refreshed.
- Runners: `codex/pr40-fp16-20260926` (root gates and external replay) and
  `codex/pr40-ecosystem-fixes-20260925` (cross runtime and discovery).
  Inspect each HEAD and logs before advancing; never alter a running snapshot.
  The root runner finished shard 1 at `432314f`, then advanced to `f80fb8f`
  for shard 4 under `_out/ci-repair-f80fb8f-shard4/`. The ecosystem runner
  finished shards 10, 17 and 24 at `692fad1`, then advanced to `d7fa0e2`
  for shard 29 under `_out/ci-repair-d7fa0e2-shard29/`. Never mix these
  report snapshots.
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
  gmgo module paths at v0.1.1. Shard 1 at `432314f` completed with 147
  selected = 116 passed + 30 source N/A + one Puter invalid-source skip;
  zero failures and 2,443 successful translations. Shard 4 at `f80fb8f`
  has checkpointed the `gitee.com/zhaochuninhefei/gmgo@v0.1.1` skip and
  remains partial.
  A skip is not a translated library. Rerun their shards on a clean snapshot.
- Large multi-package candidates now compile one package per translator
  process, releasing its LLVM objects and output before the next package.
  Rerun `wasm2go` candidates to confirm this removes `signal: killed`; do not
  recast a resource failure as source N/A or an invalid-source skip.
- A subsequent local change splits object-validation LLVM modules at 32,768
  assembly instructions or 128 functions, whichever comes first; a single
  oversized function remains whole and is still fully checked. The `p8` module
  has 2,179 functions and about 1.39 million amd64 assembly lines, so the
  function-only limit was not a reliable memory bound. Focused chunk tests
  pass, including complete coverage and cross-chunk raw TEXT references;
  the actual `p8` candidate still needs replay. The strict benchmark passed
  all 184 applicable files on five targets with no N/A in 25 target-seconds.
- A direct `pythonwasm2go/p0` Windows/ARM64 object replay passed in 39 seconds
  with a 3.3 GB maximum resident set. Its previous three-package translator
  process is still being exercised by shard 1 at `432314f`; per-package
  isolation is newer and needs a clean-shard rerun. Generated probe outputs
  were removed after measurement.
- Selected `go-highway@v0.0.12` Darwin/ARM64 failures from the old CI passed
  against current code: BF16 image 2/2, SME matmul 3/3, NEON matmul 1/1.
  These are direct replays, not a whole-module or shard pass.
- Shard 29 at `d7fa0e2` reached `go-highway@v0.0.12` and failed in two AMD64
  GoAT-generated files. Their raw QUAD opcodes decode, but the referenced
  CPI0_4 and CPI1_1 constants are absent. Go 1.27's own assembled objects
  place the fixed RIP targets beyond all TEXT symbols. A new exact-version,
  SHA-pinned invalid-source proof distinguishes this from invalid ARM64
  encodings and retains a specific skip reason. Focused proof tests pass;
  the complete candidate and shard still need replay on the new revision.
- `simd`'s apparent raw branch `0x540be400` is an inline numeric constant
  after the function body, not evidence of invalid source. Keep its two
  `parseInts` files failed until the pool-address and alias proof is sound.
- Complete shard 17 at `6e03524`: 159 selected = 126 passed + 32 source N/A
  + one failed simd. Report is retained in the ecosystem runner under
  `_out/ci-repair-6e03524-shard17/shard-17.json`.
- Complete shard 10 at `692fad1`: 173 selected = 142 passed + 31 source N/A,
  no failures. Shard 24 at the same revision: 164 selected = 134 passed +
  29 source N/A + one failed old `celliott/gvisor` dependency resolution.
  The latter produced both 404 and transient 429/503 responses, so it was
  correctly kept failed rather than relabelled source N/A. Later local code
  recognizes 429 as infrastructure and retries only transient network
  failures twice; that needs a clean-shard replay.
- Validated updater published that evidence via `5cec520`, merged at
  `e67462c`: 4,783 candidates, 4,624 pending, 126 passed, 32 N/A, one failed;
  incomplete and unverified. Evidence is stale against later implementation.
  Do not combine reports from different source/tool/ledger snapshots.
- Historical shard 25 evidence remains on
  `codex/pr40-shard25-evidence-preserved-20260926`; do not mix it into this run.

## Remaining work

1. Finish current shards and rerun failure shards on one clean revision.
   `simd` needs a sound derived-pool-pointer alias and range proof; do not
   decode its numeric pool as executable branches.
2. Resolve other actual CI failures: Knoxdb/forks need cross-TEXT ABI0
   shared-frame/register semantics; JIT libraries need native code layout and
   address/entry contracts; large `wasm2go` modules need bounded memory even
   when a single package contains thousands of generated functions.
3. Keep unproven upstream defects failed. Only the five confirmed exact
   versions above qualify for `skipped_invalid_source`; private Apple
   instructions alone are not proof of invalid source. Continue other
   applicable forms and candidates normally.
4. Resolve the historical `skywire@v1.3.69` unavailable source without
   pretending a failed download is a pass. Run all 32 discovery shards with
   identical frozen provenance, refresh the validated assembly ledger, then
   batch-push to the allowed fork. Its current source fingerprint is stale.
5. Shard 0's old runner lost communication, confirmed by its check annotation.
   It is infrastructure failure, not a pass and not proven OOM. Rerun it.
6. Clean owned generated IR/candidate caches and containers after use; retain
   JSON reports/logs. Do not delete global module caches or unrelated containers.
