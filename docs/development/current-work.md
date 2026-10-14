# Current work: PR 40 checkpoint

Checkpoint: 2026-10-14 20:34 UTC. Primary source `a50042b5`.
This is a handoff, not a ledger or a claim of complete CI/stdlib success.

## Scope and authority

This repository owns translation, shared source/tool observations, official
instruction coverage and independent tests. External scanning, ledgers and
compatibility batches belong to [llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
Do not start a local index/module scan or restore its gates to translator CI.

Push only to an allowed `cpunion` or `dastai` fork. Never directly push, merge,
close, approve or change upstream state. Keep PR40 draft until its published
head passes required tests, review and coverage. LLVM22 only; Go1.27.2 and actual
Go1.20.14 root compatibility. Missing tools fail; no fallback or hidden skips.
Read [instruction rules](instructions.md) and [validation](validation.md).
Freeze source, tools and documentation during a tested snapshot. Preserve old
failures as failures; do not increase budgets, crop sources or shrink denominators.

## Latest completed CI

[38615686212](https://github.com/cpunion/plan9asm/actions/runs/38615686212)
at frozen fork `d4baabba64ce67051f616ee8a681324b65e6731d` completed FAILURE
naturally at19:26:26 UTC:58 jobs,44 success,14 failure. Failed aggregates are
not additional defects. All25 priority jobs pass, then complete Windows modules
run;18 individual test roots fail across the main module, rich CLI and shared
SDK frontend. Simple CLI complete-module aggregate passes. Do not relabel the
run or its failed shards as successful.

Strict downloaded receipts: source127/127 roots PASS; native58/58 really execute
and PASS132.944s, eleven inventories/twenty mandatory axes, zero skips,
missing/duplicate/unfinished terminals. These do not certify complete Windows
modules or future PCMP guard-page tests.

Earlier `265c5a07` run38613446840 completed FAILED:24 priority jobs pass, one
LF-specific original bit-scan sentinel fails, full modules do not start.
`03cbab33` fixes only a policy copy while retaining original translation bytes;
its nine roots/36 children pass Go27/race/Go20 plus both vets and source seals.
Earlier `732c8c15` run38605615421 also remains FAILED. Its three classes are
repaired: native policy CRLF handling, independent maps public-CLI owner and
four independent NUM API owners. They all really pass on Windows at `d4baabba`.

## Current repair batch

The contribution's published fork branch still has head `24851df2`; do not
confuse the diagnostic branch with PR40. Current upstream main is `e2164ac0`.
No source or scanner history is being force-pushed.

| Class / failed root | Current disposition |
| --- | --- |
| 386 unsigned32 NUM LEA: `TestX86ResultNumericalLEAThreeAPIFiveTargetOriginalAndLLVMObjects` | `183b5304` integrated as `15b104a9`. Shape-only pointer grammar defers to existing typed NUM LEA; reaching pointer rules stay unchanged. Both SDKs/race38 roots/657 children pass, including unchanged104/154 rows, five targets and13 object batches. |
| Old go-bits scalar source | `c2023e0f` initializes prior scan destinations, keeps an independent unknown-prior negative and checks real native bridges rather than poison-producing intrinsics. |
| CLD/STD operand diagnostic | `5598ab3f` retains the exact no-operand diagnostic; invalid forms stay rejected. Combined direction/ecosystem13 roots/446 children pass. |
| Windows continuation directory mutation | `a9848df1` avoids `t.Chdir` pinning the very directory being renamed. Actual Go1.27 testing source confirms the old retained directory handle. Three complete local CLI roots/nine children pass; actual Windows replay pending. |
| Shared SDK no-DATA deadline | `58d73c76`/`a50042b5`: six independent complete Count/IndexByte × Text/Module/GoModule owners, original five-minute budgets. All six pass35.563/35.744/46.073/40.093/36.295/52.225s with final source Verify; portable ownership policy passes Go27/race/Go20 and both vets. An initially mis-targeted mutation test stays FAILED in its separate receipt. |
| FP storage addresses | Frontend worktree has an unsealed denial-only typed FP-write shape shared by private observer/caller effects. No numerical/address/native permission. Preserve all156 original formats with a real initialized I32 result and retain old missing-result negatives. Five-target typed objects pass on Go27; final Go20/race/vet seal still needed. |
| Legacy PADD/PSUB incoming reads | Numeric worktree `2d38de0b` adds only paired conditional result bytes for complete17-name legacy grammar. Original m128 native instruction preserves alignment/dead faults. Full object tests independently expose missing386 SSE2 metadata; `a78997f8` shares the complete grammar with feature inference. Latest six-root qualification is pending. |
| FP PUSH/FLAGS family | Pure TDD in recovery/TLS tree; preserve all42/39 status writers. Model only bounded numerical private value stack. Existing uninitialized X/K fixture inputs must not acquire bytes from unknown state. |
| Legacy/VEX EXTRACTPS stores into independent FP inputs | Still failed; requires the actual complete logical FP write/read model, not permission to mutate native caller storage. |
| Borrowed-frame continuations | `TestX86ContinuationBorrowedFrame` and `TestX86ContinuationLLVM` still fail at table-address effects. Reuse/authenticate the actual whole continuation graph; never loan helper signatures or bypass the result proof. |
| Physical/custom bool domains | Raw-bound source-stack bool tests, explicit386 register helper and synthetic no-RET I1 terminator still fail. These are distinct from ordinary logical ABI0 FP; only their independently validated producer may hand off a complete result proof. |
| Other old source assumptions | Zero-frame SP-relative MXCSR fixture and shared tuple result-pointer helper remain failed. Check original frame/language/byte roles before changing either implementation or fixture. |

No current batch is ready to push until its relevant complete checks are sealed.
Fork-only `eff7e8c5` stages28 priority lanes (six new no-DATA owners), keeping
127 source priorities/native58, original budgets and no concurrency cap. Genuine
missing-lane RED precedes45 passing policy tests/actionlint. It is not pushed and
does not itself execute new test roots. Merge only verified source leaves before
the next batched fork run; let that frozen run finish naturally.

## Worktrees and ownership

Use persistent idle trees; inspect status/processes/remotes before changes.
Basenames intentionally avoid private machine paths.

| Worktree basename | Current role |
| --- | --- |
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | Primary `codex/pr40-external-corpus-split-20261011`, source `a50042b5`; integrated repairs, completed no-DATA validation. |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | Diagnostic `codex/pr40-windows-owned-process-probe-20261013`; published `d4baabba` failed, local `eff7e8c5` stages future lanes. Do not bring fork-only workflow back to contribution. |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | NUM LEA sealed183, legacy integer conditional read/feature leaves2d38/a789 under root qualification; namespace WIP separately stashed. |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | PCMP fault leaf485 sealed separately, FP storage-address class unsealed; preserve five owned dirty files. |
| `plan9asm-wt-pr40-sdk-tls-20261013` | Parameter-only admission WIP stashed, recovery agent owns PUSH/FLAGS; inspect before resuming after model capacity failures. |

One heavy SDK/object/native/benchmark owner at a time. Small pure gates may run
in parallel. A model/connection error is not a test failure; inspect any still
running owned command before edits or retries. Do not interrupt another owner's
frozen work, discard stashes or mix unverified leaves.

Important retained WIP: parameter admission exact stash
`1a59aba60e8852c56be0ea0dc1be905b7a6dc17a` and older9ac6f538; namespace exact
stash `7fb7ed50d95a2d9dbb9486c0da3ef1198ccfb94d`. Use immutable IDs, not a moving
stash index. Parameter-only reachable public false passes are real; new dead
escape/NOP negative expectations were incorrect and stay separate failed tests.
Keep existing sealed NOP and reachable-CFG semantics; no new source-wide regime
merely to satisfy an incorrect expectation.

## Remaining source/semantic boundaries

PCMP test-only leaf `485cb2fd` is NOT integrated: required Linux owner
`TestCrossLinuxRuntimeMatrixX86PackedStringOriginalNativeMemoryFaults`, all eight
original Go encodings, zero/full length × live/dead result, NULL/partial/full/
unaligned guarded reads and LLVM O0/O2. Source/object controls pass; actual
Linux fault execution remains pending. A Darwin host skip is not a runtime pass.
Physical MXCSR dead/fault parity remains separate and unresolved.

Namespace WIP only associates actual selected two-argument go:linkname bodyless
declarations with native TEXT using source/compiler/object identities. Complete
equal/compare SDK captures genuinely reject at the old same-package prefix check.
Do not replace it with a runtime.* exemption, fake helper declarations or a new
observer frontend. Hidden DX closure extent, composite ABI carriers and native
entry/caller memory still need their separate producers.

A read-only size audit confirms all865 migration-deleted scanner/ledger paths
remain absent. `cmd/plan9asmscan` deliberately remains the official/std classifier.
The contribution is still large (about595k added lines at this checkpoint);
independent Go/LLVM/C oracles are not accidental duplicate implementation.
After failures close, possible bounded simplifications: pure CLI target-triple
mapping, identical canonical-GP post-transfer intersection, and shared original
Go parsing/import setup. Preserve separate authority/lifetime contracts.

## Official coverage and benchmark

Fresh frozen `0fa63506` passes official five-architecture classification gates on
Go1.27 and actualGo1.20. Original supported/context/unsupported denominators and
fingerprints match; Go27 has zero unsupported/parse errors, historical Go20
retains one AMD64 unsupported form. Classification is not complete execution
or full standard-library file success.

Latest strict complete benchmark remains FAILED at `e88e77ce` (productionc1):
156/184 original files pass,28 fail, zero N/A/unattempted; all five targets run.
38625/33, AMD6445/55, ARM27/30, ARM6440/47, wasm19/19.
Target wall total289s passes original300s-per-target/900s-total limits.
Different caches do not establish an algorithmic speedup. Remaining contracts
include actual entry/TLS/g, reflect frames, preempt flags, VDSO/syscall frames
and ARM64 result-tail stores; do not restore unchecked admission.

## External inventory (server only)

Last trusted readback generated17:54:33 UTC:22,315,363 index entries,
1,073,867 unique module paths,6,811 assembly-bearing version records,
1,123,033 scanned records,16,889 retryable failures,1,004,339 latest records and
1,117,209 cgo scan records (not cgo-positive count). These are NOT translator
or runtime passes. All581 ranges checked contiguous and their sum equals index
entries. Derived history endpoint2025-10-01 and incremental endpoint
2026-10-11T13:22:43Z; unfinished windows cannot advance cursors.
Later progress requires fresh remote manifest readback. Resolve server details
from private operator configuration; never commit private hosts/paths.
Do not install heavy native-test toolchains on the inventory server implicitly.

## Evidence

Primary ignored `_out/register-result-repair/` retains
`ci-d4baabba-{source,native,main-part*,root-part*,llvm-part*}/`,
`nodata-independent-58d73/` (failed mutation control) and
`nodata-independent-v2/` (complete six-owner PASS),
`full-ci-gobits-directory/`, `official-0fa63506/`,
`bitscan-original-crlf/`, `parent-7e393-ci-partition/`.
Numeric tree: `_out/sdk-numeric-maps/{numerical-lea-unsigned,legacy-integer-native-read}/`.
Frontend tree: `_out/register-result-repair/packed-string-fault/` and FP receipts.
Raw inventories, every terminal, source seals, tools and frozen heads are
authoritative. See `git show 0fa63506:docs/development/current-work.md` for the
previous detailed history rather than appending old timelines here.
