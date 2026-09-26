# Current work: PR 40 CI and ecosystem assembly coverage

Replace this checkpoint when resuming. Keep scan/coverage funnel tables in the
[draft contribution PR](https://github.com/xgo-dev/plan9asm/pull/40), not here.
Read [instruction development](instructions.md), [validation](validation.md),
and [report provenance](discovery-verification.md) before acting.

## Branches and CI

- PR 40 is open and draft on `cpunion:codex/expand-ecosystem-corpus-20260913`.
  Its remote head is still `4cf5ade`; push only to that allowed fork.
- Upstream `main` at `7cc8c0f` is already an ancestor of development. Fetch and
  check again before final verification; never rebase a running frozen tree.
- Old-head CI run `36087649450` has 74 successful and 20 failed jobs: 19
  discovered-library shards plus their aggregate. Base tests, standard library,
  race, coverage, benchmark, curated libraries and cross-runtime jobs passed.
- Current development is `codex/pr40-fp16-20260926`, through `1ea0da2`, 80
  code/checkpoint commits ahead of the remote head before this document update.
  `codex/pr40-ecosystem-fixes-20260925` is a separate frozen full-test checkout
  at `22e9519` (its full suite has now finished). Inspect processes before editing. Do not confuse
  local fixes with a new successful CI run.
- `codex/pr40-current-ci-evidence-20260925` retains 31 old Linux CI reports.
  Shard 0 is missing. Keep that evidence separate from new-source or macOS
  reports; do not import mixed provenance or hand-promote ledger statuses.

## Newest verified changes

- `9833f21`: raw FP16 scalar moves and six scalar arithmetic/min/max operations;
  packed FP16 arithmetic was added at `4ef1ead`. The root Go 1.27.1 suite at
  `9833f21` passed (617 seconds). The root suite at `4ef1ead` also passed.
- `22e9519`: all 30 raw FP16 FMA variants, sharing the existing typed FMA3
  table. Covers 132/213/231 order, packed/scalar, X/Y/Z, masks/zeroing,
  broadcasts, rounding, LLIG, source-local RIP constants and invalid encodings.
  Shared FMA lowering now suppresses masked-off memory reads and maintains
  overlapping register views. FP16 X/Y forms require AVX512VL as well as FP16.
- `1ea0da2`: raw packed/scalar FP16 square roots. Same-width conversion uses a
  shared opcode spec, masked memory loader and coherent vector stores.
- These families pass focused Go 1.20/1.27 tests and LLVM 22 objects for
  Darwin/Linux/Windows amd64 and Linux/Windows 386. Portable runtime oracles
  check FMA order/fused rounding, scalar lane preservation, high-lane clearing,
  masks, null inactive sources and unaligned guard-page boundaries. FMA ran on
  ARM64 and x86_64 via Rosetta; square root ran on ARM64. This is execution of
  lowered IR, not proof of native AVX512-FP16 hardware execution.
- At `1ea0da2`, the official Go 1.27 five-arch gate passes with no unsupported
  observed forms or parse errors. Both nested CLI suites, root vet, focused
  existing FMA/same-width regressions and supported-op extraction pass. The
  full root suite at `22e9519` also passed (606 seconds); its log is
  `_out/full-root-fp16-fma.log` in that frozen checkout. The square-root commit
  still needs the final current-head full run, not just its focused tests.

## External-library diagnostics (not ledger passes)

- `simd@v1.21.1`: all 73 amd64 files now translate and compile. The full shard
  17 replay at `aacbcc6` finished with 159 selected, 126 passed, 1 failed and
  32 N/A. The remaining candidate is simd: ARM64 has appended literal pools
  decoded as instructions and additional SVE2/FABD encodings. See its frozen
  `_out/current-local-shards/shard-17.json` in
  `codex/pr40-shift-rip-20260925`; do not mix it with old Linux CI reports.
- `go-highway@v0.0.0-dev9`: ARM64 configurations pass (31 Darwin, 21 Linux,
  21 Windows files). Its amd64 diagnostic still has 1/7 files passing. FMA
  and square-root failures are gone; the next errors are MAP6 `VCVTPH2PSX`
  (`62 f6 7d 48 13 ...`) and `VDPBF16PS` (`62 f2 76 48 52 ...`). Three GoAT
  sources also reference omitted constant pools. Never invent data from their
  comments. Diagnostic reports/logs are `_out/highway-sqrt.*`.
- `go-highway@v0.0.12` passes its ARM64 matrices (576 Linux, 577 Darwin,
  576 Windows files). Both case-distinct outfix candidates and affected
  pathtracer-ocl configurations pass local replays. Mazarin has 64 compiled
  objects plus one Go-asmdecl-backed N/A and no translation failures.
- Bounded LLVM modules and reduced concurrency fix the observed killed-worker
  cases locally: pythonwasm2go, spidermonkeywasm2go and
  spanneranalyzerwasm2go/p8. Actual large-source replays peak around 2.4–3.6 GB;
  the killed three-file Windows candidate passed the full candidate pipeline.
  New-head CI verification is still required.

## Remaining failure classes and next actions

1. Continue the go-highway precision-conversion and BF16 dot-product families
   with TDD and complete raw formats. Go 1.27 does not name these FP16/BF16
   forms: consult Intel encodings and LLVM 22 definitions, not comments alone.
2. For simd ARM64, distinguish reachable raw code from appended data and resolve
   actual source-local PC-relative references before adding the next SVE family.
3. Knoxdb/forks need a real cross-TEXT ABI0 shared-frame/register model. Indirect
   table jumps preserve the caller's live registers and FP result area. Merely
   accepting helper FP offsets compiles incorrect behavior.
4. GopherJRE, sharkie and gojit use exact raw code addresses/layout or markers.
   Do not translate their bytes to a different layout while claiming support.
5. gmgo/gmsm include invalid emitted ARM64 words (some SM4 comments describe
   byte-swapped instructions); puter and fiber/ai have other raw encodings.
   Validate actual bytes. Skywire's inaccessible version download remains a
   retryable failure, not source N/A or a passed candidate.
6. Finish current-head full gates, replay fixed shards in one frozen snapshot,
   and publish audited assembly-ledger updates using the tool. Source changes
   invalidate earlier passes. Push a substantial verified batch, then inspect
   new CI and review; keep PR draft until all completion gates pass.

Keep LLVM 22 only. Disposable corpus candidate caches/objects are cleaned by
the runner; keep only diagnostic reports needed for the next reproduction.
Do not commit private paths, hostnames, archives, cgo inventory or stale runs.
