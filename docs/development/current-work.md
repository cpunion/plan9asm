# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). This checkpoint is a
handoff, not a completion certificate. Inspect the actual frozen HEAD, complete
raw reports, inventories and terminal owners before accepting a result.

## Contribution boundary

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to cpunion.
Do not update the upstream-connected head until fork current-head CI and review
pass; never push to or merge upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40).
Root alone publishes assembly-ledger updates.

Published `c09eff1e`, run `37711752851`, finished failed: four priority shards
failed and two reached their deadlines. Downstream skips and failed aggregate
checks are not passes. Batch verified repairs before another push.

## Frozen evidence and development branch

The complete ordinary module run at `7842042a` naturally closed **failed**.
All eight root parts, corpus/rest owners and aggregate finished without timeout
or cleanup failure. The complete root inventory is 5,353, not the raw listing's
line count: 5,313 pass, 39 explicit cross-runtime skips and one failure.
The corpus-driver passes 652 roots and 2,524 children; all eighteen remaining
packages pass 589 roots and 3,687 children. The aggregate correctly rejects the
failed root part. Independent root raw/inventory and terminal input/owner audits
retain the failure; they are not a passing full-module certificate.

The sole failed root is
`TestGoAssemblerOracleARM64BranchListingRejectsUnboundMetadata`: its old negative
fixture rejects the historical four-class format, which the matcher intentionally
supports. Test-only `d97a6c41` replaces that negative with a three-class record and
adds explicit four-/five-class positives. Its complete four-root scope passes
128 ordinary, 128 race and 126 actual Go-1.20 nodes, all independently rechecked.
This does not transfer a full-module pass to a changed HEAD.

The next integration branch is `codex/pr40-ci-integration-20261009`, based on
`d97a6c41`. Preserve frozen `7842042a`, its original reports, and all component
worktrees. Latest upstream main `e2164ac0` is already an ancestor. This is still
development integration, not a verified combined head. The earlier `2ae09eb7`
full failure also remains preserved. Earlier passes establish only their exact
component revisions:

- `4617d12d`: complete explicit IMUL source/caller family, zero-operand
  diagnostics, actual ARM64 lazy-tool routes and backend physical-source prefix.
- `8c60cde3`: complete x86 aliases, scalar positive-zero rewrites, frontend
  physical-line/label roles and mode32 operand-class diagnostic metadata.
- `e50f590a`: exact MOVL/Q address-to-LEAL/Q diagnostic joins, including I386's
  distinct EXTERN/STATIC role. Three affected roots pass; no ISA/ABI permission
  is obtained from a diagnostic spelling.
- `7b05cb4c`: an explicit trailing backend-diagnostic route binds complete
  ARM64 TEXT attributes, frame/argument values and the exact unaligned-frame
  error. Original 34-node RED becomes 34-node GREEN. The complete ordinary
  oracle package passes separately; its default diagnostic route is unchanged.
- `e34c78c6`: shared target CPU command-environment construction fixes foreign
  inherited CPU keys without sanitizing output or weakening profile validation.
  Seven focused roots and the original four instrumentation roots pass.
  The separate 34-root finite rerun ends with 33 passes and one incomplete
  generated-header root. Its six-minute per-root deadline is not the original
  fixture's fifteen-minute or full module's forty-five-minute budget. Preserve
  that failed scope without calling it a semantic failure or a full-suite pass.
- `8061bd59`: explicit wasm unknown-word lexical roles preserve complete
  original statement/suffix binding. Component ordinary/race/Go-1.20 oracle
  scopes pass; this is not wasm runtime or LLVM object coverage.
- `348e27b9`: ARM64 B/BL aliases, canonical P/W suffixes,
  complete historical four-class/current five-class directories, and actual
  selected-SDK bare From versus register/special-operand lexical roles. A bare
  From symbol prints zero offset; it is not a resolved label-PC permission.
  Complete ordinary/race/actual Go-1.20 component scopes pass.
- `dc2e4918`: private x86 whole128 observer joins exact same-Void-source/private
  caller/normalized-ordinal read16 evidence. Its complete 26-root scope passes
  ordinary/race/actual Go-1.20, and shared aliases pass ordinary/race separately.
  Public/direct/nonVoid, aligned, scalar/half, native-store and unknown-extent
  boundaries remain unchanged. Only its six reviewed production/fixture
  commits are integrated here; shared oracle prerequisites were already present.
- `69253f0e`: complete x86 alias operand joins retain Go's TYPE_FCONST value,
  canonical seventeen-digit printing and signed zero. The complete shared
  16-root scope passes ordinary/race (498 nodes each); actual Go-1.20 selects
  11 roots and passes 450 nodes. Its five build-tag exclusions are not skips.
  Actual logged Go commands number 151/151/108, all independently closed.
  No instruction, source N/A or ABI permission is added.
- `fd1b9dc9`: a literal NOSPLIT TEXT and complete straight-line deltasp join
  binds I386 PARAM += delta+4 and AUTO += delta to the original operands,
  aliases and full backend class tails. Ordinary/race pass 134 nodes each;
  actual Go-1.20 selects three portable roots and passes 84 nodes. Unknown
  flow, CPP, SP writes, symbolic frames and overflow remain rejected. Its
  ordered test/fix commits are integrated as `4fd28774` and `07a1d672`.

The combined rerun at `cddbc85e` naturally ends with 52 of the original 54
failed roots passing. Two roots still fail on four diagnostic-join children:
alias FLOAT formatting and mode32 FP rebasing. Retain that failure. A new
combined gate at `07a1d672` passes all original 54 roots plus seven new roots:
61 roots and 2,401 children, 2,462 pass with zero fail/skip/incomplete. Main's
independent fresh readback closes full raw EOF, source/tools/binary and all
144 logged actual-Go children. This is a finite scope, not full-module or CI.

Generated-header planning now has a conservative condition-free fast path.
Its query-only RED at `bf8ae5b2` retains 113 passes and 13 failures; the
successor passes all 126 nodes. Full source/include/header guards, CPU-token
fallback, Go-only partners and instrumentation expanded-output scopes remain.
The actual consumer at `1d9f6b99` passes 27 Builtin nodes across eleven targets;
seven separate query-only controls pass on Linux/amd64. Planning queries fall from
123 to 11; all eleven final headers and LLVM objects remain required. Main's
fresh readback confirms the complete reports and frozen inputs. Initial full
Go-list exports and physical temporary LLVM objects were not retained: do not
claim a fresh physical-object audit or external-library success. The earlier
canonical-target helper failure remains preserved. The five reviewed commits
are integrated through `83d0274b`, not yet whole-module verified.
The actual Go-1.20 query-model check at `1d9f6b99` also passes all 126 nodes,
including Main's fresh readback. Its earlier wrapper setup failure remains
preserved with all tests unattempted; it was not a CPP failure. The corrected
reader binds complete command-package build metadata and Go-1.20's actual
worktree/VCS rules, rather than granting a generic missing-VCS exception.

Successful numerical-consumer fixtures now report actual source Close errors
instead of discarding them. The source-bound RED at `6f22c04a` exposes the
directory-identity error; `772d2504` passes ordinary and race, twelve nodes each,
including the preserved replacement-directory canary. Main independently
rechecks both complete inventories. These test-only commits are integrated as
`c308c511` and `e4e22527`; they confer no source/ABI permissions or ledger pass.

ARM64 MOV memory-row coverage now compares the selected Go table's complete
attributes and physical order, exact unsigned aliases and independent typed
shapes. Its same-count mutation RED has one pass and seven expected failures.
The test-only successor at `4328896d` passes 306 ordinary and 306 race nodes;
actual Go-1.20 passes 261 nodes using its ten-field/twenty-row schema, not skips.
Main independently rechecks all three complete inventories and frozen inputs.
The ordered commits are integrated as `84e923a6` and `a14f9f33`. This projection
does not prove overlapping encoder-row selection, LLVM objects or ABI/runtime
behavior; the inner Go-env command has no independent command-receipt proof.

## Next work, in order

1. Freeze the integrated source after inspecting its actual clean HEAD.
   Component passes never transfer to the new combined HEAD. Preserve every
   original failure and keep the component source, tools and reports frozen.
2. Run all original failed roots, complete
   affected families, the exhaustive original two-worker module plan, nested
   CLIs, vet/build, official encoder/decoder, strict SDK/benchmark and required
   native CI. Let all owners finish, including failures; recheck complete raw
   EOF, inventories, source/tools/binaries and owned processes independently.
3. Repair remaining failures in separate persistent successors. Do not rewrite
   a running snapshot or turn incomplete/killed scopes into success.
   The next corpus-driver rule class is a known integer input's occupied narrow
   MOVB/W/L read on 386/amd64. Test-only `39a92596` fixes its three-root, 64-node
   inventory and retains actual Go-object/listing and complete combined lint
   output. Its first launch fails in the input worker's Go-template escaping,
   before compilation or test execution: this is setup failure, not a TDD RED.
   The separate helper successor establishes a real RED: 48 filter failures,
   four failed parents and twelve passing controls, with no setup/unknown
   failures. Production successor `9530e21` binds the exact original diagnostic
   and readonly native width to a declared integer parent. The original scope
   now passes all 64 nodes; the expanded uintptr/386-half scope passes all 119
   nodes. The complete expanded race scope also passes 119 nodes. All three
   scopes are independently rechecked. The actual Go-1.20 run fails with 19
   passes and 100 failures: every actual-object fixture stops at its shared
   header/listing guard, before observing the filter. Preserve that failure and
   establish the exact historical object/listing identity; LLVM semantic gates
   also remain pending. Synthetic pointer components, incoming writes, padding,
   partial results and unknown effects remain unproved. The width observer now
   covers the official MOVL/MOVQ/MOVD/VMOVD/VMOVQ GP/X memory rows and the
   separate MOVL/MOVQ/MOVD MMX memory rows. Actual Go accepts MOVQ FP/X on
   both x86 architectures; its old unknown-form assertion was incorrect.
   Independent RED gates reproduce four omitted GP/X rows and 64 omitted MMX
   rows. The complete 20-row GP/X and 96-row M0-M7 load/store matrices pass.
   An actual Go package additionally checks thirteen occupied declaration
   types on both architectures, including integers, floats, named types,
   arrays, structs and complex64. Native width alone does not initialize any
   bank, authorize parameter writes or prove translation/runtime success.
   Another actual library failure is the normalized raw-byte/AUTO replay join:
   reparsing decoded instruction text loses private decoder identity and comments.
   Test-only `82d046c0` adds an independent raw-times-AUTO regression. Its first
   run passes the four-node Go byte directory but fails its named-input
   prerequisite before the mutation inventory runs. The next fixture reaches
   the eleven target failures but incorrectly expects raw RET to reach the
   observer instead of its normalization rejection. Preserve both failures.
   Corrected `6dda63fa` establishes the complete 45-node RED: 32 passing controls
   and 13 failures (eleven target children and two aggregation parents), with
   independent complete readback. Repair the private complete original-byte/
   order/boundary join, not a broad DeepEqual bypass or new native permission.
   The replay seam is now implemented privately in `Func`: it retains the
   original raw source, decoder checkpoint, and final normalized instruction
   snapshot, then reparses/decodes and compares all three before AUTO proof.
   The independent 45-node GREEN gate on `0ccbc807` completed naturally with
   45/45 nodes passing, including all eleven former targets; the race run also
   passed. The original RED artifacts remain retained.
4. Continue the SDK numerical source/object seam, dynamic reflect frames,
   native entry/TLS/continuations and full result/body contracts by typed rule
   class. The old go-fuse LoadUint16 failure remains a real ARM64 result-tail
   backing gap: parameter-padding and tail-call proofs do not authorize its
   wider result store. Never clip the store or invent a padding slot. Then
   run discovering libraries and all 64 external shards on the
   designated validators and publish only matching complete schema-11 reports.

The strict five-target SDK baseline still attempts all 184 files: 167 pass and
17 fail, with zero skip/N/A/unattempted. Official five-arch classification has
zero unsupported/parse errors but includes Context forms; classification is
not complete LLVM, ABI or runtime success. SDK numerical capture remains
query-only: its genuine CRC consumer failure is not integrated or promoted.

The non-sharded local `go test ./... -count=1` at `7bf1132e` is not a green
full-suite result. The root package and `cmd/plan9asmcorpus` each hit their
10-minute package timeout while their Go-oracle/LLVM subprocess contexts were
being cancelled; `cmd/plan9asmtest` also recorded three tests with an empty
executable path. These are retained infrastructure/resource failures, not
assembly-ledger passes or skips. The raw AUTO 45-node gate and the complete
GPSIMD transfer-family tests were run separately and passed.

## Resource and ledger rules

No local background index/discovery/external corpus. Inventory stays on the
designated source-only scanner; external compilation/execution belongs on
separate disposable validators. Local first-party fixtures and corpus-driver
unit tests remain allowed. cgo inventory stays outside this repository.

Use task-owned caches for new verification. A selected shared lazy tool vanished
and disk exhaustion caused real reader/setup failures; retain those failures.
Check available disk before compilation. Do not delete selected tools or cache
entries until every consumer and independent reader closes. A prior cache size
or manifest does not authenticate its current bytes. Dispose LLVM modules before
Contexts and serialize heavy full-suite/container gates when resources are tight.

The committed inventory reaches the requested cutoff. A fresh read-only remote
check finds no active scanner/controller or held writer lock; the historical
service exited successfully at the cutoff. Newer incremental records have been
synchronized into the independent inventory, not this contribution. Its manifest
ranges are contiguous and extend the committed prefix, but only the cutoff and
latest traffic journal endpoints have been independently checked. Independent
local sync still uses an old path-only deduplication binary, whereas current code
preserves the latest version for each path-plus-major line. Do not import that
normalized output as canonical: rebuild the reviewed sync tool and remerge the
existing records, preserving the contiguous prefix without rescanning. Historical
local union records also need explicit provenance. Inspect those exact inputs
before importing; never duplicate a writer or replace its live binary.
Status, lookup and publication stream sharded records rather than loading the
ledger payload. cgo records remain outside this repository.

The read-only remerge was subsequently completed with the rebuilt streaming
tool and the path-plus-major rule. Its staging result is 21,257,856 contiguous
index entries, 1,051,454 unique paths, 6,700 matched exact versions, and 84,225
assembly files. It restored 126 remote-only independent-major records and kept
one historical local-only record; no same-major local record was superseded.
This is discovery staging only: it did not download or compile external modules,
and it did not promote any assembly result into the committed ledger. The
committed assembly ledger remains 6,067 entries (3 passed, 6,064 pending).
The larger 6,700 denominator is a corrected merge of newer remote discovery and
major-line retention, not a downgrade of previously verified passes.

The assembly ledger is incomplete and unverified. Only automatic publication
of complete, matching source/scan/tool/file/target/profile/consumer proofs may
promote entries. First-party passes, Context and provisional exclusions are not
module passes. Keep scan/coverage funnel tables in the PR body, not this guide.
