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

Development branch:
`codex/pr40-closed-ci-class-integration-20261007`.
Read its actual HEAD/status. Root alone writes ledger. This batch integrates
the independent repairs below; ancestor receipts do not certify this new head:

- `cf3d5ef6` unifies MOVL/MOVQ/Go-MOVD/VMOVD/VMOVQ grammar, original-source
  validation, CFG routing, private FP result initialization, lowering and caller
  byte effects. Actual Go-backed RED preceded implementation. Current and old
  Go diagnostics each pass 63 roots and 2,634 children without Fail/Skip,
  including the formerly failing FloatMulDiv/MXCSR root. Keep complete table
  epochs, all register/address axes, FP holes, native dead-read faults and
  whole-byte/flags/canary oracles. Portable VEX is not physical AVX512 proof.
- Bounded multiple-SCC composition integrates the independent `5fe8133f` RED
  and `f64afdb8` repair. Each reducible SCC keeps the existing exact induction,
  immutable bound, unique entry/latch, no-early-exit and visit-budget contract.
  Independent current/old diagnostics each pass 62 roots and 958 children;
  these are ancestor evidence, not combined-head passes. Do not enable a
  general cyclic hull or extrapolate to real dynamic-stride/sentinel loops.
- `ca9df057` reconciles escaped CFG labels and exact earlier diagnostics,
  keeping stronger edge/store assertions. Ordinary Go/SDK object observation
  still supplies no native-memory Binder. Dead or unused opaque pointer
  transport is distinguished from unauthorized numerical result consumption.
  Current/old diagnostics each pass 238 nodes. The empty final-label positive
  actually compiles, links and calls its DATA-retained native function.
- `3cd3bdc7` RED and `664eda94` repair cover source MOVL/Q -> LEAL/Q in an
  unframed, stack-unobserving raw near-RET leaf. Both complete views recheck
  return/memory constraints; no stack/caller/fixed-local/raw-text or sibling
  decoder certificate is inherited. Current/old diagnostics each pass 5,412
  nodes, including original Go bytes and LLVM numerical execution at both
  backend optimization levels. The private custom ABI is numerical only.
- `cefdce60` RED and `8f7dcb84` preserve the actual private raw FS/DS/ES/RIP
  operands and carry x87 environment/BCD pointer address spaces through all
  loads/stores/GEPs. The public parser, raw-return and extent guards stay
  strict. Eighteen unused nonignored segment controls remain Context/zero IR;
  named FLDENV's actual stock-Go backend defect is not source-N/A evidence.
  The independent current formal scope passes 193 roots/5,018 children;
  obtain the actual old terminal receipt and independent review/readback.

The full frozen `ad570209` root suite completed all 32 partitions: 23 roots
failed and 39 existing host-inapplicable roots skipped. It is a failure baseline,
not a full CI pass. All 23 failures now have focused repairs, not a passing
successor full-root certificate. Frozen `6c98f32e` passes independently read
current/old GP/SIMD integration scopes: 7,377/5,566 nodes respectively, with
0 Fail/Skip. Its separately widened SCC safety scope passes 1,020 nodes on
each SDK. The narrower SCC sub-scope is 337 nodes, never relabel it as 1,020.
Race, complete source/tool/dependency/environment guards, vet/build and real
owner/PGID closure passed. These scopes overlap; do not sum them as unique
tests or whole-corpus coverage. Microbenchmarks are not the strict stdlib gate.
Freeze this successor, rerun the exhaustive root/nested suites, five-arch
official gate, strict benchmark and original-library replay before publishing.

## Integrated ancestor and remaining source proofs

Frozen `ad570209` integrates the earlier ARM64 scratch/input and transitive
same-package call proofs, CPP companion/header consumer closure, full ordinary
Go-body/SDK-type frame observation, compact original CPP resource index, and
complete eight-name mask/vector B/W/D/Q caller and alias effects, plus the
complete immediate-shuffle grammar and caller effects.

Its independently read current scopes pass 4,343 nodes (full CLI 1,804,
ARM64 1,264, CI policy 7, mask family 346, shuffle 507, YMM/MMX 415).
Its old formal scope failed on the Go 1.20 maskless-zeroing encoder epoch;
`6108450a` fixes the independent fixture, preserving reserved-form rejection.
Race, complete source/tool/dependency/environment guards, vet/build, nested
CLI checks and genuine producer/process-group closure passed. Its official
five-architecture classification/fingerprint gate passed, not complete encoder
or execution coverage. CPP resource microbenchmark passed, not the full stdlib
benchmark.

Original Weaviate v1.39.6 replay at `ad570209` still fails ten AMD64 files:
four first errors moved from VPSHUFD to VMOVD; six remain cyclic joins.
The bucket/version still fails and admits no passing object certificate.
Companion CPP mismatch and mask/vector errors no longer appear in that replay.
Partial file/object success is not a passing version.
Retain failed reports and rederive every successor receipt from its own source.

The source-only cyclic audit authenticates a fresh exact Weaviate download and
retains seven function/SCC projections. The older failed report lacks per-file
hashes; a fresh download cannot invent them. Whole-function unique-backedge
proof is too narrow for these functions. Compose only independently proved
bounded SCCs first; dynamic indexed/cursor/sentinel loops still need their own
machine-width relational and pointer-region proof. Never enable unbounded
cycle hull merges, increase caps or discard original return/stack checks.

## Running corpus, ledger and inventory

Frozen `529019e1` owns `_out/priority-shard{5,29,36}-bitfield-replay`.
Shard 5 closed after 96 explicit candidates and remains partial; shard 29
closed after all 106. Both failed. Shard 36 is still running at this checkpoint.
Read actual owner/process handles and `closure.json`; the old status utility's
missing `owner.result.json` check is not producer closure evidence.
Obtain terminal producer and independent-reader results before promotion.
Historical reports cannot certify later implementation or documentation heads.

Committed assembly ledger remains three independently published Snappy passes,
6,064 pending, no complete shard, `complete=false, verified=false`.
Only the automatic publisher may update exact matching ledger records.

Inventory retains 215 continuous ranges from 2025-10-01 through
2026-09-22T22:43:08Z and 6,067 assembly-bearing latest-per-path candidates.
Newer standalone inventory is unsynchronized; cgo stays outside this repository.
Scan/coverage funnel tables belong in the PR body, not this checkpoint.

## Completion gates still open

Repair/replay original libraries: native SP/caller backing, imported/generated
headers, instrumentation and foreign-type closure, greater-than-64 profile
planning, native FFI/JIT entries and remaining dynamic loop effects.
Unknown errors are not source N/A; proposed exclusions still need review.

Latest strict stdlib benchmark `f8ff1d9f` failed: 184 attempted, 172 passed,
12 failed; wasm 19/19. Dynamic reflect/caller frames, runtime continuations/TLS,
ARM64 result padding and SDK body closure remain.

Full root/nested CLI tests, official gates, strict benchmark, all 64 discovery
shards, cross-runtime jobs, exclusions review and current-head CI are incomplete.
Clean only exact owned caches after all consumers close; preserve unowned
orphans/shared/live caches. LLVM22 only, no wasm blockaddress or tool Skip.
