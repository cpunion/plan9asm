# Current work: PR 40 checkpoint

Checkpoint: 2026-10-14 21:43 UTC. Primary source `cf3dcf5d`.
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
| FP storage addresses | `3febe438` integrated as `8cfe6ed3`: denial-only mapped address effect, all156 original formats with a real initialized I32 result and original missing-result negatives retained. Go27/race/Go20 pure50 roots/2370 children; both SDKs two object roots/221 children, five Go targets/60 LLVM objects, vets/seals pass. No numerical/address/native permission. |
| Legacy PADD/PSUB incoming reads | `2d38de0b`/`a78997f8` integrated as `2f0f6d8b`/`77c2bcae`. Complete17-name grammar supplies only paired conditional result bytes and complete386 SSE2 inference. Original m128 alignment/dead faults remain. Go27/Go20/race11 roots/563 children, both vets, source seals and actual discovering CLI root/two children pass. |
| FP PUSH/FLAGS family | `9f7e61e1` integrated as `f6ef6030`: bounded separate private numerical bit stack, balanced depth, exact GP/FLAGS/DF/ID effects. All42/39 original status writers pass five targets/three APIs; initialize only each row's actual X/K inputs before POPF, not unrelated native rows. Go27/Go20 complete9 roots/636 children and pure6/race/vets pass. No new native-status effect or increased budget. |
| Legacy/VEX EXTRACTPS stores into independent FP inputs | `5984054e`/`61f800f7` join the existing logical field writer to the private whole-source byte proof. All immediate rows, missing-byte/pointer/bool/padding negatives; Go27/race/Go20 portable16 roots/940 children pass, five actual Go targets and three-API IR optimization objects. Full18-root attempt remains FAILED (old uninitialized span fixture plus two host skips); initialized span fixture retains all rows. Two actual runtime roots remain host-inapplicable locally: Rosetta is unavailable. Require native CI; portable passes are not execution. |
| Borrowed-frame continuations | `TestX86ContinuationBorrowedFrame` and `TestX86ContinuationLLVM` still fail at table-address effects. Reuse/authenticate the actual whole continuation graph; never loan helper signatures or bypass the result proof. |
| Physical source-stack bool | `5e590dca`/`f0b82176` integrated as `8eb4218e`/`cf3dcf5d`. Original and normalized source, declared layout, access map, owner and numerical/canonical byte facts are sealed together; nonzero arguments and default Options preserve the same boundary. Both SDKs35 roots/1216 children, race8 roots/525 children, original CI roots,20 actual Go objects and84 LLVM object axes pass. Native execution remains pending, not implied by object success. |
| Explicit386 helper bool | Independent helper call ABI remains Void with all actual ArgRegs. The complete16-SETcc pure cohort and original helper/object roots pass initially; source/helper publication identity guard and final frozen qualification remain in progress. No fake I1 helper declaration or renderer inlining. |
| MXCSR source roles | `4580b59b`/`8e588fe2` integrated as `e01da058`/`0b9d36f9`. Exact old zero-frame/uninitialized mixed source retained as Context. All4 instructions use valid initialized AUTO/FP, two DATA loads retain actual readonly bytes; typed native pointer/static stores remain Context. Both SDKs/race pure14 roots/631 children; both SDKs five actual Go objects/60 LLVM objects, exact one root/65 children pass. No production permission or Void bypass. |
| Synthetic zero terminator | Test-only `a4bcecfb` integrated as `2b5ecc8e`: invoke the actual private zero emitter rather than pretend a NOP/no-RET source writes ABI0 results. Both SDKs16 pure roots/185 children and complete40 LLVM object types pass; original malformed I1/source negatives remain. No native execution claim or production bypass. |
| Shared tuple result pointer | Local `76552a1e`/`6ba3a2e6` supplies one actually contiguous, bounded numerical backing only after the whole private direct-tail source-byte proof. Original opaque pointer-bits source remains Context; the transport fixture derives an actual numerical predicate. Initial three-API, object, independent-root/KeepFunc and actual-host portable value tests pass. Final frozen Go27/Go20/race/object qualification is running; not integrated or certified yet. |

Fourteen of the18 old failing roots have repaired, leaf-qualified source paths;
this is not a current-head CI pass. Four old roots remain unsealed: two
continuation, custom bool helper and tuple helper. Parent qualification is still
required: the union at8decf884 passes Go27/race/Go20 with81 roots/4916 children,
but predates the synthetic-zero and physical-bool leaves. No current batch is
ready to push until its relevant complete checks are sealed.
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
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | Primary `codex/pr40-external-corpus-split-20261011`, source `cf3dcf5d`; integrated qualified repair leaves; updated parent union pending. |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | Diagnostic `codex/pr40-windows-owned-process-probe-20261013`; published `d4baabba` failed, local `eff7e8c5` stages future lanes. Do not bring fork-only workflow back to contribution. |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | Tuple-region leaf6ba3a2e6 frozen qualification; namespace WIP separately stashed. |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | FP, MXCSR and synthetic-zero leaves integrated; read-only tuple review. PCMP fault leaf485 remains separate. |
| `plan9asm-wt-pr40-sdk-tls-20261013` | Parameter WIP stashed; stack leaf9f7 qualified. New local-helper bool branch preserves all old history. |
| `plan9asm-wt-pr40-sdk-numeric-maps-20261013` | Physical-bool leaves fully qualified and integrated; new continuation repair branch preserves old history. |

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

Fresh frozen `8eb4218e` passes official five-architecture classification gates on
Go1.27 and actualGo1.20. Original supported/context/unsupported denominators and
fingerprints match; Go27 has zero unsupported/parse errors, historical Go20
retains one AMD64 unsupported form. Classification is not complete execution
or full standard-library file success. Go27 observes only21 distinct386 corpus
operations and classifies all121 wasm forms as Context: no claim that every
encoder name or runtime behavior is implemented follows from a green gate.

Latest strict complete benchmark remains FAILED at `e88e77ce` (productionc1):
156/184 original files pass,28 fail, zero N/A/unattempted; all five targets run.
38625/33, AMD6445/55, ARM27/30, ARM6440/47, wasm19/19.
Target wall total289s passes original300s-per-target/900s-total limits.
Different caches do not establish an algorithmic speedup. Remaining contracts
include actual entry/TLS/g, reflect frames, preempt flags, VDSO/syscall frames
and ARM64 result-tail stores; do not restore unchecked admission.

## External inventory (server only)

Last trusted readback generated21:35:55 UTC:22,622,410 index entries,
1,077,773 unique module paths,6,847 assembly-bearing version records,
1,130,822 scanned records,16,917 retryable failures,1,008,425 latest records and
1,124,998 cgo scan records (not cgo-positive count). These are NOT translator
or runtime passes. All602 ranges checked contiguous and their sum equals index
entries. Derived history endpoint2025-10-01 and incremental endpoint
2026-10-12T10:22:43Z; unfinished windows cannot advance cursors. Service active;
no local scan process. A missing remote jq was an inspection-tool error, not a
scan failure; a bounded manifest-only JSON check independently verified ranges.
Later progress requires fresh remote manifest readback. Resolve server details
from private operator configuration; never commit private hosts/paths.
Do not install heavy native-test toolchains on the inventory server implicitly.

## Evidence

Primary ignored `_out/register-result-repair/` retains
`ci-d4baabba-{source,native,main-part*,root-part*,llvm-part*}/`,
`nodata-independent-58d73/` (failed mutation control) and
`nodata-independent-v2/` (complete six-owner PASS),
`ps-logical-store/{red,green-first,green-v2,final,final-v2}/`,
`full-ci-gobits-directory/`, `official-0fa63506/`,
`bitscan-original-crlf/`, `parent-7e393-ci-partition/`.
Numeric tree: `_out/sdk-numeric-maps/{numerical-lea-unsigned,legacy-integer-native-read}/`.
Frontend tree: `_out/register-result-repair/packed-string-fault/` and FP receipts.
Raw inventories, every terminal, source seals, tools and frozen heads are
authoritative. See `git show 0fa63506:docs/development/current-work.md` for the
previous detailed history rather than appending old timelines here.
