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

Continue on `codex/pr40-verification-checkpoint-20261007`.
Main owns integration and is the only assembly-ledger writer. Earlier separate
agents are not active; inspect process ownership before taking over an artifact.

The parent `e98a7342` contains the reviewed shared typed x86 families, strict
source/memory roles, pointer identity without pointee permission, scalar FP
bit views, complete 32-register vector masks and bounded CFG payload. Private
scalar views do not inherit the public pointer-input domain.

The integration adds ARM64 scratch-call closure from actual unchanged Go source,
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

Exact native Go package prefixes now join the selected package, symabis and all
compiler/assembler archive bodies. Dots in the last package-path component use
Go's byte-wise percent grammar; normalization never decodes unknown packages,
function names or ABI selectors.

CPP consumption registers each source once per graph construction. Captured
sources also retain a private immutable parser-built String-resource index;
file, hash, full original and directive mutations invalidate it. Restored JSON
uses the lexer. Consumers retain independent maps and complete source-hash
validation. This introduces no portable authority or excluded unit/resource.

Expired operations now retain their deadline even if a callback returns nil or
an unrelated source error. Ordinary/mode CPU planning receives invocation-local
contexts, checks loops and source reads, and preserves deterministic offline
replay. Never retain setup contexts in plans reused by later target budgets.

## Closed evidence and live work

These results belong only to the named frozen snapshots, not this successor:

- `e98a7342`: both current Go 1.27.1 and official Go 1.20.14 passed 195 affected
  root/corpus tests and 5,733 children, race, full root-module vet/build and
  independent physical source/tool/environment/process checks. No Fail or Skip.
- ARM64 `f8ff1d9f`: both SDKs passed all 24 affected tests and 871 children,
  including the shared package-prefix byte grammar, race, module vet/build and
  independent frozen-input checks. Native host tests executed unchanged Go and
  LLVM O0/O2 mathematical/full-X27 controls, not the native Linux counterpart or
  a complete external-library version.
- Original merged x86 domain/budget failures and old-SDK symabis failures remain
  under ignored artifacts in their own frozen leaves.
- CPP `b66692de`: all CPP tests passed with race and independent checks:
  current SDK 49 tests/208 children; historical SDK 47/181. Both gates bind all
  eight root/nested dependency source directories, selected SDK and LLVM inputs.
  Query-only current-SDK allocations fell from about 65,900 to 182 and focused
  benchmark time from 4.82 ms to 0.58 ms. Earlier quadratic RED `94c662b1` and
  repeated-source RED `e2738630` remain independently checked. Microbenchmarks
  do not certify external versions, full jobs or ledger passes.
- Deadline RED `4dedd225` and repair `8f9099f9`: both SDKs independently checked
  all three tests/three children, preserving actual expired-callback failures.
- Planning-read RED `c012be23` and repair `e9b99c7b`: both SDKs independently
  checked 12 tests/27 children. The RED continued five source reads after
  cancellation on all five architectures; the repair stops further reads.
- Combined `f547724d` root/corpus producers terminated successfully, but do not
  promote them: the current-SDK corpus guard omitted four nested dependency
  directories and the combined independent readback remains incomplete. Run a
  fresh combined gate with all eight directories bound before and after.
- A separate frozen `e98a7342` four-candidate shard-5 replay is still running.
  Its controlling agent exited while its actual runner remained alive. Inspect
  PID/PGID, raw reports and closure evidence; a console PASS or missing owner
  receipt cannot be promoted into a ledger pass.

Separate frozen `f547724d` and `b66692de` four-candidate shard-5 replays also
remain live under ignored `shard5-priority-replay` and
`shard5-source-index-replay` outputs. Controllers retain their actual exit
receipts; inspect them before taking over. Keep all 11 default targets, exact
candidate membership and complete-shard pending denominator. A console PASS
or initial partial checkpoint cannot update the assembly ledger.

Arrow's newest CPU sample moved from repeated String lexing to complete-source
hashing and original CPP replay. Optimize that rule-class cost without omitting
sources, profile pairs, consumer checks or source/tool provenance. Preserve
older samples; do not mutate or borrow caches from live replays.

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

The latest strict benchmark on `f8ff1d9f` failed: 184 attempted, 172 passed and
12 failed, no N/A or unattempted files. Wasm passed 19/19. The same snapshot's
five official classification gates passed with zero unsupported/parser errors;
Context classifications are not ordinary lowering or runtime passes. These
gates must still be fixed and rerun on the complete successor.

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
