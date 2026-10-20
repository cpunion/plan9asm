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

Fork run [39462713494](https://github.com/cpunion/plan9asm/actions/runs/39462713494)
finished naturally at `51d546835`: 117 jobs = 104 PASS + 13 FAIL, with no
cancelled or skipped job. Of the preceding run's 27 failed Go roots, 24 now
actually pass and three still fail. This is not a successful CI certificate.

The complete raw-owner audit identifies 15 unique failed Go roots: 13 own
CPU/XCR0 preflights report unavailable AVX512, while two are software failures.
These are the original conformance program's SHUFPS source-effect closure and
the original scaled-round C report's Windows CRLF assertion. The latter reports
35,328 observations and zero mathematical failures at both optimization levels.
A separate coverage-upload failure uses Python 3.15 beyond PyO3's supported
maximum; its module's Go qualification passed. Aggregate failures add no new
test root. Do not attribute these three software/infrastructure issues to ISA.

The user deferred unavailable AVX512 execution to pending-hardware acceptance.
This is not skip, pass, emulated native execution or whole-CI success. Original
owners still require their actual CPU/XCR0 receipt; do not infer missing ISA
from a SIGILL alone or assume different hosted allocations have equal features.

## Recent qualified scopes and staged repairs

Each receipt certifies only its recorded snapshot and scope, with unchanged
complete source/tool/environment seals. Roots and children overlap: never add
them or turn object/portable checks into original native execution claims.

- At `51d546835`, all 33 complete final commands pass and all sixteen official
  x86 SDK classification reports reproduce their reviewed fingerprints without
  changes. These receipts do not certify later integration heads.
- The complete original bigmod now actually passes Windows in 72.80 seconds;
  the ASM digest remains `32fd0923150ddd1206a59270c4349e22cb7bbb6c82508e6fa0e1c959e0c57660`.
  This is source/archive compilation evidence, not native execution or a
  controlled performance ratio. Capture still forces compiler archives; only
  repeated live selection omits `-a`. Revoke, live-hash and Close checks remain.
- Staged CI repairs let Windows schedule after ownership while retaining all
  ten original final-success dependencies. Current formal SDK scopes use the
  audited Go1.27.2, not a floating patch rejected by the observation contract.
  Historical SDKs' once-per-run exact-version resolution remains unfinished;
  do not describe the current pin as a complete patch-drift mechanism.
- CRLF has an independent 25-child RED with exactly the legal CRLF child
  failing. The minimal repair accepts only the two exact complete LF/CRLF
  reports; 25 format children plus the original O0/O2 mathematics pass on
  current Go at `06d305eaf`. No trimming, extra output or reduced count is valid.
- Coverage-upload setup selects Python3.13 only after Go qualification and
  completion evidence. Python, PATH pip and the installed CLI are checked;
  original OIDC/files/fail-on-error inputs remain unchanged. Its independent
  RED and full 250-test Node GREEN are not a real Windows upload certificate.
- SHUFPS/SHUFPD/VSHUFPS/VSHUFPD need one complete typed two-source family
  repair and independent original-source regressions; do not fix only SHUFPS
  or modify the original conformance ASM/C to bypass its proof failure.

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
All sixteen final reports reproduce this reviewed baseline at `51d546835`.
Fresh comparison remains required after subsequent translator changes.

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

The last complete strict benchmark is `51d546835`: all 184 target/file pairs
attempted, 133 PASS and 51 FAIL, zero N/A or unattempted. Complete seals and
timestamped completion/report joins match; original failed evidence remains.
Target wall times are 24/262/2/7/1 seconds for 386/amd64/arm/arm64/wasm,
total 296 seconds. The unchanged 300-second target and 900-second total budgets
qualify on this host, but the benchmark still fails its content checks.

The remaining failures comprise eight ordinary-frame, nine helper/code-address,
four SDK inventory, 21 special runtime/OS-entry, eight other-contract and one
unresolved math/big graph case. A complete bigmod pass does not remove them.
The older 52-failure receipt differs only by that bigmod pass; no source was
deleted or relabelled. The independent frame-header clobber RED is retained,
but has no production repair or native-execution qualification yet.
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
