# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit a running frozen source, tool or ledger snapshot.

## Contribution and published CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to
`cpunion`. Do not promote the upstream-connected branch until current-head
fork CI and review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
remains Draft.

Published fork source `47ea649d` failed all six priority discovery shards and
the aggregate in completed run `36942816488`. Policy preflight passed;
downstream skipped jobs did not pass. Historical reports remain evidence,
not success for local repairs. Preserve failure-first scheduling and all
64 required shards.

## Current local integration

Use branch `codex/pr40-reviewed-families-parallel-integration-20261006`.
Its reviewed parent `688267e8` integrates the earlier `82bccde5` mechanism
repairs, complete query-only CPP fixture transport and packed unsigned averages.
The branch additionally integrates the bounded exhaustive worker implementation
and its separate true-RED regression. Main is the sole assembly-ledger writer.

The underlying repairs preserve exact original source, actual SDK/tool/config
roles, native backing, ABI0 declarations, initialized bytes, code-address
continuations, and profile-scoped Go/CPP/LLVM provenance. Typed grammars and
effect consumers share complete forms without granting new source permissions.
Raw decoding keeps encoder identity and architecture-specific effects separate.
Wasm never receives `blockaddress`. LLVM 22 is mandatory.

Closed evidence at exact earlier snapshots:

- `82bccde5`: current SDK 93 roots / 1,186 children; historical SDK 92 roots /
  1,186 children. No failures, skips or missing terminals.
- `688267e8`: current SDK 133 roots / 1,903 children and historical SDK
  132 roots / 1,903 children passed complete focused race, module vet and build
  guards. Both closed runs were independently read twice.
- Worker leaf `005018f6`: both current and official historical SDKs passed
  the complete runner package, 30 roots / 44 children each, with race, full
  module vet/build and closed process ownership. Actual CLI tests prove bounded
  concurrency, exhaustive execution, failure continuation and coverage merging.

These are revision-specific mechanism/package gates, not full current-head
module, native Linux, external-library or ledger passes. Ignored `_out/`
artifacts retain raw JSON, original failures, complete source/SDK/tool manifests,
actual environments, command receipts and independent reader results.

## Parallel family work

Existing agents own separate persistent leaves:

- Sixteen x86 packed integer comparison opcodes: complete Go encoder rows,
  physical mode32 banks, dispatch identity, memory faults and whole-byte effects.
- Complete LOCK scalar RMW: native flag capture, carry initialization,
  namespace/segment identity and REX byte-bank aliases. A source spelling such
  as AH must not silently become the REX-selected SP byte.
- ARM64 ABI0 result-tail audit: Go rounds the allocation to its register width,
  but observable padding is not automatically a valid typed slot or backing
  permission. A test-only audit must preserve the missing public caller contract.

Review each entire final diff and independently recheck its frozen evidence
before integration. Do not cherry-pick a still-running source or infer combined
success from separate leaves. Required Linux native original-Go/LLVM mathematical
and fault counterparts remain distinct from Darwin/Rosetta diagnostics.

## Full-suite failures and required work

Original frozen source `7a9e69ac` completed 32 root partitions with 4,634 listed
roots: 4,583 passed, 11 failed, 40 skipped. The failed architecture fixtures,
mode32 vector forms, CPU IR assertions and missing separate LLVM 22 wasm linker
have dedicated repairs. Preserve the original failed report.

Its unpartitioned corpus-driver race package expired at 45 minutes. The actual
complete 547-root inventory is being independently checked in 32 partitions;
partial results and the known CPP fixture failure are not a package pass.
Use the reviewed ordinary runner's bounded `-workers` option for fresh suites,
retaining identical flags, complete actual inventory and original deadlines.

The strict five-target benchmark at `7a9e69ac` attempted all 184 files:
170 passed and 14 failed. Remaining contracts include dynamic reflect caller
frames, native runtime/context-switch entry state, reserved closure registers,
syscall continuation/backing, and observable ARM64 result padding. They cannot
be fixed by inventing frames, granting padding, or borrowing llgo's different
runtime provider. LOCK repairs require fresh benchmark replay.

Before publication, freeze the complete batch and run root and every other
module package, both nested CLIs, race/coverage, all five official gates,
strict standard-library/benchmark, all 64 discovery shards and required native
runtime jobs. Keep prior timeout/failure evidence. Classification coverage is
not complete encoder-form, object, native-runtime or library coverage.

## Inventory and ledger

The committed continuous Index interval is `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. It is not the whole Index or later updates.
Standalone inventory has newer unsynchronized records; cgo remains outside
this repository. Do not import or resume scans into a ledger used by a live gate.

Use the automatic writer only on independently validated fresh reports with
matching source/ledger/tool provenance, then independently read its result.
Missing shards remain pending, genuine failures remain failed, reviewed skips
retain exact reasons, and inspection failures remain retries. Never copy old
passes into a new revision. Keep funnel tables in the PR body, not this guide.
Remove owned package/build caches only after every writer and descendant exits.
