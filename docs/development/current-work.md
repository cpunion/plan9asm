# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). CI repair has priority:
do not start new module-index inventory scans while failures remain.

## Contribution and frozen worktrees

- PR 40 is OPEN/DRAFT at remote head `4cf5ade`: last live inspection found
  74 successful checks and 20 failures, including discovery shards and the
  aggregate.
  Run `gh pr view` again for current state. No later local repair was pushed.
- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`, never upstream.
  The user requests complete repairs before a batch push. Keep the PR draft.
- Upstream main `7cc8c0f` was fetched and is already an ancestor.
- Develop on `codex/pr40-arm64-raw-20260926`. Inspect
  status, worktrees and running processes before editing. Do not push until
  the failing CI classes are fixed and the assembly ledger is refreshed.
- Development is on `codex/pr40-arm64-raw-20260926`, not pushed. At clean `e7d12c7`, shard 24
  completed 164 selected = 134 passed + 29 N/A + one superseded skip;
  shard 25 completed 137 selected = 109 passed + 27 N/A + one superseded
  skip. Both had zero failures. At clean `9aa59b1`, shard 29 completed
  151 selected = 127 passed + 23 N/A + one invalid-source skip, zero
  failures. These are different provenance snapshots and cannot be combined.
- Exact historical gVisor and Skywire mirrors are now pinned as
  `skipped_superseded`, never passed. Their newer canonical projects and
  project-identity evidence are in `testdata/corpus/superseded-modules.json`.
  Same-path `@latest` updates replace old candidates through discovery.
- Corpus tool builds now use `-trimpath`: same-revision translator and runner
  builds in two clean worktrees produced identical SHA-256 values. Earlier
  e7 shard reports used path-dependent binaries and cannot be aggregated;
  replay required on a later shared frozen revision. Shard 22 at `2d17842`
  completed 163 candidates with two failures: fiber/ai's private AMX word and
  wagon's unexpanded `GO_ARGS`. The latter came from `runtime.GOROOT()` being
  empty in a `-trimpath` translator. A regression first failed, then passed
  after resolving the matching Go toolchain root. A focused exact wagon replay
  passed its Darwin/AMD64 assembly through LLVM 22. A later shard-22 replay
  at `e509472` was deliberately interrupted before completion: the runner
  itself still used empty `runtime.GOROOT()` for source probes. No result from
  that interrupted run counts as corpus evidence. A shared toolchain-root
  helper now fixes translator, runner and scanner; freeze and replay shard 22.
- The user authorized a distinct, not-passed private-extension skip for the
  exact fiber/ai v0.1.2 Darwin/ARM64 Apple AMX file. The manifest pins its
  source hash and opcode. The new runner verifies current Go assembly and
  still translates the other files. A focused local replay passed all three
  remaining Darwin/ARM64 files; log `_out/fiber-private-replay.log` in the
  development worktree. Old shard 22 at `2d17842` independently confirmed
  Mazarin passed (91 translations, 3 source/ABI N/A), while fiber failed only
  on that AMX file. The new classification needs a frozen full-shard replay.
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
  Subsequent frozen shard 24 and 25 replays passed giant `wasm2go` candidates.
  Do not recast a resource failure as source N/A or an invalid-source skip.
- A subsequent local change splits object-validation LLVM modules at 32,768
  assembly instructions or 128 functions, whichever comes first; a single
  oversized function remains whole and is still fully checked. The `p8` module
  has 2,179 functions and about 1.39 million amd64 assembly lines, so the
  function-only limit was not a reliable memory bound. Focused chunk tests
  pass, including complete coverage and cross-chunk raw TEXT references;
  the actual `p8` candidate passed at `9aa59b1`. The strict benchmark passed
  all 184 applicable files on five targets with no N/A in 25 target-seconds.
- A direct `pythonwasm2go/p0` Windows/ARM64 object replay passed in 39 seconds
  with a 3.3 GB maximum resident set. Its previous three-package translator
  process was exercised by shard 1 at `432314f`; per-package isolation is
  newer and still needs a clean-shard rerun. Generated probe outputs
  were removed after measurement.
- Selected `go-highway@v0.0.12` Darwin/ARM64 failures from the old CI passed
  against current code: BF16 image 2/2, SME matmul 3/3, NEON matmul 1/1.
  These are direct replays, not a whole-module or shard pass.
- Shard 29 at `d7fa0e2` reached `go-highway@v0.0.12` and failed in two AMD64
  GoAT-generated files. Their raw QUAD opcodes decode, but the referenced
  CPI0_4 and CPI1_1 constants are absent. Go 1.27's own assembled objects
  place the fixed RIP targets beyond all TEXT symbols. A new exact-version,
  SHA-pinned invalid-source proof distinguishes this from invalid ARM64
  encodings and retains a specific skip reason. The `9aa59b1` Go 1.27.1
  replay classified it as `SKIP_INVALID_SOURCE`, and the shard completed
  with zero failures.
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
  correctly kept failed rather than relabelled source N/A. The new 9aa59b1
  replay confirms the old 2018 source is not buildable with current Go even
  when its legacy import path is aliased to the downloaded module. The fork
  shares commit `3b895abd3b05` with canonical `google/gvisor`, whose 2026
  version is separately scanned and assembly-bearing.
- Shard 4 at `f80fb8f` completed: 168 selected = 129 passed + 38 source N/A
  + one exact gmgo invalid-source skip, zero failures. Shard 29 at `d7fa0e2`
  completed: 151 selected = 125 passed + 25 N/A + one pre-fix go-highway
  failure. The large `spanneranalyzerwasm2go/p8` candidate passed 6/6 object
  translations on that snapshot. Both are diagnostic-only for current source.
- Clean `9aa59b1` passed full root tests, nested CLI tests, vet, build,
  official five-architecture coverage, strict benchmark (184/184) and stdlib
  corpus. Clean `e7d12c7` passed full root tests in 610 seconds after three
  concurrent heavy shards had caused an earlier 20-minute timeout. A focused
  root test and package tests confirmed the timeout was load-related. The new
  toolchain-root tests and corpus/scanner packages pass on Go 1.27.1; the
  trimpath regression also passes on Go 1.20.14. Full root and shard tests
  remain required on a clean new revision. Do not call the complete external
  corpus verified yet.
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
   versions in the invalid-machine-code manifest qualify for
   `skipped_invalid_source`. The distinct private-extension exception is
   authorized only for its pinned file/target and is not a pass. Continue
   other applicable forms and candidates normally.
4. Run all 32 shards with identical frozen provenance, refresh the validated
   assembly ledger, then batch-push to the allowed fork. The current evidence
   snapshot is stale. Reuse neither path-dependent e7 reports nor reports
   from older revisions in that aggregate.
5. Shard 0's old runner lost communication, confirmed by its check annotation.
   It is infrastructure failure, not a pass and not proven OOM. Rerun it.
6. Clean owned generated IR/candidate caches and containers after use; retain
   JSON reports/logs. Do not delete global module caches or unrelated containers.
