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
No repair batch has been pushed since that run. Rechecked failures include
instrumentation profiles, source applicability, FFI/JIT calling contracts and
native state, not just opcode spellings.

## Integration and ownership

Continue from `codex/pr40-verification-checkpoint-label-grammar-20261007`.
Its production parent is `879ae479`; reports certify their exact named parent,
not this documentation successor. Main is the only assembly-ledger writer.
Earlier agents are inactive. Inspect process ownership before adopting artifacts.

The inherited repairs retain typed x86 source/memory roles, scalar FP bit views,
32-register vector masks and bounded CFG payload. ARM64 scratch closure requires
unchanged sources, actual package selection, symabis, compiler objects and SDK
bodies; metadata never supplies incoming GP/NZCV/R27, a native frame or ABI.
Exact Go package prefixes use byte-wise percent encoding, including final-component dots.

CPP source graphs are registered once and retain immutable parser-built indexes.
Every complete original is still hashed. Expired deadlines and cancellation
remain invocation-local, including loops and source reads. First-token source
rejection requires exact original bytes and an actual selected Go assembler
diagnostic; included fragments are not standalone roots.

Native-header layouts now cover scalar and descriptor values, builtin unsafe
properties, bounded receiver expressions and concrete generic instantiations.
Uninstantiated generic declarations emit no value layout. These facts do not
grant imports, allocation, pointee access, native ABI or instrumented authority.
General imported-declaration query domains remain unfinished.

The new grammar repairs are independent of the discovering Snappy version:

- Parse consumes complete identifier-label prefixes, optional suffixes and
  function-local pending labels using the official Go line grammar. DATA/GLOBL
  do not consume pending labels; TEXT starts their new function; EOF ignores
  unbound labels, as the selected Go assembler does.
- CFG builders reserve every original alias before generating entry/fallthrough
  names. Synthetic roles are explicit, not inferred from a source-name prefix.
  ARM/x86 emit the actual reserved entry name. ARM labels are injectively encoded
  in the block namespace, separate from register storage and SSA temporaries.
- Opcode heads consume `Ident ('.' Ident)*` before operands on ARM and x86.
  Spacing, comments and CPP macro origins cannot turn suffixes into operands.
  Source inspectors use the same token grammar; wasm does not acquire suffixes.
- The native scratch source producer now uses the ordinary consumer's token CPP
  frontend. Whole-file and Raw equality, original source hashes, native objects,
  actual declarations, call closure and late-mutation checks remain unchanged.

## Closed evidence

Both lanes use current Go 1.27.1 or official Go 1.20.14 and exact LLVM 22.
All eight root/nested dependency directories and actual inventories are guarded.
Independent readers rehash physical inputs, verify full environments, genuine
command exits and PID/PGID closure, and reconcile every test outcome.

- `e98a7342`: affected root/corpus 195 tests and 5,733 children per SDK.
  `f8ff1d9f`: ARM64/package-prefix 24/871 per SDK; actual host Go and LLVM
  O0/O2 mathematical/full-X27 controls passed. These are historical scopes.
- `460a9184`: whole gotoolprofile package plus full CPP/PreCPP: current
  269 tests/1,695 children; historical 240/1,404, no Fail or Skip. Unsafe,
  receiver-budget and generic-property RED controls retain actual Go objects.
- Label RED `4c1c31ed`: 78 expected failed nodes and 15 controls per SDK.
  Entry/pending-label green `84a14f35`: 4 tests/115 children per SDK.
  Reader SHA-256: `dac9044b2657c9c99a5cd619cbc3d9c27999ec04637f781d14430c41adf9ad0b`.
- SSA-role RED `393c00cc`: original Go accepted the unchanged fixtures;
  ARM/ARM64 LLVM parsing failed for colliding source labels and register slots.
  Green `390c7f6f`: affected parser/CFG 64 tests/362 children per SDK.
- Suffix RED `18b5e697`: all 26 nodes failed after actual original/expanded
  Go assembler objects. Green `7ef3a3f6`: affected parser/CFG 65 tests/388
  children per SDK, no Fail or Skip, race, full root-module vet/build.
  Reader SHA-256: `aa19989b1fcdcd6c59bba47723f9db3327546d75f92820860d2b2af3827e574d`.
- CPP identity RED `370ff75c`: three actual target sources failed exact binding
  for the mismatched frontends. Green `879ae479`: full native scratch scope
  22 tests/214 children per SDK, no Fail or Skip, race, root-module vet/build.
  Reader SHA-256: `d831096f4e3bd327e6a1a2116161869cd9dfd25a175308e68d3ae252aa2ad9cd`.

Each recent focused owner closed eight stages. Header resource-query measurements
remain about 0.3 ms and 117 allocations on current Go; these are microbenchmarks,
not whole-job or library-pass certificates.

Actual Snappy v1.0.0 replays retain genuine failures, not passes. The latest
`879ae479` replay passed the former parsing/binding stages and now fails the
ARM64 native scratch observer contract at CALL emitLiteral. One candidate was
selected from bucket 34's 73 versions; the other 72 remain pending. All 11 owned
stages closed, the production progress reader and independent physical reader
passed, but the candidate failed. Reader SHA-256:
`1c48b4754541ab8865ebaa8268e44142abf3302397ca0253ff4efca2e769f044`.
No third-party execution or ledger pass is implied. Rebuildable private build
caches were removed only after consumers closed; reports, binaries and guarded
dependency/source inputs remain.

An older `e98a7342` replay remains orphaned and unclosed. Never fabricate its
original exit or kill it merely because its controller exited. Preserve unowned
runners and caches. Older `f547724d` combined producers omitted four nested
dependency guards; their exits are not independently certified passes.

## Next work and completion gates

Five-architecture official classification reruns are running on frozen
`879ae479`; obtain both terminal owners and independently read retained reports.
Context classification is not ordinary lowering, object or runtime coverage.

Continue the scratch observer failure as a complete source/native-state rule
family, and close ordinary imported-header query domains. Arrow still retains
a >64 required CPU-profile planning failure; parquet retains original FP/ABI
obligations. Do not raise budgets, erase proposals, copy old passes or turn
unknown diagnostics and unsupported translations into source N/A.

The latest strict standard-library benchmark remains failed on `f8ff1d9f`:
184 attempted, 172 passed, 12 failed, zero N/A or unattempted; wasm passed 19/19.
Remaining classes include dynamic reflect/caller frames, runtime incoming and
continuation state, observable ARM64 result padding and exact SDK body closure.

Rebuild a frozen successor, then verify every root-module package, both nested
CLIs, race, all five official gates, strict benchmark, all 64 discovery shards
and required native runtime jobs. This integration has not passed those whole
gates. LLVM 22 is mandatory; wasm must never receive blockaddress.

## Inventory and ledger

Committed continuous Index coverage remains 2025-10-01T00:00:00Z through
2026-09-22T22:43:08Z: 18,758,534 records, 965,347 unique paths, 952,900 inspected,
16,390 retries and 6,067 assembly-bearing latest-per-path versions.
Newer standalone inventory is unsynchronized; cgo stays outside this repository.
No current-head external passes have been promoted in this repair batch.

Only the automatic writer may promote fresh independently validated reports
with matching source/ledger/tool provenance. Missing shards remain pending,
failures remain failed, reviewed skips retain reasons. Do not import into a
live ledger. Independently read writer output and keep funnel tables in the
PR body. Remove only owned caches after every consumer exits.
