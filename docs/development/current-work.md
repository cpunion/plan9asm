# Current work: PR 40 CI and ecosystem assembly coverage

This is a replaceable handoff. Keep scan and coverage funnel tables in the
[draft PR body](https://github.com/xgo-dev/plan9asm/pull/40), not here.

## Branches and provenance

- PR 40 is open and draft. Its fork head is `4cf5ade` on
  `cpunion:codex/expand-ecosystem-corpus-20260913`, rebased onto upstream
  `main` at `7cc8c0f`. Push only to the allowed fork; never push or merge
  upstream.
- The frozen PR-head CI run is `36087649450`. Do not import its reports into a
  checkout with different source content.
- `codex/pr40-current-ci-evidence-20260925` is a separate worktree at the PR
  source revision plus audited ledger checkpoints. Run
  `scripts/update-assembly-ledger.sh _out/current-ci-shards 32` there after
  downloading each new shard report. Its `verified` flag must stay false while
  shards are missing or failures remain.
- `codex/pr40-ecosystem-fixes-20260925` is the development branch. Its
  unpushed fixes include bounded compile-only LLVM modules and CI concurrency,
  complete ARM64 BF16/FP16 and SVE/SME families, x86 raw
  half-vector memory moves and AVX-512 BF16 narrowing, plus the Go 1.27
  ARM64 MSR PSTATE immediate family and source-local RIP constants for
  broadcasts, VMOVD/Q and packed X/Y/Z loads. Keep this branch separate from
  the frozen PR-head evidence worktree.

## Validated work and live CI

- At `4668f72`, the full root `go test ./... -count=1 -timeout=20m` passes with
  Go 1.27.1 (root package 564 seconds). The official Go 1.27 five-arch
  coverage gate, x/arch ARM64 corpus gate, root vet, both nested CLI suites,
  and focused Go 1.20 compatibility tests pass. Privileged MSR semantics are
  checked by comparing Go and LLVM 22 machine-code words; runtime execution
  is not claimed.
- Exact current-branch CLI replays prove `go-highway@v0.0.12` passes all
  three ARM64 targets (576/576 Linux, 577/577 Darwin, 576/576 Windows), and
  `go-highway@v0.0.0-dev9` passes its three ARM64 configurations (31/31
  Darwin, 21/21 Linux, 21/21 Windows). The latter's amd64 configurations
  still fail, so that candidate is not passed. `mazarin` now has 64 compiled
  objects over four targets, one evidence-backed asmdecl N/A and no failure.
  Both case-distinct `outfix` candidates pass 1/1; `pathtracer-ocl` passes
  its focused affected configurations. These are local diagnostics, not
  frozen shard reports or ledger pass updates.
- The new RIP constant decoder has focused Go-assembler and three-target LLVM
  22 object tests. A current-branch `simd@v1.21.1` amd64 diagnostic moves
  from 9/73 to 13/73 successful assembly files. The complete
  `nary_avx2_amd64.s` passes; several other files advance past broadcast and
  packed-move constants to their next unsupported form. The candidate remains
  failed; this is not a ledger pass or a frozen-shard report.
- The bounded-module CLI was tested on actual `pythonwasm2go@v0.4.0`
  22 MB ARM64 sources, including about 7,700 functions, on Linux and Windows
  targets, and on a 24 MB `spidermonkeywasm2go@v0.2.5` amd64 source for Darwin.
  These translate and compile with LLVM 22 and peak at roughly 2.4–3.6 GB
  resident memory. The exact three-file Windows configuration that CI killed
  also passed through the complete one-candidate discovery pipeline (3/3
  files, roughly 3.4 GB peak RSS). Objects are deleted after verification;
  local diagnostic IR directories are removed after measuring. A third killed
  candidate, `spanneranalyzerwasm2go/p8@v0.2.0`, passed its full one-candidate
  Linux/amd64 replay (1/1 file, 2.6 GB peak RSS).
- At 31 of 32 available CI shard artifacts, the old source revision has
  3,797 passed, 23 failed, 804 source/target not applicable and 159 pending
  exact module versions out of 4,783. The missing shard 0 was replayed
  locally for diagnostics (130 passed, 1 failed, 28 N/A). Its macOS LLVM and
  translator hashes differ from Linux CI, so its report must remain separate
  from the 31 CI artifacts and cannot be imported as a completed CI shard.
  The development branch must replay all shards after integration because its
  source fingerprint differs.
- All non-external CI jobs have passed on `4cf5ade`. Discovered-corpus shard
  0 lost its runner without an artifact. Do not report it as passed.

## Remaining CI failure classes

- Large generated assembly caused worker termination in `pythonwasm2go`
  and `spidermonkeywasm2go`; the local bounded-module replay addresses the
  observed peak but needs new CI verification.
- `go-highway@dev9` amd64 still contains several raw RIP-relative loads.
  Its GoAT-produced BF16 file refers to `LCPI` constants without defining a
  constant pool; Go's assembled first load points past that object's TEXT.
  Do not synthesize constants from comments or mark it passed. Other x86
  files in that candidate still need independent diagnosis.
- `sharkie`, `simd`, `gojit`, and the locally replayed shard-0 candidate
  `GopherJRE` depend on exact raw-byte layout, external entry points or
  PC-relative references. `GopherJRE` hand-encodes `LEA RIP+9` across three
  named instructions before an indirect jump. Translation that changes the
  byte layout must not be counted as semantic support. For `simd`, remaining
  amd64 failures include wider vector constant-pool loads, SSE2 PC-relative
  forms and additional AVX-512 encodings.
- `knoxdb` and its forks tail-jump into shared-frame helper `TEXT`s that write
  the original caller's FP result. A cross-`TEXT` ABI0 frame model is needed;
  accepting the FP offset in isolation would compile wrong behavior.
- `gmgo`, `gmsm` and `puter` contain raw ARM64 `WORD` values LLVM 22 does not
  decode. In `gmsm`, the SM4 comments describe the byte-swapped valid
  instruction, not the emitted word. `fiber/ai` has a private raw encoding.
  Preserve the actual source bytes when deciding whether a form is valid.
- `skywire@v1.3.69` has an invalid version download and its original GitHub
  repository is currently inaccessible; keep it failed/retryable rather than
  pretending it was tested. `mazarin` has an ARM64 argument-size
  mismatch rejected by Go's `vet -asmdecl`, so its source applicability must
  remain evidence-backed.

## Next actions

1. Reproduce each remaining deterministic failure before changing its
   instruction or ABI family. Compare against Go assembler, Go vet/build and
   LLVM 22, then add Go-accepted format, cross-target object and runtime
   tests as needed.
2. Keep old shard 0 local replay diagnostic-only; it cannot mix with the 31
   Linux CI reports. Do not import current-branch manual replays into that
   frozen old-head ledger. A new complete 32-shard run must share one source,
   toolchain and ledger fingerprint.
3. Once the remaining fixable failures are addressed, integrate the development
   branch into the contribution branch, push a substantial verified batch to
   `cpunion` only, then inspect new-head CI and review. Update PR-body counts
   from validated reports and keep the PR draft until its gates pass.
