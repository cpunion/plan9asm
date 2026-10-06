# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit a running source, tool or ledger snapshot.

## Contribution and published CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch before current-head fork CI and
review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft.

Published source `47ea649d` failed all six priority discovery shards and the
aggregate in completed run `36942816488`. Its policy preflight passed;
downstream skipped jobs did not pass. Preserve failure-first scheduling:
shards 5, 29, 36, 38, 39 and 42 must all pass before remaining jobs run.
The aggregate still requires all 64 fresh reports.

## Current integration

Continue on `codex/pr40-go-rule-final-validation-20261007`.
Main owns integration and is the only assembly-ledger writer. Earlier separate
agents are not active; inspect process ownership before taking over an artifact.

The parent `e98a7342` contains the reviewed shared typed x86 families, strict
source/memory roles, pointer identity without pointee permission, scalar FP
bit views, complete 32-register vector masks and bounded CFG payload. Private
scalar views do not inherit the public pointer-input domain.

This successor adds ARM64 scratch-call closure from actual unchanged Go source,
package selection, ABI declarations, compiler-produced objects and SDK bodies.
Source-wide native GP/NZCV/R27 CFG checks reject hidden incoming state.
Same-package stack bytes and generated frame restores remain independently
bounded; SDK entry seeds require the real complete ABI0 wrapper/body tuple.
Old `gensymabis` local names are canonicalized exactly as the selected compiler
does and must match captured original definitions. Object metadata alone is
never ABI admission. Both root translation and the actual CLI use the closure;
owned captures are verified late and closed.

Main reviewed the complete ARM64 aggregate and retained original false-admission
and false-rejection regressions. Integration preserves the parent's x86 source
normalizer and all runtime-oracle lifecycle consumers.

CPP consumption now lexes each complete original once per graph construction,
using an invocation-local resource index across all units. No persistent cache,
unit/resource exclusion or new source permission is introduced. Every new
consumer still independently validates its complete original source hashes.

## Closed evidence and live work

These results belong only to the named frozen snapshots, not this successor:

- `e98a7342`: both current Go 1.27.1 and official Go 1.20.14 passed 195 affected
  root/corpus tests and 5,733 children, race, full root-module vet/build and
  independent physical source/tool/environment/process checks. No Fail or Skip.
- ARM64 `1b0934c9`: both SDKs passed all 21 affected root tests and 317 children
  with race, full module vet/build and independently checked frozen receipts.
  Native host tests executed unchanged Go and LLVM O0/O2 mathematical/full-X27
  canary controls. This is not the required native Linux counterpart or an
  external-library pass.
- Original merged x86 domain/budget failures and old-SDK symabis failures remain
  under ignored artifacts in their own frozen leaves.
- CPP `94c662b1` independently reproduced quadratic allocations on both SDKs.
  Repair `2f4d82da` passed the two independent complexity/freshness tests on both.
  Its 16-unit query-only benchmark reduced current-SDK time from 73.29 ms to
  5.22 ms and allocations from about 1.05 million to 65,862. Historical SDK
  time fell from 91.49 ms to 6.41 ms. These are microbenchmark observations,
  not external-package, whole-job or ledger passes.
- A separate frozen `e98a7342` four-candidate shard-5 replay is still running.
  Its controlling agent exited while its actual runner remained alive. Inspect
  PID/PGID, raw reports and closure evidence; a console PASS or missing owner
  receipt cannot be promoted into a ledger pass.

Arrow's sampled older runner repeatedly lexed the entire CPP source graph inside
per-unit/profile replay. Run a fresh complete-target candidate replay with the
repair, preserving the older evidence. Do not mutate the running older replay.

## Required completion gates

Freeze and rebuild the complete successor before combined verification.
Run every root-module package, both nested CLIs, affected race gates, all five
official architecture gates, strict standard-library/benchmark, all 64 discovery
shards and required native runtime jobs. Preserve original deadlines and actual
inventories; use the exhaustive runner's bounded workers, not an enlarged
single-job timeout. The current integration has not passed these whole gates.

Earlier complete root/corpus-driver passes do not certify this successor.
Strict benchmark gaps still include dynamic reflect/caller frames, runtime
entry/continuation state and observable ARM64 result padding: never invent
backing, initialization or padding. Official opcode classification is distinct
from encoder-form, object, native-runtime and external-library coverage.
LLVM 22 is mandatory; wasm must never receive `blockaddress`.

## Inventory and ledger

The committed continuous Index interval remains `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. Newer standalone inventory is unsynchronized and cgo
stays outside this repository. Never import into a ledger used by a live gate.

Only the automatic writer may promote independently validated fresh reports
with matching source/ledger/tool provenance. Missing shards stay pending,
failures stay failed, reviewed skips retain exact reasons, and inspection
failures stay retries. Never copy revision-stale passes. Independently read
the writer's result. Keep funnel tables in the PR body, not this checkpoint.
Remove owned package/build caches only after every writer and descendant exits.
