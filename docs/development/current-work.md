# PR 40: current checkpoint

Updated 2026-10-21. This is a bounded handoff, not a current-head completion
certificate. Read [instruction contracts](instructions.md) and
[validation](validation.md) before continuing. Preserve all ignored failed
reports and their original source/tool/environment seals.

## Scope and publication

This repository owns instruction translation, complete typed families and
first-party source/ABI observations and tests for 386, amd64, arm, arm64 and
wasm. External discovery, inventories, ledgers and compatibility batches live
in [llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
Do not restore those workloads here; no inventory scanner runs locally.

Push only to an allowed `cpunion` or `dastai` fork. For other owners, contribute
through the allowed fork's PR; never directly push, merge, approve or close
upstream PRs. Recheck upstream divergence before publishing a coherent batch.

PR40 remains draft at `24851df2`; local integration and fork staging are not
that contribution head. The sole formal review thread is resolved, not reviewer
approval. Retain its orphaned-IR and symlink-include regressions. Do not mark
ready until exact-head CI, review and coverage qualify.

## Latest completed online run

Fork run [39437363187](https://github.com/cpunion/plan9asm/actions/runs/39437363187)
finished naturally at `e7c6b5fa7`: 117 jobs = 101 PASS + 16 FAIL.
The complete raw-owner audit identifies 27 unique failed test roots.
Local repairs below do not subtract from those failures. No current integration
head has subsequently completed successful CI.

Keep software failures separate from actual-allocation CPU/OS-state failures.
Late jobs also exposed Go patch drift; the pinned-tool repair needs its own
new run. Neither tool drift nor a translation error is hardware unavailability.
Required owners and the final aggregate must still reject failures or missing
outcomes; diagnostic scheduling does not weaken qualification.

The user deferred unavailable AVX512 execution to pending-hardware acceptance.
This is not skip, pass, emulated native execution or whole-CI success. Original
owners still require their actual CPU/XCR0 receipt; do not infer missing ISA
from a SIGILL alone or assume different hosted allocations have equal features.

## Recent qualified local scopes

Each receipt certifies only its recorded snapshot and scope, with unchanged
complete source/tool/environment seals. Roots and children overlap: never add
them or turn object/portable checks into original native execution claims.

- `89992764d`: complete 25-root / 804-child repair combination passes current
  Go in 26.344 seconds, with no failure, skip or incomplete terminal. This is
  not minimum-Go, race, original native execution or later-head qualification.
- `4ed846d96`: the eight static-storage capture/live-selection roots pass with
  59 children in 4.867 seconds. Initial capture still forces actual compiler
  archives; repeated live selection alone omits `-a`.
- At that performance snapshot, SDK CPU's eight targets and SDK Globals'
  five targets pass as two roots / 13 children in 22.976 seconds. These are
  actual source/archive/report checks, not instruction-runtime execution.
- Five TOCTOU, final-callback revoke and concurrent Verify/Close roots pass
  with 12 children in 178.061 seconds. No capability cache, weaker live hash,
  restored revoked lease or omitted final guard is authorized.
- Complete original bigmod compiles once in 45.29 seconds with the same ASM
  digest prefix `32fd` and archive digest prefix `894c`; full digests remain
  in the ignored receipt. Earlier normal runs were 177.680 and 201.710 seconds.
  These observations are not a controlled speedup measurement, successful CI
  or a replacement for the complete strict benchmark.

## Latest replay repair and official classification

The independent origin-only mutation scope has an actual 24-child RED at
`dff66cd11`. Its minimal three-line production repair is `91e530f73`:
retain complete original raw VPERMIL producer replay even when a mutable origin
field is removed. Do not invent named/NUM permission to bypass that boundary.

At `91e530f73`, the complete 31-root / 9,514-child current-Go combination now
passes in 41.123 seconds: zero failure, skip, incomplete terminal or stderr,
and identical complete before/after seals. It does not certify minimum/race,
a later documentation/integration head, original native execution or CI.

The independent VPERMIL audit joins all sixteen Go1.20–Go1.27 x86 reports.
Each AMD64 SDK changes exactly 45 input-deficient forms from supported to
Context. All 386 forms, encoder/form identities and every unrelated metric
remain unchanged. Baseline-only `059d894ac` records those audited classifications.
Context is not initialized-program success; do not count it as executed code.
Fresh final-head comparison of all sixteen reports is still required.

VPERMILPS/PD now share complete immediate/variable typed grammar, selected-byte
effects and exact E4NF native reads, including dead/Void and zero-mask access
obligations. Original complete source/Go objects, public unbound Context,
private emitter formats and initialized five-target objects stay distinct.
Keep real mode32 physical-bank restrictions and Go named-grammar restrictions
separate. Scanner and isolated supported-op extraction have explicit roots.

The original `TestAMD64ConformanceLLVMRuntime` was attempted on this ARM64
host. Its existing AMD64/Rosetta guard skipped because Rosetta was unavailable
to the actual probe. The strict runner rejected that receipt. Do not record
this root as a local pass or replace its original ASM/C with object-only evidence.

## Strict benchmark remains failed

The last complete strict benchmark is `4d129bcd3`: all 184 target/file pairs
attempted, 132 PASS and 52 FAIL, zero N/A or unattempted. Complete seals and
timestamped completion/report joins match; original failed evidence remains.
Target wall times are 56/505/5/11/2 seconds for 386/amd64/arm/arm64/wasm,
total 579 seconds. AMD64's 505 seconds exceeds its unchanged 300-second budget.
The unchanged 900-second total limit does not excuse a failed target limit.

Source/frame, helper/code-address, inventory and special runtime/OS-entry
contracts remain distinct blocking layers. A complete bigmod pass does not
promise other full SDK files pass or remove this benchmark's failures.
Do not raise budgets, remove failed sources or relabel them N/A to qualify.

## Non-negotiable implementation boundaries

- Consult full official Go/ISA rows; fix a complete typed family and formats,
  not a package or individual failing fixture. Use independent TDD priors.
- Actual initialized inputs, architectural undefined results, missing input,
  format rejection and required source context are different contracts.
  Go rejection requires exit 1 plus a bound diagnostic, not panic or timeout.
- Raw continuations need immutable whole-producer replay and physical
  generation/architecture identity, not just spelling or an origin pointer.
  SDK literal/numerical owners precede native-family hooks.
- Preserve native memory width, faults, dead accesses, flags and FP/MXCSR
  effects. Private mathematical instrumentation is not original ISA execution.
  Do not grant ordinary frame/caller/ABIInternal/aggregate roles to fill gaps.
- Preceding source facts alone may justify dead components. Future or
  shape-only facts cannot flow back into reachable states. Keep the shared
  16,384-visit / 64MiB budget and independent accounting controls unchanged.
- Use exact LLVM22 and selected Go versions; missing required tools fail.
  No `!llgo` exclusions, tool fallback or wasm `blockaddress`.
- Only one heavy local Go/LLVM owner runs at a time. Release LLVM objects and
  owned processes promptly. Freeze source, documentation and used tools while
  a receipt runs; never reuse it for edited code or overwrite failed evidence.

## Next verification and publication

1. Freeze the final integrated head including this checkpoint. Its complete
   current-Go, actual minimum-Go and race qualification remains unfinished;
   root minimum is Go1.20, rich CLI minimum is Go1.24, current SDK is Go1.27.
2. Complete the updated 33-command gates: build/vet, scanner actual inputs,
   isolated extraction, policies, shared state and official five-arch coverage.
   All original roots, command counts and time limits remain mandatory.
3. Independently compare all sixteen final-head classifications and complete
   fingerprints; preserve the audited VPERMIL delta and historical Go1.20
   unsupported form. Re-run the unchanged strict benchmark separately.
4. Stage one coherent allowed-fork batch and watch every job to natural
   completion. Exact-head CI/native outcomes, review and coverage are still
   unfinished. Keep pending hardware explicit and PR40 draft until qualified.

Older detailed family receipts and failed attempts remain in Git checkpoints
before this compression and ignored raw evidence. Consult that history when
needed; do not reintroduce cumulative duplicate tables into this handoff.
