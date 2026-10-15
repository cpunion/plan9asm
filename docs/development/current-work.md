# Current work: PR 40 checkpoint

Checkpoint: 2026-10-15 02:23 UTC. Qualified source observations `51bb6a5d`;
published fork diagnostic `fcff729f` is a separate, still-running snapshot.
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

## Active CI and latest completed run

[38667476177](https://github.com/cpunion/plan9asm/actions/runs/38667476177)
at fork `fcff729f` is still running. Let it finish naturally before another
push; do not cancel unrelated release jobs. All31 priority jobs pass. Simple
CLI completion now passes its fresh32-root inventory, all8 shards, complete
module/environment/coverage checks and coverage upload, with zero root/child
skips. This is actual Windows evidence for the PATH repair, not yet a complete
three-module CI pass. The other CLI and main module remain pending.

Downloaded native63 independently passes63 roots/169 children in149.992s,
all15 inventories/25 mandatory owners/eight children, zero failure, skip,
missing, duplicate or unfinished terminals. Original PCMP fault execution
covers all32 rows/34 probes at O0/O2. Retain downloaded evidence under ignored
`_out/register-result-repair/ci-fcff729f/`. A local watch TLS timeout is transport
failure, not a CI failure; reread actual job states before retrying the watch.

The diagnostic includes the qualified namespace and Windows repair batches
through contribution `2f0a4a4b`; it does not include later panic observations
`51bb6a5d`. Never use one snapshot's receipts to certify the other.

[38661041122](https://github.com/cpunion/plan9asm/actions/runs/38661041122)
at fork5657d2ab completed FAILURE naturally:64 jobs,60 success/four failure.
All31 priority jobs pass. Three module aggregates fail only on actual PATH;
root part9 has the independent shard25 deadline failure described below.
The remaining-package owner passes20 complete packages/282 named roots,
zero failures/skips. Actual raw roots show its time is dominated by
`internal/sdkfrontend` (32 roots/1517.004s), not the runner itself
(`cmd/plan9asmtest`,30 roots/108.552s). No extra failure class was found.
Do not call the complete run green or rewrite the old failed evidence.
Downloaded native63 independently passes63 roots/169 children in140.887s,
all15 inventories/25 mandatory owners/eight children, zero skips or gaps.
The new original PCMP fault owner really executes all32 rows/34 probes at
both O0/O2, including original VEX. Downloaded Windows source priority retains
the exact original127 roots/3947 children and passes131.348s; both former DF
failures pass with no reported skip. This artifact is Go verbose output, not
a JSON event trace. The Linux native artifact has strict JSON event audits.
Fork5657's non-GitHub source is identical to earlier contributionaa8287d5;
its completed run did not include the later namespace source.

The downloaded CLI producer/aggregate plans differ only in actual PATH:
hosted Git ucrt64 versus mingw64 and optional WSL entries. The repair explicitly
selects a minimal tool search path in both driver roles, retaining strict
scope equality. A synthetic command-transport RED reproduces this drift;
wrong/missing explicit LLVM and failed tool queries remain hard failures.
These synthetic controls are not actual Windows execution evidence.

Root part9's shard25 really expires at1800s. Its complete readonly DATA consumer
root passes all119 operations/three targets/three APIs/O0/O2 but takes1210.90s,
leaving insufficient time for later roots. Split that same matrix into three
independent complete target roots, with a real scheduling-policy RED and
disjoint32-shard/16-owner verification. All original12 batches/48 children and
72 module-object axes remain required. Frozenb66 qualification now passes all
three complete target roots/48 children:172.471/164.300/155.115s, every original
119-operation producer byte unchanged. Runner19 roots/22 children pass Go27,
race and actualGo20; both SDK vets, all3817 tracked inputs and selected tools,
HEAD and clean-status seals pass. Main's19 Node policies pass; the integrated
forkc95's63 policies/actionlint also pass. The active run above is the required
actual Windows replay; the old run remains FAILED. Evidence is under ignored
`_out/register-result-repair/windows-module-path/`.

[38648974765](https://github.com/cpunion/plan9asm/actions/runs/38648974765)
at frozen fork `7003d395` completed FAILURE naturally at00:08 UTC:32 jobs,
30 success, one failure and one skipped. Native62 and both Windows owned-process
SDKs pass. Source priority fails; complete Windows modules do NOT start this
cycle. Do not relabel that run or the skipped stage as successful.

The sole failed root is
`TestX86DirectionWholeSourceCFGJoinsAndFinalEffectRevocation`, both mode-specific
accepted-flags-restore children. Private-stack allocation incorrectly suppressed
DF revocation for an independently accepted POPF. Leaf98292316, integrated as
7f686ed4, clears old DF before transfer; only checked saved bits establish a new
fact. Six Go PUSHF/POPF rows, both modes, saved DF/status/ID and joins have real
independent REDs and complete regressions. Clean leaf qualification: pure18
roots/1170 children Go27/race/Go20; original exact127 source-priority roots/3947
children PASS63.173s; original42/39 status family five-target/three-API objects
on both SDKs. Vets and source/tool/head/status seals pass. These leaf receipts
do not certify the newer integrated source or complete Windows execution.

Downloaded native62 really executes62 roots/167 children PASS141.218s, fourteen
inventories/twenty-four mandatory owners and six source-stack/continuation ISA
children, zero skips, missing, duplicate or unfinished terminals. Both FP-field
owners and both source-stack/continuation modes execute. New PCMP fault
execution remains pending, not implied by this older native62 success.

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
Upstream `e2164ac0` is already an ancestor; no rebase is needed for that revision.

New integrated leaves, fully qualified independently before integration:

| Class | Integrated disposition |
| --- | --- |
| LOCK/CMPXCHG | `3b457517` as `432e39bd`: typed conditional-CAS transfer and source-bound same-block LOCK map; all67 AMD64/49 mode32 forms and six CAS names. Pure67/67/66 roots pass Go27/race/Go20; both SDKs three tool roots/1288 children, all299 Go LOCK oracles and objects pass. Conditional AX high halves and pair-CAS ZF-only effects stay exact. |
| Original386 source/generated FUNCDATA | `bca8c046` as `c99504d4`: complete42-TEXT source, actual compiler/archive and exact generated stackmap/owner/slot observations. No memory/native/terminal permission. SDK50+literal11+internal13 tool roots and portable37 roots pass; race/Go20/vets/seals pass. Former61-root timeout stays FAILED; disjoint complete groups retain five-minute budgets. |
| PCMP original native memory faults | `485cb2fd`/CRLF-only `d82ab956` as `5a74cb01`/`0d4afb5b`: eight names, zero/full length, live/dead outputs and NULL/partial/full/unaligned reads. Six truly pure policy roots/43 children pass Go27/race/Go20. Actual native O0/O2 remains pending. |
| POPF direction | `98292316` as `7f686ed4`, exact source127 and complete status-object replay as above. |

Commit and freeze the complete batch, then run fresh integrated-head affected
pure/tool groups, original source127 and official coverage before publishing.
At cleanaa8287d5 that coupled gate now passes375 unique current-SDK package/root
identities: root pure222/8060 children plus internal18/201 on Go27/race,
Go20 root221/8052 plus internal18/201, exact source127/3947, all affected
CAS/status/PCMP objects, SDK31+11 and internal13 tools. Both SDKs' official
classifications/continuation23, vets and all3809 tracked input/tool/head/status
seals pass. Complete benchmark still fails156/184; these are distinct claims.

Namespaceec21 is fully qualified on cleandd628058 (ec21 plusaa): pure16/96
pass Go27/race/actualGo20; exact22 SDK roots/71 children over two packages and
ten original five-minute commands pass, including all six independent NoDATA
owners, old denials, complete three-OS assembly and CRC/maps object axes.
Both vets and complete source/tool/head/status seals pass. opt/libLLVM's two
tool hashes explicitly use earlier aa benchmark tool-before evidence and own
post-run checks at the exact same installation; this is tool provenance only,
not reuse of aa source results. The qualified namespace merge is integrated
and included in the active fork run. Coupled `ca02d8ea` gates pass pure237
roots/8144 children on Go27/race, actualGo20 pure236/8136, source130/3951,
both official five-architecture gates, continuation23/811, vets and seals.

Panic observations are integrated as `c5337326`, `80a70903` and `51bb6a5d`.
Frozen clean `51bb6a5d` passes the deduplicated union: root pure252/8218 on
Go27/race and250/8196 on actualGo20, internal18/201 on all three lanes,
source130/3951 in98.107s and the exact two new tool roots/five children in4.106s.
Both SDK vets, formatting, complete tracked inputs, selected tools, HEAD and
clean-status seals pass. These are observation/source/object checks, not a
terminal capability or runtime result. The coupled receipts are under ignored
`_out/register-result-repair/combined-panic-51bb6a5d/`.

The following older repair rows retain their leaf-time dispositions; latest
actual native/Windows receipts above supersede their earlier pending labels.

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
| Borrowed-frame continuations | `acba4717`/`4d5e5e22` integrated as `408fa11d`/`696d63ab`. The actual coalesced whole CFG seals original/private instruction tuples, tables, declarations and final resolver callbacks. Separate opaque code/table identities never become NUM; a single shared budget and all-successor result-byte proof cover both modes. The 386 indirect memory jump reads its actual four-byte word. Same-root code joins retain only that root, not an invented exact helper. Both SDKs23 roots/811 children, pure17/428 and race/vets/seals pass; new five-target objects include15 Go/180 LLVM axes per SDK. Dedicated native execution remains pending. See [contract](x86-explicit-continuation-results.md). |
| Physical source-stack bool | `5e590dca`/`f0b82176` integrated as `8eb4218e`/`cf3dcf5d`. Original and normalized source, declared layout, access map, owner and numerical/canonical byte facts are sealed together; nonzero arguments and default Options preserve the same boundary. Both SDKs35 roots/1216 children, race8 roots/525 children, original CI roots,20 actual Go objects and84 LLVM object axes pass. Native execution remains pending, not implied by object success. |
| Explicit386 helper bool | `e3fd3479` integrated as `d129bdcc`. Actual helper remains Void with its ArgRegs; sealed whole original graph, helper transport, selected source/signatures and publication identity prove the root byte without copying its frame. Go20 pure55 roots/1090 children and Go27 race55/1090 pass; objects/current original root Go27 seven roots/58 children, Go20 six/53, both vets and source seals pass. The one extra physical-oracle root has an explicit go1.27 build role. Earlier58-root cohorts are tool-backed, not pure or isolated-performance evidence. No fake I1 callee or renderer inlining. |
| MXCSR source roles | `4580b59b`/`8e588fe2` integrated as `e01da058`/`0b9d36f9`. Exact old zero-frame/uninitialized mixed source retained as Context. All4 instructions use valid initialized AUTO/FP, two DATA loads retain actual readonly bytes; typed native pointer/static stores remain Context. Both SDKs/race pure14 roots/631 children; both SDKs five actual Go objects/60 LLVM objects, exact one root/65 children pass. No production permission or Void bypass. |
| Synthetic zero terminator | Test-only `a4bcecfb` integrated as `2b5ecc8e`: invoke the actual private zero emitter rather than pretend a NOP/no-RET source writes ABI0 results. Both SDKs16 pure roots/185 children and complete40 LLVM object types pass; original malformed I1/source negatives remain. No native execution claim or production bypass. |
| Shared tuple result pointer | `76552a1e`/`6ba3a2e6` integrated as `9580ca12`/`0f356ad8`: actual contiguous numerical backing requires whole private direct-tail source-byte proof. Original opaque pointer-bits source remains Context; transport uses a real numerical predicate. Both SDKs28 roots/1485 children, race21/1271, original CLI one/three, vets/seals pass. New four-width/five-target source owner compiles60 LLVM object axes per SDK; actual-host portable values execute all four IR/backend axes. Native x86/Go execution is not implied. |

Original failed-root replay at frozen408fa11d now passes21/21 portable roots
and493/493 children, with no skips, missing/duplicate/unfinished terminals or
changed source/tools. Eighteen old roots map to23 current roots:15 unchanged
portable owners, six independent no-DATA successors and two original host-runtime
owners. The last two are NOT executed locally. Neither portable success nor
object compilation establishes actual Linux or Windows execution.

Parent503f4076 passes156 pure roots/6223 children on Go27/race/Go20, plus the
two separately owned Go/LLVM tool roots/192 children on both SDKs. Both vets,
fresh inventories and source/tool/head/status seals pass. The earlier81-root
group included these two tool-backed owners: do not call that group pure or
isolated performance evidence. Final integrated696d63ab qualification adds the
17 disjoint continuation roots:175 total,173 pure plus the same two tool owners.
At clean46517f30 the complete175-root gate passes:173 pure roots/6651 children
on Go27/race/Go20, two tool roots/192 children on both SDKs, vets and seals.
The fresh official gate and continuation23 also pass both SDKs. These older
receipts do not certify the newer integrated batch; run its coupled gate next.

Fork-only582dcc55 stages28 SDK priority lanes,127 source priorities and native63,
without changing original budgets or adding a concurrency cap. Fifteen exact
native inventories require25 mandatory owners and eight dispatch children,
including the new original PCMP fault owner and its O0/O2 children. The strict
JSON auditor additionally rejects child terminals after their parent and final
package terminals without the exact package identity; both have real REDs.
All61 Node policies and actionlint pass on clean582 with source/tool seals.
This has NOT executed native63; older actual native62 success is separate.
Merge only verified main source into the diagnostic branch, never the reverse;
batch push the allowed fork and let that frozen CI run finish naturally.

## Worktrees and ownership

Use persistent idle trees; inspect status/processes/remotes before changes.
Basenames intentionally avoid private machine paths.

| Worktree basename | Current role |
| --- | --- |
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | Primary `codex/pr40-external-corpus-split-20261011`; source observations51bb and coupled gate qualified, not yet pushed. |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | Published diagnosticfcff; active CI has all priorities and simple CLI completion passing. |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | Clean qualified argument-planner8607 leaf, not integrated: calculation alone must not grant ABI permission. Old stash preserved. |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | Clean fully qualified POPF982 leaf; integrated. No heavy process. |
| `plan9asm-wt-pr40-sdk-tls-20261013` | Terminal-CFG WIP on a new branch at51bb; old qualified4c44 branch/receipts and stashes preserved. Sole heavy lease. |
| `plan9asm-wt-pr40-sdk-numeric-maps-20261013` | Clean forkfcff/native63 staging mirror; never merge fork workflow into contribution. |

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

PCMP test-only leaf `485cb2fd` and CRLF-onlyd82 are integrated: required Linux owner
`TestCrossLinuxRuntimeMatrixX86PackedStringOriginalNativeMemoryFaults`, all eight
original Go encodings, zero/full length × live/dead result, NULL/partial/full/
unaligned guarded reads and LLVM O0/O2. Source/object controls and actual
Linux fault execution pass at5657. A Darwin host skip is not a runtime pass.
Physical MXCSR dead/fault parity remains separate and unresolved.

Namespaceec21 associates actual selected two-argument go:linkname bodyless
declarations through the existing live source/SymABI/compiler-object observer
and complete TEXT archive. Actual unchanged equal/compare captures now pass;
four live mutation controls permanently revoke stale options. Committed-head
complete capture, old positives, six NoDATA owners and seals now pass ondd628.
This is declaration association, not complete file translation. No runtime.*
exemption, fake helper declaration or second observer is permitted. Go20 tests
the portable framework, not a fabricated SDK producer.

The separate386 single-panic terminal consumer is still WIP/RED. The integrated
observation now binds complete typed Go declarations, compiler/linker sources,
full compiler archive and every original CALL in the complete42-TEXT source.
Two actual tool roots/five children pass; standalone386 ELF compilation/linking
does not execute that ELF. Typed body observation alone is not terminal
authority. The next consumer must preserve original CALL/unwind, cut only its
proved zero-normal-successor edge, validate all original and normalized source,
and retain independent alternate entries after the CALL. Share the16k-visit/
64MiB budget and permanent/final-callback revocation. No source truncation,
fake ABI0 Go body, noreturn or nounwind attribute is admitted.
Hidden DX closure extent, composite ABI carriers and native entry/caller memory
still need separate producers.

The separate AMD64 argument planner8607 is qualified but not integrated. It
retains typed grouping/paths, Go's9GP/15FP banks and whole-argument rollback;
it supplies no source, initialized-byte, spill, DX or caller permission. Actual
compiler ABI wrappers are not guaranteed for every declaration. Inventory the
real full compiler archive before designing a consumer; do not manufacture
SDK references or wrappers. Shared ABI readers need their own observed source
roles, full lifetime checks and a coherent downstream consumer before merging.

A read-only size audit confirms all865 migration-deleted scanner/ledger paths
remain absent. `cmd/plan9asmscan` deliberately remains the official/std classifier.
The contribution is still large (about595k added lines at this checkpoint);
independent Go/LLVM/C oracles are not accidental duplicate implementation.
After failures close, possible bounded simplifications: pure CLI target-triple
mapping, identical canonical-GP post-transfer intersection, and shared original
Go parsing/import setup. Preserve separate authority/lifetime contracts.

## Official coverage and benchmark

Fresh frozen `ca02d8ea` passes official five-architecture classification gates on
Go1.27 and actualGo1.20. Original supported/context/unsupported denominators and
fingerprints match; Go27 has zero unsupported/parse errors, historical Go20
retains one AMD64 unsupported form. Classification is not complete execution
or full standard-library file success. Go27 observes only21 distinct386 corpus
operations and classifies all121 wasm forms as Context: no claim that every
encoder name or runtime behavior is implemented follows from a green gate.

Latest strict complete benchmark remains FAILED at `ca02d8ea`:
156/184 original files pass,28 fail, zero N/A/unattempted; all five targets run.
38625/33, AMD6445/55, ARM27/30, ARM6440/47, wasm19/19.
Target wall total312s (AMD64254s) passes original300s-per-target/900s-total limits.
New386 atomic source/LOCK observations only advance to an independent CALL
contract; no additional complete-file pass has been established.
Different caches do not establish an algorithmic speedup. Remaining contracts
include actual entry/TLS/g, reflect frames, preempt flags, VDSO/syscall frames
and ARM64 result-tail stores; do not restore unchecked admission.

## External inventory (server only)

Last trusted readback generated2026-10-15T00:21:33Z:22,775,671 index entries,
1,080,162 unique module paths,6,877 assembly-bearing version records,
1,134,359 scanned records,16,946 retryable failures,1,010,978 latest records and
1,128,535 cgo scan records (not cgo-positive count). These are NOT translator
or runtime passes. All616 ranges checked contiguous and their sum equals index
entries. Derived history endpoint2025-10-01 and incremental endpoint
2026-10-13T00:22:43Z; unfinished windows cannot advance cursors. Service active;
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
