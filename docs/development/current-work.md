# Current work: PR 40 CI and ecosystem coverage

Read [instruction development](instructions.md), [validation](validation.md),
and [report provenance](discovery-verification.md). Keep progress funnels in
the [draft PR](https://github.com/xgo-dev/plan9asm/pull/40), not this checkpoint.

## Branches and verification

- Contribute only through `cpunion:codex/expand-ecosystem-corpus-20260913`.
  Remote head remains `4cf5ade`; upstream main `7cc8c0f` is an ancestor.
  Both were fetched again during this batch. No new push or CI pass yet.
- Old CI run `36087649450`: 74 successful jobs, 19 failed discovered shards
  and one failed aggregate. Base/root, standard library, race, coverage,
  benchmark, curated libraries and cross-runtime jobs passed on that old head.
- Active development: `codex/pr40-arm64-raw-20260926`. Commits through
  `c144677` add complete raw SVE floating comparisons, unary operations and
  predicate counts; fix floating opcode masks, reduction source registers,
  Z/V/F aliasing and fixed-point conversion destination width. Focused Go
  1.20/1.27, three-platform LLVM 22 objects and required QEMU checks pass.
- Full-root runner: `codex/pr40-fp16-20260926`. Full suites at `97c89a0`
  and `052a7f6` passed (631 and 602 seconds); the latter also passed the
  official five-arch gate. Inspect Git/logs before treating later changes as
  fully tested. Never advance its source while a verification is running.
- `codex/pr40-ecosystem-fixes-20260925` retains complete shard 10/32 evidence
  for `74b02be`: 173 selected, 142 passed, 31 source N/A, zero failures and
  2,889 successful translations. Its report is
  `_out/current-local-shards/shard-10.json`; preserve that historical report
  if advancing the checkout for another frozen verification.
- `codex/pr40-shift-rip-20260925` retains completed shard 17 at `aacbcc6`:
  159 selected, 126 passed, one failed (simd), 32 source N/A. These results and
  the old Linux reports are historical evidence, not current-source passes.
- The committed assembly ledger from `d779df4` (merged at `4a57deb`) publishes
  complete shard 10 evidence for `74b02be`. Subsequent code changes make it stale until
  validated reports for the new source replace it. Never hand-promote statuses.

## Completed implementation batches

- Raw FP16 arithmetic, moves, all FMA variants, square roots, packed/scalar
  precision conversions, BF16 dot products and VMOVW. Full roots through
  `74b02be` pass. FP16 directed rounding is implemented explicitly because
  LLVM 22 AArch64 constrained narrowing did not honor static rounding modes.
- SVE INDEX decodes all four immediate/register combinations and B/H/S/D
  widths, without applying Go's named forced-D encoding to raw words.
  CNT/INC/DEC preserve all patterns and multipliers. Required QEMU execution
  checks six vector lengths (128 through 2048 bits), wrapping arithmetic and
  unknown predicate patterns; independent LLVM MC checks guard encoding bits.
- Scalar FABD/FMULX cover H/S/D and coherent upper-lane clearing, sharing the
  scalar binary spec. Native ARM64 and QEMU compare finite/zero/subnormal/
  infinity/NaN results against architectural instructions.
- Raw floating comparisons cover register and zero operands; unary operations
  cover all 17 Go families and their distinct merging/zeroing size fields.
  Required QEMU oracles exercise arithmetic, comparisons and ordinary unary
  semantics over six SVE lengths. Newer zeroing operations use a cleared-
  destination SVE1 reference; sized-rounding and counter forms have object
  checks, not a claim of native execution on an unavailable newer ISA.
- PCNTP covers both ordinary predicates and SVE2.1 counters. PTRUE unnamed
  patterns produce empty predicates as specified by Arm. Fixed GPR-to-float
  conversion uses destination precision independently of integer width;
  all signed/unsigned W/X-to-S/D combinations have runtime oracles.
- Unlabelled trailing ARM64 pools use raw control-flow reachability and a
  conservative load-only address-use proof. Pool aliases share one contiguous
  global, preserving positive and negative offsets. The load-only proof now
  follows branches/loops until every live address is killed. Escaping addresses,
  indirect calls and reachable invalid words remain failures.

## External diagnostics and next blockers

- `simd@v1.21.1`: all 73 amd64 files passed the earlier replay. New ARM64
  diagnostics pass 22/47 files on each of Darwin/Linux/Windows (66/141 total).
  Remaining files include more complex pool address lifetimes and raw SVE
  predicate logic, compact/selection, floating min/max, ordinary memory and
  permutation encodings. See `_out/simd-cntp.{json,log}` in the active tree.
- `go-highway@v0.0.0-dev9`: ARM64 diagnostics passed; amd64 now passes 4/7
  files on each of Darwin/Linux/Windows. Three source files reference absent
  RIP constant pools. Do not invent data from comments. See the frozen
  runner's `_out/highway-fp16-scalar-convert.{json,log}`.
- Knoxdb/forks need a real cross-TEXT ABI0 shared-frame/register model.
  GopherJRE/sharkie/gojit observe exact native code addresses/layout/markers.
  Loosening frame checks or changing code layout would miscompile them.
- gmgo/gmsm, puter and fiber/ai contain other raw words requiring encoding
  validation (including byte-swapped comments and private Apple instructions).
  Skywire's inaccessible version remains a retryable fetch failure, not N/A.
- Earlier local fixes cover case-distinct outfix, pathtracer, mazarin and large
  wasm2go LLVM memory exhaustion. They still need current-head CI confirmation.

Finish current full gates, continue the failing families with red/green tests,
then replay all shards in one frozen source/tool/ledger snapshot. Publish
evidence with `scripts/update-assembly-ledger.sh`, push a verified batch only
to the allowed fork, and keep the PR draft until every completion gate passes.
Do not resume discovery scanning while CI failures remain. Temporary candidate
packages/objects must be cleaned; retain diagnostic reports, not full caches.
