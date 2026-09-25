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
- `codex/pr40-ecosystem-fixes-20260925` is the development branch. It has
  unpushed complete ARM64 BF16 matrix, FP16 FMA/unary, scalar FCVTZS/FCVTZU,
  raw LDPSW pair, and bounded compile-only LLVM-module changes through
  `e7f0d2b`. The last change also limits CI matrix concurrency to eight.

## Validated work and live CI

- LLVM 22 and Go 1.27 focused tests, the nested `plan9asmll` suite,
  `go vet` in the root and nested CLI, and `actionlint` pass on the
  development branch. Cross-target LLVM object and native ARM64 runtime
  fixtures pass for the new instruction families.
- The bounded-module CLI was tested on actual `pythonwasm2go@v0.4.0`
  22 MB ARM64 sources, including about 7,700 functions, on Linux and Windows
  targets, and on a 24 MB `spidermonkeywasm2go@v0.2.5` amd64 source for Darwin.
  These translate and compile with LLVM 22 and peak at roughly 2.4–3.6 GB
  resident memory. The exact three-file Windows configuration that CI killed
  also passed locally (3/3 files, roughly 3.4 GB peak RSS). Objects are
  deleted after verification; local diagnostic IR directories are removed
  after measuring.
- At 29 of 32 available CI shard artifacts, the old source revision has
  3,560 passed, 21 failed, 757 source/target not applicable and 445 pending
  exact module versions out of 4,783. The evidence branch validates and
  commits these partial ledger states; the development branch must replay
  all shards after integration because its source fingerprint differs.
- All non-external CI jobs have passed on `4cf5ade`. Discovered-corpus shard
  0 lost its runner without an artifact; shards 12 and 29 were still running
  at the last check. Do not report these three as passed.

## Remaining CI failure classes

- Large generated assembly caused worker termination in `pythonwasm2go`
  and `spidermonkeywasm2go`; the local bounded-module replay addresses the
  observed peak but needs new CI verification.
- `go-highway` has 20/31 Darwin ARM64 files passing after the current family
  work. The other 11 are primarily SME ZA/matrix and streaming-memory forms.
  Model complete instruction families with register/state semantics.
- `knoxdb` and its forks tail-jump into shared-frame helper `TEXT`s that write
  the original caller's FP result. A cross-`TEXT` ABI0 frame model is needed;
  accepting the FP offset in isolation would compile wrong behavior.
- `sharkie`, `outfix`, `gojit`, and `simd`/`pathtracer-ocl` use address-sensitive
  raw x86 bytes, mixed directives or layout-dependent control flow. Do not
  silently decode, skip or relocate bytes whose PC-relative meaning changes.
- `gmgo` and `puter` contain raw ARM64 `WORD` values that LLVM 22 does not
  decode, even with relevant features enabled. `fiber/ai` has a private raw
  encoding. Do not label these instructions supported without preserving
  correct runtime register and memory semantics.
- `skywire@v1.3.69` has an invalid version download; keep it failed/retryable
  rather than pretending it was tested. `mazarin` has an ARM64 argument-size
  mismatch rejected by Go's `vet -asmdecl`, so its source applicability must
  remain evidence-backed.

## Next actions

1. Finish the old run; import and audit all artifacts in the evidence worktree.
   For the lost shard 0, obtain a same-source complete rerun rather than
   fabricating a result. Keep the ledger synchronized with actual reports.
2. Reproduce each deterministic failure before changing its instruction or
   ABI family. Compare against Go assembler, Go vet/build and LLVM 22, then
   add Go-accepted format, cross-target object and runtime tests as needed.
3. Once the development batch addresses the failures, integrate it into the
   contribution branch, run focused and official five-arch gates, then run
   all 32 external shards on one frozen source fingerprint.
4. Push a substantial verified batch to `cpunion` only, update the PR body
   from validated counts, inspect new-head CI and review, and keep the PR
   draft until its completion gates pass.
