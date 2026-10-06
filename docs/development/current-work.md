# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit a running source, tool or ledger snapshot.

## Contribution boundary

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to
`cpunion`. Do not promote the upstream-connected branch until current-head
fork CI and review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
remains Draft.

Published fork source `47ea649d` failed all six priority discovery shards and
the aggregate in completed run `36942816488`. Policy preflight passed;
downstream skipped jobs did not pass. Preserve failure-first scheduling and
all 64 required shards. The read-only upstream main check still resolves to
`e2164ac0`; inspect again before final publication.

## Current integration

Use branch `codex/pr40-rule-and-module-integration-20261007`.
Its `8125c3dd` checkpoint contains the reviewed complete comparison/LOCK and
memory-role rules, grouped symbol/PUSH source roles, FP observation fixtures,
scalar self-XOR byte effects, ABI0 byte-region boundary fixtures, and exhaustive
main-module CI ownership with real LLVM/Node/wasm dependency observations.

The shared self-XOR effect recognizes only identical validated register
operands, preserves exact written bytes and flags, and no longer consumes
undefined input for a true zero idiom. It grants no caller or memory authority.
The ABI0 fixtures retain observable native result tails; they do not authorize
discarding them. The module runner exhaustively owns root, corpus and all other
main-module packages; nested CLIs and required native jobs remain separate.
Its portable verifier uses captured byte identities without requiring the
producer's filesystem to remain mounted at aggregation time.

Reviewed leaf evidence is revision-specific, not combined success:

- `927e6d6e`: complete first-party corpus-driver race suite, 32 partitions,
  548 roots / 1,904 children, no failures/skips/missing terminals. This is not
  discovered external-library compilation.
- `0c020fc1`: complete grouped-symbol/PUSH gates passed both SDKs. All five
  official architecture scans closed on Go 1.27 and complete Go 1.20; the
  historical maskless EVEX zeroing exception remains explicit.
- `197a3976`: complete FP-width observation gates passed both SDKs,
  eight roots / 288 children; the stale unknown-width assertion was corrected.
- `85213ffd`: scalar self-XOR narrow and complete affected-family gates passed
  both SDKs; the complete gate has 64 roots / 6,079 children, no failures/skips.
- `b820b331`: ABI0 byte-region and actual Windows executable-naming fixtures
  passed both SDKs; 573 children per SDK. Linux test selection and Windows PE
  build observations do not certify execution on those platforms.
- `07c0e088` and `51a7369b`: exhaustive module and dependency mechanisms passed
  both SDKs, 45 runner roots / 81 children plus seven policy roots per SDK,
  full module vet/build and independent frozen readbacks. The corrected
  ignored auditor rejects unknown compiler build diagnostics. Cold-download,
  missing-runtime and retired-producer-origin failures remain retained.

Ignored `_out/` artifacts retain actual inventories, raw JSON, command exits,
source/SDK/tool manifests, effective environments and closed process ownership.
A transient late PID probe failed during main's independent dependency review;
the original failure is retained and a fresh complete readback closed.
Do not kill a process merely because its numeric PID matches an old receipt.

## Remaining family work

Existing agents own separate persistent leaves:

- Complete unnamed x86 TYPE_ADDR and ordinary grouped-address grammar:
  preserve Go's absent-star Index/Scale0 role instead of coercing it to Scale1.
  Real MOVL/Q and LEA counterexamples require the whole source grammar repair.
  ARM/ARM64 parsing and named SB/FP/AUTO permissions must not be widened.
- Scalar FP byte views: MOVSS/SD and VMOVSS/SD must use exact declared byte
  spans across integer/floating fields and suboffsets, preserving bit patterns,
  initialized result bytes, input permissions and holes.
- ARM64 snappy source CFG: reproduce the exact large-frame encodeBlock LR
  failure and compare original Go prologue/epilogue and all reaching paths.
  Official instruction tables alone do not prove continuation preservation.

Review each complete final diff and independently recheck its frozen evidence
before integration. Do not cherry-pick live sources or infer combined success
from separate leaves. Main is the sole assembly-ledger writer.

## Whole-suite and native boundaries

The original full-root run at `29cdea5f` closed all 32 partitions with
4,718 actual roots: 4,677 passed, two failed, 39 host-inapplicable skipped;
55,712 children passed. Both failures have later focused repairs above, but
the unchanged original report remains failed. A fresh complete integration
replay, every other main-module package and both nested CLIs are still required.

The strict five-target benchmark at `0c020fc1` attempted all 184 files:
172 compiled and 12 failed, no skipped or unattempted files. Remaining contracts
include dynamic reflect callers, native runtime/context-switch entry state,
closure registers, syscall/TLS continuation/backing and ARM64 observable
result tails. Do not invent frames/padding or borrow llgo's different provider.
The `fa88a9ca` source/object audit reproduces both Or32 and And32 on all three
APIs and five ARM64 targets; 30 blocked translations are not LLVM objects.

Before publication freeze the complete batch and run root and all other module
packages, both nested CLIs, race/coverage, official/stdlib/benchmark gates,
all 64 discovery shards and required native runtime owners. Required Linux
original-Go/LLVM mathematical and fault counterparts remain distinct from
Darwin/Rosetta or emulated Linux diagnostics. Wasm must never use blockaddress.

## Inventory and ledger

Committed Index continuity is `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`, not the entire Index or later updates. Standalone
inventory has unsynchronized newer records; cgo remains outside this repository.
Do not import/resume scans into a ledger used by a live gate.

The older external 64-shard owner at `6ed5be69` is still live. Preserve its
cache and reports; its progress cannot certify the current integration.
Use the automatic ledger writer only on independently validated fresh reports
with matching source/ledger/tool provenance, then independently read the result.
Missing coverage remains pending, genuine failures remain failed, reviewed skips
retain exact reasons and inspection failures remain retries. Keep scan/coverage
funnel tables in the PR body, not this guide. Remove only owned caches after
every writer and descendant has ended.
