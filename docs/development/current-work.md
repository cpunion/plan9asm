# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git; never edit a live source, tool or ledger snapshot.

## Contribution and CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch until current-head fork CI and
exclusions review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
stays Draft.

Published head `47ea649d` and completed run `36942816488` still fail the six
priority discovery shards (5, 29, 36, 38, 39, 42) and aggregate. Policy passed;
downstream skipped jobs did not pass. This local batch is not yet pushed.

## Continue here

The development branch is
`codex/pr40-amd64-complete-source-handoff-green-20261007`.
Read its actual HEAD and status before resuming. Main is the only ledger writer;
earlier agents are inactive.

Latest repairs:

- ARM64 logical shifted, ADD/SUB shifted and ADD/SUB extended forms share one
  physical GP/NZCV effect grammar, including WZR/XZR, shifts and SP boundaries.
- CPP resource queries index exact original edges once instead of scanning all
  resources per unit. Unknown edges fail; public replay still validates and
  hashes the whole original graph, including unconsumed sources.
- AMD64 frame observation typechecks complete ordinary Go bodies, values and
  stock-SDK imports. Only bodyless ASM declarations enter its frame table.
  SDK source/export graphs remain bound; foreign imports stay Context.
- The CLI passes every actual selected Go source and retains actual-load
  import type identities and source hashes. Failed verification revokes its
  independent binding. Pointer/slice fixed-local transport, caller regions,
  dynamic frames and native JIT/FFI contracts remain unproved.

The old observer regression expecting blanket body/import rejection failed
after the expanded observation domain. Its replacement tests accept those
independently observed numeric frames while retaining nested, foreign-module,
DATA and pointer rejection. Obtain genuine terminal results for this branch;
a test edit or an ancestor pass is not current-head verification.

## Closed ancestor evidence

These scopes certify only their named frozen revision. Independent readers
rehash full tracked source, SDK/tool/dependency inputs and environments and
check genuine exits and PID/PGID closure.

- `5395402a`: affected ARM64 native/scalar/raw-pool scope, 150 roots and
  3,094 children per SDK; current Go 1.27.1 and official Go 1.20.14 passed
  without Fail/Skip, with race, root-module vet/build and actual object controls.
- `2209057f`: expanded current-SDK CPP corpus scope, 94 roots/418 children,
  all 512 passed, including actual Go/header/LLVM consumer and mutation tests.
- `c07c03ae`: compact CPP resource index, 15 focused passes per SDK.
  Current 256-unit microbenchmark is about 21x faster than the retained RED,
  not evidence that the complete CI benchmark passes.
- `56717ae5`: full AMD64 frame/fixed-local root scope, 29 roots/115 children
  per SDK, all 144 passed. Pointer observation still fails numeric-only binding.
- `b2a75d68`: actual CLI RED has four failed children; original Go exported
  every full source. Two lack syntax handoff and two hit blanket import rejection.
- `b7b4a06d`: CLI scope passed 42 nodes; observer scope retained three failures
  (two stale blanket-rejection assertions plus their parent). Keep raw failures.
- Original golang/snappy v1.0.0 replay on `c07c03ae` compiled 12 LLVM objects
  across six applicable configurations with all 11 targets retained.

## Running corpus and ledger

Frozen `529019e1` owns the three replay directories
`_out/priority-shard{5,29,36}-bitfield-replay`. Check actual owner/process handles
and reports. Shard 5 uses explicit candidates and remains partial.
Obtain terminal producer and independent-reader results before promotion;
never infer whole-shard coverage from running counts.

Committed assembly status remains three independently published Snappy passes,
6,064 pending, no complete shard, `complete=false, verified=false`.
Documentation and implementation successors need fresh source-matching reports.
Only the automatic publisher may update matching ledger records; no hand edits.

Inventory retains 215 continuous ranges from 2025-10-01 through
2026-09-22T22:43:08Z and 6,067 assembly-bearing latest-per-path candidates.
Standalone newer inventory is unsynchronized; cgo stays outside this repository.
Scan and coverage funnel tables belong in the PR body, not this checkpoint.

## Remaining completion gates

Repair and replay original failing libraries, including actual native SP/caller
backing, imported/generated headers, instrumentation and foreign-type closure,
greater-than-64 profile planning, and native FFI/JIT entry contracts.
Never raise budgets, remove proposals or relabel unknown errors as source N/A.

Latest strict stdlib benchmark `f8ff1d9f` remains failed: 184 attempted,
172 passed and 12 failed; wasm passed 19/19. Dynamic reflect/caller frames,
runtime continuations/TLS, ARM64 observable result padding and SDK body closure
remain. Context classification is not successful lowering/object/runtime proof.

Full root/nested CLI tests, official gates, benchmark, all 64 discovery shards,
cross-runtime jobs, reviewed exclusions and current-head CI are not complete.
Preserve the old unowned orphan/caches; clean only exact owned caches after all
consumers close. LLVM 22 only; no wasm blockaddress or required-tool Skip.
