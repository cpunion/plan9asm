# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit a running source, tool or ledger snapshot.

## Contribution and CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch before current-head fork CI and
review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft.

The published head remains `47ea649d`. Completed run `36942816488` failed all
six priority discovery shards and the aggregate. Policy preflight passed;
downstream skipped jobs did not pass. Shards 5, 29, 36, 38, 39 and 42 must all
pass before remaining jobs run; the aggregate still needs all 64 fresh reports.
This local repair batch has not been pushed.

## Integration and ownership

Continue from `codex/pr40-verification-checkpoint-bitfield-20261007`.
Its production ancestor is `529019e1`; ledger-only successors `70a2182d`
and `50202ae5` share its non-derived semantic source. Reports certify their
named source, never an edited documentation or implementation successor.
Main is the only assembly-ledger writer. Earlier agents are inactive.

Retained repairs include typed x86 source/memory roles, scalar FP bit views,
32-register vector masks, bounded CFG payload, and exact Go package-prefix
encoding. CPP graphs retain immutable parser-built indexes and hash every full
original. Deadlines and cancellation are invocation-local. Source rejection
requires original bytes and an actual selected Go assembler diagnostic;
included fragments are not standalone roots.

The ordinary and native scratch producers now share token CPP and Go line
grammar. Complete label prefixes, optional suffixes, pending labels and original
aliases survive parsing. CFG synthetic roles are explicit and collision-free.
Consecutive empty original-label blocks fall through without inventing an exit,
register fact, frame or ABI permission. All text/module/GoModule tests consume
the actual modified alias source, not an unchanged baseline.

The latest fix closes the complete scalar ARM64 SBFM/BFM/UBFM encoding family:

- One typed decoder validates W/X width, reserved fields, rotate and mask ranges.
- Its physical GP effects drive R27 scratch, native input and pool-write checks.
  BFM reads its old destination; SBFM/UBFM overwrite it. Register 31 means zero,
  never SP; native input closure still rejects LR effects.
- NZCV preservation uses the same physical grammar, not an alias-name whitelist.
- Go's scalar `MOVW R1,R4` becomes `SXTW X4,W1`. Lowering already supported
  it, but a separate native-effect whitelist rejected it. This consumer drift,
  not an absent opcode, explains why official lowering coverage missed the bug.

Native-header properties cover scalar/descriptor layouts, builtin unsafe,
bounded receivers and concrete generic instances. Uninstantiated generic
declarations emit no value layout. These facts grant neither imported queries,
pointee access, instrumented authority nor native calling contracts.

## Closed evidence

Both SDK lanes use current Go 1.27.1 or official Go 1.20.14 with exact LLVM 22.
All eight root/nested dependency directories, SDK tools/sources, LLVM headers
and required runtime/linker inputs are guarded. Independent readers physically
rehash inputs, verify full environments, genuine exits and PID/PGID closure.

- Alias RED `6a2faacd`: actual Go objects accepted, 10 expected failed nodes
  per SDK. Green `88d2fb85`: full native scratch, 23 roots/257 children.
- Scalar native-input RED `37324b37`: four real source fixtures demonstrate
  the same SXTW effect gap, without inventing pointee or runtime authority.
- Bitfield RED `b394f8fb`: 40 actual Go encoding rows, 33 rejected and seven
  controls passed per SDK. Complete base/alias/extension/shift formats are
  derived independently from Go encoding cases.
- Green `529019e1`: affected native scratch, bitfield and raw-pool scope,
  146 roots/2,282 children per SDK, no Fail or Skip; race, root-module vet/build.
  Includes affected Go/LLVM objects, host runtime oracles and 138,240 synthetic
  physical encoding checks. Synthetic checks are not individual object proofs.
  Reader SHA-256:
  `04f4d4a6ce3abcf9f98caa60c4afa185b295cc4f4909eaf750f424289bd0f903`.
- Current five-architecture classification on `529019e1`: 9,460 formats,
  9,201 supported, 259 Context, zero Unsupported or parser errors. Context is
  not ordinary lowering, object or runtime coverage. Historical Go 1.20 retains
  its reviewed single AMD64 VADDPD.Z source-format baseline exception.

Three original Snappy versions passed all applicable source/profile/LLVM object
scopes on `529019e1`: golang/snappy v1.0.0 and google/go-snappy v1.0.0 each
compiled 12 objects across six targets; Beeketing/snappy v0.0.1 compiled six
across three. The full 11-target denominator remains recorded. Each producer
closed 11 stages and its independent reader passed; no third-party execution
is implied. Provenance and physical translator/runner hashes match.

The automatic publisher and separate readback closed eight stages and retained
3 passes/6,064 pending, zero failed records, three partial shards (5/34/59).
Independent reader SHA-256:
`696d52740068d4a8b0967deb01fb1942ccf63ba6b3cb695e99597535a74b4008`.
No complete shard or whole-corpus pass is claimed. Non-derived edits invalidate
current-head freshness: rebuild/replay before publishing another ledger.

## Next work and completion gates

Frozen `529019e1` is replaying all 96 candidates of priority shard 5 under
`_out/priority-shard5-bitfield-replay`. Obtain the terminal owner and independently
read the actual reports. Explicit candidate scopes remain partial reports; do
not infer whole-shard/aggregate completion. Distinguish original asmdecl errors
from translator errors; never convert unknown failures to source N/A.

Remaining mechanism classes include ordinary imported-header query domains,
instrumentation/foreign-type source closure, FFI/JIT state and calling contracts,
Arrow's greater-than-64 CPU-profile planning domain and parquet FP/ABI obligations.
Do not raise budgets, erase proposals, copy old passes or fake native authority.

The latest strict standard-library benchmark remains failed on `f8ff1d9f`:
184 attempted, 172 passed, 12 failed, zero N/A or unattempted; wasm passed 19/19.
Dynamic reflect/caller frames, runtime incoming/continuation state, observable
ARM64 result padding and exact SDK body closure remain.

Whole root/nested CLI suites, all five official gates, strict benchmark,
64 discovery shards and required cross-runtime jobs have not all passed this
integration. Preserve the old orphaned `e98a7342` runner and unowned caches;
do not fabricate an original exit. Remove only owned caches after all consumers
close. LLVM 22 is mandatory; wasm must never receive blockaddress.

## Inventory and ledger

Committed continuous Index coverage remains 2025-10-01T00:00:00Z through
2026-09-22T22:43:08Z: 18,758,534 records, 965,347 unique paths, 952,900 inspected,
16,390 retries and 6,067 assembly-bearing latest-per-path versions.
This is the committed inventory, not a fresh server query. Newer standalone
inventory is unsynchronized; cgo stays outside this repository.

Only the automatic writer may publish matching fresh source/ledger/tool reports.
Missing candidates/shards remain pending, failures remain failed, and reviewed
skips retain reasons. Never import into a live ledger. Keep scan/coverage funnel
tables in the PR body, not these guides.
