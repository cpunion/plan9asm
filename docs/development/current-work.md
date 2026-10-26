# PR40 closeout

Updated 2026-10-26. This handoff identifies scope and required work, not a
current-head completion certificate. Read [instruction contracts](instructions.md)
and [validation](validation.md). Preserve original failed receipts.

## Current priority: finish amd64 first

The public `f2e6a88b5` [CI run](https://github.com/xgo-dev/plan9asm/actions/runs/40317000859)
finished naturally: 325 jobs, 247 successful and 78 failed. All failures are
accounted for: 57 jobs share ten failed test roots, four exceed the exact-selector
argument limit, and 17 completion jobs correctly reject failed prerequisite parts.
The local repair batch preserves exact shard membership and the argument limit,
restores cross-width MXCSR targets, retains generated CPP and dynamic-frame
diagnostics, resolves physical source paths, and routes complete typed ABI0
forwarders through their existing call lowering. LLVM22's actual EVEX extension
families now drive the native ISA checker. Original positive source witnesses
and genuine rejection controls remain; no failed receipt was relabeled.

The integrated production source `48f14c4cf` passes 387 root-package roots /
22,198 children with zero failures/skips in 805.401 seconds. Separate current-SDK
checks pass shared preprocessing (9 roots / 31 children), CLI integration
(35 / 205), source profiles (6 / 45), readonly storage (6 / 65), owned processes
(12 / 47), and actual runner/selector tests (6 / 4), all with zero failures/skips.
The same process fix has an independently audited Linux
[receipt](https://github.com/cpunion/plan9asm/actions/runs/40339136434): 11 roots /
44 children, Go1.27.2, LLVM22.1.8 and Node22.23.4, including the actual failed
module cancellation test. Its fork-only diagnostic files are not part of PR40.
These results still require the complete new public CI run before qualification.
Go1.20.14 additionally passes the affected 64 roots / 2,716 children with no
failures/skips; one planned root is explicitly Go1.27-only. Builds, root vet,
all five official baselines, the ARM64 reference-family gate, tracked-source
formatting and actionlint pass. Later diagnostic/doc-only commits leave the
qualified Go source and module files unchanged.

The retired dated-branch diagnostic workflow and its scheduling-only policies
are removed, reducing the patch by 4,157 lines. All 79 semantic auditor tests
remain, as do the main/reusable CI workflows and Go tests. Node22.23.4 passes all
573 remaining policy tests with no skips. Historical receipts remain auditable
from the fork history; its extra Windows Go1.20 stress profile is not claimed as
an equivalent current main-CI run.

The complete Go 1.27.2 / LLVM22 Linux/amd64 object benchmark at `bd10bd410`
passes **37/55**, with 18 failures and zero N/A/unattempted cases. SHA256 is
restored alongside GCM, SHA512 and SHA3; no previously passing file regresses.
Wall time is 68 seconds (66.883 tool seconds); this is not a repeated speedup
measurement. SHA1 now reaches conditional selection between a readonly DATA
address and an incoming pointer. Its whole-file failure remains unresolved.
Ordinary private helper calls remain separate unresolved work, not runtime
entry exclusions.

The integrated Go 1.27.3 / LLVM22 receipt at `698edc367` passes
**244 roots / 11,064 children**, zero failures/skips, in 315.957 seconds.
Every root and child ran and terminated exactly once. It includes exact SHA,
VEX block transfers and logical zero dependencies, packed rotate inputs and
writeback, commuted pointer addition, XCHG, readonly arithmetic, target guards,
and actual retained frame-state budgets.
The packed rotate implementation replaces its duplicate parser with the shared
typed grammar (production net +10 lines); commuted ADD adds 29 production lines.
The original frame-state storage and work limits remain unchanged: only actual
CFG boundary states count toward storage. Go 1.20.14 passes
241 roots / 11,025 children at the same source, zero failures/skips, in
200.611 seconds. Its three absent roots each have an explicit Go 1.27 build
constraint; none was skipped during execution. Root build/vet, formatting,
actionlint, all five official-corpus baselines and the full 679-test Node22.23.4
policy suite also pass on this source. Later production batches require their
own qualification; documentation-only successors preserve the tested Go files.

The 16 selected Go 1.20–1.27 x86 reports at `9f649eebd` change only 78 amd64
classifications: 71 rotate forms and seven SHA forms now require their missing
input/target context. Restoring just those 78 statuses reconstructs each prior
coverage fingerprint exactly. All encoder inventories/fingerprints and all386
classifications remain unchanged. The current-SDK five-target report agrees
with the reviewed baseline: amd64 has 2,072 IR probes and 4,664 context forms,
not 100% initialized/object/runtime coverage. Initialized family and native
receipts above remain separate. The full Node22.23.4 policy gate passes679 tests
after retaining the original readonly negative-control fingerprint separately
from an appended independent test owner; its earlier 678/679 failure is retained.

The public CI's Linux native owners 4 and 14 now pass at the actual PR merge
checkout `e9c9fd902` (contribution head `f2e6a88b5`), Go 1.27.2 / LLVM22.
Their complete disjoint inventories contain 46 roots / 283 children and
36 roots / 474 children, respectively, with no failures/skips. XCHG checks
21 kernels over four inputs; readonly packed arithmetic checks 38 kernels over
four inputs. Both compare original Go execution with all three APIs at O0/O2.
Actual CPU and opcode checks passed. This qualifies these native families,
not the failed CI run as a whole or native AVX512 execution.

The earlier Linux native [receipt](https://github.com/cpunion/plan9asm/actions/runs/40305171644)
qualifies `44fe11785`: **35 roots / 430 children**, zero failures/skips, in
511.625 seconds. Source identity and the complete unique inventory were checked.
It includes all seven SHA instructions, SHA live/dead memory faults, VEX block
transfers, logical self-zeroing and commuted pointer addition against original
Go. The added AVX2 encoding-writeback check executes LLVM against an independent
register-bank oracle; it does not use original Go as its oracle.
SHA's unaligned behavior is compared with
the original Go opcode on the actual CPU; it does not assume AES's alignment
fault or relax the private-frame contract. The preceding SHA fault receipt
remains failed. That earlier native receipt predates XCHG and owned readonly arithmetic;
portable LLVM execution does not qualify native AVX512.

The XCHG batch shares scalar value reads/writes with MOV and retains the complete
width/alias grammar. Its final focused receipt passes 12 roots / 1,084 children;
the earlier wider failed receipt remains failed, and the complete integrated
receipt above now covers its corrected tests. Owned readonly arithmetic
retains 19 legacy forms and exact typed-state loop convergence, restoring the
complete SHA256 file. The uniform used-intrinsic target guard passes 20 roots /
576 children and is included in the integrated receipt above.
The audited Go 1.27.3 literal-alignment registration at `fdbec8c79` passes
19 roots / 68 children with no failures/skips; all 15 selected linker/layout
source files match Go 1.27.2 byte for byte. Unknown patches remain rejected.
The remaining four producer registrations independently audit 41 unique SDK
sources and pass 25 roots / 231 children. The new XCHG and readonly-arithmetic
native roots are selected exactly once by required Linux CI owners 4 and 14.
Local object compilation is not their original-Go native execution certificate.

Public PR40 remains draft; its earlier green checks at `728520def` do not
certify this publication. LLGo PR2694 is separately draft at `2e820f5c2`
with 131 passing checks and two release skips as of this checkpoint. Update its
exact translator pin only after publishing a verified fork revision.

### Earlier receipts

The earlier original-Go Linux native [receipt](https://github.com/cpunion/plan9asm/actions/runs/40274466633)
qualifies `c2eeeda34`: 19 roots / 294 children, zero failures/skips, in
316.992 seconds. It retains the earlier 176 VEX shape kernels and adds
logical, 41 shuffle kernels, readonly DATA and constant-index execution.
Every comparison uses three APIs at O0/O2. The same source's Go 1.20.14
focused receipt passes 25 roots / 708 children, zero failures/skips, in
74.720 seconds. Neither receipt certifies later source revisions.
The newer `d62b84c22` minimum-SDK receipt passes 29 roots / 646 children,
zero failures/skips, in 50.523 seconds, including CLMUL, packed shifts,
numerical constant propagation and block replay. Its expanded original-Go native
[gate](https://github.com/cpunion/plan9asm/actions/runs/40280464705) passes
22 roots / 333 children in 346.604 seconds, zero failures/skips. The complete
artifact was independently checked against its source revision and manifest.
The newer [native batch](https://github.com/cpunion/plan9asm/actions/runs/40285422731)
passes at `e6c475a9f`: 26 roots / 370 children, zero failures/skips, in
405.035 seconds. It adds counted loops, overlapping iterations, scalar lanes
and GP-aliased private frame addresses. Its complete artifact was checked
against the exact source and manifest. The [next native batch](https://github.com/cpunion/plan9asm/actions/runs/40289414541)
passes at `fae1aa8bc`: 29 roots / 397 children, additionally covering
bounded DATA indices, early returns and separated counter updates.
The exact artifact passed independent source/inventory checks, with no
failures/skips, in 449.777 seconds. It cannot certify later source revisions.

The integrated Go 1.20.14 / LLVM22 receipt at `fae1aa8bc` passes 63 roots /
2,341 children in 79.946 seconds, zero failures/skips. The exact root inventory
and every root/child terminal were checked. It includes range refinement and
revocation, loop updates, private aliases, machine return boundaries, scalar
lanes and the earlier CLMUL/shift/GP/replay regressions. Current-SDK-only source
oracles and Linux native execution are separate gates.

The integrated `b2f3f0c81` focused receipt passes 31 roots / 1,240 children,
zero failures/skips, in 85.176 seconds. It includes whole scalar-lane functions,
counted loops, full GP/FLAGS regressions and three-OS three-API O0/O2 objects.
Candidate discovery now follows possible full-word frame-address copies; the
existing independent proof still checks reaching values and every access.
Private LEA retains checked offsets without turning pointer bytes into NUM.
Earlier failed receipts retain the missed LEA transfer and the subsequently
fixed nil-proof panic in standalone transfer tests.
The constant-index batch `74998514d` retains 120 address combinations and
30 original Go functions; its isolated receipt has 16 roots / 269 children,
zero failures/skips, across three OSes and three APIs at O0/O2.
The shared VEX shuffle batch `0a319cd38` has a separate 37-root / 946-child
receipt with zero failures/skips, including legacy/MMX/386 regressions. Its
42 complete source functions retain both X/Y widths and unaligned memory.

The subsequent block-replay optimization `544aece11` adds 38 production lines
while preserving instruction budgets, CFG meets, eight SP residues and final
independent replay. Its isolated receipt passes 26 roots / 603 children.
Twelve analysis benchmarks, each with five samples, reduce allocation counts
65.1–88.4%; the 5,000-instruction loop median changes from 514 to 105 ms and
471 to 63 MB. On the same `14ebac865` source baseline, the full 55-file pass
set and all failure diagnostics remain identical. One wall observation falls
97 to 65 seconds; it is not a repeated end-to-end speedup measurement.

Next: publish the qualified batch and watch current-head CI and coverage.
Continue conditional value joins for complete SHA1 after the shared scalar
transfer is validated; retain every alternative's memory and origin contract.
Ordinary helper call continuations still need independent implementation and
execution tests. Retain each failed receipt until its complete replacement
passes; do not relabel an unresolved source as a runtime exclusion.

The receipt chronology below retains earlier failures and narrower evidence;
it does not replace this latest snapshot or certify a later source revision.

Follow [x86 convergence](x86-convergence.md) before expanding other architectures.
Preserve shared 386 and existing other-architecture tests. The review baseline
is `728520def`: ordinary whole-source regressions, repeated family parsing and
incomplete initialized official-form coverage remain despite green PR checks.
Local batch `d324dda54` repairs ten unaligned private-scratch transport forms,
shares the packed add/subtract/D-Q low-multiply pipeline, retains the independent
legacy rule inventory, and reduces splat/mask IR construction cost. MMX and
word low-multiply retain their separate lowerers. These changes do not restore
every ordinary crypto source or qualify every official form/native instruction.

Frozen local receipts use LLVM22.1.8; counts below exclude skipped roots:

| Snapshot / gate | Go | Passed roots / children | Skipped roots |
|---|---|---:|---:|
| `d324dda54` expanded x86 regression | 1.27.3 | 81 / 4,875 | 0 |
| `b71e2e46f` complete transport CLI sources | 1.27.3 | 3 / 19 | 0 |
| `b71e2e46f` focused minimum-SDK regression | 1.20.14 | 8 / 376 | 0 |
| `b71e2e46f` complete rule reader and scanner | 1.20.14 | 115 / 2,231 | 0 |

At `b71e2e46f` three software oracles were skipped by an unrelated AMD64 launcher
requirement. The test-only follow-up now executes all 17 shadow operations,
18 masked-memory and 144 register-view cases on the host, retaining AMD64
objects and the separate native EVEX paths. Its current/minimum-SDK patch
receipts each pass 7 roots and 35 children without skips. The frozen integrated
result above includes those oracles; neither receipt qualifies native AVX512.
Original failed reports remain unchanged. An initial minimum-SDK command used
an unsupported `-p1` spelling and did not run tests; the successful receipt uses
`-p=1`.
None of these local receipts certifies a later revision or a fresh CI run.

The full Linux/amd64 standard-library object benchmark at `d324dda54` still
fails: Go 1.27.2 / LLVM 22 compiled 33 of 55 files, with 22 failures, no N/A and
none unattempted. The file inventory and failed-file set match the review
baseline. It took 35 wall seconds; this one warm observation does not establish
a performance improvement. An earlier run with GOFLAGS set failed source-profile
checks and cannot supply a translation performance measurement.

Follow-up commits retain exact STEXT/STEXTFIPS identity (`266edb094`), separate
zero-frame siblings from nonzero private-frame permissions (`9afd76c39`), and
register the independently hash-matched Go 1.27.3 Windows unwind sources
(`6d587c49d`). They do not grant private helper, native pointer or runtime roles.
Official initialized joins (`86daf969e`, `c02a8adc3`) cover 291 canonical rows
and 912 formats plus two aliases through all three APIs and LLVM22 objects on
Linux/amd64, under Go 1.27.3 and 1.20.14. This is a bounded family scope, not
whole-architecture completion or an execution certificate.

`95278dfd6` reuses typed GP effects for all arithmetic widths and replaces the
DAG-only private-frame analysis with a bounded fixed-point worklist. Numerical
LEA and sized NOP operands no longer require fabricated memory certificates.
Its Go 1.27.3 focused receipt passes 13 roots / 1,160 children; Go 1.20.14 passes
16 roots / 1,175 children, additionally covering the mixed-frame source batch.
Both have zero skips and failures. Actual objects include three target OSes
and independently captured Translate/Module/GoModule paths. Execution comparison
for this new batch remains pending; preserve the earlier failed receipts.

The complete Go 1.27.2 / LLVM22 Linux/amd64 benchmark at `95278dfd6` still has
33/55 passes and 22 failures, zero N/A/unattempted, with 41 wall seconds and
39.714 seconds reported tool duration. Diagnostic progress is not a file pass:
gcm/sha256 now reach automatic split-entry checks, p256 reaches private helpers,
and sha512 reaches frame/DATA composition. The complete failed set is retained.
Do not report this single run as a speedup over the earlier 35-second result.

`f0a32f004` finishes feature inference after validated logical/vector-move family
dispatch. The 23,168-row before/after feature outputs match; the two targeted
inference microkernels improve about 81–84%, with zero allocations unchanged.
That narrow result does not establish end-to-end translation improvement.

The next integrated batch ends at `4cd3b05bb`:

| Scope | Evidence and remaining limit |
|---|---|
| Automatic split entries (`a69fd4b7b`) | Exact generated TLS/stack-growth machinery is separate from source-written CALL/TLS. Three OSes, two frame sizes and two source flags retain actual Go objects; split entries combined with DATA still need work. |
| Native-input/private-frame composition (`ed6a9fe56`) | Typed input addresses retain faulting reads alongside private scratch. Logical arithmetic and IMUL reuse the existing scalar memory path; this does not authorize pointer arithmetic or arbitrary memory writes. |
| Packed byte boundaries (`b8dd81caa`) | Five existing raw families share physical-memory decoding: 2,067 actual Go forms, complete SHA512 880-PC and math 101-PC inventories. These are boundary observations, not translation or execution certificates. |
| SDK frame/DATA (`98033e8e8`) | Complete original DATA and source-PC observations compose under the same owner. SDK metadata resolves only against the captured root and original inputs. The focused final receipt passes 4 roots / 12 children without skips. |
| Call-local source projection (`4cd3b05bb`) | Three APIs retain exactly two live verification rounds for both one and four functions. Independent source parsing and ordered feature inference run once per call; per-function access/CFG replay and final revocation remain. |

The integrated Go 1.27.3 / LLVM 22 receipt at `4cd3b05bb` passes 6 roots / 27
children, with no failures or skips, including all 32 GP kernels through three
APIs and O0/O2 objects. That object root takes 8.5 seconds versus 58.05 in the
earlier local receipt. This is a single observation, not a repeated benchmark
or a native execution certificate. The integrated minimum-SDK receipt at
`5223375ef` passes 13 roots / 220 children in 31.235 seconds under Go 1.20.14
and LLVM 22, with no failures or skips. It retains the snapshot revocation and
work-count checks, GP/native-input objects, physical packed boundaries, source
paths and all 291 official packed/min-max rows (912 formats plus aliases).
Darwin's linker emitted its existing `-ld_classic` warning; the original
combined output is retained separately from the parsed test events.
The separate fork-only Linux native
[probe](https://github.com/cpunion/plan9asm/actions/runs/40256528274) qualifies
the same source head: 4 roots / 30 children pass with no failures or skips under
Go 1.27.3 and LLVM/clang 22.1.8. It executes the 32 GP kernels against original
Go through three APIs at O0/O2, plus the 112/84 scalar arithmetic value/fault
and FLAGS/alias kernels. The first diagnostic run failed before testing because
of a shallow-checkout revision query; keep it failed. The passing diagnostic
commit differs from `4cd3b05bb` only in its workflow, which must not enter PR40.
These baseline tests do not qualify the separately pending AVX512 execution.

Complete-source results have **not** improved: Go 1.27.2 / LLVM 22 still reports
33/55 passes and 22 failures, no N/A/unattempted. The frozen before/after source
projection observations take 40s at `98033e8e8` and 44s at `4cd3b05bb`; do not
claim an end-to-end speedup. SHA512 now reaches an exact parsed-source mismatch
at the CLI handoff; AES/GCM and SHA256 still need split+DATA composition.

The subsequent integrated code batch ends at `ea3915207`:

| Scope | Evidence and remaining limit |
|---|---|
| Complete CLI input (`f5ea05e03`) | Canonical preprocessing precedes source/signature binding. Private DATA accepts the existing injective `<>` or `$local` renderer without weakening original identity. Focused current-SDK source/object checks pass; complete SHA512 next reaches an unregistered VEX effect. |
| Input origins (`4a069268b`) | Full-word affine arithmetic and comparisons preserve an explicitly typed input origin, not numerical address bits or pointee extent. Control-flow joins intersect origins instead of requiring dead spill facts on every path. |
| Shared legacy effects (`5d5d0e0df`) | Six packed families share typed byte/alias/alignment effects and a call-local immutable syntax plan. The agent receipt passes 18 roots / 1,486 children. Production grows by 110 lines; full-proof allocation counts decrease, but noisy timings do not establish a speedup. |
| Physical shuffle and split+DATA (`6b35e55fb`, `04bdd25ac`) | Complete shuffle boundary decoding and exact generated-entry/body relocation partitions compose. Current-SDK agent receipt passes 16 roots / 320 children, including 246 original Go forms, 36 objects and complete SHA256 source capture. Capture is not whole-file translation or execution. |
| Defined GP bank (`ea3915207`) | BP/R14 remain unknown at entry but may hold explicitly defined values. LEA and scalar MOV reuse complete machine grammar with separate source roles. The independent O2 fault RED exposed a dropped dead BP read; native-input MOV now uses the existing exact-width volatile load/store paths. Private-stack and incoming-g permissions are unchanged. |

The frozen integrated minimum-SDK receipt for `ea3915207` passes 58 roots /
2,082 children in 83.509 seconds under Go1.20.14 / LLVM22, with no failures or
skips. It includes 92 GP object kernels, 28 legacy binary kernels, all 291
official packed/min-max rows, five-target LEA and scalar-move regressions,
source revocation and 22 actual-host scalar fault kernels through three APIs
at O0/O2. The last host oracle is not original AMD64 execution. Earlier current
receipts retain the genuine BP-role and dead-access failures; a later broad
current receipt still failed one SP address-only classification, repaired before
this minimum-SDK run. The separate current-SDK Linux native
[gate](https://github.com/cpunion/plan9asm/actions/runs/40264352972) now passes
13 roots / 178 children, with no failures or skips, at source `ea3915207`.
It compares the 92 GP, 28 binary and 14 affine kernels with original Go through
three APIs at O0/O2 (4,140, 840 and 56 rows respectively). Go 1.27.3 and
LLVM/clang 22.1.8 were used; AVX512 and later VEX changes are not qualified.

The VEX batch at `6015b1be7` shares existing family parsers with the fixed-frame
effect consumer, including upper-zero byte semantics and exact memory widths.
Its independent current-SDK receipt passes 32 roots / 1,385 children, with zero
failures/skips: 146 official Zvex rows produce 354 original-Go source witnesses
and three-API O0/O2 objects. This is not exact encoder-row selection or native
execution evidence. Candidate detection remains allocation-free; its focused
median improves from 34.704 to 8.071 microseconds, not an end-to-end speedup.
Production grows by 166 lines; counterpart-family compaction remains pending.

The complete 55-file benchmark at `6015b1be7` still passes **33/55**, with
22 failures and no N/A/unattempted cases (39 wall seconds; 37.130 tool seconds).
SHA512 has advanced from a missing VEX effect to ordinary Go `_K` address
binding; SHA256 still lacks private-DATA address composition. GCM reaches the
automatic frame machine-shape check. The no-DATA ordinary-relocation capture
prerequisite at `84163349c` observes complete original objects on three OSes;
it does not grant global address or memory permissions.

Subsequent verified batches (not yet a new complete-source or native receipt):

- `ff15c4d8b` binds ordinary Go variable addresses through the existing actual
  typecheck, layout and alias checks. Mutable globals stay external, never
  private constant DATA. The agent receipt passes 11 roots / 192 children,
  including 54 three-platform LLVM objects. Its mutable-global native oracle
  is compiled but not yet executed.
- `43fc0e320` validates the signed 64-bit ADJSP delta rather than only positive
  ADD/SUB immediates. Actual Go frame128 failed on all three OSes before the
  fix; the integrated stack/global receipt passes 10 roots / 270 children in
  122.199 seconds, with zero failures/skips and altered-width/register/delta
  rejection retained.
- `71472ce1d` adds 31 VEX counterpart members (62 official rows / 186 source
  forms) through the shared legacy/VEX binary effect builder. The original
  354 forms remain, for 540 original-Go witnesses with three-API O0/O2 objects.
  The agent receipt passes 38 roots / 1,878 children in 115.170 seconds, plus
  three min/max roots / five children. Six duplicate wrappers are removed;
  typed effect code shrinks by 21 lines, total production grows by 19 lines
  including physical decoder adapters. Neither count is native execution or
  exact encoder-row-selection evidence.

Next: finish VEX counterpart effects and their independent original-byte joins,
repair complete-source frame boundaries, then rerun the 55-file benchmark.
Compose ordinary mutable Go global
addresses with existing captured declaration identities; never reclassify a
Go variable as private readonly assembly DATA. Complete current/minimum-SDK
integrated checks and initialized official-family joins. Keep source-written
CALL/TLS and private helpers distinct from assembler-generated entry machinery.
Do not add another per-proof opcode whitelist or generalize to another arch.
The CI chronology below is historical, not the current local batch's status.

## Scope and authority

PR40 owns instruction translation and first-party tests for 386, amd64, arm,
arm64 and wasm. External scanning, inventories, ledgers and compatibility batches
live in [llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
Do not run an inventory scanner locally or restore those jobs to translator CI.

Push only to an allowed fork, then update its contribution PR. Never directly
push, merge, approve or close upstream PRs. The user's CI-rerun exception permits
workflow/job reruns for every repository; it does not authorize cancellation or
other upstream mutations. Keep [PR40](https://github.com/xgo-dev/plan9asm/pull/40)
and staging [PR5](https://github.com/cpunion/plan9asm/pull/5) draft until actual
current-head tests, CI, review and coverage qualify. Resolved threads are not
reviewer approval.

## Historical CI repairs

Published head `6a39cdd831e3` is still under validation in
[run39999357637](https://github.com/xgo-dev/plan9asm/actions/runs/39999357637).
Its race31 receipt records the internal 3600-second command deadline, not an
instruction assertion or proven OOM. Three already-passed complete roots used
47 minutes in that shard. The local repair distributes the complete numerical
revocation and ARM64 immediate-scratch owners alongside the existing79 owners;
it preserves every root/API/target and the original deadlines. The full runner
regression and679 Node policy tests pass locally; current-head CI remains required.

LLGo PR2694 separately exposed complete386 SDK math rounding rejection, Linux
self-host OOM and a WASM worker finalizer timeout. The closed x87 control-word
scope now passes all70 register/control/layout cases, the complete SDK source
on three targets, Go1.20 regressions and LLGo's complete Windows386 math archive
locally. Native original-Go/LLVM execution is required in the dedicated x87 job;
object success is not that execution certificate. Preserve all failed receipts.
The Linux memory and WASM finalizer failures remain separate LLGo work, not
reasons to exclude assembly sources or claim CI completion.

## Earlier integration baseline

Published head `0d8676fdd233` completed
[run39955468231](https://github.com/xgo-dev/plan9asm/actions/runs/39955468231)
naturally: 325 Actions jobs, 322 successful and three failed. Two jobs failed
while installing LLVM dependencies, before Go tests; the Go1.26 module
completion then rejected the missing root10 receipts. The local installer
repair uses official Ubuntu mirrors and treats partial index updates as errors.
Its complete Node policy suite passed 676 tests; real installation still needs
the next CI run. The earlier Go1.20 root4 deadline failure did not recur.

The integration repair shares typed logical/AES reads of declaration-bound Go
globals and bounded indexed reads of complete private numeric RODATA. Direct
global reads have independent LLVM22 value/canary execution; indexed tables
have complete-family objects and actual LLGo SDK maps archive checks, not a
new indexed-table runtime certificate. KMOV byte effects preserve unrelated
index ranges. Raw KMOV shares the existing exact-source replay and volatile
native read/store path; unknown inputs and mode32 physical GP-qword forms
remain explicit context failures, not skipped or successful forms.

LLGo PR2694 remains separate and draft. Complete maps and Windows386 atomic
objects now pass locally, and a real Go1.26.10 native smoke passes. The new
Windows386 atomic runtime roots still need their actual CI lanes. Go1.22.2
native smoke has exposed a further ARM64 bytealg selected-declaration failure;
keep that failed receipt and repair the integration rather than excluding it.
These are local checks, not a current-head CI completion certificate.

## Earlier completed runs

Published head `fbb288364141` completed
[run39936529939](https://github.com/xgo-dev/plan9asm/actions/runs/39936529939)
naturally:325 jobs,323 successful and two failed. The Go1.20 root4 job exhausted
its unchanged45-minute shard deadline; its own artifact retains43 incomplete
roots. The Go1.20 completion's own log rejects that same failed/incomplete part.
All other producers and completions passed. The local scheduling repair balances
the existing complete heavy families in ordinary and coverage modes as well as
race; seven native conversion owners bring the closed family inventory to79.
The actual inventory and all deadlines remain unchanged. Full runner regression
passed103 roots and313 children, and656 Node policy checks passed. These local
results do not change the failed published run or qualify the next batch's CI.

The next translator batch also repairs zero-iteration REP/REPN comparison
branches. It admits only bounded source continuations whose two successors have
the same complete constant result without observable accesses; it does not invent
incoming FLAGS. Independent tests cover the complete SCAS/CMPS width, prefix and
direction family, actual386 SDK IndexByte source, LLVM22 objects and portable
values. CLD/STD retain virtual DF on portable targets and physical effects only
on matching x86 backends. Required native Linux counterparts remain separate.
Revalidate the integrated snapshot before publishing.

Published head `a5504a23c768` completed
[run39916058698](https://github.com/xgo-dev/plan9asm/actions/runs/39916058698).
Its terminal outcome remained stable for225 seconds:325 jobs,290 successful
and35 failed. Own-job logs identify two stale test
expectations: loaded typed bool arguments are now valid, and I32Clz is now
implemented, so its empty-stack diagnostic is underflow rather than unsupported.
The local corrections retain upper-register/unknown-input rejection and test an
actually unknown WASM opcode separately. Independent bool execution covers the
real typed argument path; do not merely delete the old checks. Two runner-loss
failures have unavailable logs and are not attributed to OOM or source errors.
The35 failures comprise22 direct test owners, two lost runners and11 downstream
completion failures. Each completion's own logs retain its failed parent or
missing artifacts; they establish no third instruction failure. Preserve the
lost runners' missing logs and nonterminal step observations separately from
the completed workflow. Publish only after validating the next frozen batch.

Contribution head `da6521352eb1` completed
[run39899167605](https://github.com/xgo-dev/plan9asm/actions/runs/39899167605)
naturally: 325 jobs, 213 successful and 112 failed, stable for more than three
minutes. The failed jobs comprise 101 direct parent shards with 13 distinct
test roots and 11 strict aggregation failures propagating those failed parents.
Each aggregate's plan and tool/environment observations match its failed
parent; this is not evidence of a separate plan-identity defect. The local batch
repairs raw-shift facts, relocation fixture budgets, SVE test declarations and
argument mismatches, and TEST flag expectations. Revalidate the integrated
snapshot before publishing; the old failed run remains failed.

[Run39882172980](https://github.com/xgo-dev/plan9asm/actions/runs/39882172980)
at contribution head451bf650 finished naturally with all325 jobs successful.
Its terminal state was observed unchanged for more than five minutes. This
certifies that published revision's job outcomes, not newer local changes,
every instruction form, unavailable native hardware, or reviewer approval.
The patch-coverage check also succeeded. Keep the PR draft during review repairs.

The failed historical run39716083123 at head8e658698 remains recorded separately:
213 jobs, 171 success, 34 failure and eight cancelled. Its logs exposed
interrupted Go/Node downloads marked successful, a non-retryable
MSYS2 directory collision, complete CLI package deadlines, Go1.26.10 Windows
producer rejection, conversion execution/proof deadlines and artifact-upload
failures. Some cancellation causes and a failed race shard's first test were not
recoverable from the original logs. Do not turn those unknowns into hardware
failures. Attempt2's actual race root12 rerun passed259 roots and8851 children,
with eight explicit skipped roots retained. Its only newly executed downstream
completion still failed on the original root14 plan. This does not establish
the missing original failure's cause or rerun carried-forward jobs. Keep both
attempts' original receipts immutable.

Windows native FMA3 and scalar-precision failures had their own unavailable
AVX512 CPU/OS preflight receipts. Those historical failures remain failed.
Windows CI now explicitly selects the optional AVX512 execution policy described
in [validation](validation.md#native-and-cross-runtime-semantics): new runs may
report capability-unavailable SKIP while retaining mandatory object checks.
The default and explicit `required` mode still require native execution.
Do not infer another owner's capabilities from those receipts, LLVM host names,
a signal alone or a passing object test.

## Review repair batch

The review of head39e71ec61 identified real regressions as well as misleading
coverage interpretation. Published451bf650 already includes the EXTR operand
order and vector structure-memory repairs. Subsequent local changes add the
ten named ARM64 move-wide/bitfield aliases, demand-selected ARM64 declarations,
shared target checks, named SVE count feature propagation, mixed-bool frame
roles, and bounded reads of ordinary Go globals from supplied declarations.
Validate the actual final batch before publishing; older focused passes do
not certify a later snapshot.

Whole-file regression checks must accompany independent instruction tests.
The math/big carry chain exposed TEST's data-independent CF/OF writes. go-hex
also exposed omitted TEXT argument extents, packed shift/mask bit precision,
and length-dependent initialized-byte facts across control-flow joins. Keep
each remaining failure explicit; fixing one diagnostic is not whole-file
success. Do not grant numerical pointer access, arbitrary global writes or
initialized scratch bytes to make these cases pass.

The scan report now separates IR-text probes, LLVM object evidence and actual
execution. Context is not supported, registered manifests are not fresh test
results, and WASM's isolated probe is explicitly not run. Initialized-context
reprobes now have an explicit `mask-logical` profile: all32 typed opcodes and
214 concrete occurrences from Go1.27.2's official amd64 corpus passed IR and
LLVM22 object compilation with synthetic initialized inputs. Original source
classification and fingerprints remain unchanged. Runtime is still `not_run`;
unregistered context occurrences retain their exact sources without credit.
Other initialized-context families, broader386 forms, remaining WASM control/SIMD support, and
complete-source performance still need work. External inventories stay in the
compatibility repository; do not restore scanner jobs here.

The typed WASM integer-count family covers all six I32/I64 CLZ, CTZ and POPCNT
names in Go's opcode table, including defined zero-input behavior. Its independent
oracle checks 798 inputs using original Go/WASM and LLVM-generated WASM across
three APIs, two targets and both IR/backend optimization settings. These are
family-specific execution checks, not evidence for other WASM opcodes or SIMD.
Minimum-Go execution requires the selected SDK's own WASM JavaScript helpers;
auto-downloaded toolchain modules can omit them. Use the complete same-version
official SDK, not another version's helper or a runtime skip.

The complete eight-op WASM F32/F64 Add/Sub/Mul/Div family shares the typed binary
grammar and quiet-NaN handling without fast-math. Independent original-Go and
translated-WASM checks cover25088 input pairs, three APIs, two targets and both
IR/backend optimization axes. Current and official minimum Go passed the same
13 roots and238 children. This is not evidence for BrTable or SIMD.

Named x86 integer EQ/GT comparisons propagate initialized bytes only for complete
paired lanes. Partial result knowledge does not shorten a memory operand or
grant another lane a value. Raw encodings, standalone queries, K results, MMX
and immediate comparisons retain their strict input boundaries. Independent
portable LLVM execution covers the 24 legacy/VEX register forms; the matching
48 register/memory functions have three-target object checks. These checks do
not certify original native x86 execution or the remaining go-hex CFG contracts.
For all eight SDK classification snapshots, precisely sixteen VEX X/Y register
forms now permit an unused result in the isolated void IR probe. Their result
bytes remain unknown: extracting a returned value or using an unbound memory
address still fails. This reviewed classification delta is not new runtime
evidence; preserve the independent probe/result boundary regression.

The local declaration-bound global-read extension shares the scalar ALU and
explicit IMUL grammars, retaining exact volatile widths. Executed independent
value/flag/canary checks found and repaired missing byte/word logical PF writes.
Same-origin pointer subtraction now also recognizes the original typed FP
parameter, without granting unrelated pointers numerical or dereference roles.
The actual xxhash assembly file translates and compiles to an LLVM22 object;
this is not an ecosystem ledger or original native execution certificate.

Loaded GoModule declarations now use the existing AMD64 typed argument planner
for virtual aggregate/FP argument leaves. The call-local role supplies no
native ABI, spill home, closure or pointee permission and expires at return.
Internal resolution must not repeatedly format all syntax or query all layouts:
complete source checks remain at call boundaries, including after caller layout
callbacks. Preserve mutation/revocation and internal-work-count regressions.
Closing that call must also drop its package, syntax, type, plan and callback
references. Returned signatures retain only a closed private tombstone, not the
entire input graph. Public signatures and caller-owned LLVM modules are unchanged.

The simple CLI now forwards loaded imported declarations and uses one bounded
whole-source GoModule fallback only after a context failure. Its complete CLI
suite passed38 roots and39 children without skips. The new explicit-context
entry lets this caller dispose its LLVM module and context together, while the
legacy API retains its existing ownership contract. The rich CLI's bounded
selected-source bridge now passes the actual complete math/big file and LLVM22
object, same-package/imported declarations on386/amd64, four complete bytealg
private-tail sources and output-time mutation controls (seven roots,16 children).
The latter exposed and fixed an incorrect raw-TEXT/public-metadata count check:
validated private tails are consumed before public entries are returned. Exact
defined global ELF function symbols, not textual IR quoting or internal helper
counts, now check the public inventory. It adds no Go query or AST cache;
unrelated native/source roles retain their existing route. The integrated full
rich-CLI suite at `5a8bbb8c8` failed after1035 seconds:387 roots passed and one
root failed, with no skips. Its two CRC32 negative checks now unexpectedly
succeed through the ordinary declaration route. Determine whether those inputs
are legitimate ordinary reads or a mismatched selection before changing the
assertions; retain the failed complete-suite receipt.
The follow-up independent RED separates a genuinely loaded v2 profile, which
passes, from four mutations of the captured environment, which wrongly pass.
The fix includes the exact ordered environment in the existing selection hash;
it does not normalize duplicate keys, reload packages or add a baseline-only
restriction. CRC32's legitimate ordinary reads have an actual object positive,
while missing-owner rejection still publishes nothing. The shared reader and
source/profile controls pass16 roots and63 children in the repair worktree;
the complete integrated CLI at `cc19ba5da` then passed389 roots and2155 children,
with no failures or skips, in926 seconds. This does not erase the earlier failed
complete-suite receipt or qualify later edits and other modules.

The strict full-source benchmark at published `a5504a23c768` still fails:
linux/amd6432/55 passed in280 wall seconds; linux/arm6440/47 in7 seconds.
All102 files were attempted and30 failed. The integrated rerun at `5a8bbb8c8`
passed33/55 amd64 files in64 wall seconds and40/47 arm64 files in5 seconds.
All102 files were attempted,29 failed and none were N/A. The math/big file now
passes and the four bytealg private-tail sources retain their original success.
This is an improvement, not a passing benchmark or a claim to main's speed.
Five ordinary crypto failures need physical private-frame storage, byte
initialization and alias handling; they are not runtime-entry exclusions.
Other private-helper and runtime contracts remain separate failed observations.

LLGo PR2694 remains a separate draft at `684bb1a0e524`, with an exact public fork
pin to translator `a5504a23c768`. Its latest106 visible checks are terminal:
42 failed, including41 own-log varlen closure failures and one lost runner;
three cancelled and11 skipped checks are not successes. The separately reviewed
runtime-owner candidate still needs actual compilation and execution checks.
Its producer checks now pass for four target profiles, but both real execution
fixtures stop during compilation at ARM64 atomic Or32's result-tail write.
Do not drop those observable native bytes to make the test run: the existing
native-caller boundary remains intact. This is still failed runtime integration,
not successful execution or new instruction coverage.
Its previous `aa438347ce34` run finished naturally with six successes and seven
failures; that is historical evidence only. The cgo-preamble, Windows UTF-8
collector and fork-provenance notice changes are published, but remaining ABI
failures are not resolved by those changes. Local imported-source handoff and
typed argument repairs pass the actual AMD64 compare/count/index/indexbyte
files. Equal still needs the real hidden closure producer/transport contract;
386 helpers now distinguish incoming pointers from the canonical result-byte
pointer and share ordinary scalar/extension and whole128 read effects. The
complete386 equal file now reaches its varlen closure load, not a helper-read
failure. This is still a failed whole-file observation, not runtime success.
Publish a verified translator revision before updating the exact fork pin.
Do not fake closure registers, relabel pointers as integers, or filter failed
public roots to turn these integration failures into passing tests.
LLGo's actual equality function value carries a Type pointer, whereas stock Go's
varlen consumer expects a captured byte count after the code pointer. Connecting
those layouts directly would read PtrBytes as a length and can report unequal
pointer-free values equal. A real producer/transport or complete runtime-owner
replacement belongs to the separate runtime integration work below.

## Closeout mechanisms

- Go installation requires the installer's resolved stable version and the
  actual PATH command to match the requested exact patch or historical minor.
  GOTOOLCHAIN=local forbids an implicit substitute. Installation has an
  independent ten-minute deadline; original test budgets do not increase.
- All relevant Node22 owners use one exact22.23.4 installation and verify both
  the official action output and actual command. An ambient22.23.2 caused the
  observed producer/verifier mismatch. Preserve complete PATH/tool hashes;
  never normalize them away to accept incompatible scopes.
- MSYS2's three bounded attempts retain updates, signature checks and exact
  LLVM packages. Each owns a distinct runner-temporary installation so a
  failed update's existing directory cannot make both retries fail immediately.
  The complete successful LLVM prefix moves without overwrite to one fixed
  runner-temporary path. Absolute prefix/resource checks and a native C
  compile/link/run must succeed there before publishing build bindings, so
  retries cannot change strict cross-owner tool paths. Failed attempts remain;
  no global installation is deleted or reused.
- Go1.26.10 is explicitly registered only for the Windows-unwind producer.
  Its eight complete producer inputs are byte-identical to the audited schema.
  Unknown patches, changed sources and generic compiler archives remain rejected.
- Linux CLI normal tests retain all four SDKs and both modules; coverage retains
  both current-SDK modules. One reusable workflow owns four complete root parts,
  eight shards and the actual profile/rest roles. Part0 executes the latter
  without empty runner jobs. All six reports and a fresh strict union are
  mandatory; missing evidence or failed producers fail completion.
- Native fixed-local conversion retains all336 live/dead/canary cases in seven
  independently scheduled48-case roots, original Go execution, actual ISA/OS
  preflights, all three independently captured APIs and LLVM O0/O2 execution.
- Conversion objects retain all741 forms, all three targets and twelve scheduling
  batches per target. Each batch captures two sources of at most32 functions.
  Every API keeps its original independent five-minute lease and complete
  source/object/lifetime checks. Only disposable compiler artifacts share a
  test-owned cache; no source observation, binding or Verify result is cached.
- Cancellation regressions wait for an actual root's atomic readiness signal
  outside observed source before starting the original cancellation window.
  Planning keeps its existing bound; every exit cancels and joins the owner.
  Full local runner qualification may use its existing exhaustive partitions;
  retain the failed single-binary15m observation separately. Formal main-module
  ordinary/coverage commands remain45m and race commands60m, unchanged.
- Complete readonly DATA, AES, typed transport and non-temporal consumers use
  the existing fresh owner for each API instead of sharing one five-minute
  binding across sequential consumers. GoModule receives the complete captured
  preprocessed source, including DATA/GLOBL. Negative mutation, expiry and Close
  checks deliberately keep their supplied binding; never renew it to make a
  rejection test pass. Compiler-cache setup is separate from source authority.
- Native-wide packed execution keeps all340 original kernels in six required
  top-level source-batch owners. Each owns all three APIs and pipelines with
  the unchanged actual AVX512 preflight. The six source/C batches form the exact
  original union; each owner still executes the unchanged whole Go oracle.

Independent REDs precede these repairs. A local policy, object or source-plan
pass is not a required native runtime pass or a final CI certificate.

## Verification and publication

1. Integrate one coherent batch, commit it and freeze source, docs, controllers,
   tools and environment. Work elsewhere while it runs; never certify a later
   edit with an older receipt. Only one heavy Go/LLVM owner runs locally.
2. Validate changed scheduling and tool guards with complete Node policies,
   actionlint, current/minimum/race TestCI and the complete plan9asmtest suite.
   Execute both CLI drivers' real normal/coverage owners and fresh aggregates.
   Distinguish local current-SDK transport from four-SDK Linux CI execution.
3. Retain complete conversion object unions and independent source-lifetime,
   missing-input, revocation and cleanup controls. Run required native batch
   roots on their actual Linux hardware; do not substitute portable values.
4. Run build/vet, formatting, five-architecture official coverage, complete
   instruction corpus including every WASM object, and fresh sixteen-report
   x86 SDK classification comparison. Review any delta before changing baselines.
5. Push one verified allowed-fork batch after the active run finishes naturally.
   Watch actual current-head jobs, investigate every failure and check review
   comments. Missing uploads, cancelled jobs and unexecuted hardware axes remain
   explicit, never successful coverage.

Use currentGo1.27.2, root minimumGo1.20.14, rich CLI minimumGo1.24.13 and LLVM22.
CI's shared Node is22.23.4 and cross execution uses pinned QEMU10.2.3.
Missing required tools fail; no !llgo exclusions or WASM blockaddress.

## Separate runtime integration

The retained strict standard-library benchmark had133 passing and51 failing
target/file pairs. These remain failed historical observations, not PR40
instruction successes. Full source/runtime entry contracts, including reflect
dynamic calls, TLS, system stacks and preemption, continue independently in
[runtime/ABI issue47](https://github.com/xgo-dev/plan9asm/issues/47) and the
manual/weekly SDK/runtime integration workflow. Moving their scheduling does
not resolve or relabel their failures.
