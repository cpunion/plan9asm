# PR 40: current checkpoint

Updated 2026-10-18. Keep this document a current handoff, not an append-only
history. Earlier checkpoints and original failed fixtures remain in Git.
Read [instruction contracts](instructions.md) and [validation](validation.md)
before changing a family. Preserve existing WIP, frozen branches and failed
reports; do not discard them while simplifying this checkpoint.

## Scope and publication

PR40 owns instruction translation, complete typed families and independent
tests for 386, amd64, arm, arm64 and wasm. External inventories, scanners,
ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). No inventory
scanner runs locally. Do not expand that workload in this translator PR.

Push only to the allowed `cpunion` or `dastai` fork. PR40 remains draft at
`24851df2`; verified local or backup branches do not qualify that contribution
head. The last frozen CI has finished; repair its actual failures before another
diagnostic batch, without a direct upstream push or merge. Latest upstream main
`e2164ac0` is an ancestor
of the current candidate. Do not create another worktree for routine work.

## Frozen snapshots and evidence

| Snapshot | Evidence and limits |
|---|---|
| `2d9059765` | Complete GP/SIMD native-store inputs and continuation: 42 roots / 2,627 children pass current Go, actual Go1.20 and race; identical 2,669 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Both-SDK build/vet and nineteen scanner roots, five-arch classification, four CLI extraction roots, 132 CI policies and tracked formatting pass. Required native execution and complete CI remain separate |
| Same code head | The eight-SDK audit changes only six isolated AMD64 forms per SDK supported-to-Context for missing inputs; 386 classifications, unsupported counts, encoder rules and unrelated metrics stay unchanged. Original RED retains 44 failed terminals, including the original overwritten-store program; the portable byte oracle already passed before this repair |
| `cdbc92b85` | Complete multiply-add input effects and MMX format contracts: 69 roots / 772 children pass current Go, actual Go1.20 and race; identical 841 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Both-SDK build/vet and eighteen scanner roots, five-arch classification, three CLI extraction roots, 132 CI policies and tracked formatting pass. Native Linux execution and complete CI remain separate requirements |
| Same code head | The sealed eight-SDK audit changes only 27 isolated AMD64 multiply-add forms per SDK from supported to Context for missing inputs; 386 classifications, encoder rules, unsupported counts and unrelated metrics stay unchanged. Original TDD at `87ae8fb66` retains 126 failed terminals: both missing-input acceptance and initialized-vector rejection |
| `3d26bf2cc` | Complete scalar-shift memory effects: 23 roots / 5,793 children pass current Go, actual Go1.20 and race; identical 5,816 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Both-SDK build/vet and seventeen scanner roots, five-arch classification, three CLI extraction roots, 132 CI policies and tracked formatting pass. Complete native Linux execution remains required |
| Same code head | Two earlier race attempts remain failed/incomplete: one timed out including compilation, one timed out during inventory. The fresh complete retry passed within unchanged budgets. A separate Linux AMD64 original-Go zero-count diagnostic passed 84 cases (76 RMW read-only faults, eight BMI2 reads); it is not complete LLVM native qualification |
| `e95a10314` | Complete CPU-query implicit effects: 40 roots / 1,593 children pass current Go, actual Go1.20 and race; identical 1,633 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Original named/raw runtime sources compile through all three APIs on all five x86 targets. Native Linux AMD64/386 execution remains a separate required counterpart, not a local pass |
| Same code head | Both-SDK build/vet and sixteen scanner roots, five-arch official classification, three CLI extraction roots, CI policies and tracked formatting pass. Backed up in the allowed fork; neither complete CI nor the strict benchmark has qualified this source |
| `dab3bb581` | Preserved CPU-query TDD RED: four roots / 136 children, three failed roots and 89 failed terminals with unchanged seals. The original IA32 program fails at CPUID, the implicit NUM descriptor is absent, and Void accepts missing inputs. The simple initialized-result object root already passes; it is a positive control, not a repaired failure |
| `5df8cde37` | Eleven original failed format/witness roots plus nine initialized public-object roots pass current Go, actual Go1.20 and race: 20 roots / 2,362 children, identical 2,382 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Production, CI and classification baselines are unchanged; this repairs test contracts, not eleven new instruction implementations |
| `83f21268e` | The latest CI's exact 167 failed roots were fully replayed locally: 10 pass, 68 fail, 89 explicit host skips, zero absent/pending/incomplete. Original output hashes, complete run/terminal inventories and both owner source/SDK/tool/environment seals audit unchanged. This failed local replay does not certify unexecuted native counterparts or later repairs |
| `252239cde` | Complete mask/vector NUM effects: 30 roots / 734 children pass current Go, actual Go1.20 and race; all 764 terminal identities match with no fail/skip/incomplete and unchanged source/SDK/LLVM22 seals. Both-SDK build/vet and fifteen scanner roots, five-arch classification, three CLI extraction roots, CI policies and tracked formatting pass. The original native x86/MXCSR counterpart remains required, not a local pass |
| `4bcd33be9` | Packed private-parameter stores: 48 roots / 2,031 children pass current Go, actual Go1.20 and race; all 2,079 terminal identities match, no fail/skip/incomplete, unchanged source/SDK/LLVM22 observations. Build/vet, twelve scanner roots on both SDKs, five-arch classification, three CLI extraction roots, 132 CI policies and tracked formatting also pass |
| `e7e90b7c9` | Ten historical unbound-format failures plus eight initialized/source-boundary owners pass current Go, actual Go1.20 and race: 18 roots / 605 children, identical 623 terminals, no fail/skip/incomplete. Original source rows remain; public Context and emitter-only objects are not runtime passes |
| `8e82392cd` | Floating logical-field addresses: 57 roots / 3,871 children pass all three lanes, identical 3,928 terminals, no fail/skip/incomplete and unchanged source/SDK/LLVM22 observations. Build/vet, twelve scanner roots on both SDKs, five-arch classification, three CLI extraction roots, 132 CI policies and tracked formatting pass; backed up in the allowed fork |
| `d5bdb0399` | All 179 historical failed roots re-observed locally: 29 pass, 58 fail, 92 explicit host skips, zero absent/pending/incomplete. Original output hashes, all run/terminal identities and both source/SDK/tool/environment seals audit unchanged; selected aggregate remains failed, not a full/native CI pass |
| `0cb14d2fa` | Integer self-comparison and mandatory Void repair: 54 roots / 7,277 children pass current Go, actual Go1.20 and race; all 7,331 terminals match, zero fail/skip/incomplete, unchanged source/SDK/LLVM22 observations |
| Same code head | Build/vet, twelve scanner roots on both SDKs, five-arch official classification, three CLI extraction roots, 132 CI policies and tracked formatting pass; staged in allowed fork branch `codex/pr40-integer-self-compare-closing-20261018`, not the contribution head |
| `bc79fd0f` | Combined named/raw NOP repair: 64 roots / 8,606 children pass current Go, actual Go 1.20 and race; all 8,670 terminal identities match, zero fail/skip/incomplete, unchanged source/SDK/LLVM 22 observations |
| Same code head | Three additional current-Go SDK producer roots / eight children pass normal and race; they are not part of the Go1.20 denominator |
| Same code head | Build, vet, eleven scanner roots on both SDKs, five-architecture official classification, three CLI extraction roots, 132 CI policies and tracked formatting pass |
| `aec6c45a` | BZHI/BEXTR/BLS affected union: 26 roots / 620 children pass all three lanes; eight-SDK audit changes only four isolated AMD64 BZHI forms per SDK supported-to-context, no unsupported drift |
| `92acfbb2` | Paired native-input repair: 44 roots / 3,522 children pass all three lanes; the original latest failed priority selection is retained |
| `06eedf847` | Latest diagnostic [run 39065963275](https://github.com/cpunion/plan9asm/actions/runs/39065963275) finished naturally: 81 jobs pass and 17 fail. All 16 Windows root parts and their completion fail |
| Same published run | Audited actual native execution: 73 roots / 228 children pass, zero fail/skip/duplicate/incomplete; these receipts do not certify later BZHI/NOP repairs |
| Same latest run | Complete raw root union: 16 parts / 64 shards, 6,998 roots: 6,568 pass, 167 fail, 263 explicit host skips, zero incomplete. Every part shares the actual head/source and exact inventory; disjoint starts/terminals match all 6,998 names. This is a failed raw union, not canonical all-platform qualification |
| `8ca976c6` | Previous full denominator: [run 39054965721](https://github.com/cpunion/plan9asm/actions/runs/39054965721), 6,986 roots: 6,544 pass, 179 fail, 263 explicit host skips, zero incomplete |

Do not subtract newer focused successes from that completed CI denominator.
Canonical complete modules, all-platform execution, the strict benchmark,
coverage and fresh review must qualify the final contribution head.
Classification, emitter-only objects, initialized public translation, portable
execution and native execution remain different claims.

## Latest repairs

The six named GP/SIMD stores reuse the complete transfer grammar and require
their exact four/eight input bytes and incoming pointer address even for Void.
Paired native continuation preserves numerical/storage facts; it grants no
future memory read, caller, physical MMX or raw-entry permission. Existing
unaligned volatile emission keeps overwritten stores observable. Original
read-then-store source is retained as a separate initialized five-target
positive; uninitialized format rows retain Go/public-Context/emitter objects.
The complete 42-root qualification above includes portable byte/canary execution,
not a claim of native execution or a reduction of the old online CI count.

PMADDWL, PMADDUBSW, VPMADDWD and VPMADDUBSW now share their existing complete
grammar with mandatory numerical inputs, result definedness and paired native
reads. The production repair adds three lines and removes a duplicated two-line
fallback; it changes no arithmetic emitter or memory permission. Independent
register/memory, X/Y/Z, mask/merge/zero and mode32 controls retain all 56 input
shapes. An unused bool or Void return cannot supply absent inputs. Original
unbound shuffle/multiply-add sources remain actual-Go/public-Context/emitter-only
format controls; independently initialized programs retain three public APIs,
five-target objects and both IR/backend optimization axes. Closed MMX state,
raw encoding and actual caller boundaries retain their independent owners.

The scalar-shift memory batch retains all 148 AMD64 and 110 IA32 independent
memory witnesses, including explicit Context for the four unbounded W double
shifts per mode. Group2/double shifts preserve exact volatile read/write effects;
BMI2 sources preserve exact volatile reads, including mode32 Q's dword width.
Complete addresses, numerical count/data and carry remain mandatory even for
Void. The original `doubleShift32/64` conformance bodies now produce objects
without source rewriting; portable execution checks all byte counts, FLAGS,
unaligned memory and complete canaries through all three APIs and both IR/backend
optimization axes. The full Linux native runtime owner is added but not locally
qualified. Original RED at `e57ab2805` remains unchanged.

The sealed eight-SDK audit changes only two 386 and 77 AMD64 isolated memory
forms per SDK from supported to Context because their address/inputs are absent.
Encoder rules, unsupported counts and every unrelated metric remain unchanged.
The original MMX format RED at `3d26bf2cc` remains failed. Its initial contract
repair exposed the separate mode32 multiply-add input bypass, subsequently
reproduced by the independent full-family tests above. No original source was
silently given EMMS or replaced by an initialized program.

CPUID and XGETBV now share their existing complete zero-operand grammar and
implicit dword input/output spec with numerical definedness. Inputs are checked
before any overwrite, output writes revoke old pointer/address identities, and
unrelated GP, SIMD, mask, FLAGS, DF and private-stack facts remain unchanged.
Exact named/raw identity checks allow the paired native continuation without
granting a caller region, native entry, source stack or fixed-local role.
Missing/partial inputs, pointer-as-data, unused bool parameters and incomplete
CFG joins remain Context even for Void functions. Original bare format sources
remain actual-Go/public-Context controls; initialized counterparts retain real
native opcodes in LLVM objects. The complete original runtime sources, including
FLAGS and MXCSR observations, have all-five-target object coverage, not local
native execution. The forty-root qualification is recorded above.

The sealed eight-SDK audit changes exactly the isolated AMD64 CPUID/XGETBV
forms from supported to Context per SDK; all other metrics and encoder rules
remain unchanged, including unsupported counts. The 386 official corpus has
no classification change; independent named/raw regressions still cover both
modes. No benchmark or completed online CI failure count is reclassified.

At `5df8cde37`, eleven failures from the complete 167-root replay retain their
original assembly and actual Go objects. Unbound named format fixtures now
require public Context before independent emitter objects; masked subtraction
also retains its conditional volatile reads. MOVNTO checks both real native
stores rather than obsolete generic LLVM stores. The complete official binary
table witnesses must still reach the actual dispatcher even when the separate
probe correctly needs inputs. That syntax observation alone claims no LLVM
object or execution. Nine unchanged initialized public-object roots retain the
compare, logical, CLMUL, ternary and non-temporal positive counterparts.
The batch removes a net 49 test lines, changes no production or CI gate, and
does not turn any actual whole-program failure into a skip.

The eight mask/vector B/W/D/Q moves now share their existing complete EVEX
grammar with numerical result effects. Reverse moves need only each lane's
sign byte; forward moves require actual K0-K7 data and clear wider aliases.
FLAGS, DF, GP state, unrelated vectors and masks remain unchanged. Mandatory
Void checking closes the falsely accepted missing-input case reproduced at
`521ef9c48`; the same RED also rejects legitimate initialized programs.
The original unbound Go sources remain as Go/public-Context/emitter-object
controls. Separate initialized 48-program batches retain all three public APIs,
five x86 targets and both IR/backend optimization axes. Actual-host execution
uses the discovering fixture's unchanged complete source and observes every
vector byte, both masks, FLAGS and input/output canaries. It does not replace
the native x86/MXCSR or cross-runtime owner. The first attempted local cohort
retains that explicit host skip as failed qualification; the disjoint 30-root
host-applicable union is qualified at `252239cde`.

The sealed Go1.20-Go1.27 audit changes only 24 isolated AMD64 mask/vector forms
per SDK from supported to Context: their input data is absent. All other
metrics and encoder rules remain unchanged, including unsupported counts.
Initialized complete programs still translate and execute. A classification
correction is not lost instruction support or a whole-suite runtime pass.

Logical numeric field addresses now include float32/float64, without widening
integer-only return or native-frame roles. ABI0 allocation padding does not
extend a field or create initialized bytes. The independent copy tests retain
five-target objects, all three APIs, 80 negative boundaries and exact-bit/private-
copy execution. Original RED includes a falsely accepted uninitialized floating
result; that hole is closed at `8e82392cd`. The original SqrtByAddress program
still needs its separate complete SQRT arithmetic-effect repair.

The Windows diagnostic's complete root owners have real failures, not queue or
installation failures. Its exact 167-root queue is the next CI authority;
the older local 179-root replay does not supersede it. The private-parameter
VMOVDQU store defect from part3 is now
independently reproduced across 41 complete unaligned store shapes. A bare
unknown-input Void store additionally reproduces false acceptance on all five
targets at `867cf9f17`. `d85723a5b` shares existing move effects for numeric
logical stores and routes every valid packed FP store through whole-source
checking. Pointers, bools, holes, native alignment and SP keep their boundaries.
The discovering program's complete source is byte-identical after extraction.
Its three-API actual-host execution passes 257 input tuples and full canaries
under both IR and backend optimization axes. All 82 original Void/integer-result
functions per target remain in the objects. The complete 48-root affected union
and final gates pass at `4bcd33be9`; this portable evidence does not certify a
new native Windows run or change the completed online 167 failures.

At `e7e90b7c9`, ten old format tests retain their complete original Go objects
and explicitly reject missing inputs through all three public APIs before
checking emitter-only objects. The NOTB oracle now requires exact volatile
byte loads/stores and an LLVM object. The net test-code change is smaller.
The original attempted 19-root qualification retains its explicit amd64-host
runtime skip as a failed qualification. The disjoint 18-root host-applicable
cohort does not replace or certify that required native counterpart.

All 24 fixed/immediate integer EQ/GT comparison names share the same validated
same-vector dependency rule. Self comparisons need no incoming numerical vector
bits; actual masks, different inputs, memory, raw origins and MMX keep their
separate requirements. Retain operands for emitter allocation, existing FLAGS
and upper-alias rules. Complete Go rows/attributes, all imm8 controls and banks,
three-API five-target objects and whole-byte/mask/FLAGS/canary execution pass.
Actual-host portable execution is not native AVX512 execution.

The scanner gate then found a real mandatory-descriptor bypass: the Void/no-
result shortcut ran before `requiresResultProof`. True RED `f50765b07` has
91 failed terminals, five original Go object receipts and no fixture/tool error.
Moving the shared descriptor check before that shortcut fixes immediate compares,
CRC32, implicit MUL/IMUL and the remaining registered mandatory effects without
another opcode list. The two new roots / 157 children pass at `7d0ab042d`.
Three old unbound format roots fail their obsolete public-success expectation;
`0cb14d2fa` preserves every original source row as Go/public-Context/emitter-only
object coverage. Existing initialized public owners remain independent.

The sealed Go1.20–1.27 audit changes exactly eight AMD64 self-initializing formats
Context->supported, 92 isolated missing-input formats supported->Context, and
one 386 IMULL format supported->Context per SDK. Encoding rules and every other
metric remain unchanged, with no new unsupported forms. Baseline changes retain
that distinction; a form classification is not execution of every instance.

## Strict standard-library benchmark

The complete run at `bc79fd0f` finished naturally with exit 1 and unchanged
source/SDK/LLVM 22 observations. All 184 original target/file pairs were attempted;
132 compile and 52 fail, with zero N/A or unattempted. Compilation is not runtime
execution. The result is failed, not a qualified performance pass.

| Target | Compiled / total | Failed | Wall seconds |
|---|---:|---:|---:|
| linux/386 |17/33|16|73|
| linux/amd64 |29/55|26|709|
| linux/arm |27/30|3|5|
| linux/arm64 |40/47|7|12|
| js/wasm |19/19|0|5|

Total target wall time 804s meets the 900s total budget, but AMD64 exceeds its
300s target budget. A one-second diagnostic sample captures substantial
SHA256/file-validation work; inspect `benchmark-bc79fd0f/amd64-sample.txt`
before assuming instruction lowering is the bottleneck. The sampled diagnostic
cannot certify a performance improvement. Full reports remain under ignored
`compile-benchmark-bc79fd0f` and `benchmark-bc79fd0f`.

## Immediate work

1. Repair the 52 complete standard-library failures by actual shared rule class,
   retaining their unchanged complete sources and the original 300s/target and
   900s/total budgets. Current failures include source-input effects, private
   helpers/custom entries, AUTO/native frames, TLS/implicit state and typed
   result backing. A Context rejection is not successful compilation; determine
   why a real original Go program lacks its translation contract before adding
   a new factory, exception or validation layer. Investigate repeated checking
   using the captured performance sample without removing mutation, lifetime
   or source-boundary failures. A concrete next candidate is the parser's
   conflation of omitted TEXT arguments (`$0`) with explicit zero (`$0-0`).
   Go's assembler retains ArgsSizeUnknown for the first; the private Void extent
   check rejects both. `subtle/xorBytes` reaches this error. Reproduce before
   changing it; keep explicit mismatch and native-frame/address controls.
2. Use the completed `06eedf847` run's exact 167-root failure index. All 16
   parts / 64 shards have been downloaded and audited against their identical
   actual source/head/inventory. Its complete local replay at `83f21268e`
   has 10 passes, 68 failures and 89 host skips; eleven of those failures now
   have the separate focused qualification above. Preserve both failed runs
   and their host skips;
   source fixes staged afterward are not online passes. Avoid another complete
   diagnostic run while its known failure classes remain unaddressed.
3. Repair only remaining observed rule classes. Preserve discovering sources,
   independent complete Go rows/attributes, initialized positives and negatives.
   Use shared typed grammar/effects rather than library-specific exceptions.
   Existing failure classes include scalar input/effect gaps, MMX-region
   composition and stale unbound-format expectations; the completed CI failure
   index is the authority, not an estimated remaining count. Use the latest
   167-root replay instead of the older 179-root local cohort. Host skips
   remain unexecuted native counterparts, never repaired tests.
   This selected queue does not revalidate the old complete suite's passed roots.
   Its main classes are numerical/field-address effects, MMX closure, raw/native
   read composition, custom ABI/stack roles and obsolete unbound-format contracts.
   The first old local failure remains `TestGoABI0PaddedSqrtObjects`. Its logical
   floating-field address class is now qualified; SQRT arithmetic is separate.
   Packed private-copy stores and all eight mask/vector data effects are now
   qualified, as are the complete CPU-query implicit NUM and exact named/raw
   continuation effects. Scalar-shift memory effects now have the complete
   focused qualification above, followed by the complete multiply-add and MMX
   format-contract union and GP/SIMD native-store inputs/continuation. Next close
   native block-extraction stores and remaining MMX region composition failures. Preserve unmodified
   failed sources and independently initialized positive programs. Do not assume that a whole
   format batch's first MMX rejection establishes every sibling's input checks;
   retain per-mode positive and missing-input controls.
   CPU-query numerical effects do not supply the separate custom entry of the
   runtime386 `cpuidProbe`. Preserve failed receipts, native runtime owners and
   the already-passing simple query-result control; focused successes are not
   subtracted from a completed run.
   Keep compiler/runtime producer gaps separate from instruction support. No
   mnemonic-only exception, padding permission or native-frame backing follows
   from a valid encoding. Inspect the final error, not a preceding successful
   tool/CPU preflight log: Windows conformance reaches an unclosed three-operand
   SHLL memory effect, not an AVX512 hardware-availability failure.
4. Require current-head complete CI, benchmark, coverage and fresh review before
   ready. Keep ABI/context failures visible; do not weaken the benchmark or
   treat Context classification as successful program compilation.

## Reproducing and finding receipts

Use the commands and LLVM 22-only environment in [validation](validation.md).
Run focused RED first, then the complete affected union on current Go, actual
Go 1.20 and race. One heavy local Go/LLVM owner runs at a time; develop in an
existing idle worktree while its tested source stays frozen.

Resolve branch locations with `git worktree list`. Latest ignored receipts use
`gp-simd-store-qualified-{contracts,results,objects}` at `2d9059765`;
`gp-simd-store-qualified-2d905976.json` seals the complete 42-root three-lane
union. Final gates use `gp-simd-store-final-gates-2d905976`; the eight-SDK audit
is `gp-simd-store-classification-audit`. Original RED uses
`gp-simd-native-store-red` at `e2e2406fa`; the related replay's two obsolete
contract failures remain preserved, with original sources and separate
initialized positives retained in the qualified batch. Earlier receipts use
`multiply-add-qualified-{contracts,multiply,physical,initialized}` at `cdbc92b85`;
`multiply-add-qualified-cdbc92b8.json` seals the complete 69-root three-lane
union. Final gates use `multiply-add-final-gates-cdbc92b8`, and the sealed
eight-SDK audit is `multiply-add-classification-audit`. Initial RED uses
`multiply-add-input-red` at `87ae8fb66`. A first orchestration attempt stopped
on its invalid mixed-case phase name after the contracts group passed; the
version-two driver retained that receipt and completed the exact full union.
Earlier receipts use `scalar-shift-memory-qualified-complete` at `3d26bf2cc` for both normal SDKs and
`scalar-shift-memory-qualified-complete-retry2` for the complete race lane;
`scalar-shift-memory-qualified-3d26bf2c.json` seals their exact three-lane union.
The earlier race and `-cached` attempts remain failed. Final gates use
`scalar-shift-memory-final-gates-3d26bf2c`; the eight-SDK audit uses
`scalar-shift-memory-classification-audit-v2`. The independent native spot check
is `zero-shift-native-result.json`, not the full native root. Fresh MMX format
RED uses `mmx-original-format-red` at that same code head. Earlier receipts use
`cpu-query-result-red` at `dab3bb581` for the original CPU-query failures.
The latest completed qualification uses `cpu-query-qualified-{core,state}` at
`e95a10314`; `cpu-query-qualified-e95a1031.json` retains the complete forty-root
three-lane union. The sealed eight-SDK audit is
`cpu-query-classification-audit-v2`; the first audit attempt retains its strict
report-format assertion failure, not a translator failure or a completed audit.
Final gates use `cpu-query-final-gates-e95a1031`.
Earlier completed qualification uses
`unbound-official-formats-{first-green,initialized,native}`; the complete
20-root three-lane seal is `unbound-official-formats-qualified-5df8cde3.json`.
`latest-failures-83f21268/raw-outcomes-audit.json` retains the complete latest
167-root local replay, including every actual failure and host skip separately.
Earlier receipts use
`integer-self-compare-final-{core,effects,grammar,mandatory}`;
`integer-self-compare-qualified-0cb14d2f.json` records the exact disjoint union.
The eight-SDK audit is `integer-self-compare-classification-audit` at `7d0ab042d`;
later changes before qualification are only test contracts and their baseline.
The local 179-root replay is `previous-failures-d5bdb039`; its
`raw-outcomes-audit.json` retains every actual failure and skip separately.
The floating-field RED is `logical-floating-field-address-red` at `d0505fd99`;
the 57-root seal is `logical-floating-qualified-8e82392c.json` and final gates
use `logical-floating-final-gates-8e82392c`. Format-contract qualification uses
`unbound-format-contracts-host-applicable-qualified-e7e90b7c.json`.
Windows complete-owner artifacts use `ci-39065963275-windows-root-*`.
The exact latest completed raw root union is
`ci-39065963275/root-union-audit.json`. Packed-copy RED/initial green receipts
use `packed-logical-parameter-store-*`; the unchanged-program portable/typed-
field selection is `packed-private-copy-complete-first-green` at `cdcd98cc6`.
Final qualification is `packed-private-copy-qualified-4bcd33be.json`; final
gates use `packed-private-copy-final-gates-4bcd33be`.
Mask/vector RED uses `mask-vector-numerical-red` at `521ef9c48`;
qualification is `mask-vector-qualified-252239cd.json`, with final gates under
`mask-vector-final-gates-252239cd`. The eight-SDK full-form audit is
`mask-vector-classification-audit-v2`. The next selected original-CI replay
uses `latest-failures-<head>` and `rerun-latest-ci-failures.cjs`; its raw
pass/fail/skip outcomes are not a new complete all-platform CI denominator.
Earlier receipts use `nop-source-state-qualified-{nop,stack,identity}` and
`nop-current-sdk-producers`; `qualified-nop-source-state-bc79fd0f.json` records
the exact union. Earlier cohorts use `bitfield-qualified-*` and
`paired-native-qualified-*`. CI artifacts are retrievable from the linked runs;
local reports are useful evidence but do not become tracked repository data.
