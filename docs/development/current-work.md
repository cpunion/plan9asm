# Current work: PR 40 checkpoint

Checkpoint: 2026-10-16 00:15 UTC. The notes below this section are historical
and retain their original source scope; they are not the current task queue.

## Active queue

The repaired candidate `354ea4ef` has been pushed normally to the allowed fork.
[Run38824961157](https://github.com/cpunion/plan9asm/actions/runs/38824961157)
is still running/queued. Its independent native x87 job passes and the original
artifact independently replays as 4 roots/26 children, no skips or incomplete
terminals, with the exact GitHub SHA256 digest. Complete-module CI is not yet
successful. The run must
finish naturally before another repair push. Its non-GitHub tree is exactly
the qualified main source `d0388d27`; fork-only diagnostic policy is separate.

The last completed allowed-fork [run38760793656](https://github.com/cpunion/plan9asm/actions/runs/38760793656)
at `191fbeaf` ended naturally: 78 jobs, 66 successful and 12 failed. Ten
execution jobs contain 13 distinct failing test roots; the two other failures
are strict complete-module aggregates refusing those failed partitions.
No cancellation or rerun established success. The complete simple CLI and its
coverage upload pass; main/rich coverage uploads are skipped, not passed.
All 41 SDK priorities, native65 and ARM cross execution pass their own scopes.

Independent complete Windows raw audits retain main's 22 packages, 7,043 root
and 166,564 child executions, with 12 failed roots/45 failed children and 263
explicit host-inapplicable root skips. The simple CLI has 32 roots/29 children,
all passing without skips. The rich CLI has 334 roots/2,095 children, one
failed root and one explicit host-inapplicable child skip. Host skips do not
establish native execution on the skipped architecture.

| Failing class | Current bounded repair or remaining work |
| --- | --- |
| Complete Maps public CLI deadline | `ccf1a8f9` preserves the selected content-addressed Go build cache; fresh source/proof observations and complete two-file package remain independent. Its exact CLI passes locally in 58.23 seconds. Windows performance still needs CI. |
| CLC/STC/CMC numerical effects | Complete shared carry grammar and CF-only effects; the complement requires incoming CF. Existing unrelated FLAGS/DF remain unchanged. |
| MOV/FP helper metadata and wide complex store fixture | Actual result schemas, complete source witnesses and initialized GP input replace invalid metadata/unknown-entry assumptions; production checks are not loosened. |
| Native XADD/XCHG result continuation | Uses the existing complete typed sideeffect emitter; preserves faults/order and grants no reusable memory or declaration-query authority. |
| Partial FP parameter stores, BIT shadow writeback and trap-only source | Qualified `7753b575` is merged. Separate typed logical-shadow and all-path trap roles preserve the incoming-write, pointer/bool and complete-source boundaries. Actual Linux runtime still needs replay. |
| SB LEAW/L/Q address effects | Qualified `e5523905` is merged through `cf07263b`. Unknown-address writes kill destination identities without inventing numerical bytes or dereference permission. A separate old AUTO cost-selector gap remains unresolved. |
| x87 initialized words, ENV and exact signed integers | Merged `bda3e19c` passes 139 complete affected/retained owners on Go27, actual Go20 and race: 417 root executions/40,950 children, zero skips, complete ordered terminals and matching source/tool seals. Both whole-root-module vets pass. These are not whole-module tests, native execution or general FP80 arithmetic receipts. |
| DATA relocation address values | Qualified `1f39d8db` is merged through `62784911`: 54 affected owners on both SDKs and 30 source/model owners on race, 138 root executions/3,497 children, two whole-root-module vets and tracked formatting pass. Generic incoming pointers and code addresses gain no numerical or dereference permission. |
| Complete 386 runtime fixture | The original 18 TEXTs, manual signatures and C driver remain unchanged. The BX-only source retains its missing-scratch Context; the separately proved bounded counterpart now supplies the Windows runtime owner. All three original/corrected/bounded Go objects, eight LLVM22 axes and 28 single-mutation controls pass in the merged portable set. Actual Windows execution remains pending. |
| Strict complete standard-library benchmark | `00cfc843` finishes naturally with unchanged complete source/tool seals and all 184 original files attempted: 146 pass/38 fail, zero N/A. Total target time is491s; AMD64410s still exceeds the original300s target budget. Compared with `ab5ad450`, two files recover and six regress. These failures remain failures; focused owners do not replace this gate. |

The independent original-runtime boundary pair genuinely fails at `ab0395d3`:
BX-only scratch has no backing frame, and the new conditional Go oracle lacks
the selected SDK include path. `02ac0895` fixes the test driver and activates
only the bounded complete 386 counterpart; its two roots/26 children pass.
The full conditional counterpart retains the original C checks and returns
fourteen typed scalar functions instead of granting native stores from a pointer.
Three APIs/two IR levels execute all fourteen predicates with independent
NaN/zero/subnormal bit checks on the actual host in explicit Software mode.

Expanded integration also reproduces five stale byte-budget expectations after
the native string-counter state was added. The test-only repair counts that
one logical byte, distinguishes source-scan and reaching-state charges, and
checks the exact limit versus one byte short. No production limit is raised.
Complete old no-argument x87 memory fixtures retain original Go acceptance and
explicit public Context checks; separately labeled emitter units retain every
original instruction and LLVM object check, without claiming executable inputs.

All eight actual SDKs independently compare complete AMD64 reports before and
after the repair. Only eight FCMOV forms and FRNDINT change from supported
isolated probes to context-required; every other form, denominator and encoder
rule remains unchanged. `368991ae` updates exactly three AMD64 baseline fields
per SDK. The independent regression genuinely reproduces 60 false acceptances
on the old complete source; its 30 initialized positives still pass there.
At `d0388d27`, all 120 children pass separately on actual Go27, Go20 and race,
without skips and with complete before/after seals. Both actual SDKs then pass
the five-architecture official gate and all three selected scanner roots.
The existing Go20 malformed maskless EVEX.Z corpus row remains unsupported;
this is not a new regression or a full execution-coverage certificate.

Allowed-fork candidate `354ea4ef` retains the original 41 SDK priorities and
native65 job, and adds a separate native x87 owner. Its four required roots
include original-Go/LLVM amd64 and 386 I64 execution, conditional moves and
hardware status checks, with all 26 backend/architecture children mandatory.
All 102 Node policies and three workflow lints pass freshly at that exact
candidate. The new run above owns actual execution; the old completed failed
run remains failed. PR40 is still draft at its older head, pending fork CI
and the separate strict benchmark/standard-library work.

The benchmark's new opaque-pointer return regression is fixed locally through
`2e7f3821`, merged as `b3d900d9`. The first test-only run has a genuine private
signature setup failure; the corrected fixture then reproduces all15 failed
translation paths while all30 negative controls pass. The complete bounded
address/MOV/copy role keeps `x86ResultScalarBytes` unchanged and grants no NUM,
caller or pointee permission. All31 affected owners/2,723 children pass on each
of Go27, Go20 and race, without skips. Both merged whole-root-module vets pass.
The original complete Go27 checktest package then compiles both files on each
of386 and amd64: four original files, zero failures/N/A. This is not a fresh
184-file benchmark or native pointer-result receipt; the38 failed benchmark
files above remain the latest complete result. No second push interrupts CI.

The separate compact/expand leaf retains its current-source RED at `0d45f567`:
four roots/2,448 children expose216 missing register effects,1,620 rejected
initialized API paths and540 wrongly accepted missing-input/result paths.
`bbf06955` shares complete Go `_yvcompresspd`/`_yvexpandpd` grammar across all12
opcodes, the two emitters and numerical descriptors. Its first focused run is
in progress. Memory forms, old unbound compile-only fixtures, complete SDK
objects, compatibility, native execution and baseline classification still
need qualification before this leaf can be merged or published.

The FS/GS first-GREEN set at `951a68a0` passes 17 unique roots and 30,902
children, retaining all 212 AMD64 or 114 386 TEXTs per target. Its explicit
old-source Context and independent read-only counterpart close the old unbound
CX-store fixture without granting a native-store contract. The first expanded
qualification nevertheless fails at command 21 on actual Go20 encoding
observations; four later commands, including both vets, remain unattempted.
That failure is retained, not a qualified or native-execution receipt.

Actual selected Go20/386 drops FS/GS Index prefixes on dynamic-GP addresses;
current Go27 emits them. The old encoder also prints numeric register enums in
four genuine rejected-address diagnostics. Test-only `692c94a2` binds those
observations to the actual selected assembler/parser source and exact diagnostic
writer. Unknown schemas fail. Both SDKs independently pass its first three
roots/83 children; historical no-prefix objects do not prove equivalent current
segment semantics. The unchanged current-Go source-capability gates still
exclude historical objects from their private permission factories.
Fresh 120-owner qualification at `692c94a2` finishes naturally: all 25 commands,
322 root executions and 115,591 children pass, including both whole-module
vets. The complete Maps SDK owner passes in 224.657 seconds under its original
300-second budget. Root independently rereads ordered raw terminals and all
24,384 source/SDK/tool paths. Native execution and full-module tests are separate.

Two more complete producer schemas are genuinely missing on Go21/24 before
assembler-witness execution. Their independent failed three-root/zero-child
setup receipts remain failed. Test-only `e550b99d` adds the two actual selected
source tuples without changing translator permissions. Its three source/byte
owners pass on every actual supported minor SDK, Go1.20.14 through 1.27.2:
24 root executions/664 children, no skips and complete before/after seals.
`5e3bc51c` merges the complete string/segment leaf into the main development
branch; this combined source still needs its own current-head qualification.

The x87 input-obligation RED at `14e81e89` contains 1,170 real Void false
acceptances; the complete 3,420-child matrix also retains 540 initialized
positives. The narrow `78ffe059` routing repair removes all 1,170 false
acceptances without weakening the independent storage/result checks.

ENV test-only `cfb43daa` then independently reproduces further obligations:
four public roots/75 children fail in 69 children; the full-read root fails
all 12 children because only two nonvolatile words are read; whole raw-replay
identity fails its two authentic positive children while 28 mutation/query
controls pass. New Void controls expose the missing source/input route.
`b67f9c48` preserves one complete volatile 28-byte read, and
`a8713543` pairs Software word observations and raw FCLEX with their complete
source proof. Unknown Tag/TOP never becomes an initialized value or an empty
stack: the independent bounded counterpart has an actual final FINIT. Original
complete ENV source and its unknown-shape-at-RET control remain unchanged.
An independent two-mode logical-FP backing regression then reproduces two
false accepts: initialized scalar fields are not a contiguous native alloca.
`91dfc7af` restricts the new paired read to actual reaching native pointers;
it grants no FP/AUTO/DATA backing. Three additional independent pointer-reader
positives fail before the bounded argument/select/phi reader correction.
At `79b2d6cf`, all 16 roots/316 children pass: 60 LLVM22 objects over five
targets/three APIs/two IR levels/two backend levels; six actual-host Software
executions covering 393,216 status inputs with canaries; and four live/dead
guard-page executions checking exact 28-byte access. These are not physical
x87 or original-Go environment execution. After integration and explicit old
source-boundary/counterpart corrections, `abeeaf57` passes 27 roots/908 children.

Complete Runtime386 scratch TDD at `11f7ca69` independently passes all 18
source bodies, three complete Go objects and eight LLVM objects. Its 28
single-axis negative children nevertheless include 26 false acceptances; only
two reject, at an independent earlier AUTO boundary. Preserve the failed
four-root/59-child receipt. After composing the actual word/input proof,
`9247bef4` passes that complete four-root/59-child set, removing all 26 false
acceptances while retaining the legitimate complete 18-TEXT source. Objects are
not Windows execution; the bounded native counterpart is activated locally,
but its actual Windows execution still requires the next CI run.

Frozen `0c80d8a6` passes nine carry/exchange/fixture owners on each of Go27,
actual Go20 and race: 27 root executions/378 children, no failures or skips.
This is a focused result, not whole-module, native-runtime or final PR
qualification. Vets and broader affected-family qualification remain required.

The combined `1c883d32` snapshot separately passes the exact 37-root
DATA/BIT/LEA/carry/fixture integration set on Go27 and actual Go20:
74 root executions/5,238 children, no failures, skips, gaps or duplicates.
Both original five-minute commands retain identical source/SDK/tool seals.
This is not race, whole-module or native-runtime qualification.

The later combined `3a560149` separately passes whole-root-module `go vet ./...`
on current Go and actual Go20, under the original three-minute commands and
unchanged source/SDK/tool seals. These vets do not certify subsequent candidates.

Root independently rechecks the two newly merged leaves' original inventories,
raw ordered terminals, source/SDK/tool identities and vet exits. BIT/FP/trap's
Go27/Go20/race scopes pass 252 root executions/16,382 children. SB LEA's same
lanes pass 185 root executions/7,584 children. Both retain zero skips, missing
or duplicate terminals; neither is a whole-module or native-runtime receipt.
The combined head still requires its own fresh qualification before publishing.

DATA's independent final audit rereads 106 bound reports and 24,362 current
source/SDK/tool files. Its original failed snapshots, including the historical
Go local-symbol formatter mismatch, remain failed. The reader selects the
actual SDK lexer producer and exact observed assembler package, not a version
guess or arbitrary package suffix. Objects do not establish native execution.

The x87 follow-up still needs a real bounded scratch producer and an exact
extended-value representation. I64-to-double loses integers above 2^53.
Test-only `c0870164` independently passes all 96 original-Go format/byte cases
for the seven-name, nine-row integer/rounding family. Its complete I64 identity
source, without any control-word probe, reaches LLVM O2 through all three APIs
on five targets: 90 numerical leaves pass, 60 return incorrect constants and
15 return poison. Parent failures are not additional numerical cases. These
are genuine Software-mode numerical failures, not native-execution evidence.
The total-conversion follow-up `0df839a6` removes all 648 poison results in the
expanded five-direction boundary test. Its independent I64 identity diagnostic
still fails: all 165 numerical leaves execute, 90 pass and 75 return incorrect
constants, with no remaining poison. The subsequent `a0594c03` stores the exact
signed integer payload/tag in each of the eight runtime values, preserving
complete push/pop/copy/exchange/selection transport instead of an adjacent-pair
cache. All 165 identity leaves then pass. The extended whole-stack test also
exposes the missing conditional source-effect class; its independent complete
14-name form/input matrix has a retained 308-child RED before `431d3d7c`.
That repair uses one typed grammar for all conditional forms and requires both
reaching values and the complete actual predicate FLAGS. The `40db2550` batch
passes 12 roots/2,559 children, including 120 Hardware/Software LLVM22 objects
and six actual-host Software executions. Neither tagged integer/binary64
transport nor those tests establish general extended-precision arithmetic;
mixed fractional arithmetic and all remaining consumers still need review.
Software mode still requires an actual source control-word producer; its API
default is unchanged. Auto/Hardware must still observe actual hardware. Do not
turn the existing private alloca initializer into an implicit entry fact.

SCAS/CMPS retains the independent failed self-CMP and exact caller-write-footprint
matrices. The unchanged SDK IndexByte replay now reaches the existing unresolved
PC-relative branch guard, not an observed FLAGS failure; its expected next
Context is updated only after preserving the actual three-API failed trace.
That observation does not establish a complete SDK-file pass.

Main now includes the qualified commuted-pointer read leaf through `67c38bc2`
and immediate packed comparison leaf `605c8615`. The latter's Go27/Go20/race
affected tests pass 44 root executions/5,020 children. Its complete discovering
SDK package remains FAILED2/3: the next source effect is VPCOMPRESSQ. That is
not a complete SDK success. The strict original Index leaf independently passes
all three APIs and 12 LLVM22 object axes, without a native-execution claim.

The x87 logical-value leaf `0b541d6b` has only first focused GREEN so far:
14 roots/1,888 children, with Go-row/byte observations and bounded storage
controls. Go20, race, backend objects, complete original hypot and native
execution remain pending. Its original failed snapshots remain failed.

Earlier merged `ab5ad450` passes the exact 692-owner union: 1,690 root
executions/85,181 children across its selected SDK/race lanes. Its strict
benchmark is nevertheless FAILED150/184, with 34 failed files and AMD64
484 seconds above the unchanged 300-second budget. Neither that receipt nor
the subsequent focused leaves certify a newer combined snapshot. Do not
raise budgets or turn Context into a file pass.

Keep one local heavy command owner at a time, including inventories that build
test binaries. Release immediately after natural completion. Reuse existing
idle persistent worktrees, preserve immutable refs/failed reports, and freeze
each tested source. Stage only qualified batches in the allowed fork; do not
push a new diagnostic while its current run is active. PR40 remains draft at
`24851df2`; never directly push or merge upstream. No local inventory scanner
or external ledger workload belongs in this repository.

## Historical checkpoint: 13:13 UTC

Checkpoint: 2026-10-15 13:13 UTC. Fork diagnostic `5b44378d` completed naturally
with six failed jobs in three classes: CRC32 result/read proof, a string-IR
reader's CRLF handling and a shared Compare source deadline. Main contains the
qualified BP, string/LODS and exact-FP-byte/implicit-product leaves through
`60b269e2`. Their receipts do not certify the subsequent merged snapshot.

The actual combined run at `941f641a` FAILED on complete original CRC32 after
26 successful commands (1,188 root and37,258 child executions). Later commands
did not run. Preserve its failed raw output and frozen before/after seals.
Compare scheduling RED `a2f72363` and repair `bd8e6607` give each API an
independent original five-minute source lease. Focused frozen `4a8fda37`
qualification passes all seven commands/seven root executions/33 children,
including all three complete-source APIs/12 LLVM22 objects, retained Equal
Context, Go27/race/Go20 ownership checks and six SDK vets. Final source/tool
seals match. This does not certify the newer merged source.

Strict string-reader repair `3f1bdf2e` is merged through `3724e438`: the exact
original14 owners plus three reader roots pass17 roots/2,337 children, including
264 backend object children. The portable reader pair separately passes all
three SDK/race lanes with240 children each. Genuine REDs preserve both CRLF
false failures and96 extra-memory false accepts; neither is weakened into
success. Immutable tested Git blobs and current selected tools are audited.

CRC32/caller-projection/canonical-composition leaf `9adca3c4` is now merged.
Its exact new17 owners pass Go27 and actual Go20 with2,936 children per lane;
race passes15 owners/2,834 children. All49 root executions/8,706 children,
six SDK vets, formatting and frozen seals pass. Root independently rechecks
all raw terminals,52 bound report files and24,353 source/SDK/tool inputs.
This includes five original-Go objects and60 LLVM22 axes per SDK, not native
execution. Historical `de9ef3e3` independently passes the complete SDK CRC32
root plus all four failed CLI owners (five roots/34 children); that receipt
does not certify the newer9adc or merged head. Fresh merged replay is required.

The previous complete main qualification at `68ef8aa3` passes548 unique owners;
its strict benchmark remains FAILED158/184 with26 failures, zero N/A and zero
unattempted files. Classification, objects and executed semantics stay distinct.

## Current blocking scope

1. Preserve the complete BP repair: `903274da` passes 28 unique owners on each
   of Go27 and actual Go20, plus 16 on race and both SDK vets. Its 74 commands
   have 10,018 child executions, no failures or skips, and exact frozen seals.
   Complete 69-name memory objects and 32-member CMOV objects compile on both
   mode32 targets through three APIs. SP and amd64 BP remain physical-frame
   negatives. Main contains equivalent commits; this is not native execution.
2. Preserve qualified exact-byte/implicit-product leaf `bbe4fcc6`: all29 commands
   pass, with229 root executions and20,053 children, zero failures/skips/gaps.
   This includes34 semantic/control roots on Go27, race and actual Go20;
   57 retained roots and six tool owners on both SDKs; CLI extraction, both
   five-architecture official gates and three vets. Root independently audits
   every raw inventory/terminal and all3,882 tracked sources,20,452 SDK-source,
   tool and driver files, before/after/current identities and five stashes.
   Scalar MOVB/W/L/Q now transports exact partial/interior/cross-field bytes;
   original REDs retain534 wrong returned values and183 valid-view rejections.
   Implicit MUL/IMUL B/W/L/Q shares the complete Go grammar and consumes whole
   NUM/canonical proof even without result stores or through other storage
   routes. Independent120+120 false-pass REDs remain failed at their old heads.
   No DIV, raw/source-stack, caller or native-memory permission is added.
   The complete original19-TEXT fixture remains an explicit Context negative:
   its zero-frame physical SP, live memory POP and PUSHAL/POPAL lack backing.
   Its other18 unchanged TEXTs have explicitly partial object coverage; an
   independent bounded named-AUTO/MUL positive has real Go/LLVM objects,
   three-API value checks and five negative controls. Neither is a complete
   original-arith pass. Failed630 and build-failedb119 reports remain failed.
   Native implicit-MUL RED `74eeb615` now retains all28 kernels:84 IR0 native
   loads are nonvolatile and84 O2 loads disappear across three APIs. Its other
   three controls pass; the real Go test exits1, not success. This test-only
   leaf is not merged or repaired. Native fault execution remains separate.
   Objects do not certify native execution. Other arithmetic FP consumers and bound
   source-stack/fixed-local roles retain their independent contracts.
3. Preserve qualified LODS leaf `74e145d6`: 38 unique owners, 105 root executions
   and 9,697 children pass across Go27, actual Go20 and race, plus six vets.
   Genuine RED `183f7121` isolates exactly 12 AMD64 LODSL empty-repeat wrong
   values after observer mutation controls pass. Production `c9b03e61` selects
   the complete old AX on zero iterations. Each SDK/race lane compiles 24 LLVM
   objects from all 84 original functions. Only zero-state observations are
   checked; bare/nonzero execution and native execution remain unclaimed.
   Preserve failed earlier observer attempts separately from this genuine RED.
4. Preserve qualified9adc: full CRC grammar and mandatory NUM obligations,
   exact volatile reads, actual-caller low-mask projection, native-word INC/DEC
   and consumption of the existing mixed-bool canonical role. Genuine REDs
   retain public false passes, rejected valid sources and the failed6fc full
   SDK proof. Corrected Q-row tests distinguish arbitrary full-width input
   from a legitimately canonical zero-extended BSWAPQ result. No generic PTR
   bytes, native bounds, raw/source-stack or store permission is added.
5. Freeze the merged source, documentation and tools. Rebuild the exact union
   of all previous548 owners and every qualified BP/string/FP addition without
   dropping SDK lanes or enlarging original3m/5m budgets. Bind llvm-objdump as
   well as the other actual tools. Rerun fresh inventories, raw terminal audits,
   official gates, vets and the strict benchmark; no older receipt certifies it.
   The failed941 union contains669 unique owners over171 planned commands.
   Add the three new reader roots, Compare ownership root/two new API roots
   and all17 CRC/composition owners explicitly. The old Compare root owns
   only text; its other two complete APIs must retain independent scheduling.
   All five previously failed CRC SDK/CLI owners remain required in this run.
6. Stage the qualified batch in the allowed fork, without merging diagnostic
   workflows back into main. Policy `97cf2398` retains all37 prior early lanes
   and adds two independent five-minute Compare lanes (39 total); its93 Node
   policies and three workflow lints pass. Follow-up `ae7368f8` retains all39
   lanes and appends all three strict reader roots to the existing string
   lane without changing its five-minute budget. Fork-only `34138d3c` adds
   one17-owner CRC/composition lane (40 total) and the existing CRC runtime
   root as required native65:17 inventories/27 mandatory owners, retaining
   all17 required child paths and original budgets. Its97 policies/three
   lints pass; actual Linux execution remains pending. The current Darwin
   ARM64 host lacks Rosetta, so it cannot certify that runtime via a skip.
   Push once after merged qualification and let the new run finish naturally. Draft
   PR40 is not ready; its contribution head remains `24851df2`.

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

The latest [run38729263837](https://github.com/cpunion/plan9asm/actions/runs/38729263837)
at `5b44378d` ended naturally FAILURE:42 jobs,35 successful, six failed and one
skipped. Four failed priorities expose the same CRC32 missing result/read proof;
the other two are the string-IR reader and Compare deadline above. The complete
Windows reusable workflow was blocked by those prerequisites: none of its three
full modules, aggregates or coverage uploads executed. Do not count them as
passes or compare this42-job denominator to a fully expanded71-job run.

Complete current-head raw audits retain295 unique priority owners and14,320
children, six failed roots/56 failed children, no skips or terminal gaps in
executed SDK lanes. Native passes64 roots/208 children,16 inventories,26
mandatory owners and17 required children. ARM64 QEMU passes four required
children/512 inputs. Objects and executed adapter semantics do not establish
whole Go/ABI0 caller execution.

### Historical checkpoints

[Run38703136882](https://github.com/cpunion/plan9asm/actions/runs/38703136882)
at `e6a29969` ended naturally FAILURE:71 jobs,63 successful and eight failed.
Its nine failing BP fixture roots/385 children are repaired through `6925918c`.
The complete main inventory had22 packages/6976 named roots/155190 children,
263 explicit host-inapplicable root skips and no child skips. The simple CLI
had32 roots/29 children with no skips; the rich CLI had333 roots/2095 children
and one explicit ARM host child skip. These historical outcomes do not certify
the newer snapshot or the unexecuted complete modules in the latest run.

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

The qualified CRC leaf `codex/pr40-crc32-result-native-read-20261016` remains
frozen at9adc while primary source receives fresh merged qualification.
Test-only follow-up `codex/pr40-implicit-multiply-native-read-regression-20261016`
retains28 dead/zero-result native-read kernels and genuine failed74ee output;
it is not merged or repaired at this checkpoint. The string reader's qualified3f1b ref
and reports remain preserved. Fork-only scheduling is
the preserved native65 fork branch (resolve `34138d3c` with Git). Qualified BP, mixed-FP and
string-zero-count refs remain preserved. Resolve existing worktree directories
with Git; do not create disposable temporary clones.

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
