# Current work: PR 40 checkpoint

Checkpoint: 2026-10-15 05:12 UTC. Fork diagnostic `79e484c9` has finished
successfully; main's argument/BSWAP focused gate at `e56a7dda` also passes.
Neither receipt establishes complete PR40 or standard-library success.

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

## New main source, not yet published in that CI

| Class | Qualified scope and remaining boundary |
| --- | --- |
| Complete ARM64 CLI oracle, `8f4b3396` | Adds the missing pinned-QEMU Linux counterpart to the existing cross owner. Local complete native oracle executes512 inputs; strict four-child auditing, runner Go27/race/Go20, actual tool dependencies, policies and seals pass. Linux/QEMU execution of this new source still requires the next fork run. |
| BSWAPL/Q, `2797cc31` | One complete Go-table register grammar feeds lowering, byte/canonical effects and opcode extraction. Unknown result bytes fail before all nonvoid consumers; ordinary AMD64 ABI0 bool needs an independent canonical role. Go27/race281 roots/9233 children, Go20 279/9211, both SDKs' complete grammar and five-target Go/LLVM objects and official gates pass. |
| Ordinary386 BP, copied as `ec593a57` | Allows ordinary386 BP without admitting SP or AMD64 BP. Both SDKs' banks/objects pass; original42-TEXT atomic source still returns its independent CALL Context, not file success. |
| Typed AMD64 Go argument carriers through `cc0e26f6` | Actual selected compiler rules, original archive/SymABI and complete source bind9 GP/15 FP banks, grouped values and whole-argument rollback. All independent pure/race/Go20 guards pass. This extends argument transport only, not result, closure, spill, native entry or caller memory authority. |
| Actual LLVM entry inspection, `e56a7dda` | Checks real function identity and grouped LLVM types rather than printed identifier quotes. Nine parser/type controls pass Go27/race/Go20; complete original Compare3-TEXT source passes all three translation APIs in97.437s. Equal correctly retains hidden `8(DX)` Context in all three APIs. Neither is a new object/native/file-benchmark receipt. |

`e56a7dda` integrates the qualified production leaves above. Its focused gate
also passes both SDK vets and before/after3861 tracked-input/19 selected-tool
seals. A whole coupled-source gate is still required after the next test batch.

## Next work, in order

1. Finish the frozen test-only BSWAP native adapter leaf and its exact both-SDK
   Go grammar/five-target objects. The adapter compares actual Go opcode bytes,
   LLVM and an independent byte reference; it is not a new Go ABI0 entry grant.
   Actual Linux AMD64/386 execution requires an exact required native CI owner.
2. Add LLVM22 O0/O2 objects to the existing complete original Compare owner,
   without another source capture, cropped TEXT or increased five-minute budget.
   Keep Equal's closure boundary explicit. Investigate any newly confirmed
   mixed-result canonicality defect with independent class-level RED/GREEN.
3. Build a fresh deduplicated coupled manifest at the final committed head.
   Preserve independent six NoDATA owners and complete CLI Count/IndexByte
   owners, every previous source/namespace/tool root, both SDKs and race/vets.
   An inventory is preparation, never evidence of execution.
4. Qualify fork-only priorities and strict audits for the new ARM64 cross oracle
   and BSWAP execution. Batch-push only verified source; let CI finish naturally.
   Never merge the fork diagnostic workflow/auditors back into main source.
5. Rerun the strict five-target benchmark at that head, resolve remaining real
   failures, then update draft PR40 and require current-head CI/review/coverage.

Do not infer a speedup from changed caches. A static audit finds fixed repeated
verification stages, not per-instruction disk verification, in Compare. No
measured algorithmic improvement has been established; do not cache mutable
permission facts or remove before/final callback guards speculatively.

## Ownership and preserved leaves

Use existing persistent worktrees and inspect their status before edits.
The primary branch is `codex/pr40-external-corpus-split-20261011`. One heavy
SDK/ASM/LLVM/object/native/benchmark owner runs at a time. Small audited pure
groups may overlap. Do not interrupt another owner's frozen snapshot.

The BSWAP test branch is `codex/pr40-bswap-definedness-20261015`; its production
leaf is integrated, later native tests are separate. The argument-carrier branch
`codex/pr40-amd64-go-argument-carriers-20261016` is frozen at `cc0e26f6` and
integrated. Fork staging is `codex/pr40-fork-native63-staging-20261015`;
published diagnostics use `codex/pr40-windows-owned-process-probe-20261013`.

Preserved `f5bc1eb2` fixes only Original386's inactive GOAMD64 observation:
the key must exist, active GO386 remains sse2 and all raw observations stay
unchanged. Seven roots/81 children pass Go27/race/Go20 and actual Go metadata;
actual SDK capture/CLI/object qualification is still pending. It is not merged.
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

Latest strict complete benchmark remains FAILED at `ca02d8ea`:

| Target | Complete files passed / total |
| --- | --- |
| linux/386 | 25/33 |
| linux/amd64 | 45/55 |
| linux/arm | 27/30 |
| linux/arm64 | 40/47 |
| js/wasm | 19/19 |
| Total | 156/184;28 failures, zero N/A/unattempted |

Target wall total312s meets the original300s/target and900s total budgets.
No newer benchmark pass follows from an individual Compare source test.
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
Leaf receipts remain in their respective worktrees. Preserve the failed
`a1a6e9dc` Compare printer-spelling assertion and the corrected `e56a7dda`
receipt separately. Only the actual tested head and scope are certified.
