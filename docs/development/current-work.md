# PR40 closeout

Updated 2026-10-26. This handoff identifies scope and required work, not a
current-head completion certificate. Read [instruction contracts](instructions.md)
and [validation](validation.md). Preserve original failed receipts.

## Current priority: finish amd64 first

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

Next: finish automatic entry-layout, ordinary typed-pointer/private-scratch and
readonly DATA composition, then rerun complete sources. Keep source-written
CALL/TLS and private helpers distinct from assembler-generated entry machinery.
Complete the GP execution oracle and remaining initialized official family joins,
then remove repeated parsing and measure whole-source/generated-program cost.
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
