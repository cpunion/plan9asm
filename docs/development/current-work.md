# Current work: PR 40 checkpoint

Checkpoint: 2026-10-15 09:36 UTC. The latest fork diagnostic completed naturally
with nine stale BP test-root failures, not a tool or timeout failure. Six roots
and their complete object matrix are repaired and locally integrated; three
remain in development. Main also contains the qualified string-memory/REP-CFG
repair. Do not publish this partially integrated batch as qualified.

The previous complete main qualification at `68ef8aa3` passes548 unique owners;
its strict benchmark remains FAILED158/184 with26 failures, zero N/A and zero
unattempted files. Classification, objects and executed semantics stay distinct.

## Current blocking scope

1. Finish test-only repairs for PSHUFB, signed integer-to-float conversion and
   CMOV's old blanket386 BP negatives. They retain actual Go acceptance, all
   applicable widths/masks/targets and three APIs; SP and amd64 BP remain
   physical-frame negatives. The other six repaired roots plus complete64-name
   BP objects pass21 unique owners on Go27/actual Go20 and13 on race at
   `ce65e120`. Main contains the equivalent commits `c3f6b1b3`/`ec654367`.
2. Complete scalar MOVB/W/L/Q logical-FP byte transport in the TLS leaf.
   Test-only `6caaaca0`/`94e72a08` reproduce534 actual wrong returned values,
   183 valid-interior rejections and69 passing controls. `51964b2a` now passes
   all786 semantic cases plus229 boundary/status controls on Go27. This is
   only a quick gate: full objects, Go20, race and retained roles remain due.
   It fixes named GP MOV views, not all arithmetic/compare FP consumers or
   specialized source-stack/fixed-local roles. Never retag legal tuples Context.
3. Finish LODSL zero-count TDD in the string leaf. `f66bc165` and test-only
   target-layout correction `57eb5c48` remain failed because the first observer
   assumed incidental aggregate constant folding. No production repair is
   certified. Preserve all84 original source functions and prove the observer
   with mutation controls before claiming the target-only wrong-high32 RED.
4. Integrate qualified leaves, update this checkpoint, then freeze source,
   documentation and tools once. Retain every previous548 owner, ten new string
   owners and every new BP/FP/state root; deduplicate package/root ownership
   without dropping SDK lanes. Rerun official gates, vets and strict benchmark.
5. Stage the complete batch in the allowed fork, preserve all34 old early CI
   lanes and add early checks for these real failures. Push once and let the
   entire new run finish naturally. Draft PR40 is not ready; its contribution
   head is still `24851df2` at this checkpoint.

One local heavy SDK/ASM/LLVM/object/native/benchmark command owner at a time.
Release its lease immediately after natural completion, before read-only
archiving or further edits, so another existing worktree can make progress.

## Scope and authority

This repository owns translation, source/tool observations, official instruction
coverage and independent regressions. External scanning, ledgers and package
compatibility belong to [llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
Never run its index/module inventory locally or restore those gates here.

Push only to an allowed `cpunion` or `dastai` fork. Do not directly push, merge,
close or approve upstream PRs. PR40 remains draft at `24851df2`; diagnostic
branches are not the contribution head. Base `e2164ac0` is already an ancestor.
Read [instruction rules](instructions.md) and [validation](validation.md).
Use LLVM22 exactly, Go1.27.2 and actual Go1.20.14 for portable root compatibility.
Missing tools fail. Freeze all tracked inputs, tools and documentation during
qualification. Preserve failed snapshots, deadlines and complete denominators.

## Latest completed CI

The latest [run38703136882](https://github.com/cpunion/plan9asm/actions/runs/38703136882)
at `e6a29969` ended naturally FAILURE:71 jobs,63 successful and eight failed
(seven root parts plus their correctly failed aggregate). Independent complete
raw audits find exactly nine failing main-module roots and385 failing children,
with no missing, duplicate or unfinished terminals. All are the mode-aware BP
test-fixture class above; every other job succeeds, including native64/208
children, ARM QEMU and both nested CLI aggregates/coverage uploads.

The complete main inventory has22 packages/6976 named roots/155190 children,
263 explicit host-inapplicable root skips and no child skips. The simple CLI
has32 roots/29 children and no skips; the rich CLI has333 roots/2095 children
with one explicit ARM host child skip. These are exact-run outcomes, not
certification of newer local source or an all-platform executed pass.

### Historical checkpoints

[Run38691683296](https://github.com/cpunion/plan9asm/actions/runs/38691683296)
at `de238372` has67 successful jobs, one failed root part and one remaining
package owner at this checkpoint. All37 priorities passed. Independent raw
audits confirm native64 roots/201 children, ARM64 QEMU four children/512 input
rows, and complete Compare21 children/12 LLVM objects. These are current-run
executed scopes, not certification of the newer local source.

The failed part has exactly one failing root,
`TestX86ABI0ResultTupleScalarTypesAndObjects`, across all five targets. Original
Go objects pass; the new mixed-I1 preflight incorrectly rejects valid typed bool
input copies. Do not fix this by changing those positives to Context. Test-only
`845d9c92` reproduces120 rejected positives with240 negative controls passing.
Production `f46199bf` supplies a bounded ordinary ABI0 canonical byte proof,
retaining raw2/unknown/join/layout rejection and physical I8 intermediate storage.
Its20 portable roots/1246 children pass Go27/race/Go20, and the unchanged tuple
object root passes Go27. The combined object group remains failed because four
new BP fixture roles lacked `textflag.h`; `acae3606` corrects those inputs and
adds physical-byte/layout controls. Final merged qualification is still required.

Earlier fully completed run:

[Run38679026699](https://github.com/cpunion/plan9asm/actions/runs/38679026699)
at `79e484c9` ended naturally SUCCESS:68/68 jobs, including35 priority jobs,
both ownership SDKs, native63 and all three complete Windows module aggregates
and coverage uploads. No cancellation or rerun is being mistaken for success.

Downloaded current-head native JSON independently passes63 roots/169 children,
15 inventories,25 mandatory owners and eight required children, with no skips,
missing, duplicate or unfinished terminals. The repaired complete ARM64 carrier
matrix independently passes27 children in priority118.805s and in its actual
covered Windows shard155.07s. Its parent-owned immutable build cache fixes the
previous repeated observer compilation without sharing mutable source owners.

The simple CLI completes32 roots without root/child skips. The rich CLI completes
332 roots with one explicit host-inapplicable ARM64 native child. Main completes
all22 packages; its root-package inventory contains6347 roots and263 explicit
host-inapplicable root skips. These are successful strict host-module results,
not an all-platform executed pass. Earlier `41251144`, `fcff729f` and other
failed diagnostic runs remain failed; historical detail is retained in
`git show d214b793:docs/development/current-work.md` and ignored raw reports.

## Earlier implementation evidence

| Class | Qualified scope and remaining boundary |
| --- | --- |
| Complete ARM64 CLI oracle, `8f4b3396` | Adds the missing pinned-QEMU Linux counterpart to the existing cross owner. Local complete native oracle executes512 inputs; strict four-child auditing, runner Go27/race/Go20, actual tool dependencies, policies and seals pass. Linux/QEMU execution of this new source still requires the next fork run. |
| BSWAPL/Q, `2797cc31` | One complete Go-table register grammar feeds lowering, byte/canonical effects and opcode extraction. Unknown result bytes fail before all nonvoid consumers; ordinary AMD64 ABI0 bool needs an independent canonical role. Go27/race281 roots/9233 children, Go20 279/9211, both SDKs' complete grammar and five-target Go/LLVM objects and official gates pass. |
| Ordinary386 BP, copied as `ec593a57` | Allows ordinary386 BP without admitting SP or AMD64 BP. Both SDKs' banks/objects pass; original42-TEXT atomic source still returns its independent CALL Context, not file success. |
| Typed AMD64 Go argument carriers through `cc0e26f6` | Actual selected compiler rules, original archive/SymABI and complete source bind9 GP/15 FP banks, grouped values and whole-argument rollback. All independent pure/race/Go20 guards pass. This extends argument transport only, not result, closure, spill, native entry or caller memory authority. |
| Actual LLVM entry inspection, `e56a7dda` | Checks real function identity and grouped LLVM types rather than printed identifier quotes. Nine parser/type controls pass Go27/race/Go20; complete original Compare3-TEXT source passes all three translation APIs in97.437s. Equal correctly retains hidden `8(DX)` Context in all three APIs. Neither is a new object/native/file-benchmark receipt. |
| Complete Compare objects, `9f74695f` | All three APIs retain all three original TEXTs and compile12 LLVM22 objects across original/O2 IR and O0/O2 backend axes. Its21 children pass in96.387s; Equal remains Context. Final source/tool seals and both SDK vets pass. The subsequent default CLI benchmark independently compiles the entire original Compare file successfully. |
| BSWAP native fixtures through `afc18aa4` | Integrated as `a12697f4`; both SDKs' complete Go grammar, five-target objects and freestanding C/ASM adapter objects pass. Go27/race284 roots/9290 children and Go20 282/9268 pass. Actual original-Go-opcode/LLVM execution on Linux AMD64/386 remains pending, not established by adapter compilation. |
| Shared mixed-I1 boundary through `abf771ac` | Genuine REDs reproduce noncanonical bool aggregate false passes both with BSWAP and in180 MOV-only combinations. The shared preflight and direct-module fallback require an existing complete canonical result role; a copied fixed-local token cannot supply it. Leaf Go27/race287 roots/9774 children, Go20 285/9752, both SDKs' grammar/objects/adapter, official gates and seals pass. Independent code review reports no concrete finding. Ordinary AMD64 ABI0 scalar-I1 remains a separate pre-existing boundary. |
| Mode-aware logical BP through `76a4c52c` | Source CFGs, paired native loads and the bounded private value stack distinguish ordinary386 BP from SP/AMD64 frame BP. Shared complete stack forms preserve only unwritten GP identities; whole helper CFGs are included before unchanged budget accounting. New pure controls pass at `f46199bf`; expanded original-Go/LLVM fixtures must be rerun after their header repair. LOCK remains an independent complete input/address proof boundary. |
| Inactive original386 profile, `d4c3307d` | Qualified live capture and portable framework are merged. GOAMD64 must be observed but is inactive for Original386; active GO386 and strict DATA/NUM/AMD64 profiles remain exact. This does not solve the complete atomic source's independent CALL boundary. |
| Typed string grammar, `fa45ba44` | A shared complete20-name descriptor validates implicit Ynone forms, modes, suffixes and operand counts. Six new roots/524 children plus retained grammar/objects pass both SDKs/race at the leaf. Existing CMPS/LODS host skips are not native passes; REPN source/memory/result authority is unchanged. No code-size reduction is claimed. |

At `870d011f`, independent merged-source qualification passes514 unique owners
across112 commands, retaining all source/namespace/SDK/CLI owners, both SDKs and
race. Its official gates and before/after source/tool seals pass. That receipt
cannot certify the subsequent BP/profile/string/tuple integration.

## Retained gates

The current ordered plan is at the top of this document. Preserve all six
NoDATA owners, complete CLI Count/IndexByte sources, both SDKs, race/vets,
official forms and before/final source/tool seals. Object compilation is not
native execution; Equal retains its closure Context. Native64's17 required
children remain exact. Never merge fork diagnostic workflows/auditors into
the main source branch.

Do not infer a speedup from changed caches. A static audit finds fixed repeated
verification stages, not per-instruction disk verification, in Compare. No
measured algorithmic improvement has been established; do not cache mutable
permission facts or remove before/final callback guards speculatively.

## Ownership and preserved leaves

Use existing persistent worktrees and inspect their status before edits.
The primary branch is `codex/pr40-external-corpus-split-20261011`. One heavy
SDK/ASM/LLVM/object/native/benchmark owner runs at a time. Small audited pure
groups may overlap. Do not interrupt another owner's frozen snapshot.

Active leaves are `codex/pr40-bp-fixture-peer-families-20261015`,
`codex/pr40-mixed-bool-fp-span-red-20261016` and
`codex/pr40-x86-string-zero-count-state-20261016`. Resolve their existing
worktree directories with Git; do not create disposable temporary clones.

The BSWAP branch is `codex/pr40-bswap-definedness-20261015`; its qualified
native tests and shared mixed-I1 guard are integrated through `abf771ac`.
The argument-carrier branch
`codex/pr40-amd64-go-argument-carriers-20261016` is frozen at `cc0e26f6` and
integrated. Fork staging is `codex/pr40-fork-native63-staging-20261015`;
published diagnostics use `codex/pr40-windows-owned-process-probe-20261013`.

The inactive-profile leaf is now qualified at `d4c3307d` and merged; its live
capture, six retained source observers and seven-root/81-child portable roster
pass with exact before/final seals. Only two portable roots are new, not seven.
Preserved `e47ca042` is a private386 terminal-CFG consumer, not a public factory
or activation. Complete original CALL/unwind/alternate-entry boundaries and
permanent revocation remain required; no noreturn/nounwind or source truncation.
Do not discard either leaf, any older WIP, stashes or failed reports.

## Official classification and strict benchmark

Both SDKs' unchanged five-architecture classification gates pass at the BSWAP
leaf. Go27 has zero unsupported/parse errors; Go20 retains one known malformed
AMD64 maskless EVEX.Z encoder form. Classification is not full encoder-name,
LLVM-object, runtime or complete standard-library coverage: the observed386
corpus has21 distinct operations and121 wasm forms remain Context.

Latest strict complete benchmark remains FAILED at `68ef8aa3`:

| Target | Complete files passed / total |
| --- | --- |
| linux/386 | 26/33 |
| linux/amd64 | 46/55 |
| linux/arm | 27/30 |
| linux/arm64 | 40/47 |
| js/wasm | 19/19 |
| Total | 158/184;26 failures, zero N/A/unattempted |

Target wall times46/286/2/6/2s and total342s meet the original300s/target and
900s total budgets. Its184 exact target/file pairs equal the `870d011f` scope;
only the entire original386 Compare file changes to success, with no regression.
Do not infer a speedup from one second of wall-time difference. Retain the
command's failed exit1 and all26 failures rather than reporting performance
success as compile success.
Remaining roles include TLS/g/entry, reflect dynamic frames, preempt flags,
VDSO/syscall frames, hidden closure context and ARM64 result-tail backing.

## External inventory: server only

Fresh manifest generated2026-10-15T05:09:35Z contains23,202,445 index entries,
1,090,783 unique module paths,6962 assembly-bearing version records,
1,149,060 scanned records,17,082 retained retry failures and1,021,860 latest
records. Its1,143,236 cgo scan records are not the cgo-positive count.
All640 recorded ranges were independently checked adjacent and their entry sum
matches the manifest. Derived history boundary is2025-10-01; incremental boundary
is2026-10-15T04:27:23Z. The head-only service naturally ended with exit0 after
reaching its captured start-time target. Do not call this a scan to a later now.
These counters are not translation, LLVM-object or runtime passes.

Resolve server details from private operator configuration, never tracked docs.
Sync only a coherent completed export or under the writer lock, not a live
ledger. No local scanner or new ledger files belong in this repository.

## Evidence

Ignored `_out/register-result-repair/` retains `ci-79e484c9/`,
`arm64-fallback-cross/` and `combined-arguments-bswap/` with frozen manifests,
fresh inventories, raw Go JSON, final audits and before/after source/tool seals.
The complete failed benchmark is `benchmark-final-870d011f.../`, including its
schema-2 results and identical before/after source/tool/head/status manifests.
Leaf receipts remain in their respective worktrees. Preserve the failed
`a1a6e9dc` Compare printer-spelling assertion and the corrected `e56a7dda`
receipt separately. Only the actual tested head and scope are certified.
