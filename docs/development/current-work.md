# PR 40: current checkpoint

Updated 2026-10-19. This is a handoff, not a completion certificate.
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

[39258001772](https://github.com/cpunion/plan9asm/actions/runs/39258001772)
is still active at `e890858c5`; do not cancel or supersede it. The previously
failed `SDK priority (completed-ci-repair)` now passes. Its original CAS/AES
program keeps the restored unknown-register domain; an unknown dead AES output
does not destroy independent predicate FLAGS. No complete Windows-module pass
is certified until its actual owner artifacts and aggregate finish.

All six Linux/native priorities and both Windows ownership jobs pass.
The downloaded product/CMOV/AES/vector-test numerical artifact independently
audits four roots and fifteen execution children: nineteen passed terminals,
no skips or incomplete outcomes. Vector tests retain all six original named/raw
optimization paths. CMOV executes original Go opcode bytes; wide product/AES
portable LLVM values are not original AVX512 execution.

The preceding run `39251569206` at `755bbb710` finished naturally with
64 passed jobs, one failed CAS/AES root and the dependent Windows job skipped.
Preserve that failed evidence; the later priority pass does not rewrite it.

The last complete Windows main-root inventory is still the older
[39228817887](https://github.com/cpunion/plan9asm/actions/runs/39228817887)
at `17669bdbd`: 7,288 roots = 6,954 pass + 70 fail + 264 explicit host skips,
zero absent/incomplete. Its failed rich-CLI owners additionally retain five
failed roots; their partial inventory is not the main-module denominator.
Do not subtract later focused repairs from those historical counts.

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
`ci392515/run-final.json`, and original raw test streams.
Actual filenames may be in another persistent worktree; inspect worktrees first.
Evidence is not tracked repository data.

## Immediate next work

1. AES/vector-test repairs are published at `e890858c5`: all 24 final gates,
   eight-SDK classification audit and the old sixteen portable priority roots
   passed locally. Finish the current online run and inspect its exact head.
2. IFMA input classifications and isolated CLI extraction pass current/minimum
   Go and race at `997d70af5`. Its eight-SDK audit changes only eighteen AMD64
   mask-bearing probes per SDK to Context; 386, all unrelated opcodes, encoder
   and unsupported metrics stay unchanged. The baseline-only update is local
   at `300b51833`. Finish its final gates before publication. The next native
   numerical audit additionally requires the unchanged original IFMA source/C
   oracle at backend O0/O2 and optimized-IR O2; portable LLVM is not AVX512.
   Batch qualified repairs only after the preceding run has ended.
   Never cancel or supersede active diagnostics to obtain a newer result.
3. Once priorities pass, collect the complete Windows module artifacts.
   Audit actual inventories, disjoint ownership, raw terminals and strict
   module completion before claiming a full pass or publishing coverage.
4. Continue remaining original failure classes: qword bit lookup,
   funnel/align, floating dot, FPCLASS/GETEXP/GETMANT/SCALEF/range/fixup/rounding,
   scalar conversions/moves, native POP/MOVNTI, and physical SP/BP/g-register
   or undefined-W-shift fixtures. Reproduce first, consult complete Go rows,
   share typed grammar/effects and repair a class rather than one library.
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
