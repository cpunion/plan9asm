# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit a running source, tool or ledger snapshot.

## Contribution and CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch before current-head fork CI and
review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft.

The last published head is `47ea649d`. Completed run `36942816488` failed all
six priority discovery shards and the aggregate. Policy preflight passed;
downstream skipped jobs did not pass. Preserve failure-first scheduling:
shards 5, 29, 36, 38, 39 and 42 must all pass before remaining jobs run.
The aggregate still requires all 64 fresh reports. No repair batch has been
pushed since that run.

## Integration and ownership

Continue from `codex/pr40-verification-checkpoint-closed-20261007`.
Main owns integration and is the only assembly-ledger writer. Earlier agents
are inactive; inspect process ownership before taking over any artifact.

The parent retains reviewed typed x86 source/memory roles, scalar FP bit views,
32-register vector masks and bounded CFG payload; ARM64 scratch-call closure
uses unchanged Go sources, actual package selection, symabis, compiler objects
and SDK bodies. Source-wide GP/NZCV/R27 checks reject hidden incoming state.
Metadata alone is not ABI admission. Exact Go package prefixes use byte-wise
percent encoding, including dots in the last path component.

CPP graphs register complete sources once. Captured sources retain a private
immutable parser-built resource index; changed file/hash/original/directives
invalidate it, and restored JSON uses the lexer. Every complete original is
still hashed; units do not redundantly hash that same already-verified source.
Cancellation remains invocation-local, checks loops and source reads, and
retains expired callback deadlines without fabricating live cancellation.

Original assembly roots now have a separate first-token rejection protocol.
Only exact original bytes plus an actual selected Go assembler rejection may
establish source N/A. Included fragments are not standalone roots; comments,
newlines, semicolons, identifiers and CPP controls preserve their normal rules.
Native descriptor layouts now cover pointer, slice, string, map, channel,
function, interface and unsafe-pointer values without following backing stores
or granting allocation, memory-access or ABI permission.

## Closed evidence

Results certify only their named frozen snapshots, not this successor.
Both SDK lanes use current Go 1.27.1 or official Go 1.20.14, exact LLVM 22,
all eight root/nested dependency source directories and actual inventories.
Independent readers rehash physical inputs and check command exits, SDK,
environment, PID/PGID closure and every root/subtest outcome.

- `e98a7342`: affected root/corpus aggregate 195 tests and 5,733 children per
  SDK; race and full root-module vet/build passed, no Fail or Skip.
- `f8ff1d9f`: ARM64/package-prefix aggregate 24 tests and 871 children per
  SDK; actual host Go and LLVM O0/O2 mathematical/full-X27 controls passed.
- `b66692de`: full CPP current 49/208, historical 47/181 tests/children.
  Immutable resource-index RED `e2738630` remains checked.
- Deadline RED/green `4dedd225`/`8f9099f9`: 3 tests/3 children per SDK.
  Planning-read RED/green `c012be23`/`e9b99c7b`: 12/27 per SDK.
- `f63317e4`: full CPP plus entire gotoolprofile package: current 251/1,534;
  historical 228/1,323. One-pass digest RED `04b95176` preserves the genuine
  relative allocation failure. Query microbenchmark allocations fell from
  about 65,900 to 117 and query time from about 4.82 ms to 0.32 ms; these are
  microbenchmark observations, not full-job or external-library passes.
- First-token `8aa7d134`: full CPP/PreCPP current 58/255, historical 52/199.
  Independent reader SHA-256:
  `a17765ab93dacfd5bf510e6941aa29814606aefba8187dec5465fae5d03d30a8`.
  Actual Ace v1.44.0 replay closed successfully with 22 Go-rejected source
  configurations, one N/A candidate, zero translations/passes; 95 other shard-5
  candidates remain pending. This does not update the assembly ledger.
- Descriptor RED `7c9183eb`: both SDKs reproduced two failing roots/ten
  architecture children after actual Go compiler objects were produced.
  Descriptor green `d399a5d0`: entire gotoolprofile package current 204/1,330,
  historical 183/1,146, no Fail or Skip, race, module vet/build and 8 closed
  stages per SDK. Independent reader SHA-256:
  `89ef4d98d03bb2daf770faa1d3f0af31d741338ba28c6d416c1b078b1a08a056`.

Four-candidate shard-5 replays `f547724d` and `b66692de` have both terminated:
one pass and three failures, not complete-shard success. The newer replay took
23m12s versus 70m44s for the older run; exact source/tool guards differ and this
is not a current-head benchmark certificate. Arrow retains a >64 source-required
CPU-profile planning failure; parquet retains original native FP/ABI obligations.
Ace's former wrong-stage parser failure is addressed separately above.

An older `e98a7342` replay runner remains orphaned and unclosed. Never fabricate
its original exit receipt or kill it merely because its controller exited.
Preserve the unrelated 64-shard runner and its descendants/caches as well.
Combined `f547724d` corpus producers omitted four current-SDK nested dependency
guards; their successful exits are not independently certified passes.

## Remaining work and completion gates

Next close ordinary native-header imported-declaration query domains using
actual source/import/module/compiler roles. Start with the builtin unsafe
property family; do not raise CPU budgets, erase proposals, borrow instrumented
authority or relabel unresolved layouts as source N/A.

The latest strict standard-library benchmark on `f8ff1d9f` is still failed:
184 attempted, 172 passed, 12 failed, zero N/A or unattempted. Wasm passed 19/19.
Remaining classes include dynamic reflect/caller frames, runtime incoming and
continuation state, ARM64 observable result padding and exact SDK body closure.
Five official classification gates had zero unsupported/parser errors; Context
classification is not ordinary lowering, object or native-runtime coverage.

Freeze/rebuild the successor, then run every root-module package, both nested
CLIs, race, all five official gates, strict benchmark, all 64 discovery shards
and required native runtime jobs. This integration has not passed those whole
gates. Preserve actual denominators and original deadlines. LLVM 22 is mandatory;
wasm must never receive blockaddress.

## Inventory and ledger

Committed continuous Index coverage remains 2025-10-01T00:00:00Z through
2026-09-22T22:43:08Z: 18,758,534 records, 965,347 unique paths, 952,900 inspected,
16,390 retries and 6,067 assembly-bearing latest-per-path versions.
Newer standalone inventory is unsynchronized; cgo stays outside this repository.
No current-head external passes have been promoted in this repair batch.

Only the automatic writer may promote independently validated fresh reports
with matching source/ledger/tool provenance. Missing shards stay pending,
failures stay failed, reviewed skips retain exact reasons. Never copy stale
passes or import into a live ledger. Independently read writer output and keep
funnel tables in the PR body. Remove only owned caches after all consumers exit.
