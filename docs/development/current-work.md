# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Inspect actual HEAD/status
and terminal owners; never edit a live frozen snapshot. This is a checkpoint,
not a current-head completion certificate. Older evidence remains in history.

## Contribution boundary

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to cpunion.
Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft; do not
promote its fork head until current-head fork CI and review pass. Root alone
publishes assembly-ledger updates.

Published head `c09eff1e`, run `37711752851`, finished failed without agent
cancellation. Four priority shards failed; two reached the job deadline and
uploaded partial checkpoints. Aggregate and completion checks failed; skipped
downstream suites are not passes. Inspect the six original reports with their
frozen tools, retaining the partial denominators and missing work. Previous
run `36942816488` on `47ea649d` also failed. Batch the successor before pushing.

## Frozen validation

Continue on `codex/pr40-following-batch-20261008`. Running component gates stay
frozen in separate persistent worktrees. Overlapping focused passes cannot be
added into a full-suite count.

Frozen `026f8cfc` completed all 32 root partitions: 5,021 roots and 87,348
children passed, no failure/incomplete node. Another 39 host-inapplicable
Linux runtime roots were skipped and still require their native/cross CI axes.
Independent raw JSONL/list, full source/tool/stamped-runner and recorded
owner/group readback agrees. This is not the later combined head. Preserve the
prior `40ff602f` full-root report with its 17 ARM64 scratch-path root failures.

Its other-package/nested-CLI/first-party corpus-driver gate naturally closed
failed: 2,011 roots passed, one retained two-minute LLVM loader root failed,
8,625 children passed, zero skip/incomplete. Both corpus-driver lanes completed
all 32 partitions. Independent full-source/tool/raw/list/owner readback agrees;
the later loader repair does not upgrade this historical failure. These unit
tests are not local external-library scans or corpus compilation.
The published `c09eff1e` sixteen-stage affected gate passed, but did not certify
the complete root/module suite or external corpus.

Frozen `39109359` closed seven following-batch affected stages: corpus-driver
ordinary/race each 555 nodes, root scopes each 223, and full storage-package
scopes each 219 passed; full vet and stamped corpus-driver build passed.
Independent complete source/tool/list/raw/owner readback agrees. These are
first-party affected scopes, not current-head full suites or library passes.

The aggregate command now audits reports once and reuses that exact progress
for strict verification and stored-ledger comparison. Subsequent publication
hardening is integrated at frozen `9c786dae`: ordinary/race affected checks each
pass 71 roots and 302 children, with no failure, skip or incomplete test.
Full vet, stamped CLI build, 29 CI-policy tests, actionlint and shell syntax
checks pass. Independent full source/tool/raw/owner readback agrees. These are
affected scopes, not the complete suite or new external-library certificates.
Diagnostic phase events are uploaded separately from corpus reports.

The affected publication scope is repeated at frozen `00ff1d5f`: current Go
ordinary/race each pass the same 373 nodes, and Go 1.20 passes all 70 roots and
198 children in its actual inventory, with no failure, skip or incomplete node.
The preceding Go 1.20 run retains a real long-TempDir-name fixture failure;
bounded subtest labels repair the fixture without changing its CLI assertions.
Independent full-source/tool/list/raw/owner readers close all three runs.
Its separate exhaustive ordinary corpus-driver owner naturally closed failed:
31 partitions passed, one failed; 629 roots and 2,293 children passed. The
single failure compared Darwin's temporary-directory alias spelling with Go's
physical source selection. The test-only physical-directory repair preserves
the complete original selection and source-rejection assertions.

The fresh immutable `ea2fcb4c` corpus-driver owner passes all 32 partitions,
630 roots and 2,301 children, with no failure, skip or incomplete node. The
eight additional children are the now-reached source-rejection controls, not
new external-library passes. Independent full source, 53 actual tool roles,
12,051 SDK files, complete inventory/raw EOF and owner/group readback agrees.
This certifies that exact ordinary first-party owner, not the later combined
head, race suite, complete module or external corpus.

## Integrated rule-class repairs

The predecessor already integrates shared GP/SIMD, scalar-lane, AES, CMOV,
horizontal integer and dword-product grammars/effects; source CFG/result
requirements; exact native widths; raw segment/x87 and Go-frame address roles;
and planning cancellation propagation. Preserve independent Go-byte, numerical,
fault, canary and optimization axes. Logical FP bytes do not grant native
backing/alignment. Rosetta cannot certify Intel alignment faults.

The successor additionally integrates these independently frozen components:

- ARM64 scratch-owned directory canonicalization and fixture preconditions:
  `eb4e9b12`, ordinary/race each 1,270 nodes pass. Source/tool redirects remain
  invalid; no caller path is canonicalized into a new authority role.
- Requested-key-only superseded lookup, including matched replacement payload:
  `f1e3d892`, ordinary/race each 168 pass. The fixture sampled about 140 MB RED
  versus 2.5 MB GREEN, not a whole-ledger constant-memory guarantee.
- Ordered streaming directory-collection fan-in: `bd51d81d`, full Discovery
  ordinary/race each 287 pass plus vet/build/readback. Its CLI fixture measured
  101.1 MB RED versus 3.85 MB GREEN, and the 80-member/FD64 regression passes.
  Initial/final EOF, source-order folds, counts/ranges, generation/member guards
  and atomic publication remain mandatory.
- Complete register XCHG effects and AX-self NOP versus ModRM semantics:
  `84798f6f`, ordinary/race each 2,301 pass. Partial/high-byte/REX writes,
  simultaneous reads, opaque identities and original runtime axes stay distinct.
- Actual SDK internal/cpu typed numerical storage observation: `8879d43a`,
  ordinary/race each 120 pass plus independent five-target archive/type readers.
  Storage extent is not a CPU constant, initialization or global-read permission.
- Exact downloaded original source-query observation: `317a3cd7`,
  ordinary/race each 199 pass, historical REDs and independent full-source/raw/
  binary/Go-driver readback, vet and stamped build retained. Each foreign module
  has its own actual exact download/ZIP/member/source guard. No copied synthetic
  owned-main role. Hash agreement is not an independent SumDB signature.
  Downloaded BindFixedLocal still returns explicit Context and no binding.
- Request-owned stable LLVM loader source: `39556bfb`, ordinary/race each
  31 pass. Every scope still builds and executes its own binary and captures
  actual tools/environment/library; no proof/result cache. Parallel creation,
  cancellation and foreign temporary cleanup/publication guards pass.
  A fresh empty-cache unchanged-budget check passes; historical two-minute F
  remains failed evidence, not a relabeled success.
- Per-target ordinary profile budgets: `2c46bdb3`, ordinary/race each 46 pass.
  Each actual target is bounded by 64 unique validated profiles, not one
  module-global limit. The full 123-profile generated-header projection fixture
  is separately corrected in `c0570fa1`; its ordinary/race each 50 pass preserve
  all eleven target scopes and independent original-Go/header consumers.
- Original Go TEXT/code/relocation/source-PC observation: `aae04cd8`, full
  storage-package ordinary/race each 206 pass, plus actual SDK archive readers.
  Wasm logical PC remains separate from physical bytes. This sealed producer
  observation does not itself grant frame, SB-read or native permissions.
- Complete scalar-shift grammar and numerical effects: `471d95f7`,
  ordinary/race each 8,128 pass, all 148 independently enumerated Go encoding
  rows and actual Go/LLVM optimization axes. Defined flags, undefined-count
  behavior, physical modes and register-only numerical permission stay distinct.
- Complete selected-source CPP transport closure: `d5fe3277`, focused
  ordinary/race each seven pass, retaining actual 32 MiB RED and source-only,
  inactive/transitive include and query-isolation controls. The expanded old
  leaf retains the generated-profile assertion failure and interrupted race
  stage; it is not a completed expanded-suite certificate.
- Selected-shard-only candidate payloads and exact exception-file interests:
  `08a04af9`, focused ordinary/race each 157 pass, plus vet/stamped build/readback.
  Real foreign private/native manifests exposed 115–147 MB startup RED; the
  same first-party operation is about 9 MB ordinary / 11 MB race after repair.
  Global module/version membership, duplicate unions, target filtering, exact
  subsets, foreign aliases and unselected/empty-interest EOF checks remain.
- Complete broadcast grammar and numerical effects, including K-data forms:
  `d3b2d06a` closes the earlier `1d55e698` family and case-normalization repair.
  Final ordinary/race each 2,331 nodes pass, with 111 independently checked
  Go encoding rows, three-API/five-target objects and original Go/LLVM runtime
  axes. K-data numerical caller authority is still explicit Context. Physical
  WIG dword behavior and cleared wider aliases do not grant native memory,
  full-vector definedness or an unobserved AVX512 runtime capability.
- Task-owned restored Go build-cache connection and separate checkpoint upload
  budget: both discovery matrices preserve fresh source/tool/profile/object
  checks. Configuration regressions and focused ordinary/race cache/checkpoint
  controls pass. No measured corpus speedup is claimed; timed-out work fails.
- FP alias rejection diagnostics: `513b3c4c`, current ordinary/race each 542
  nodes pass; historical Go 1.20 directly affected lanes each 162 pass.
  Actual warning anchors and CPP/native/result causes are retained without
  changing permissions, command receipts or pass/source-exclusion decisions.
- Optional streaming phase journal: `4ecd89ca`, focused ordinary/race each
  111 nodes pass, plus vet/build and independent EOF/source/tool/owner readback.
  Its bounded writer retains handled child errors separately from the original
  top-level outcome. Requested I/O or corrupt-tail failures invalidate reports;
  timing records never certify candidates. The initial wrong-CWD RED remains
  invalid evidence; the separately frozen original-driver reproduction is RED.
- Frozen aggregate publication: `9c786dae` closes the real late-input,
  occupied-output, changed preparation, growing-file and replaced-inode REDs.
  Read-only preflight resolves actual filesystem aliases, including case
  aliases without excluding case-sensitive filesystems. Preparation closes
  complete bytes before final source/scan/stored-ledger guards and publishes
  without overwrite. Physical ledger checks stream bounded reads, not another
  semantic audit; cleanup preserves foreign replacements. Five actual CLI
  fixtures retain strict failure exits and reject unauditable trailing data.
- Typed numerical SB source/object/storage joins: `fc56d73f`, full storage
  ordinary/race each 219 pass and numerical/FrameSource lanes each 216 pass,
  with independent SDK archive/source/read-width/addend checks. All 22 scalar
  MOV/CMP/extension shapes are observational only at this checkpoint. No
  Options, LLVM load, native address or caller authority is granted yet.
- Independent downloaded frame factory: frozen `91a24297`, full affected
  ordinary/race scopes each pass 5,519 nodes, with independent readers. Exact
  module/version download, ZIP/source identity, original compiler and assembly
  objects, complete declarations/CPP and numeric fixed-local binding remain
  separate prerequisites. Shared concrete download-boundary lifecycle at
  `671ddc92` additionally passes 96 ordinary/race nodes each; changed or foreign
  files survive cleanup. Cache `Module.Main` fields grant no owned-source role.
- Actual downloaded CLI consumption: `a3d613b9`, ordinary/race each pass all
  ten selected roots and 41 children, including seven invocation/load/resolver
  identity negatives and existing fixed-local CLI regressions. The preceding
  actual original-Go-export RED proves the missing independent download route.
  A later expanded fixture RED retains the leaked host CPU environment; only
  explicit test-target setup is repaired. Production environment checks remain
  strict. Independent full-source/tool/list/raw/owner readback closes both gates.
- ARM64 original byte-region producer: frozen `cfd6c447`, integrated through
  `4366ed95`. Ordinary/race each pass 18 roots and 500 children; full shared
  storage scopes each pass 25 roots and 265 children. Go 1.20 framework scopes
  each pass six roots and 429 children, retaining their exact Darwin linker
  warning. Auxiliary vet/build/Node/shell checks and the unchanged actual SDK
  census are independently closed. Five OS original objects, all nine wide
  scalar result-store combinations, original lease cancellation, source/PC/
  relocation and native bytes are observations only, not LLVM permission.

Normal sharded ledger startup/status, active-batch lookup, bounded fair retry
selection, publication and transfer stream. They still validate every row,
counts, continuity and generation. Legacy monolithic JSON/gzip imports and
explicit full-report reference helpers are separate compatibility paths.

## Remaining shared failures

Source-query proof is not an actual frame/object certificate. SDK numerical
storage is not a per-instruction SB-read certificate. The independent original
TEXT/relocation/source-PC and typed x86 read producers above are closed. The
next bounded stage consumes only factory-bound instruction-local numerical
reads through all three APIs, mutable volatile loads and original runtime
witnesses. Standard/external source roles and generated headers remain Context
until their separate source roles are closed.
Do not turn an OpSym spelling, DATA record, BSS extent or source/report JSON
into initialized/native memory. Preserve wasm logical PC versus code-byte axes.

Separate component worktrees currently own these limited consumers:

- Numerical reads: seal complete declarations/options and join independent
  original numerical/frame objects with exact TEXT, code, source-PC and ordered
  relocations. Keep default no-relocation refusal. Rejoin instruction tokens
  after fixed-local reparsing; never copy old certificates or enter numerical
  owner callbacks while holding a frame lock. The actual consumer, copied-local
  owner guard and mutable/fault runtime scopes are still on the component leaf;
  its focused passes do not certify the root combined head or native Linux CI.
- Downloaded frames: share bounded exact-download physical source guards, then
  use an explicit independent factory/binding route. Ordinary Go `Module.Main`
  and an empty version do not prove an owned-main role: cache CWD has the same
  fields. Default owned routes exclude SDK/cache roots; an exact-download route
  must retain its own complete original archive/cache identity. The factory and
  CLI route above are integrated; expanded actual LLVM22 objects/runtime still
  require Windows SEH metadata closure and the exact required Linux fixture.
- ARM64 result transport: the original byte-region producer above is closed.
  Next reproduce and consume the complete same-carrier native result effect,
  initialized tail bytes and canary, not just a narrow typed return. The three
  APIs use the original bounded source lease; LLVM Context controls resources,
  not cancellation. Logical result bytes and rounded frame allocation alone
  do not grant a padding slot or an open native tail-call effect.

These components require complete ordinary/race and independent reader closure
before integration. Their producer or focused passes do not promote the ledger.

Current CI failure classes include FP/native/CPP/result binding, cached source/
object roles, x86 caller/stack joins, ARM64 native ABI/frame state, profile
planning and generated/instrumentation/resource scopes. Repair entire classes
with independent first-party regressions, not library allowlists. The module-
global profile limit and selected-package CPP duplication have source-level
repairs above; discovering external versions still require fresh matching-head
remote reports. Never drop required inputs, raise bounds to hide an error, or
turn observation-only storage/TEXT data into native authority.

Frozen `026f8cfc` official five-arch classification and Go-accepted ARM64
decoder gates pass, with zero official unsupported forms/parse errors.
Context classifications, encoder tables and runtime proofs remain distinct
measures, not 100% object/runtime coverage. Never bless a changed fingerprint.

Its strict five-target benchmark attempted all 184 files: 167 pass, 17 fail,
zero N/A/unattempted; target wall total 28 seconds under unchanged 300/900-second
budgets. The standard target-matrix gate also fails. Remaining classes include
SDK symbol reads, dynamic reflect frames, native entry/TLS/continuations,
ARM64 result padding and whole SDK body closure. Preserve the earlier strict
`55ecfe9d` failure; speed does not certify compilation.

A later independent recheck at frozen `1a0c2d71` again attempts all 184 files:
167 pass, 17 fail, zero N/A/unattempted, build three seconds and target wall total
27 seconds. Official five-arch classification and Go-accepted ARM64 decoder
commands pass; the standard target-matrix gate still fails. Complete tracked
source, SDK source/include inventory, actual tools, logs, reports and all five
owner/groups were independently read back. The decoder dependency itself was
not part of the initial SDK snapshot. This is a closed failing baseline, not
the forthcoming consumer integration or an external-library certificate.

Historical scalar compatibility hit a pre-test link limit with zero tests
attempted. The mode32 extreme-SP witness has an LLVM22 O0 allocator failure,
not source N/A. Exact official SDK namespace refusal and valid sibling objects
have their own frozen five-target proof; earlier timeout evidence stays failed.

After components close, freeze the combined source and run affected/full root
and other packages, both nested CLIs, historical framework compatibility,
official tables, strict stdlib/benchmark and required fork/native CI axes.
Missing tools, unknown errors and timeouts fail. Dispose LLVM modules before
contexts; clean only exact owned caches after every consumer closes.

## Corpus, ledger and inventory

No local background external inventory/corpus. Inventory belongs on its
designated source-only server; compile/execute external code only on separate
disposable validation executors. Local first-party fixtures remain allowed.

The committed assembly ledger still has three historical Snappy passes and
6,064 pending, no complete shard, complete=false and verified=false. Focused
instruction tests never advance it. Only the automatic publisher may update
records from complete reports with matching frozen provenance.

Committed inventory retains 215 continuous ranges, the 2025-10-01 history
cutoff and head 2026-09-22T22:43:08Z. The requested cutoff is reached;
history_complete instead means the entire official index back to its epoch.
The bounded two-worker standalone incremental service has newer unsynchronized
records. Inspect its actual owner/manifest/journal; never hot-replace its binary,
start a second writer or merge an unvalidated live snapshot. cgo stays outside
this repository. Cursors advance only after complete contiguous publication.

Recheck discovering samples and all 64 shards after systematic repairs.
Partial reports, Context and provisional exclusions are not passes. Funnel
tables belong in the PR body. Never hand-upgrade records or claim CI/review/
coverage completion before every required gate closes.
