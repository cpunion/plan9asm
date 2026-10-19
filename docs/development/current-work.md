# PR 40: current checkpoint

Updated 2026-10-20. This is a handoff, not a completion certificate.
Earlier checkpoints remain in Git; do not duplicate their accumulated tables
or overwrite ignored failed evidence. Read [instruction contracts](instructions.md)
and [validation](validation.md) before changing a family.

## Scope and publication

PR40 owns translation, complete typed instruction families and independent
tests for 386, amd64, arm, arm64 and wasm. External inventories, scanners,
ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
No inventory scanner runs locally.

Push only to a cpunion/dastai fork. PR40 remains draft at `24851df2`;
staged branches do not qualify that contribution head. Last checked upstream
main `e2164ac0` is an ancestor of this candidate; recheck before publishing.
Reuse persistent idle worktrees and preserve existing changes.

## Latest online evidence

[39285088077](https://github.com/cpunion/plan9asm/actions/runs/39285088077)
finished naturally at `47e44aa7d`: 83 jobs pass and 16 fail, none cancelled.
Fifteen main-root owners fail and their aggregate correctly rejects the run;
both CLI module owners and aggregates pass. The complete sixteen-owner main
union is 7,379 roots: 7,063 pass, 52 fail and 264 explicit host skips, zero
absent/incomplete. Every root is owned exactly once. The independent ignored
`ci392850/root-union-outcomes-audit.json` binds the raw events and inventory.
Compared with the preceding complete run, eighteen old failures now pass and
there are no newly failing roots. Local later repairs do not rewrite this
online inventory or qualify a whole-suite percentage.

Scalar lane/POP priority's exact
sixteen roots and 846 children pass, including all four original Windows
runtime failures. Its independent raw audit has zero skips or incomplete
outcomes. The numerical product/CMOV audit passes seven roots and 24 children,
including the original CMOV FP-span Go/LLVM execution and portable IFMA/lookup
C oracles. Portable integer values are not native AVX512 qualification.

The downloaded Linux floating-compare CPU receipt reports `AVX512F=0`,
`AVX512VL=0` and `XCR0=7`. That runner cannot qualify the new complete
AVX512 conversion, floating-helper or physical 512-bit/DF oracles. Keep the
required capability checks and native owners; object compilation and portable
values must not replace original-opcode execution. A capable execution
environment remains unresolved. Earlier SSE/AVX native passes are not evidence
of AVX512 support.

[39258001772](https://github.com/cpunion/plan9asm/actions/runs/39258001772)
finished naturally at `e890858c5`: 81 jobs pass and 17 fail, none cancelled.
All 65 priorities, including `completed-ci-repair`, pass. Both CLI modules pass
their complete Windows owner/aggregate checks. All sixteen main-root owners
fail, and the main aggregate correctly rejects those failed plans.

The complete main-root union is 7,326 roots: 6,992 pass, 70 fail and 264 explicit
host skips, zero absent/incomplete. Every root is owned exactly once. The
ignored `ci392580/root-union-outcomes-audit.json` retains all raw-backed outcomes;
it is an inventory diagnostic, not a complete-module or coverage certificate.

All six Linux/native priorities and both Windows ownership jobs pass.
The downloaded product/CMOV/AES/vector-test numerical artifact independently
audits four roots and fifteen execution children: nineteen passed terminals,
no skips or incomplete outcomes. Vector tests retain all six original named/raw
optimization paths. CMOV executes original Go opcode bytes; wide product/AES
portable LLVM values are not original AVX512 execution.

The preceding run `39251569206` at `755bbb710` finished naturally with
64 passed jobs, one failed CAS/AES root and the dependent Windows job skipped.
Preserve that failed evidence; the later priority pass does not rewrite it.

The older complete Windows main-root inventory at
[39228817887](https://github.com/cpunion/plan9asm/actions/runs/39228817887)
at `17669bdbd`: 7,288 roots = 6,954 pass + 70 fail + 264 explicit host skips,
zero absent/incomplete. Its failed rich-CLI owners additionally retain five
failed roots; their partial inventory is not the main-module denominator.
Ten old failures now pass, while ten previously passing roots fail under the
new checks. Diagnose both sets; do not subtract later local focused repairs
from either historical inventory.

The sole formal PR40 review thread is resolved. The no-orphaned-IR regression
`TestAMD64RejectedLegacyVectorDestinationDoesNotMutateIR` covers PSUBUSB,
PABSD and PTEST; the separate symlink-include escape regression also passes.
The exponential discovery search moved outside this repository and the
readability parentheses are fixed. Resolved comments are not reviewer approval.
Current-head coverage and complete CI remain unqualified.

## Qualified focused batches

These scopes overlap. Their counts cannot be added into a whole-suite percentage.
Each receipt binds its own unchanged source and tools; all listed three-lane
cohorts have zero failed, skipped or incomplete terminals.

| Snapshot | Complete focused evidence | Important limits |
|---|---|---|
| `579c4cba1` | MOVNTI, 36 floating helpers, conflict/funnel and raw compact/expand: 64 roots / 1,472 children, current Go, actual Go1.20 and race; identical 1,536 terminals | Complete typed input, mask, source-identity and native-access obligations. Cross-target objects and portable value/fault checks do not qualify original AVX512 or MOVNTI execution |
| `013181561` | Integrated ALIGN, DIV/IDIV, mode32 raw logical returns, original scalar formats and floating-owner policies: 53 roots / 2,130 children, current Go, actual Go1.20 and race; identical 2,183 terminals | Bound raw RET composes only with the existing register-only leaf result; no entry or memory permission. Original undefined division status remains unavailable. Required native MOVNTI/ALIGN/DIV and eight floating programs still need the next online run |
| `47e44aa7d` | Raw IFMA/lookup and shared decoder: 23 roots / 378 children, current Go, actual Go1.20 and race; identical 401 terminals | Complete original identities, initialized raw five-target objects and physical mode32 mask rows; no raw/native entry or missing-input permission |
| `9769cc2c1` | Scalar lane/POP stores and original format repairs: 26 roots / 1,219 children, current Go, actual Go1.20 and race; identical 1,245 terminals | Complete selected-byte native stores, Void/unused-bool obligations, exact volatile POP widths, private stack balance/bit/origin controls and five-target objects. Local x86 runtime skips are excluded from this certificate; original execution remains required online |
| `66dcd0d31` | CMOV FP views: 2 roots / 183 children, current Go, actual Go1.20 and race; identical 185 terminals | Original uninitialized formats remain Go objects plus Context; independent initialized AX/FLAGS sources own complete, partial, cross-field and pointer/hole checks. No new caller or FP storage permission |
| `1cebb3c38` | Branch hints and original CMOV/KVM formats: 5 roots / 15 children, current Go, actual Go1.20 and race | All sixteen conditions and both hints compile on five targets with actual word-aligned results. The original signed-division compound test still fails at IDIV; mode32 hypercall objects do not supply a logical scalar/native entry contract |
| `9738b0376` | Raw immediate compares: 14 roots / 657 children, current Go, actual Go1.20 and race; identical 671 terminals | Eight signed/unsigned B/W/D/Q members, complete Go rows, initialized X/Y/Z five-target raw objects, selectors/predicates and missing/native/standalone controls. Physical mode32 arity is distinct from the named frontend |
| `313f7afc5` | Named qword lookup plus portable numerical programs: 13 roots / 39 children, current Go, actual Go1.20 and race; identical 52 terminals | Both opcode families, complete independent attributes, every input byte, write-only K versus merging vector results and E4/E4NF reads. IFMA/lookup original C oracles execute on the actual AMD64/ARM64 host at three optimization paths; not native AVX512 |
| `eac2dcdb4` | IFMA: 16 roots / 382 children, current Go, actual Go1.20 and race; identical 398 terminals | Complete two-opcode Go rows, X/Y/Z initialized named/raw five-target objects, actual mode32 masks, every meaningful source/accumulator byte and independently nonzero upper aliases. Native numerical execution remains required in the next CI |
| `e6d340706` | AES/CAS: 22 roots / 2,471 children, current Go, actual Go1.20 and race; identical 2,493 terminals | Restores the existing nontrapping unknown-register domain without initializing unknown bytes. Original CAS predicate remains independent; observable AES data/control uses still reject. Native addresses and complete reads remain mandatory |
| `475ffc681` | PTEST/VPTEST/VTESTPD/VTESTPS: 15 roots / 72 children, current Go, actual Go1.20 and race; identical 87 terminals | Complete independent Go rows, five-target three-API objects, every-byte/state controls and exact original dead memory opcodes through IR/backend O0/O2. Native execution remains required in the next CI |
| `805a2e667` | CMOV: 16 roots / 1,075 children, current Go, actual Go1.20 and race | Complete conditions and widths, old destination, predicate and unconditional native reads. Dead named GP results may remain unknown, never initialized |
| `ae7f42184` | Rich-CLI tail/declarations: 13 roots / 246 children, current Go, minimum Go1.24 and race | Actual root/terminal declarations precede private-node heuristics; ABI selectors and strict graph/dependency checks remain |
| `4d1416874` | Raw DWORD private controls: 16 roots / 1,891 children, current Go, actual Go1.20 and race | Exact data-changing products preserve existing DF/ID/private storage, not inert-NOP or caller authority |
| `39378d6b6` | MOVBE/BMI2/oracle repair: 34 roots / 1,658 children, current Go, actual Go1.20 and race | Complete six MOVBE canonical/alias names, original Go-rejected imm64 stores, forty oracle aliases, values/faults and source boundaries. Eight-SDK audit and 22 final gates pass at later baseline-only snapshots |

Preserve the failed AES/CAS RED at `075c9832c` and the earlier overstrict AES
qualification at `a731d124d`; the latter is not the restored-domain certificate.
PTEST RED `fdaded1c0` and first GREEN `f9ceee21e` remain failed. The latter
found missing isolated native carrier features as well as uninitialized old
format fixtures. Keep original formats as Go/emitter objects plus Context,
separate from initialized public objects.

The eight-SDK audit at `796dfb1d9` changes exactly fourteen AMD64 vector-test
probes to Context and restores twenty pure-register AES probes per SDK.
No AES memory form, unrelated opcode, 386 form, encoder or unsupported metric
changes. The one historical Go1.20 unsupported form is unchanged; current Go's
five-architecture gate still requires zero unsupported forms and parse errors.

Ignored receipts include `windows-aes-cas-qualified-e6d34070.json`,
`windows-vector-test-qualified-475ffc68.json`, prior family qualification files,
`windows-ifma-qualified-eac2dcdb.json`,
`windows-qword-lookup-qualified-313f7afc.json`,
`windows-raw-mask-compare-qualified-9738b037.json`, and original raw test streams.
Actual filenames may be in another persistent worktree; inspect worktrees first.
Evidence is not tracked repository data.

## Immediate next work

1. Continue the exact fifty-two-root failure list from the completed run. The
   new local repairs are not online passes. Prioritize new regressions as well
   as the old runtime families; retain complete source, Go/emitter objects and
   explicit Context for genuinely unbound formats, with independent initialized
   public counterparts. Never relabel a valid initialized program as a format.
   The original CMOV fixed-local object and runtime roots also exhausted their
   five-minute capture context. They now retain all384 object kernels per target
   in six independent batches, and all387 runtime kernels in seven batches.
   Source/ABI capture and the original Go/C oracles remain mandatory. Independent
   policies reject missing or monolithic owners; the complete batch union is
   not yet qualified. Only Linux/Darwin batch00 has actual current-Go objects.
   Scalar lane and POP repairs have their own Windows priority, including the
   unchanged original runtime failures. Do not count the new local cohort as
   a successful online run or subtract it from the historical seventy failures.
2. The eight-SDK audit at `579c4cba1` changes exactly 286 AMD64 input-deficient
   probes per SDK, exclusively within the 52 MOVNTI/conflict/funnel/floating
   names. All 386 forms and unrelated fields are unchanged, with no additional
   unsupported form. Baseline-only commit `262e140b1` records the audited
   classification; ALIGN/DIV need their own next audit. Context is not an
   initialized program pass. Preserve the complete positive object matrices.
3. The integrated 53-root current/minimum/race receipt is
   `windows-native-vector-new-qualified-01318156.json`; the earlier 64-root
   receipt is `windows-native-vector-families-qualified-579c4cba.json`. Ignored
   source/tool-sealed raw streams remain in persistent worktrees. Finish new
   scanner/isolated-CLI checks and required native completion ownership, then
   rerun final gates on an unchanged batch before an allowed-fork push.
4. Complete named/raw conversion, duplicate, segment-query, blend and
   half-vector repairs are integrated locally, not qualified online. The
   eight-SDK audit at `db0abbfc8` allows exactly 62 family names: each SDK
   changes 290 AMD64 probes from supported to Context, zero 386 probes and
   no unrelated metric. Baseline-only `f834e4377` records that exact audit.
   Its complete current-Go conversion legacy union passes 62 roots and
   1,914 children; minimum/race and integrated final gates remain unfinished.
   Preserve failed and different-head receipts without combining them into
   a current-head certificate. Half-vector values/objects at `063ef9c19`
   pass 19 roots; the separate indirect ownership repair at `d25f0e9fd`
   passes three roots and 219 children. Full native DF/512-bit execution is
   still required. Dot/horizontal and raw-move candidates still fail their
   first complete checks; integer min/max physical mode32 is under repair.
   Continue scalar reciprocal, CLMUL, shuffle, ternary, raw compare/broadcast,
   physical SP/BP/g and undefined-shift observations. Retain original source
   and separate format objects, initialized values and ABI authority.
5. Preserve the separately qualified logical integer return batch
   `53a93cfd4` and incomplete floating MIN/MAX candidate `7009ff18f`.
   They are not merged into this repair candidate.
6. After fork CI, review comments and coverage qualify, update the contribution
   head and PR body. Do not merge/approve upstream or mark the PR ready early.

## Benchmark and verification discipline

The last strict standard-library benchmark at `18bb77a09` remains failed:
all 184 target/file pairs were attempted, 129 compiled and 55 failed, no N/A.
AMD64 took 392s, exceeding the unchanged 300s target budget; total 504s is below 900s.
No new benchmark or controlled speedup has been certified.

Use Go1.27 and exact LLVM22; root-module minimum is actual Go1.20 and rich-CLI
minimum is Go1.24. One heavy local Go/LLVM owner at a time. Commit before
validation; freeze tracked files, documentation and used tools for each run.
Develop in a different idle worktree while a snapshot runs. Missing tools fail,
never skip. Native host exclusions retain a required cross/native counterpart.
A format object, Context classification, runtime value, ABI proof and complete
CI receipt are distinct evidence; none can silently substitute for another.
