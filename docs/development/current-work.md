# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git. Never edit, rebase or import records into a running verification tree.

## Contribution boundary

Repair batches belong in fork Draft PR 4, targeting
`codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to `cpunion`.
Do not promote its upstream-connected branch until current-head fork CI, review
and exception-policy gates pass. Upstream PR 40 remains Draft.

The last inspected fork run is `36662878534` at `3aeb6b7c`: discovery shard 39
failed Intel's raw UMONITOR decoding, followed by the aggregate. Subsequent
repairs need a new run; old successful jobs are not current-source evidence.
Keep scan/coverage funnel tables in the PR body, derived from audited reports.

## Evidence boundaries

Imported Index ranges are continuous from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. This interval is consumed, not the whole Index or later
incremental updates. Failed inspections remain retries. Standalone inventory
and cgo records remain outside this repository. Do not resume scans while the
current assembly verification consumes this immutable ledger.

Historical source `754cbebb` completed all 64 shards with five failures. Later
`6b0c20a4` reports exposed ordinary profile, generated-header and scope defects.
Both are historical diagnostics, not passes to copy into a new ledger.
Only matching, clean stamped tools and complete current-schema reports may
update the assembly snapshot. Metadata capture alone is never translation PASS.

Clean frozen source `b7938b27` passed the full root suite in about 20 minutes
38 seconds, all five architecture instruction-table gates, and both nested CLI
suites. These results prove that revision, not later producer or instruction
changes. Its unmodified strict standard-library and benchmark gates failed.
The separate exhaustive object walk produced 171 successes and 13 failures;
all 44 official profiles were attempted and failed their object gates despite
zero classified unsupported instructions. These are not performance or runtime
passes. Missing contracts include dynamic reflect FP frames, native runtime
entries/continuations, hidden ARM64 closure registers and wasm packed PCs.

## Integrated repairs

Development branch: `codex/pr40-guarded-integration-20261001`.
Integrated changes include declaration-backed ABI0 frames, exact bool FP
storage, complete raw return-width decoding, bounded static-read contracts,
ordinary schema-10 profile consumers, shared-file custom-tag scope closure,
authenticated proxy metadata for legacy ZIPs, and real data/BSS objects.

Source-order CPP preprocessing and complete typed `go_asm.h` generation have
actual Go compiler/assembler oracles. Generated-header discovery now registers
raw include edges, retains legal CPU proposals when definition presence is
unknown, captures actual package metadata, independently recompiles it at the
consumer, and binds executed file/target/profile/tag scopes to offline replay.
Same-module imported Go source is checked against original ZIP bytes through
final objects. Macro-only files require an actual same-scope Go empty object
and a real LLVM object; their generated include directory is explicit.
Cross-host comparison validates both original proofs before projecting only
physical output digests. Fresh multi-host external replays remain a gate.

The automatic writer can replace only an intact legacy pending-only queue;
ordinary status readers still require v2. Unknown fields, old outcomes and
damaged shards prevent replacement. Large JSONL records round-trip without
Scanner's unrelated 64 KiB token limit. Neither migration nor a storage unit
fixture provides compilation evidence.

ARM64 typed prefetch/DCZID/ZVA lowering retains exact native-emission and Go
oracles. ZVA granule comes from the checked source protocol, never an inferred
constant. Fresh private SP call-frame backing survives only when original and
normalized whole-source proofs exclude escaped stack/FP addresses. Unknown
calls and pointer-bearing incoming FP frames remain conservative.

Typed DATA addresses use real native relocations and wasm 64-bit memory/table
relocations, signed data addends, actual LLVM 22/LLD objects and Node execution.
They never become zero placeholders or truncated low-word pointers. Source-bound
Go-mode wasm function PCs now use static low-16-bit plus table-index relocations;
immediate addresses share that contract. MOVB/H/W/D use complete unsigned memory
widths while preserving 64-bit register/constant/address values. Actual Go,
LLVM/LLD objects and first-host-memory Node checks cover this batch, including
portable current-toolchain JS helper invocation. Unknown logical-PC origins or
unrepresentable static addends remain Context. No constructors or `blockaddress`.

## Required next gates

The generated-header and wasm address batches are integrated and reviewed.
Freeze a new clean source revision. Rebuild stamped tools and run the affected
full suites, strict official corpus, benchmark and all 64 Discovery shards.
Use bounded parallel batches, separate report paths, and release owned caches
only after every writer exits. Never relabel or mix old-source reports.

Run the automatic assembly-ledger writer after its shared audit succeeds.
Retain genuine failures and pending candidates until their contracts are
implemented. Verified completion requires every shard and zero failures;
source applicability, audited skips, object compilation and runtime execution
remain separate claims. Provisional native-layout/JIT exceptions still require
review before promotion.

Reflect's `makeFuncStub()` declarations have no explicit arguments, but native
callers supply dynamically sized argument frames and closure/register state.
Do not invent ordinary `GoArgs`, FP bounds or register values to make these
files translate. ARM native-entry machinery does not supply the x86 reflect
contract. Private tails, runtime context switches and ARM64 R26 entries
likewise need actual caller/effect evidence.

## Dependent llgo contribution

Branch `codex/pr40-assembly-user-regressions-20261002` is rebased onto upstream
main `d98b43f90`. Before creating its contribution, publish the reviewed plan9asm
repair head to the allowed fork and resolve its version with Go. Pin a public
`replace` to that exact head, never a local path or invented pseudo-version.

Exact regressions live in the nested `test/asm` module, which root
`./test/...` does not traverse. Dedicated Linux/AMD64 and Darwin CI uses the
checkout compiler, Go 1.27.1 and LLVM 22, executes every source-applicable
package even after a sibling failure, and preserves actual tool/dependency
provenance. Linux also requires pinned-QEMU ARM64 memmove execution.

Actual Linux/AMD64 Go execution passed all six issue libraries. Diagnostic
llgo execution passed huff0, CRC32, go-hex, websocket and modernc/libc Uint128.
Websocket's reported public path uses Go; a separately labeled private-kernel
oracle executed assembly. Unavailable CRC AVX512/go-hex AVX branches are not
claimed executed. None is final pinned CI evidence.

ARM64 memmove now passes actual Go and LLVM/C overlap/continuation execution,
including required Linux/QEMU execution. The two R26 closure-entry failures
remain in the whole translator suite. Purego still fails at `syscall15X`:
native C/g0 calls, callbacks, closure representation and fixed callback-address
tables need an explicit broader contract, not a guessed Go declaration.
Keep these failures visible and the contribution Draft.
