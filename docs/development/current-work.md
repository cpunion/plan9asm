# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git; never modify a running verification snapshot.

## Contribution and CI

Repairs are staged in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to
`cpunion`. Do not promote the upstream-connected branch until current-head fork
CI and review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
remains Draft.

Published source `0dad3b63` failed all six priority shards in run `36922137520`.
Their complete reports were audited and imported automatically in `c057aa97`.
This is historical evidence, not success for subsequent source revisions.
Skipped downstream root, standard-library and benchmark jobs did not pass.
Preserve the failure-first scheduling and all 64 required shards.

Published repair source `47ea649d` has a new fork run `36942816488`; its policy
preflight passed, but all six priority corpus shards and their aggregate failed.
The run is complete; downstream jobs were skipped, not passed. Its assembly
ledger was automatically reset to pending because old
reports have different source provenance; the original reports remain evidence.

Development branch: `codex/pr40-reviewed-producer-literal-batch-20261004`. Earlier
verification snapshots remain frozen in separate worktrees. Integrated repairs
are local, not yet the published fork head. Freeze a clean final batch before
rebuilding stamped tools and collecting new reports. Documentation changes
also change the source fingerprint. Put progress/funnel tables in the PR body,
not here; derive them with the shared report readers.

## Immutable inventory and evidence

Imported Index ranges are continuous from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. This interval is consumed, not the whole Index or later
updates. Inspection failures remain retries. Do not import or resume scans
while corpus verification uses that ledger. Standalone inventory and cgo
records remain outside this repository.

Only matching source/ledger/tool provenance may update assembly evidence.
Do not copy old passes into a new-head snapshot or mix host tools. Run the
automatic writer against fresh reports and re-read its result; missing shards
stay pending and genuine failures remain failed. Partial reports cannot meet
the completion gate. Never hand-edit status flags or report schemas.

## Integrated mechanism repairs

The batch preserves declaration-backed ABI0 frames, bounded stack/static-read
contracts, exact data relocations, wasm logical-PC contracts without
`blockaddress`, and schema-11 ordinary profile proofs.

Recent repairs cover ARM64 private-SP stores and bounded affine frame reads;
the typed ARM status-register family; source-order CPP EOF and relative-include
binding; actual ordinary-package roles for test-named assembly; and ZIP-bound
zero-byte selected siblings. Nonempty, whitespace/comment and macro-empty
assembly still requires its normal Go/CPP/LLVM proof.

The local batch also preserves complete declaration-backed code-address roles
through source aliases, x86 ID-flag state and operand-width POPF semantics, and
typed ARM64 call/result transport across the independent GP and FP banks. Live
private affine scratch that is neither an argument nor a result no longer
requires an artificial physical clear; unknown/escaping transport stays strict.
Focused original-Go, LLVM-object and runtime oracles exist, but do not substitute
for full current-head corpus and CI validation.

The complete raw ARM64 SM3 family has exhaustive decoding, independent LLVM-MC
encodings, alias/target object tests and independent Go/LLVM runtime oracles.
The original gmsm candidate passed its default-target production replay at
frozen `45f9e6d7`; scoped Go source rejections remain in the report. The separate
`96c1d58f` oracle observes every NZCV flag and exercises both values. Neither
report proves the subsequently integrated source without a fresh replay.

Bulk Go source errors retain their complete bounded raw diagnostic digest and
a canonical sample of 256 distinct positions. Sampling does not stop checks
for late infrastructure errors or invalid source positions. The original avo
candidate passed its default-target replay at frozen `ff488337`; source-excluded
scopes are not LLVM objects or runtime passes.

The next local batch includes complete ARM64 raw UDF decoding and explicit
integer source-class checks. An old unknown-WORD negative was a valid UDF;
preserve that failed report and its corrected negative/reachable-trap tests.
Per-analysis raw GP-effect reuse memoizes only complete-word, context-independent
effects. Its bounded performance improvement does not establish a passing
standard-library benchmark.

Ordinary proof handling now retains separate original pre-CPP and CGO-disabled
Fortran-loader refusals. Their exact source, tool, profile and affected-scope
witnesses never replace eligible object proofs. Actual test-only declarations
and missing direct callees use a separate typed signature source and pure-Go
test dependency closure; query-only metadata is not translation evidence.
The original gensimd failure and independently passing fixed-source replay are
preserved. Recheck these protocols together after merging, including their
mandatory anti-downgrade fields and exact supplemental consumption.

The reviewed x86 private caller-region producer and scalar consumer require
complete original typed callers and bounded, initialized per-access regions.
The actual CLI-load producer is integrated and retains its private unchanged
package selection and source/expansion guards through object compilation.
The scalar SIMD memory family and KMOV caller-region consumers are integrated,
including coherent X/Y/Z views and initialized, bounded writes. The complete
MOVSS/MOVSD/VMOVSS/VMOVSD floating-literal family follows Go's actual rewrite,
rounding and constant-pool choices before certifying caller effects. Scalar
ADDSS/ADDSD/SUBSS/SUBSD literals now retain their distinct finite conversion,
constant-pool and initialized-destination contracts. Legacy
MOVSD memory loads clear the upper XMM half, unlike register copies; independent
complete-byte Go/LLVM oracles retain the real pre-fix failures. The complete
packed-vector move family now shares one typed grammar with the lowerer and
caller consumer; independent entries cover every format, and Go-accepted `.Z`
memory stores retain their native invalid-opcode exception before operand reads.
Numerical/byte-canary LLVM oracles do not establish native AVX512 execution.
The FMA3 lowerer and caller consumer now share the complete named Go grammar:
register/memory, widths, initialized masks, merge/zero, broadcast and fixed
rounding. All arithmetic inputs must be initialized, including the old
destination in zeroing forms; scalar upper-byte knowledge remains distinct.
Fixed-rounding finite float32/float64 arithmetic uses an exact integer helper
because the actual LLVM 22 constrained-libcall path reproduced an ambient-
rounding error. Independent edge/cancellation/DAZ/FTZ/fenv tests do not establish
new NaN-payload semantics, raw FP16 fixed rounding or native AVX512 execution.
Caller control flow now admits finite acyclic joins and a unique source-bound
unit-stride loop whose original low-bit length gives at most three visits.
Both retain all-path initialized/native-width access proofs and independent
final replay; neither infers allocation extents from native input offsets.
Multiple, nested and unbounded loops remain unfinished. These fixtures do not
establish original Weaviate or native AMD64 runtime success.

The sealed stock-Go frame-source foundation captures actual object/selection
bytes and supports Verify/Close with explicit temporary-directory identity.
Its separate fixed-local binder accepts complete numeric-only declarations,
including typed void returns, only after the original source, native-byte
initialization and normalized control-flow contracts close. Pointer/slice
observations do not supply local-memory permission. Genuine private SP/BP
accesses still require original backing, not the virtual FP shadow. Unknown
native calls and dynamic caller frames need their own complete contracts.

Assembly rejection requires a plain observed analyzer exit, complete known
diagnostics and actual physical source attribution. Same basenames, unknown
outer failures and incomplete testless retries cannot supply a source exclusion.
Retry copies preserve exact replacement versions, relative origins and current
checksum inputs; their diagnostics retain an unambiguous original-file binding
after cleanup. Stock Go's dependency-facts vet cache can omit asmdecl analysis.
The local producer now uses actual Go-selected ordinary/test variant exports
and invokes the same SDK's official analyzer directly, with facts-only and
typecheck-success disabled. Complete source/tool/config/export guards and
independently observed unit exits remain mandatory; unknown, partial and cleanup
failures cannot become source exclusions. Logical directory aliases retain
their physical identity checks. Do not promote cached nil results or historical
reports as evidence for this source revision; fresh library reports are required.

Successful source/CPP guards retain the actual completed analyzer receipt;
failed guards retain both errors and cannot supply a source exclusion. Exact
successful module resolver notices are checked by stage, module and version,
not accepted as arbitrary stderr. Direct analyzer tools retain their complete
canonical experiment suffixes and independently bound SDK Flags/source bytes.
Unknown, changed or inconsistent tool identities remain failures.

Legitimate DATA/GLOBL-only files now have a distinct original-Go symbol-listing
proof, complete native object/source identity and exact parsed symbol/size
agreement. They do not acquire fictitious TEXT signatures or symbol-free
status. Actual LLVM objects remain necessary for a pass. The test-declaration
query/consumer intersection with DATA-only siblings now preserves the complete
native symbol/object/listing proof before query-only return. The final consumer
requires that same proof and unchanged normalized CPP. The mixed TEXT/test-
declaration/DATA regression produced real LLVM objects across five architectures.
The original Sneller version passed its default-target replay at frozen
`97e54f30`; require a new replay after later integrations, with equality intact.

Always-documentation directories require complete original Go inventories and
same-profile actual typed Go NoGoError witnesses before CPP exclusion. Ordinary
test-only packages, mixed/bad source and tool/dependency errors cannot supply
that role. Already inventoried pure-Go test-dependency ancestors now receive
their missing ZIP-bound selection headers without duplicating their directory
or ordinary package roles. Recheck all three mechanisms together in the current
integrated corpus, not just their separate frozen repair witnesses.

Foreign pure-Go test dependencies now use actual module/version/ZIP boundaries,
including nested module roots. A legacy proxy-generated `go.mod` outside the
original ZIP is a separate signed-checksum origin, never fabricated source.
Its public checksum record, inclusion proof, archive inventory and actual Go
metadata bytes remain bound through final objects. Every observed equal-byte
metadata copy retains its own canonical-path and post-object guard. Missing,
changed or redirected metadata fails; reader consistency is not a library pass.

CPP fallback uses Go's actual fixed package-CWD and GOROOT include search.
Canonical physical origins, original literal operands and exact outside-module
ENOENT witnesses are rechecked by producers, portable readers and active
consumers. Validate bound regular source identities before reading. Ordinary
missing inactive/generated headers retain their typed ENOENT error chain.
Do not accept arbitrary external include roots or borrow a pkg/include allowance
when the actual source requires its own search-miss witness. The included-header
EOF operand witness is integrated: it authenticates only the original literal;
Go and the active CPP input stack still independently reject translation-unit
EOF without a required newline. Whole original-library tests remain necessary.

The instrumentation rule foundation binds complete actual Go driver and platform
source registrations and exact early diagnostics. Standalone early observation
and same-mode pure-Go marker exports have independent source/tool and complete
trace/archive/symbol proofs, including actual codegen/buildmode routing. Separate
original-package planning/export foundations retain complete Go/ASM bytes,
actual import-C roles and the full mode/CGO/custom-tag denominator. Separate
same-mode metadata and physical CPP/LLVM object consumers retain independent
compiled Go sources, dependency exports, header bytes and complete typed ABI0
replay. Local types, aliases, constants and zero-sized fields are rechecked from
original syntax, not trusted layout JSON. Imported selectors from the stdlib
or exact same module need the separate original type-source capture, actual
same-mode pre/post graph and independent consumer recapture. Complete needed
Go bytes bind the exact import-spelling sets; paired exports/layout metadata
cannot invent an edge or replace that source anchor. Foreign, generated, cgo,
embed and generic type closures remain unsupported. The separate CLI and
corpus/report/progress/ledger consumers are
integrated locally. They require an independently replayed instrumentation
denominator, including the explicit empty requirements array, and actual
same-mode exports, metadata/header queries, CPP and LLVM objects. Paired proof
deletion or ordinary-protocol downgrade fails. Early rule acceptance is not
export evidence; neither markers nor queries are compiled-object, C/SDK,
sanitizer-runtime or final-link evidence. Frozen mechanism tests passed on
Darwin and native Linux ARM64, but establish no new external-library passes.
Never synthesize feature tags or turn CGO, SDK or dependency failures into N/A.
Mode-only CPP CPU/experiment unions remain explicit unfinished failures.

Source selection follows Go's actual legacy last-blank/block boundaries and
first-byte BOM behavior. Stored original headers and source hashes are never
rewritten. Stock-Go filename pruning removes impossible partners before custom
tag search, not final original-source checks. These repairs passed combined
full corpus race verification; later source changes need fresh reports.

## Required next work

Run the affected focused suites, root `./...`, both nested CLIs, all five
official instruction-table gates, strict standard-library corpus, benchmark
and all 64 Discovery shards. Preserve earlier timeouts/failures; do not count
individual successful tests as a full-suite pass. Use bounded parallel batches
and remove owned caches only after every writer exits.

The official opcode classification gate is not complete standard-library
object coverage. Dynamic reflect FP frames, native runtime entries/context
switches and private ABI transport remain real contracts to implement.
Provisional native-byte-layout/JIT exceptions require review before promotion.
The physical-SP/R29/private-pointer-spill proof and complete raw scalar,
GP/FP-pair memory forms, including LDPSW writeback, are integrated. Frozen family
and original Go/LLVM runtime oracles are separate evidence; complete original
library replays and integrated-source validation remain required. Never accept
arbitrary raw stack effects based on a private-frame name or package whitelist.

The indexed Snappy ARM64 proof is integrated locally. Ordinary indexed tables
and counted clear loops now require a private-DATA byte envelope for each
instruction; saved LR/SP0, mixed-origin, stale-compare and escaping-address
guards remain strict. Go-accepted scalar/GP/FP-pair memory forms have object and
initialized numerical/full-canary counterparts. Frozen agent family evidence
does not prove the later integrated source or replace its complete library,
root-suite and corpus reruns.

The LD/ST1-4 whole/replicate/lane memory effects now share their complete typed
native extents and writeback with lowering, private-frame analysis and control
flow. Actual Go encodings retain reserved lane-displacement/D1 refusals; R18,
g/R28 and saved LR cannot acquire fresh-frame permission. Go's large-frame R20
producer now has a separate original-and-normalized source/CFG proof: only an
explicit complete ABI0 scalar frame may transport the actual prologue's new
private SP. Small/NOFRAME entries do not borrow it, and ABIInternal software
values remain distinct. All GP/FP reads need real source definitions; unknown
inputs, mixed joins, escapes and saved-link corruption fail. Scalar/pair and
structure memory consume the same bounds and continuation proof.
Original Go plus independent LLVM/C initialized-data/canary
oracles run on Darwin and native Linux ARM64. Opcode/post-index dispatch is
checked before IR emission, including legacy helpers. Original Arrow failures
remain failures, not whole-library passes inferred from these family oracles.

The complete typed scalar division/remainder family now follows Go's actual
REM expansion: division writes R27 before MSUB rereads its possibly aliased
sources. W results zero-extend. Shared explicit/implicit outputs invalidate old
caller-link, code-address, private-frame and numerical-range witnesses.
Independent original-Go bytes, initialized numerical/NZCV/canary oracles,
three LLVM APIs, all affected object targets and native Linux execution retain
the real pre-fix failures. Opcode extraction includes all eight scalar forms.

The typed ARM64 immediate arithmetic/logical/compare family now derives R27
materialization from Go's full constant and register classes. Scratch writes
precede alias reads and invalidate stale link, code, frame and range witnesses.
Only actual WORD decoding supplies the private physical-instruction marker;
named pseudo instructions cannot borrow its four-byte width. Actual Go UDF
fallthroughs retain fault semantics, not numerical pass claims. The precise
X BITCON-to-ZR subset with non-equivalent old/new constant encoders remains
Context: a future producer needs the actual source encoder identity, never
the translator's build Go version. Retain its old/new byte/state and raw-PC
failure oracles; run `TestARM64GPImmediate` and
`TestCrossLinuxRuntimeMatrixARM64GPImmediate` before broader validation.

The scalar GP memory family is now integrated locally. Its shared typed grammar
models Go's ordered R27 expansion across lowering and every source-state proof.
Physical WORD identity, AUTO/FP source-frame classes and encoder-dependent or
linked-page scratch remain distinct; unknown observable scratch is Context,
not source N/A. New and existing private-frame positive/escape-negative guards
must run together. Separate clean family gates do not establish later integrated
or whole-library success.

Isolated ARM64 AUTO probes now own an explicit zero-frame NOFRAME TEXT; they
do not supply another function's frame, native continuation or encoder identity.
Same-file RET-only callees have a separate private original/normalized source
witness that narrows only scratch observation. Unknown scratch stays unknown;
typed call, escape, SP/LR, result and native-effect checks remain independent.
Run `TestARM64ScalarAUTOProbe` and `TestARM64SourceNoopCall` with the complete
scalar-memory and counted-table families after changing these boundaries.

The barrier family preserves exact machine-neutral effects and Go integer
operand classes. SB is `d50330ff`, independently checked with original Go,
LLVM-MC and all ARM64 object targets; HINT #7 is not SB. Integer scalar moves
retain distinct floating-class and writeback-overlap rejection. Self-BL escape
negatives use Go-legal SP-to-GP-to-memory transfers and original-Go branch
bytes. FP-pair numerical fixtures observe each scalar slot at its declared
width; the old MOVD source and its independently witnessed stale-reference
failure remain separate regression evidence. Do not change expected numbers
or borrow a narrower load's source-encoder effects for a wider original load.

Native oracle commands now use a bounded owned lifecycle in the repaired
consumers, including the new scalar-memory Go assembler audit. Keep its AST
consumer inventory current. Dependency-free fixture minima have actual Go 1.20
oracles; the executed repeated raw helper uses ordinary R22/R23 instead of
temporarily writing g/R18. Separate old reserved-register object proofs are not
executed-runtime coverage. The self-BL observation repair is separately reviewed:
startup, a full post-marker window and successful cancellation are independent
conditions. Its caller readiness marker is not an independent ASM-PC witness.
Require fresh combined-source host and pinned Linux/QEMU gates before promotion.

The [exhaustive root runner](../../cmd/plan9asmtest/README.md) discovers and
hash-partitions every current root test, example and fuzz seed into fresh serial
processes. Existing skips are counted separately. It does not cover other
packages, nested CLI modules or corpus gates. Historical `47ea649d` root
verification is not proof of this development branch; require a fresh complete
report. Frozen `cd5438b7` finished every root partition but exposed two KMOV
helper regressions: internal dispatch supplied the opcode while the synthetic
instruction omitted it. The complete KMOV family now binds a local opcode copy,
rejects conflicting/unknown identities before emission and preserves strict
source/caller validation. Focused Go 1.27 race and Go 1.20 checks passed. Frozen
`b57d293b` subsequently passed every root partition, full non-root and nested-CLI
race suites, vet/build and five-architecture classification; later integrations
need their own fresh complete reports. The tool's Linux
cold-fixture timeout was corrected and genuinely rerun at `2259cf73`, without
changing the production suite deadline or discarding the earlier failure.

ARM64 closure entries require an explicit immutable `{code pointer, uint64
capture}` carrier and the real hidden register ABI. The default API remains
conservative; ordinary descriptor environments are not capture sizes. The
dependent compiler must provide actual source/caller evidence, not merely
register the translator contract.

## Dependent llgo contribution

Fork Draft [PR 259](https://github.com/cpunion/llgo/pull/259) is rebased onto
upstream main `d98b43f90`. Its public `go.mod replace` resolves published
plan9asm head `0dad3b63` in published llgo source `033a81c0`. The local next batch
resolves the actual public `47ea649d` version and passes module checksum checks;
no local path or invented pseudo-version is committed. Local llgo `fafa83bf6`
and newer plan9asm APIs have been tested through ignored diagnostic modfiles.
They are not the published compiler. Publish and verify the allowed plan9asm
fork head before updating llgo's exact public pin; never commit a local path or
claim a stale public pin builds the new API.

The nested `test/asm` module has exact CRC32, huff0, go-hex, websocket,
modernc/libc Uint128 and purego regressions. Dedicated Linux and Darwin jobs
execute every applicable fixture and retain source/tool/dependency provenance;
root `./test/...` alone does not reach this module. Required Linux ARM64 memmove
execution uses pinned QEMU 10.2.3, not a skipped host-only test.

Frozen llgo `4bbc5cec` passed five actual Linux issue fixtures; purego still
failed at native `syscall15X`. Darwin was blocked by two original R26 closure
entries. A real typed caller/carrier producer is in development, not yet final
public-pin evidence. Purego needs a native C/g0/callback/address-table contract,
not an invented Go declaration or suppressed standard-library assembly.

The next llgo batch repairs the cold replacement installer lookup, formatting
and unnecessarily patch-specific fixture module minima. Those repairs are
published; current Mac/self-host/Windows-ARM64-cross failures still expose the
R26 contract. The reviewed local bridge has original-source and typed-carrier
oracles, but its provenance, bounded-process and global-name collision guards
must complete fresh public-pinned validation before publication.

Windows include-path diagnostics preserve the failing directory/stage without
weakening symlink guards. Valid quoted Windows include fixtures are corrected.
The independent native Windows canonical-source probe passed; it does not
prove the full compiler CI or other targets. Purego's native contract remains
unfinished.

## Independent runtime batch work

The separate assembly compatibility repository owns public API calls,
original-Go baselines, final-link symbols and execution. Its exact per-case
workspaces avoid central-module MVS upgrades. Reviewed family bindings reuse
independent references without duplicating probe bodies; bounded workers and
one collector retain sibling failures and exact source/tool provenance.
Authenticated legacy ZIPs use official proxy metadata plus the complete original
entry map, never an invented source `go.mod`.

Broad runtime tests still need an appropriate native ISA environment. An
independent original-Go MULX carry-preservation failure in Docker/Rosetta is
execution-environment evidence, not an ordinary instruction skip or permission
to change BN curve expectations. Keep masked/QEMU diagnostics separate.
Generated probes now authenticate only actual Go-selected package/test module
sources, not unused ZIPs from the whole declared graph. Repeated selection and
whole-byte source guards bind the exact case, target checksums and invocation
module; missing/late dependencies and old proofs fail. BN candidates use typed
API, curve and representation preflight before independently authenticated
source and real oracles. A candidate is not a pass. These runtime results do
not promote the plan9asm instruction/object ledger or prove every assembly
file in a module.
