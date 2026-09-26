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
- Active development: `codex/pr40-fp16-20260926`. Latest additions after
  `74b02be` cover raw SVE INDEX, scalar CNT/INC/DEC, scalar FABD/FMULX and
  proven unlabelled ARM64 constant pools. Focused Go 1.20/1.27, three-platform
  LLVM 22 object tests and relevant native/QEMU runtime checks pass. Inspect
  `_out/` logs and processes for the newest full-run completion.
- Frozen runner: `codex/pr40-ecosystem-fixes-20260925` at `74b02be`.
  Its full root suite passed (607 seconds), as did the official five-arch gate
  and nested CLI suite. Its full shard 10/32 replay is still running; inspect
  `_out/current-local-shard-10.log` and the atomically published report under
  `_out/current-local-shards/`. Do not edit/rebase this tree while it runs.
- `codex/pr40-shift-rip-20260925` retains completed shard 17 at `aacbcc6`:
  159 selected, 126 passed, one failed (simd), 32 source N/A. These results and
  the old Linux reports are historical evidence, not current-source passes.
- The committed assembly ledger at `569f8e3` publishes partial shard 10
  evidence for source `74b02be`. Subsequent code changes make it stale until
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
- Unlabelled trailing ARM64 pools use raw control-flow reachability and a
  conservative load-only address-use proof. Pool aliases share one contiguous
  global, preserving positive and negative offsets. Escaping addresses,
  indirect branches and reachable invalid words remain failures.

## External diagnostics and next blockers

- `simd@v1.21.1`: all 73 amd64 files passed the earlier replay. New ARM64
  diagnostics pass 21/47 files on each of Darwin/Linux/Windows (63/141 total).
  Remaining 26 files per platform include more complex pool address lifetimes
  and raw SVE compare/selection, floating min/max, ordinary memory, permutation
  and predicate-count encodings. See `_out/simd-arm64-pools-scalar.{json,log}`.
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
