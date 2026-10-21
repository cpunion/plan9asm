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

## Online run and publication boundary

The active allowed-fork contribution is draft
[PR5](https://github.com/cpunion/plan9asm/pull/5). Its formal
[run 39535494403](https://github.com/cpunion/plan9asm/actions/runs/39535494403)
uses frozen `2c208003`, not the subsequent local integration branch
`codex/pr40-ci-closeout-integration-20261021`. Let that run finish naturally;
do not cancel, rerun or publish another tracked batch while it is active.
All ten standard-library instruction corpus owners and all six ARM scan owners
have passed. Other owners still have software and independently observed
hardware failures. No final CI, coverage-upload or reviewer-approval claim is
valid until its actual completion evidence is checked.

The older naturally completed run `39462713494` at `51d546835` had 104 PASS and
13 FAIL. Its failures and the later run's raw artifacts remain historical
failures, even where a separately frozen local repair now passes. Repeated
SDK/owner failures are not independent instruction defects. Read each native
root's own allocation receipt before attributing a failure to hardware.

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
- At `2c208003`, the complete 33-command warm replay also passes: 263 Go roots
  and 26,000 children across overlapping lanes, zero failed, skipped or
  incomplete terminals. All sixteen official classifications have zero
  unreviewed delta. The complete instruction matrix retains all 44 scopes and
  144 WASM objects. These scopes are not additive and do not certify later heads.
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
  reports; 25 format children plus the original O0/O2 mathematics pass in
  separate current-Go, actual minimum-Go and race receipts at `06d305eaf`.
  No trimming, extra output or reduced count is valid.
- Coverage-upload setup selects Python3.13 only after Go qualification and
  completion evidence. Python, PATH pip and the installed CLI are checked;
  original OIDC/files/fail-on-error inputs remain unchanged. Its independent
  RED and full 250-test Node GREEN are not a real Windows upload certificate.
- Six complete typed families now cover 41 spellings: floating shuffles,
  immediate blends, implicit mask stores, legacy integer min/max, qword
  permutations and signed widening. Shared widening and element-permute
  machinery replace duplicate lowerers, without borrowing raw/ABI roles.
  Keep Type4/E4NF full accesses distinct from E5 active-lane suppression;
  MASKMOV retains its original non-temporal opcode and MMX/EMMS owner.
- At `1c326b4bb`, all 45 original public conformance entries independently
  pass source closure with the complete unchanged ASM and Go declarations.
  The same receipt's three stale test contracts still fail; preserve it.
  Later test-only repairs retain every original row, Go object, public
  Context and physical-format partition. Source closure and five-target
  objects are not original native execution or whole-CI qualification.

## Current unpushed repair batch

Each following receipt belongs only to its named snapshot. The integrated
head still needs its own full current/minimum/race, policy and official-form
qualification. Do not sum overlapping child counts or promote an object check
into a native execution receipt.

- Scalar FP-to-integer named/raw continuation closure at `b06b13d05` passes
  all three lanes: 15 roots and 7,973 children per lane. Immutable raw replay,
  native opcode/MXCSR effects and unrelated source-state domains are retained.
- Owned static MOV and immediate encoding normalization at `1f34021b8` pass
  current/race 24-root and minimum-Go 23-root scopes, each with 1,810 children.
  The original 42-row relocation program retains its own writable storage;
  an unresolved external symbol is not writable or numerical authority.
- Mandatory isolated-Void COMI/static-MOV input proof has an independent RED.
  The complete COMI RIP literal family has another 360-child RED, followed by
  the minimal existing-literal-owner repair `8fe454d0c`. Its eight-root /
  5,932-child current-Go combination passes. It retains complete original Go
  sources, public uninitialized Context and private physical-format objects.
  At `3b2737f32`, the exact original 252 COMI runtime-fixture kernels also
  pass original-Go compilation, three-API Linux objects and host C linking:
  three roots / 60 children, zero skip. Object tools no longer depend on an
  executable x86 transport; this is not original native execution.
  Minimum/race and original native execution remain separate obligations.
- At `25fea2cc6`, five roots / 394 children pass for the complete 386 BMI2
  fixture contract. The original SP-writing kernels cannot return with their
  physical stack intact: retain them as Go-accepted Context. The unchanged C
  comparison now executes an independent safe CX/DX companion, not those SP
  originals. Actual Linux cross execution still needs online qualification.
- FMA dead-source access receipts retain exact stdout and every independently
  expected QEMU fault-child stderr line. Unknown/missing/extra diagnostics,
  command failure and cleanup failure still fail. Current/minimum/race local
  cohorts pass; required ordinary Linux owners also install their actual
  AArch64 C compiler. Full Node policies pass at the isolated repair snapshot.
- Reciprocal SIB rejection at `e0dd4feb1` binds the selected Go encoder's exact
  diagnostic profile: historical numeric index/base fields are not operand
  type IDs. Current/minimum seven-root / 313-child scopes pass; race is pending.
- Go1.22 concrete Alias lacks the later immediate-RHS API. Test-only contract
  repair `5643b7823` preserves every original row and explicitly checks Context
  when that complete chain is unavailable; it never substitutes Underlying or
  Unalias or skips a row. Actual Go1.22: four roots / 37 children pass.
  The named type's underlying node counts separately in language identities,
  not in scalar argument layout; preserve that original depth boundary.
  After the independent Go1.23 RED, the corrected five-root / 37-child scope
  passes at `09b7dbca7`, including the complete COMI mandatory-effect contract.
- Fixed-local consumers now capture an independent bounded source per API.
  Wide compilation retains all 340 kernels in six complete source packages
  (64 each, last 20), all three original optimization paths and separate API
  owners. Only disposable compiler artifacts share a parent-owned cache;
  source selection, Verify, expiry and Close remain live and independent.
  Object tools no longer require Rosetta/CPU execution. Actual cold-cache
  Translate at `53683ebd7` passes in 101.70s; Module at integrated `a073f2156`
  passes in 103.40s. Each retains 30 children and 18 object paths. GoModule,
  race, complete target unions and native 340-kernel execution are unfinished.

The old combined 19-root race cohort at `2c208003` exhausted its unchanged
five-minute command budget. It remains failed, without a race-detector report;
validate its complete union in independently bounded commands, not by raising
the budget or deleting roots. Unavailable Rosetta also caused an actual local
wide-object skip before the object-tool repair; the strict verifier rejected
that receipt. Neither is a pass.

## Reviewed replay and official classification history

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
All sixteen reports reproduce that reviewed baseline at `51d546835`.
The six new families have an independent unapproved diagnostic followed by a
fresh reviewed sixteen-report audit at `06098a256`: all eight AMD64 SDKs change
the same 134 input-deficient forms to Context, and all 386 reports are unchanged.
Complete form/encoder identities, unsupported rows and parse metrics remain
unchanged. Baseline-only `c6aaad048` records exactly 24 derived fields, not a
translation or runtime pass. Fresh final-head zero-delta comparison is required.

The current batch has a separate unapproved sixteen-report diagnostic and a
fresh independently reviewed audit at `09b7dbca7`. Each AMD64 SDK changes the
same77 input-deficient forms: 18 COMI and 59 scalar FP-to-GP. All386 reports,
complete form/encoder identities, unsupported forms and unrelated fields remain
unchanged. Those small386 corpora have no observed rows for these families;
complete five-target family tests are separate evidence. Baseline-only
`277ae4d1b` records only eight AMD64 support/context/fingerprint triples. Require
a fresh sixteen-report zero-delta comparison after integration. Historical
diagnostics, the old COMI expected-effect failure and the alias depth failure
remain failed receipts, not retroactively qualified tests.

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

PR40 is scoped to instruction translation, not the remaining runtime/ABI contracts.
PR CI retains complete instruction scanning, full WASM source/object compilation
for typed-stack/control context, and its existing object/native owners.
The complete all-architecture source/object corpus and unchanged benchmark are independently
required to succeed within the manual/weekly `SDK/runtime integration` workflow;
moving their scheduling does not turn the retained 51 failures into passes.

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
   unsupported form. Complete source/object corpus and the unchanged strict
   benchmark follow independently in [runtime/ABI issue #47](https://github.com/xgo-dev/plan9asm/issues/47),
   not as a requirement to resolve all 51 integration failures in PR40.
4. Stage one coherent allowed-fork batch and watch every job to natural
   completion. Exact-head CI/native outcomes, review and coverage are still
   unfinished. Keep pending hardware explicit and PR40 draft until qualified.

Older detailed family receipts and failed attempts remain in Git checkpoints
before this compression and ignored raw evidence. Consult that history when
needed; do not reintroduce cumulative duplicate tables into this handoff.
