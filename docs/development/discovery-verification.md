# Discovery verification and standalone inventory

Read [Discovery operations](../../testdata/discovery/README.md) for the canonical
ledger format, scan commands, version ordering, traffic accounting and cursors.
Do not hand-append JSONL: `plan9asmdiscover` validates, deduplicates, semver-sorts
and publishes records automatically.

## Inventory versus testing

`scanned` proves successful exact-version inspection, including no-assembly
results. `matched` additionally retains assembly paths. `failure` remains
retryable and does not prove absence. Retain all architecture hints for future
platform queries, not only currently supported targets.

For a complete local replay, optional scheduling priorities do not reduce
coverage or reuse old reports:

```sh
PLAN9ASM_DISCOVERY_PRIORITY_SHARDS=25,26,39 \
  PLAN9ASM_DISCOVERY_PARALLELISM=4 \
  scripts/check-discovered-library-corpus.sh all 64
```

The priority indices must be unique and in range. Every remaining shard runs
exactly once, with the same bounded-batch cache and final aggregate check.

The standalone inventory runs outside this repository and also collects direct
cgo imports. Keep its program, cgo records, deployment settings and traffic
reports outside plan9asm. Its server inventories source only: do not compile or
execute downloaded code there. Consult that checkout's own `AGENTS.md` for
service/synchronization commands; never hard-code a host or local workspace.

The standalone scanner persists successful latest resolutions with a
conservative freshness boundary. Older history reuses them across restarts;
incremental updates, newly seen major-suffixed paths and retries remain eligible.
Resolving latest does not mark a failed ZIP inspection complete. Its assembly-
only export omits cgo and latest metadata for compatibility with this
repository's three record kinds.

## Cursors and safe publication

- Derive history from the first manifest range's start and incremental from
  the last range's end. An empty ledger starts at captured current time.
  Never invent a manual resume token.
- Reverse scans consume complete windows and retain equal-timestamp groups;
  incremental scans drain their entire captured interval. Reject gaps,
  overlaps and inconsistent repeated ranges.
- Keep the highest Go-semver version per `(module path, semver major)` line.
  v0/v1 can share a path, while `/v2` and `/v3` have distinct paths. All
  coexist; older versions within each line are skipped. Do not replay already
  committed ranges after changing this rule.
- Concurrent network workers feed one publisher. Do not start a competing
  writer, remove a live lock, or overwrite a ledger being tested.
- Remote snapshots/imports validate checksums, counts, sorting and ranges.
  The standalone CLI waits on snapshot/writer contention without discarding
  an inspected batch; integrity failures are not lock retries.

## Importing an assembly-only checkpoint

First finish current corpus runs. Validate the export without network access:

```sh
go run ./cmd/plan9asmdiscover -status -out-dir "$ASSEMBLY_LEDGER"
```

Inspect ranges and records against the committed checkpoint and keep an ignored
recovery copy. The writer merges with the existing destination under its writer
lock, retaining the highest semver per version line and rejecting discontinuous ranges.

For a validated checkpoint, the converter normalizes, merges, and atomically
publishes the destination. It is not a file-copy replacement:

```sh
go run ./cmd/plan9asmdiscover \
  -convert-report "$ASSEMBLY_LEDGER" \
  -out-dir testdata/discovery/ledger

go test ./cmd/plan9asmdiscover ./cmd/plan9asmcorpus -count=1
go run ./cmd/plan9asmdiscover -status -out-dir testdata/discovery/ledger
git diff --stat -- testdata/discovery/ledger
```

Review newly selected/retired versions and retries. Commit the normalized
checkpoint, then test every selected assembly-bearing version. Only one
manifest and 256 module-hashed JSONL shards belong in Git, never gzip/run trees.

## Applicability is evidence, not an escape hatch

Select by Go filename GOOS/GOARCH suffixes and `go/build.Context.MatchFile`.
Probe unsuffixed architecture-specific assembly with the current Go assembler.
Future-target queries conservatively retain unsuffixed/custom-suffixed paths
before reevaluating source constraints.

Architecture probes have bounded output, a one-minute deadline and owned
process-group cleanup. Only a positive assembler exit with an actual source
diagnostic can exclude a target. Missing tools, killed/crashed processes,
resource errors and unknown/empty tool failures remain failures. Retain the
actual target diagnostic in both ordinary source evidence and native-layout
selection plans; do not replace it with a generic rejection message. If an
empty generated-header stub lacks constants, keep the target eligible for the
real package build rather than treating the missing layout as invalid source.

For each target/tag/package group:

1. Compile that exact package and its dependencies with current Go using
   `go list -export`, not `package/...`. This compiles Go and assembly without
   demanding an executable `main` entry point or running package initialization.
   Final linking and execution belong to the separate compatibility repository.
2. Run `go vet -asmdecl`. Only concrete argument-size/FP offset/FP width
   mismatches are ABI N/A. Generic vet errors must not hide translation.
   Equal-width whole-aggregate moves remain eligible: Go's asmdecl type-kind
   comparison can reject a 16-byte MOVOU against a 16-byte array even though
   the Go compiler accepts its layout. Such a warning is not a width mismatch;
   translation and LLVM 22 compilation must still establish the outcome.
   A zero-size warning is not conclusive when the selected source's exact
   TEXT line uses omitted arguments or the historical literal `-0` form.
   Check its selected package, physical line and symbol before filtering just
   that warning. Preserve other ABI diagnostics; unknown source metadata fails.
3. Translate every applicable saved `.s` file and compile every result with
   LLVM 22 at `-O0`. IR is verified before `llc`, and an object must be
   generated; optimization is unnecessary for this compile-only corpus and
   costly for large generated files. Unsupported instructions, parser/LLVM
   errors and missing tools fail. Other gates retain their normal `llc` level.

Deterministic current-Go compiler/assembler rejection may be structured source
N/A for that package/target only. Network, proxy, timeout, process or filesystem
failures remain failures. Every N/A needs an allowed kind, affected files,
attempted targets and a diagnostic. Omitted `TEXT -args` is not explicit zero.

Keep regressions for test-only declarations, concrete referenced `go_asm.h`
layouts, undeclared tail-forwarding ABI inference and exact-package asmdecl.
Never accept arbitrary frame mismatches just to make historical modules pass.

## Reports and provenance

Corpus sharding is `sha256(module@version) % 64`, independent of module-hashed
ledger files. Parallel shards read one frozen ledger and write distinct reports:

```sh
PLAN9ASM_DISCOVERY_PARALLELISM=4 \
  scripts/check-discovered-library-corpus.sh all 64
scripts/verify-discovered-library-corpus.sh _out/discovered-library-corpus
```

The `all` command replaces stale canonical reports; save evidence elsewhere
under `_out/` first when needed. Download, applicability and each target
translation/compile operation each have a 60-minute deadline. Diagnostic
override: `PLAN9ASM_DISCOVERY_CANDIDATE_TIMEOUT`. A timeout fails.
The `all` command gives each bounded parallel batch one temporary Go build
cache and removes it after every writer in that batch exits. It then starts
the next batch with a fresh cache, preventing one full ecosystem run from
retaining tens of GiB of unrelated build objects. Failed batches retain their
reports and a failing exit status; all remaining batches and the final strict
aggregate still run. Independently invoked shards keep their
own build caches unless `PLAN9ASM_DISCOVERY_BUILD_CACHE` names an existing
absolute directory; an independently supplied cache is never deleted by the
runner.

The aggregate checks exact ledger ownership/inventory and these identities:

- `selected = passed + not_applicable + skipped_invalid_source + skipped_superseded + skipped_private_extension + skipped_native_layout + failed`;
- each target's `total_asm = success + not_applicable + failed`;
- final `failed = 0`, each candidate appearing exactly once.

A passing candidate needs a successful target and no failed applicable target.
A global N/A candidate has no applicable target. Translation counts are
per-target; applicable assembly files form a unique set.

Only an exact module version with pinned source SHA-256 and independently
checked evidence can be `skipped_invalid_source`: either LLVM 22 rejects the
evaluated raw ARM64 WORD, or an AMD64 raw RIP instruction names an absent
file-local constant and Go's own assembled object places its fixed target
outside every TEXT symbol. The exception manifest is
`testdata/corpus/invalid-machine-code.json`. The runner downloads the exact
version into a disposable workspace and rechecks every witness. A changed
file, decodable ARM64 word, missing constant proof or required tool fails the
candidate; a skip
never contributes to passed candidates or translation counts. Reports and
the assembly ledger retain the reason and witnesses; progress and aggregate
verification compare them with the current pinned manifest.

An obsolete exact mirror/fork may instead be `skipped_superseded` when the
reviewed `testdata/corpus/superseded-modules.json` pins its replacement module
and a strictly newer version already recorded as scanned. Evidence links must
establish project identity, not just a similar name. The runner does not
download or translate the old version; the report and assembly ledger retain
its reason and replacement. This is not a pass. If the replacement has
assembly, it remains a separate corpus candidate and must pass the same gate.

`skipped_private_extension` is a file/target exception, not a module-wide
translation exemption. The reviewed `testdata/corpus/private-extensions.json`
pins an exact source SHA-256, raw opcode, target and source link. The runner
requires current Go to assemble that file, excludes only the pinned
file/target from LLVM translation, and still compiles all other applicable
files. Its successful translations are retained, but the module is counted
as skipped rather than passed. A changed file, failed Go assembly or failed
other translation makes the candidate fail.

`skipped_native_layout` is reserved for byte-exact native TEXT/JIT behavior,
not a missing instruction form. This is a provisional Draft PR proposal, not
an accepted coverage policy. `testdata/corpus/native-layout.json` pins the
exact module version, source file, every selected target, source SHA-256,
symbol, Go object bytes and reviewable evidence. The runner reassembles the
file with current Go for each target, checks the byte witness inside that
symbol, and compiles every other applicable file. The candidate is never
counted as passed; an unpinned target, stale source, changed object bytes or
failed remaining translation fails the candidate.

Schema 9 retains a native-only pre-filter source-selection plan, captured
before exception filtering. It covers every discovered file with its source
hash, Go source/constraint-header inputs, and every report target's selected
tags or explicit exclusion. Verification replays Go filename/build-constraint
selection, checks the original plan against that selection and the full report
matrix, applies only the exact pinned file/targets, and requires each remaining
(file, target, tags) to be executed or covered by structured source N/A evidence.
Duplicate, omitted or added scopes and forged empty remainder counts fail.
The audited assembly ledger retains and revalidates the same proof; old native
reports without it cannot be promoted.

Ordinary source exclusions require `ordinary_selection_plan` with protocol
`exact_source_matchfile_v1`. It captures de-duplicated source SHA-256 and actual
constraint/package-clause inputs, complete package/ancestor directory names,
the exact module ZIP SHA-256 and Go h1 identity, and every target/tag decision.
Related directory files (including assembly headers) also retain byte hashes;
the producer rechecks them after package checks before releasing its workspace.
Compiler/assembler/asmdecl source rejection requires a concrete source
diagnostic, not merely a positive exit or an assembler failure footer. Unknown
diagnostics remain failures for investigation. Raw reports keep the actual
diagnostic; compact ledger summaries retain `concrete_go_source_diagnostic_v1`
with its digest and canonical basename/line/column locations, without machine
paths or arbitrary diagnostic text. This witness depends on frozen producer
provenance; a digest alone does not authenticate report bytes. Old reason-only
ledger skips cannot be promoted by inventing a witness. Selection and
empty-object proofs remain separate from failed compiler invocations.
New ordinary runs additionally capture compact `cpp_inputs` before their Go
checks: exact ZIP-bound assembly/header SHA-256, registered CPP controls and
Go's fixed package-directory/tool-include search. Go export/vet checks and
translation recheck those inputs; changed nested headers, newly preferred
headers and source-proof errors fail, even alongside a native source error.
Unbound generated includes, unsupported control expansion, include cycles and
explicit inventory bounds remain failures, not empty-object N/A. This source
guard alone does not establish feature-profile or branch coverage. Historical
schema-9 reports remain historical evidence and cannot be relabeled as
schema-10 profile-aware reporting; source-diagnostic checks separately reject
reason-only rejections. CPP-only CPU predicates propose legal profiles
and require actual Go-driver observations; exclusive assembler macros are not
cumulative build tags. The producer first observes source-tag profiles and
captures their selected CPP file union, then observes CPP-required profiles.
Each ordinary `(file, target, profile ID, custom tags)` scope must have either a
concrete source diagnostic or same-profile Go export/vet, actual package-role
selection, exact CPP graph and LLVM object consumption. Progress schema 2 and
assembly-ledger v2 preserve these scopes and a canonical shared observation
inventory. Missing scope dimensions or consumer proofs fail closed. Unknown
CPU macros and experiment predicates
without an actual package-role binding fail closed. Reachable predicate sides
do not by themselves prove complete operand-form or runtime coverage.
Every package-local custom-tag configuration also includes its shared assembly
files selected by Go MatchFile. A file tested with baseline declarations is not
evidence for declarations selected by another tag configuration. Do not spread
one package's custom tags to unrelated packages or discard shared-file asmdecl
diagnostics to make the scope denominator fit.
Ordinary profiles are noninstrumented: Go partner constraints impossible with
`race`, `msan` and `asan` false do not propose profiles for otherwise ordinary
assembly. Boolean/custom/CPU alternatives and negations remain selectable.
Assembly requiring instrumentation still fails until an actual driver profile
and its source/tool/consumer contract exist; these flags are never custom tags
and instrumentation-only assembly is not an ordinary N/A exemption.
Special/test-only roles, unproved generated headers, incompatible include
binding, no-TEXT profile variants without actual empty-object evidence, and
legacy modules lacking authenticated declared-module metadata remain failures.
The private `plan9asmll -metadata-only -feature-profile ... -report ...` query
uses a separate `actual_go_cpu_profile_metadata_v1` result: it performs no
assembly translation and cannot satisfy a translation PASS or add to counts.
It loads selected Go source and same-profile actual dependency exports, then
invokes the observed Go compiler with `-asmhdr`. Every definition name, value
and presence must match the selected-types emitter, not only referenced offsets.
The compact `actual_go_asmhdr_full_definitions_v1` witness retains package/source
roles, language/target/profile identity, complete ImportMap and dependency source
and export hashes, actual header/object hashes and canonical full definitions.
Offline replay relies on frozen producer provenance; production must compare
the actual compiler bytes and full definitions independently. This query alone
does not resolve discovery's generated-include scope or branch coverage.
Macro-only and inactive CPP variants may produce a symbol-free LLVM object only
after the actual Go assembler accepts the original source, creates a nonempty
object, and emits an empty `-S` symbol/data listing under the same target,
package and registered CPU definitions. CPP proofs retain an explicit emission
kind and a separate source/profile/Go/tool/object/listing witness. Invalid source,
nonempty native listings and missing LLVM outputs fail; this is not N/A. Go's
`nm` deliberately exits nonzero for an object with no symbols, so the proof uses
the assembler's successful listing instead of interpreting arbitrary tool errors.
If an exact ZIP has no `go.mod`, the ordinary producer may use Go's canonical
proxy-generated module directive. This is separate `proxy_go_mod` evidence,
never an added ZIP member: exact ZIP and GoMod h1 sums, metadata bytes/SHA-256,
the public SumDB-signed tree and exact record inclusion are verified by producer,
compiler consumer and offline ledger readers. The actual download metadata path
and actual package metadata origin are checked before and after loading; changed
files, missing authentication or noncanonical synthesized directives fail.
Checksum metadata reads are bounded and read-only (or fetched in memory from
the official service), with no cache mutation. Cross-host comparison first
authenticates both checkpoints and retains their exact record/metadata identity.
`no_current_go_package` describes ordinary non-test source pairing only; it
does not prove that `go list -test` could not select a test-only package.
The producer checks captured files and directory names against that ZIP;
aggregate/progress/ledger readers replay MatchFile and the custom-tag search.
Offline readers validate frozen producer/source/tool provenance; they do not
independently authenticate ZIP contents without obtaining that ZIP. Header
and full-file hashes alone are not an offline cryptographic membership proof.
Every ordinary PASS and N/A must carry this proof, even when no source rejection
is recorded. Only an active verified exception uses its separate proof protocol.
Every eligible scope must have an executed configuration or a concrete scoped
source diagnostic; omissions, duplicates, changed headers and reason-only N/A
fail. The virtual root `.` is a package directory, not a hidden-directory skip.
Reasons distinguish filename targets, disabled-cgo selection, constraints,
package-clause absence, source diagnostics and explicit target ABI evidence.
`no_current_go_package` describes the ordinary non-test Go/ASM pairing scope;
it does not prove that `go list -test` cannot select a test-only assembly package.
Ignored directories and nested modules explicitly identify recursive corpus
walk boundaries, not a claim that Go cannot build an explicitly named package.
Raw diagnostics remain in reports; the ledger keeps stable source-skip categories
and compact selection inputs, never compressed third-party source archives.

The Go object witnesses establish byte-layout dependence, not runtime success.
GopherJRE's fixed +73 entry is stale on Go 1.27.1 Linux/amd64 (+63), and Sharkie's
Run+7 lands inside a CALL displacement after Go's four-byte prologue. GoJIT's
sentinel entry and fixed private frame assumptions also require runtime review.
These defects do not justify skipping ordinary instructions or other files.

Schema 9 binds Git revision/content/dirty state, full ledger fingerprint,
translator bytes/VCS metadata, matching Go build/runtime versions and LLVM 22
version/llc bytes. Before/after capture detects mutations. All shards need
identical provenance. Dirty builds are diagnostic-only; schema-2, stale tools,
mixed revisions or missing reports cannot pass. Resolve CLI repository/tool
paths before candidate working-directory changes.
GitHub PR merge and branch revisions may differ while their complete tracked
source trees are identical. The importer accepts that exact content hash match;
the translator must still match the revision recorded in its own report.

Reports prove translation/object compilation, not execution of every external
library's own test suite. Required runtime oracles remain separate.

## Progress without false completion

Query the frozen ledger and available shard reports, including a run with no
reports yet:

```sh
bash scripts/discovery-status.sh [reports-directory] [shard-count]
```

The defaults are `_out/discovered-library-corpus` and 64. `scan-status.json`
contains validated contiguous index endpoints and scan counts.
`assembly-progress.json` binds the ledger/source and lists every selected exact
version as `pending`, `passed`, `not_applicable`, `skipped_invalid_source`,
`skipped_superseded`, `skipped_private_extension`, `skipped_native_layout` or
`failed`. Its invariant
includes every one of those categories exactly once. Missing whole shards
remain pending, including in-progress shards not yet published.

The progress reader and final passing gate share the same provenance, inventory,
target and accounting checks. Stale/mixed reports or a truncated shard fail the
query instead of silently becoming coverage. `complete` means all shards and
candidates are accounted for; only `verified` additionally requires no failures.
Status-query success is not test success: the final verifier still exits
nonzero for failures or missing reports. CI publishes this view even when a
shard fails. Reports are atomically published for concurrent status readers.

Final assembly coverage has only two accepted outcomes: tested `passed`, or
an explained skip. `not_applicable` is a source/target skip, not a pass; the
other `skipped_*` statuses retain their specific reviewed reason and witnesses.
`pending` and `failed` must remain visible during development and prevent
completion. The completion gate validates each candidate's evidence, not just
the summary's `verified` flag. Every selected assembly-bearing version must
appear exactly once, including skipped versions.

Windows publication coordinates local readers and retries transient sharing/
deletion errors for at most two seconds; persistent errors still fail and keep
the previous complete report. The required Windows job runs these concurrent
publication and external-reader regressions before the full suite.

Long shard jobs publish a partial report before the first candidate and every
eight completed candidates. A runner cancellation can therefore leave
auditable passed/failed/N/A results instead of losing the entire shard.
Partial reports mark all unreported candidates and the shard itself pending;
they never satisfy the complete-coverage gate. Publication remains atomic,
and each checkpoint rechecks the frozen source, ledger and tool provenance.

Keep test evidence separate from immutable scan records: writing a pass flag
into the hashed input ledger would invalidate that report's own provenance.
Assembly-ledger files are excluded from the source-content and dirty-worktree
fingerprints because they are derived output. To publish progress while shards
run, use a separate worktree at the same source revision for the evidence
update; leave the runner worktree and its Go binaries untouched. Do not change
the source revision or scan ledger during a shard run. After auditing the frozen
reports, publish the
diff-friendly assembly status snapshot with `scripts/update-assembly-ledger.sh`.
It writes `testdata/discovery/assembly-ledger/manifest.json` and module-hashed
JSONL shards, then reads them back against the current scan and semantic
source fingerprints. Statuses include pending and failed; `verified=true`
requires complete shard coverage and zero failures. Do not promote the
snapshot to a current pass after source or scan-ledger changes. cgo scan
records remain in the separate inventory, never in this repository.

CI compares two independently verified snapshots semantically, not by physical
host tool or output-object bytes. Each original retains its exact LLVM version
and tool hash. The comparison preserves Go versions, the verified LLVM 22 major
contract, target/profile environment, registered sources/ASTs, marker/tag selection,
module/source/package role, CPP inputs and consumed macros, all four scope
dimensions, source-diagnostic kinds/positions and exact outcomes/counts. Profile
IDs are rebound to those retained semantics. Native-layout source/decision and
exception witnesses remain exact; only host ToolTags unreferenced by every
saved source constraint are ignored. This private projection is not evidence
and cannot pass a report reader. Within one run, shards still require identical
complete physical provenance. Old reports cannot be upgraded by relabeling.

## Traffic and cleanup

Inspection reads ZIP metadata/ranges and candidate sources; only matched
versions are materialized for corpus tests. Successful exact versions avoid
repeat ZIP inspection. ZIPs up to 64 KiB are fetched whole; larger ZIPs start
with the 22-byte EOCD and exact directory ranges, with 64 KiB read-ahead.
Larger EOCD scans are fallbacks. Change thresholds only with measured evidence.

The corpus uses the shared Go download cache as a read-only file proxy. New
downloads, extracted sources and LLVM outputs live in a private candidate cache
removed on success or failure. Build caches last one shard or one bounded
parallel batch within a coordinated `all` run, as described above. Keep
diagnostics, not full packages. Replaying
an uncached exact version can download
it again; that is distinct from repeated inventory work.
Each buildable Go package is translated in its own child process. This keeps
large multi-package modules from accumulating LLVM objects across packages;
the exact version is downloaded once and candidate-level file/target counts
are still aggregated and validated. Target outputs are removed after each
successful package compilation.
Public module fetches ignore the user's global Git URL rewrites and disable
interactive Git credential prompts. A failed direct fallback remains a failure
to retry, never proof that assembly is inapplicable.
Git automatic maintenance stays in the foreground. On Unix, each captured
command owns a process group; cancellation and ordinary parent exit terminate
remaining descendants before workspace removal. Output-pipe waits are bounded
and failures remain infrastructure errors. Directory-not-empty cleanup races
receive at most 20 attempts over two seconds; persistent errors fail the
candidate rather than claiming successful cleanup.

Traffic counts response-body bytes with identity encoding, not headers/TLS/IP
overhead, separating index, latest, ZIP HEAD/range and whole-ZIP requests. Do
not attribute every speedup to deduplication without accounting for changing
numbers and sizes of newly encountered modules.
