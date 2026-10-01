# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git; never edit or rebase a tree with an active verification run.

## Authority and remote state

Repair batches belong in fork Draft PR 4, targeting
`codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to `cpunion`.
Do not promote its upstream-connected head until current-head fork CI, review
and exception-policy gates pass. Upstream PR 40 remains Draft.

The last inspected fork run is `36662878534` at `3aeb6b7c`: discovery shard 39
failed Intel's raw UMONITOR decoding, then aggregate verification failed.
Subsequent local commits are not pushed. Keep scan/coverage funnel tables in
the PR body, derived from validated evidence rather than console PASS lines.

## Evidence boundaries

The imported Index ranges are continuous from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. That interval is consumed, not the whole Index or later
incremental updates. Failed inspections remain retries. Do not rescan/import
records while verification uses this immutable input. Standalone inventory and
cgo records remain outside this repository.

Historical source `754cbebb` completed all 64 shards with five failures;
publisher `3c32ecc3` preserves those outcomes. Complete accounting is not
verified success. Read-only exact-download recovery evidence does not prove
translation, and a Go compiler crash is not source N/A.

The frozen `codex/pr40-validation-20261001` tree is at `6b0c20a4`. Its full root
run had one failing LEA fixture: the fixture constructed an immediate FP address
despite spelling a Go non-immediate operand. The subsequent fixture repair
parses real Go syntax and preserves Go-rejected immediate-FP negatives.
Its five-architecture enumeration gate passed; its strict standard-library and
benchmark gates did not. Its discovery replay also exposed proof/scope defects.
These reports are diagnostic historical inputs, not evidence for later commits.
Never relabel, mix revisions, or copy their passes into the current ledger.

## Integration and active independent work

Development branch: `codex/pr40-guarded-integration-20261001`.
Important completed local repairs include declaration-backed ABI0 frames, exact
bool FP storage, raw return-width and bounded static-read contracts, ordinary
schema-10 profile consumers, shared-file custom-tag scope closure, authenticated
proxy metadata for legacy ZIPs, and actual data/BSS object generation.

Ordinary noninstrumented profile planning no longer pairs an otherwise normal
assembly file with an impossible race/msan/asan-only Go partner. Actual
instrumentation-only assembly still needs its own driver contract.
Macro-only files require successful same-scope Go assembly with a real object
and empty `-S` listing, plus actual LLVM output; no-TEXT is not a silent pass.
The default owned build cache exists before actual Go subtool-route capture.

Source-order active-include preprocessing and typed complete `go_asm.h`
generation have Go assembler/compiler oracles for all five architectures.
Generic-alias fixtures use a separate Go 1.24 language-version file; common
header/instruction tests still execute on Go 1.20, without missing-tool skips.
The discovery producer/consumer/offline generated-header binding is being
developed independently on `codex/generated-go-header-profile-proof-20261001`.
Metadata-only header capture must never count as translation PASS.

ARM64 typed prefetch/DCZID effects now retain exact native-emission proofs.
The independent `codex/pr40-arm64-dc-zva-effects-20261001` branch additionally
models hardware-sized ZVA stores, with complete named/raw grammar, conservative
continuation-memory invalidation and actual Go/LLVM runtime oracles. Review its
original/normalized private-address gates before integration; no inferred
granule, noalias or privileged-operation exemption is permitted.

## Remaining gates and contracts

Fresh current-source exhaustive tests and all 64 reports are still required.
Generate evidence only from clean stamped tools and frozen source/scan inputs;
use the automatic assembly-ledger writer after its shared audit succeeds.
Keep failed and pending candidates visible; only complete zero-failure coverage
may become verified. Release owned caches after every writer exits.

The strict benchmark still exposes dynamic reflect FP frames and native entry/
continuation contracts. ARM NAME_AUTO/local-frame and outgoing-call companions
are not integrated; hidden ARM64 closure registers, private tail entries and
runtime context-switch functions also need real caller/effect contracts.
Never invent FP allocation sizes, signatures or constant register values, or
turn these translation failures into source N/A.

## llgo user-regression contribution

Branch `codex/pr40-user-package-e2e-20261001` uses a local replacement only for
development. Before creating the contribution, pin `go.mod` to the remotely
resolvable final fork plan9asm head; never commit a local path or invented version.
Exact regressions live in the nested `test/asm` module.

Actual Linux/AMD64 Go execution passes all six issue libraries. Actual llgo
execution passes huff0, CRC32, go-hex, websocket and modernc/libc Uint128.
Websocket's public mask uses Go in the reported version; its separately named
private-kernel oracle executes assembly. Unavailable CRC AVX512/go-hex AVX
branches are not claimed as executed. None of this is final pinned CI evidence.

Purego still fails at `syscall15X`: its native C/g0 calls, callbacks, closure
bridge and fixed callback-address table need a broader explicit runtime contract.
The user has been asked whether that bridge belongs in this or a separate PR.
The whole ARM64 llgo translator suite also retains its two R26 closure failures
and a memmove continuation fixture failure. Do not delete symbols, skip these
failures, or claim all reported issues fixed. Provisional native-byte-layout/JIT
exceptions still require review before promotion.
