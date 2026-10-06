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

Use branch `codex/pr40-complete-go-rule-families-20261006`.
Its reviewed parent `688267e8` integrates the earlier `82bccde5` mechanism
repairs, complete query-only CPP fixture transport and packed unsigned averages.
The branch additionally integrates the bounded exhaustive worker implementation,
shared x86 address/memory roles, complete packed integer comparison grammar/
effects, and the complete named scalar LOCK RMW family. Conflict resolution
preserves the unsigned-average consumers as well as comparison consumers.
Main is the sole assembly-ledger writer.

The full-root replay found the direct MRS arity fixture still used AMD64.
Its test-only repair now binds every case to its explicit architecture and
owns/disposes its LLVM context. The ARM64 result-tail audit is also integrated;
it deliberately grants no padding or caller permission. Fresh combined gates
are required for these later commits.

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
- `927e6d6e`: the complete corpus-driver package passed all 32 race partitions,
  548 roots / 1,904 children, no Skip or missing outcomes, in 41m49.664s.
- `879feae6`: both SDKs passed seven address/memory-role roots / 2,461 children.
  Real failures exposed shared classification, direct symbolic pointer sinks,
  and x87 integer-load bypasses; all legitimate memory object controls remain.
- Comparison leaf `9d7755fc`: both SDKs passed 85 roots / 8,739 children;
  main independently rechecked the frozen receipts before integration.
- LOCK leaf `b68e4d97`: both SDKs passed 101 roots / 5,612 children, no Skip;
  main's separate physical collector rechecked all 24 closed command stages.
  The leaf's before/after tool guard covers 33 selected inputs, not the full
  SDK. The new integrated replay must supply full SDK guards independently.

These are revision-specific mechanism/package gates, not full current-head
module, native Linux, external-library or ledger passes. Ignored `_out/`
artifacts retain raw JSON, original failures, complete source/SDK/tool manifests,
actual environments, command receipts and independent reader results.

## Parallel family work

Existing agents own separate persistent leaves:

- Unnamed x86 TYPE_ADDR: bind the actual frontend before encoding. Go's
  MOVL/Q progedit can rewrite to LEA, while narrow/other immediate roles differ;
  literal normalization must not grant memory or caller permissions.
- Module-wide exhaustive CI: preserve the actual package and test inventories,
  original flags/deadlines and complete package coverage; avoid replacing the
  corpus timeout with an overlong single job.
- Remaining strict standard-library contracts: reproduce the actual complete
  five-target failures, then repair only source-backed mechanisms. Caller,
  closure-register, runtime-entry and result-padding permissions remain strict.

Review each entire final diff and independently recheck its frozen evidence
before integration. Do not cherry-pick a still-running source or infer combined
success from separate leaves. Required Linux native original-Go/LLVM mathematical
and fault counterparts remain distinct from Darwin/Rosetta diagnostics.

## Full-suite failures and required work

Original frozen source `7a9e69ac` completed 32 root partitions with 4,634 listed
roots: 4,583 passed, 11 failed, 40 skipped. The failed architecture fixtures,
mode32 vector forms, CPU IR assertions and missing separate LLVM 22 wasm linker
have dedicated repairs. Preserve the original failed report.

Its unpartitioned corpus-driver race package expired at 45 minutes. Its separate
547-root partition replay retained a real CPP fixture failure. The repaired
927 snapshot's complete 548-root race replay passed; that does not promote
either old failed run or certify a later combined source.
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
