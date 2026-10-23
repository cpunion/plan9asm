# PR40 closeout

Updated 2026-10-23. This handoff identifies scope and required work, not a
current-head completion certificate. Read [instruction contracts](instructions.md)
and [validation](validation.md). Preserve original failed receipts.

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

## Last completed upstream run

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
reprobes, broader386 forms, remaining WASM scalar/control/SIMD support, and
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

LLGo PR2694 remains a separate draft at `aa438347ce34`. Its 13 workflows finished
naturally with six successes and seven failures. The cgo-preamble, Windows UTF-8
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
