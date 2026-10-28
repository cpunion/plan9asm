# PR40 closeout

Updated 2026-10-29. Finish **amd64 correctness, shared typed families and
translation efficiency first**, preserving all five architecture gates.
Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). Historical receipts remain in git history;
this page records the current candidate and its limits.

## Publication and review

PR40 remains draft. Published `4aeffe806` has its
[own CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40713618690).
Its unique inventory is **2,748 rows / 10,314 witnesses**:
2,617 of 3,568 AVX-table rows (73.3%) and 131 legacy rows. This measures
initialized original-Go bytes and LLVM objects, not native execution or every
amd64 opcode. Both endpoint SDKs passed the combined family, full rule and
five-architecture gates before publication. Complete policies and the affected
race scope also passed. Require its complete new-head CI and coverage.
The preceding `61b5737f2` round completed **325/325** successfully, with
**88.1%** statement coverage and a successful exact-head
Codecov patch check. All 325 named jobs and their check IDs match the retained
inventory. Complete race and Windows reports each cover all 22 packages;
the historical race 299-root / seven-child and Windows 279-root skips remain
skips. Local follow-up qualification and external approval remain separate.

Preceding `52c40967a` attempt 1 finished with 321 successful checks, three
failures and one cancellation. A targeted Windows rerun passed the original failed
root partition and its complete aggregate; the earlier failure remains recorded.
Attempt 2 has 323 successes and retains the original race failure and dependent
cancellation. The race-other job reached its 75-minute job limit, leaving no
complete race-module receipt. The SDK frontend now uses the existing profile-owner shards,
with no larger budgets or omitted roots. Local SDK race qualification passes
34 roots / 35 children on current Go and 18 / 27 on minimum Go; sixteen
current-only roots are not compiled by the minimum SDK. The actual complete
22-package cloud race union passes, retaining the historical 299
root and seven child skips. Coverage is 88.1% and the exact-head patch check
passes. Whole-round Windows completion remains separate. Do not call the
preceding `52c40967a` round wholly successful.

The preceding entirely green publication is `61b5737f2`:
[325/325 Go checks](https://github.com/xgo-dev/plan9asm/actions/runs/40688196703),
88.1% main-package statement coverage and a successful Codecov patch check.
Its 299 root and seven child skips remain skips. Its counts do not certify a
later revision. Earlier source failures and infrastructure retries retain their
original receipts; resolved inline threads are not external approval.

Published `790fcb4b1` adds complete signed/zero widening: 24 source
names, 108 rows and 288 witnesses. Both endpoint SDKs pass actual Go encodings,
initialized three-API objects, the full rule reader and five-architecture gate.
All prior rows, bytes and generated sources remain identical. The translator
and initializer are unchanged. Its intermediate union is **2,181 rows / 8,046
witnesses**: 2,056 AVX rows (57.6%) and 125 legacy rows.

Published `111e549f9` adds the complete 20-name broadcast family:
111 rows / 234 witnesses, with separate vector, GP, K-data and memory sources.
Both endpoint SDKs pass original-Go bytes, all three API objects, full rule
tests and five-architecture gates. All old 24 views and 269 generated source
batches remain identical. Translator semantics are unchanged. Its intermediate
union is **2,292 rows / 8,280 witnesses**, including 2,167 AVX rows (60.7%) and
125 legacy rows.

Published `e2de31937` adds complete single-source narrowing: 18 names,
108 rows and 324 witnesses. It shares a typed form across effects, result
analysis and lowering, fixing unsigned-input saturation and masked-store
accesses. Carrier width, active output tuple and old-destination requirements
are separate. The former scalar saturation loop becomes vector comparisons,
selection and truncation. Translator production grows by 96 lines; PACK keeps
its signed-input semantics.

Both SDKs pass numerical boundaries for all 18 members, masked-store fault and
64-byte canary checks, partial-carrier controls, actual Go bytes and three-API
objects. Of the 324 forms, 270 have ordinary continuations; 54 Go-encodable
memory `.Z` forms use the existing terminal UD2 contract before input access.
The old 25 views and 277 generated source batches remain identical. The local
union is **2,400 rows / 8,604 witnesses**: 2,275 AVX rows (63.8%) and 125 legacy
rows. The final combined source passed all nine publication gates.

Its baseline changes only 147 amd64 probe classifications from IR generation
to Context. Three source profiles retain all 212 original examples: 188 lack
inputs, twelve lack a BP role and twelve lack the original SP/TEXT frame.
Before/after probes preserve their original IR and current diagnostics; other
architectures and encoder fields are unchanged. Initialized AX-based witnesses
do not grant those original SP/BP examples a frame contract. Bare original
named/raw sources remain isolated format/object checks, distinct from the new
initialized API tests. Ordinary raw numerical and native AVX512 execution remain
separate work.

RCP14/RSQRT14 native F and F+VL owners are implemented under the explicit
`plan9asm_native_avx512` tag. Default checks compile original Go, three-API
objects and the C oracle fixtures; none establishes native AVX512 execution.
Use the required command in [validation](validation.md) on suitable hardware.
Numerical, MXCSR and fault acceptance remains pending; retain the two
repeated-R15 address-contract Context cases. The older packed reciprocal
owner has actual Linux CI execution, separately from reciprocal14.

The locally qualified SDK digest change retains only direct, immutable struct
field metadata within one complete digest call. It adds 32 production lines;
values, payloads and mutable source verification are never cached. Both SDKs
retain byte-for-byte serialization, fixed golden streams and existing mutation
controls. Paired complete-encoder measurements on the same real CPU, CRC32 and
maps captures improve normal time by approximately 15–38% and race time by
19–25%, with substantially fewer allocations. Capture and live Verify are
outside those timers; this is not a whole-SDK or CI acceleration claim.
Combined primary integration checks passed before publication at `61b5737f2`.

The declaration-demand batch was qualified at `fe97b6d53`, then published as
`9fede619c`. It adds 186 official
rows / 765 witnesses, avoids unnecessary SIMD feature queries for baseline GP
forms, and validates actual native inline/module assembly against the source
backend, then selects amd64/386 LLVM declarations from actual generated uses.
Its five-architecture gate passes on frozen source/tools. Complete original SDK
objects retain 41/55 passes and all 14 previous failure diagnostics. Backend
qualification includes both endpoint SDKs and both complete CLI suites below;
the later declaration change has its own scoped qualification. New-head cloud
CI remains required after publication.
The failed `004da9f44` cache-policy round and earlier failures retain their
original outcomes. Its stale Node assertion was repaired without removing
the actual Go lifecycle, ownership or cleanup checks.

[LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694), publication `9422371af`,
is rebased onto `748cf1811` and still pins the qualified public
`v0.0.0-20261027230043-3b61ac1fb97c`. All 117 contribution patches survive the
rebase unchanged. The public ZIP's 5,049 files match exact Git blobs and its
license. Adapter, assembly-build and affected Go checks retain their separate
receipts; neither those tests nor this temporary dependency resolves every
runtime/ABI issue.

All ten WASM suites passed at `df6f5eeb2` after splitting original stdlib
targets and caller profiles into independent jobs. Original cases, flags,
cache settings and 20/15/60-minute budgets remain intact. That complete round
ended with 136 successful jobs, one policy fixture failure and two conditional
release skips across thirteen workflows. The failed fixture read metadata
after file creation but before its complete write. Published `8f68e5878` fixes
publication atomically and requires the Go/Git/Node/Bash tools used by policy
tests. All 191 local policies pass without skips. Its child `9422371af`
changes only benchmark-table display, separately validated against the complete
YAML structure. Its actual cloud policy passes all 191 unique methods without
skips: cold Go1.27.0 / Node24.22.1 setup and tests take 111 seconds within the
unchanged five-minute budget. All thirteen compiler workflows now succeed;
the latest 139 jobs contain 137 successes and two conditional publisher skips.
The first round's artifact-transfer failure remains recorded. Only its failed
prepare job actually reruns; all six ZIP digests match and the candidate script
produces six YAMLs. No source, runner or timeout repair was needed for this retry.

The native Windows demo repair at `8c3555126` has actual six-profile CI coverage,
including the unchanged 4 x 5,000-operation temporary-file stress cases. Preserve earlier
timeouts and fixture failures. The P1 dependency-provenance requirement remains
open until an official upstream version replaces the temporary fork pin.
Never push or merge upstream directly.

## Latest published family batch

These five complete groups have individual endpoint-SDK qualification,
byte-identical integration and nine passing joint gates on the same frozen
primary snapshot. They are published at `4aeffe806`; its CI remains separate.

| Group | Official rows / witnesses | Change and acceptance |
|---|---:|---|
| FP/integer numeric conversions | 162 / 876 | 42 names; shared harness only; six MMX rows remain separate |
| Immediate integer comparisons | 48 / 240 | Eight signed/unsigned members; shared K-result transport; translator unchanged |
| VPTESTM/NM | 48 / 240 | Eight members; typed effects and result ownership, conditional volatile reads preserve live/dead access behavior |
| GFNI multiply and affine/inverse | 24 / 84 | Three members; shared format adapter; translator and initializer unchanged |
| Leading-immediate vector group | 66 / 270 | Eleven names across four semantic families; existing FMA3 producer and fixed-prefix encoder reused |

The published unique union is **2,748 rows / 10,314 witnesses**: 2,617 of 3,568
AVX rows (73.3%) and 131 legacy rows. Both PADDD aliases and the 24-row /
108-witness rotate overlap remain separate and are not double counted.
All earlier fields, encodings, generated source/declarations and case IDs
remain identical on both SDKs. Existing numerical owners retain their original
scope; these joins add no native AVX512 or ordinary raw execution claim.

VPTESTM/NM reproduced a missing result/effect transfer and active-memory access
being incorrectly removed or performed under a zero mask. Its shared typed
form reuses the existing conditional volatile E4 memory path. Only 62 bare
amd64 probe forms change from IR generation to Context: their original 150
examples retain undefined inputs or unbound original SP/BP roles. They are not
positive initialized objects. Other architectures and encoder data are unchanged.

The leading-immediate vector group first reproduces the formatter rejection
on both SDKs, then verifies all 66 rows / 270 witnesses with original Go bytes
and three-API objects. Strict tuple controls also fix incorrect broadcast
metadata acceptance for W/BW forms. All old thirty views and 335 generated
source/declaration batches remain identical. Align, double-shift, ternary and
byte-SAD retain their separate numerical and old-destination contracts;
byte-SAD still needs an independent fault-execution owner.

## Next local batch

Locally integrated `1cb030c44` completes PALIGNR/VPALIGNR, SHUFPD/VSHUFPD and
SHUFPS/VSHUFPS: 27 rows / 90 actual Go witnesses and three-API objects on both
SDKs. All old 31 views, 344 generated source/declaration batches and two aliases
remain identical. Its shared leading-immediate renderer removes 37 lines
against the same-coverage implementation. Including new family support, the
whole batch adds 33 harness lines and 266 test lines, with no translator,
initializer or encoder changes. Both SDKs pass the full 85-root / 720-child
rule suite, policies and five-architecture gates.

The local unique union is **2,775 rows / 10,404 witnesses**: 2,641 of 3,568
AVX rows (74.0%) and 134 legacy rows. This follow-up is not published or
jointly qualified on its final primary snapshot. Preserve the initial RED,
legacy PS/PD-prefix rejection repair and original overlapping qualification
attempts; the minimum identity and Node gates have separate serial reruns.
No translation speedup or additional native execution is claimed.

Locally integrated `2ba15117d` completes compress/expand: twelve names / 72
rows / 216 actual Go witnesses. The existing move producer and per-function
terminal inspector are reused unchanged. Its 198 ordinary forms and eighteen
terminal memory `.Z` forms remain separate; each SDK verifies all 54 terminal
function/API bodies for UD2, unreachable and no input loads. Eight additional
actual-Go displacement anchors cover both directions and element scales
1/2/4/8 without increasing the row-witness denominator.

The same-coverage unary memory renderer removes 34 lines. All old 32 views,
347 source/declaration batches and aliases remain identical on both SDKs;
the translator and producer are unchanged. Each SDK passes the complete
87-root / 964-child rule suite, objects, five-architecture and policy gates.
The local unique union is **2,847 rows / 10,620 witnesses**: 2,713 AVX rows
(76.0%) and 134 legacy rows, including 102 terminal controls. Ordinary
protected-page/dead-result fault execution remains incomplete. The setup path
and zero-test invocation failures remain recorded as setup failures.

Locally integrated `03c41000c` completes GETMANT, RNDSCALE, REDUCE, FIXUPIMM
and RANGE: twenty PS/PD/SS/SD names / eighty EVEX rows / 390 actual Go
witnesses. Both SDKs pass the original byte selections, six independent
displacement/SAE anchors, three-API objects, full rules (89 / 1,006), policies
and five-architecture gates. All old 33 views and 354 source/declaration
batches remain identical. The small format factory and limited shared
arity/SAE axes add 104 harness lines; translator, initializer and encoder are
unchanged. Scalar N4/N8 tuples do not widen the X carrier; FIXUP still needs
the active old destination even with `.Z`. Whole-family native fault/MXCSR
execution remains incomplete.

The local unique union is **2,927 rows / 11,010 witnesses**: 2,793 AVX rows
(78.3%) and 134 legacy rows. The three local follow-ups await one final
combined primary-snapshot qualification and the current public CI round
before another batch publication.

The follow-ups are now jointly qualified and published at `bebc7092d`, with
[new-head CI](https://github.com/xgo-dev/plan9asm/actions/runs/40731923496)
pending. The preceding `4aeffe806` round completed 325/325 successfully,
88.1% statement coverage and its exact-head Codecov patch check. Actual race
and Windows unions retain their 299-root / seven-child and 278-root skips.

## Local legacy GP follow-up

Integrated `b57928448` preserves all 53 scalar RMW names / 169 original source
rows. Actual Go selection and mode observations establish 167 rows / 278
shapes on amd64 and 124 / 202 on 386. Two Px1 short-register rows remain
386-only; all 45 Q-width/386 refusals have independent frontend observations.
First-match, AL/AX aliases and shadow mappings retain their original records.
The 480 cross-mode observations deduplicate to 280 physical forms.

Six physical schemas consume original opcode payloads. The producer returns
actual data and only defined FLAGS, using the existing POP/immediate-AND
recovery rules; it grants no new numerical or storage permission. Both SDKs
pass actual Go bytes, three-API O0/O2 objects, family 6 / 139, full rules
92 / 1,029, five-architecture and policy gates. All old 34 groups, 367 generated
source/declaration batches and two aliases remain byte identical. The original
O0-only report is retained separately from the final O0/O2 qualification.

This adds 167 amd64 legacy rows without overlap, for 301 locally joined legacy
rows. AVX remains 2,793 / 3,568; modes and source inventories have separate
denominators. This follow-up is not published and adds no runtime or performance
claim. Translator, shared reader, encoder, initializer and baselines are unchanged.

Integrated `f23cb32d2` completes eight shift/rotate stems across B/W/L/Q:
120 original `optab` rows and eighteen `ymovtab` double-shift records. Both
SDKs pass 138 amd64 rows / 356 shapes and 100 386 rows / 264 shapes, including
38 concrete Q/386 frontend refusals, actual Go bytes and three-API O0/O2
objects. Defined-FLAGS masks retain count-dependent and existing paired-choice
boundaries. All prior groups and runtime source/declaration/C-driver bytes
remain identical. Full rules pass 94 / 1,046, family 6 / 138, both five-architecture
gates and complete policies. No runtime execution, permission or translator
change is claimed.

The cross-mode source inventory now has 3,234 records and 11,646 deduplicated
row/instruction/encoding witnesses plus two aliases. These totals combine
modes; they are not a single-architecture percentage. The amd64 additions are
120 `optab` rows plus eighteen separately counted fallback records. Next,
validate all Jcc/SETcc source forms with input-dependent results, retaining
hint/short/near/alias selection and 386 synthesized byte-register paths.

## Current implementation

- Native machine bits and dereference permissions remain separate in existing
  typed effects. Defined addresses can participate in arithmetic; numerical
  results gain no pointer extent, frame role or source lifetime.
- Private automatic frames use the existing CFG. Ordinary analysis runs first;
  one retry partitions compatible private-offset layouts using the remaining
  budget. Role keys, SP, saved BP, all eight SP residues, ordered accesses and
  final independent replay must agree. Limits remain 2^18 steps, 64MiB logical
  work and 1MiB retained payload. This adds 380 net production lines without
  another instruction interpreter or a larger budget.
- The retry charges variable clone, meet and candidate payloads. Pure GP work
  reuses the current instruction's typed descriptor at its original late
  fallback. Special/native/private paths keep their order. A full-form cache
  was measured and rejected; mutable source verification is never cached.
- Readonly logical dependencies reuse the packed logical family and complete
  original Go DATA bytes. AND/ANDN need the second input only where the literal
  has set bits; OR needs it where the literal is not all ones; XOR needs all
  bits. Full memory access, alignment, identity and live owner checks remain.
  This repairs SHA1 with 77 net production lines and preserves the old shifted
  literal rules. Zero-filled implementation storage supplies no source facts.
- Existing compare and GP/X transfers share typed byte effects. Terminal RET
  validates each incoming state before dropping dead private identities.
  Ordinary compatible dead private offsets now have a whole-source positive
  test; mismatched roles, saved BP, undefined bytes and restoration still reject.
- Byte-definedness uses a bitset. The four-family packed normalization pilot
  and shared feature projection remove 76 production lines. SDK verification
  reuses its fresh dependency query, reducing two subprocess queries to one.
  Test-owned compiler artifacts are shared with independent source leases and
  bounded cleanup; user-global caches are never cleared.
- Baseline GP feature inference reuses existing RMW, LEA and direct-branch
  grammars after logical/vector routing. Physical forms and separately bound
  targets retain their paths. No source permission or mutable cache is added.
- Backend validation follows actual inline-assembly operands and actual
  nonempty module assembly, including WASM relocations. Matching source modes
  take a fast path; portable empty compiler barriers remain legal. The official
  LLVM binding is v0.10.1 in all three modules. Its missing module-assembly
  getter uses a 23-line LLVM22 C API length bridge, without parsing IR or adding
  linker flags. The repair adds 81 net production lines.
- AMD64/386 declaration selection derives immutable catalogs from the original
  emitter and reuses the ARM64 global-name scanner. Six ordinary declarations
  and source-dependent SHA/string/rounding helpers retain their owners. Direct
  module success adds no lookup or scan. This adds 68 net production lines,
  without another signature table, opcode registry or source/caller cache.
- Backend checks normalize target triples with LLVM22 once, including short
  WASI triples and reordered architecture aliases. All guards use the same
  normalized observation; the emitted module target and original input remain
  unchanged. The small C bridge releases both owned strings.
- Native LOCK/CMPXCHG tests bind their original instructions to matching
  backends. Backend rejection controls remain separate from extent, FLAGS,
  alias and malformed-source negatives, so an early target error cannot mask
  a semantic acceptance regression. Production admission is not weakened.
- BF16 tests inspect actual conversion and masked-load calls, with the six
  original rows plus three full-memory forms. Unused declarations are not
  evidence of an emitted memory operation. Real consumer module fixtures query
  the selected LLVM binding and replacement instead of pinning another version
  while copying the repository checksum file.
- Reciprocal14 uses the existing typed native floating helper for all eight
  members. Independent actual O0/O2 objects retain two operations around two
  ordered MXCSR changes and preserve masked native memory operations when
  results are unused. The old pure intrinsics allowed CSE across MXCSR and
  removal of ordinary source loads. Scalar upper-copy and merging low bytes
  have separate input roles; feature projection requires F, plus VL only for
  packed X/Y. Original 304 source shapes remain as Go encoding observations,
  with initialized three-API objects and explicit mode32 high-bank context.
  Dead spec state and sixteen unused intrinsic declarations are removed.
- All 68 original reciprocal14 statements retain their operands. Sixty ordinary
  inputs and six freshly captured SP/BP frames compile through three APIs;
  two repeated-R15 pointer-index forms remain explicit address-contract Context
  cases. They are neither positive objects nor instruction-form exclusions.
  The private boundary adapter validates 132 original Go encodings without
  registering raw semantic translation. Shared frame transfer checks typed
  reads, including K, before writes and retains the original storage envelope.
  Baseline changes are limited to 43 bare forms becoming Context on each of
  three verified source profiles; encoder data and other architectures remain.

The final GP descriptor reuse avoids the rejected eager-query overhead.
Measured replay ABBA differences against the same partition implementation
before eager analysis are +4.0% time for register work and +2.4% for mixed work,
with allocation-byte differences +0.22% and +0.56%. Flat cases vary -0.05% to
+1.62%. These are residual local costs, not a whole-translator speedup.

The later baseline GP feature query removes allocations in six isolated
inference cases and reduces their time by approximately 86–88%. Complete GP
translation improves by 2.58–5.30% across paired blocks and endpoint SDKs.
The complete original P256 workload is mixed (+4.17% / -1.64%); its 35,079
source bytes, 14 TEXTs, ten public definitions and four internal helpers remain
intact, with equivalent IR/attributes and LLVM22 objects. These results do not
establish a whole-SDK speedup or generated-program acceleration.

Declaration selection is measured separately, using uninstrumented ABBA runs
with 12 unchanged-input cases and ten samples per variant/case/SDK. Across both
SDKs, the nine packed translation cases improve approximately 28–29%, GP CFG
13–17%, and annotated text fallback 46–48%. Direct control varies below 1%, with
unchanged allocations. Immutable catalog construction costs about 0.19–0.23ms
and 224–230KB once per source mode; that is not complete cold-API startup cost.
Full P256 loses about 67KB of unused declarations; its function bodies,
attributes and ELF execution sections remain equivalent. Its wall-time results
are mixed (+4.58% / -0.39%), so no complete-SDK speedup is claimed.

## Qualification of this batch

| Scope | Result | Boundary |
|---|---|---|
| Joint Go1.27.3 / LLVM22 at `d2784db04` | **104 roots / 4,292 children PASS**, zero failures/skips | Original selected tests, unique complete terminals; frozen source/tools |
| Joint minimum Go1.20.14 | **73 roots / 3,938 children PASS**, zero failures/skips | Actual Go-selected files; 31 current-SDK source roles are absent, not runtime skips |
| Complete preflight at `d2784db04` | **573 Node policies; 26 Go roots / 299 children PASS** | Exact Node22.23.4 and unchanged CI selector |
| [Joint Linux execution](https://github.com/cpunion/plan9asm/actions/runs/40496018236) | **3 roots / 35 children PASS**, zero failures/skips | Workflow-only child `0a3eb36b2`; complete Git blobs, tools and raw terminals independently audited |
| [Private-frame Linux execution](https://github.com/cpunion/plan9asm/actions/runs/40489811812) | 5 roots / 36 children PASS | Earlier `89e53140f` source; physical FLAGS owner is original-Go-only |
| Complete original SDK objects | **41/55 PASS**, 14 failures, zero N/A/unattempted | Go1.27.2 Linux/amd64; identical 55 original hashes; SHA1 newly passes, no prior pass lost |
| Five-architecture official corpus at `d2784db04` | PASS | Go1.27.3; no baseline mismatch or parse error; classifications are not initialized/runtime coverage |
| Horizontal integer official join | 56 roots / 685 children PASS on each endpoint SDK | Independent rows, original Go bytes and LLVM22 O0 objects through three APIs |
| Packed ABS join and affected regressions at `b5ca7da71` | Go1.27.3: 60 roots / 1,048 children; Go1.20.14: 59 / 1,048 PASS | Full rule reader, every existing row join, original ABS/sign five-target objects and portable values; zero failures/skips |
| Combined ABS/FMA3 join at `d21fb3c8f` | Go1.27.3: **112 roots / 9,018 children**; Go1.20.14: **111 / 9,018 PASS** | All prior row joins, 178 initialized FMA axes, full named formats, raw/undefined-input controls and existing portable owners; zero failures/skips |
| Integrated `6ca5e4a9d` preflight | **573 Node policies; 26 Go roots / 299 children; five-architecture gate PASS** | Code identical to the qualified combined join; source/tools frozen |
| AVX move join at `72e765098` | Both endpoint SDKs: **114 roots / 1,765 children PASS**, zero failures/skips | Host-applicable scope; two x86 execution owners remain with required Linux counterparts |
| Final integrated `824f9375b` preflight | **573 Node policies; 26 Go roots / 300 children; five-architecture gate PASS** | Exact tools, frozen source; current/minimum-first creation order preserves every SDK and owner |
| Ordinary AVX floating binary at `ff2f14caa` | Both endpoint SDKs: **114 roots / 2,004 children PASS**, zero failures/skips | Complete previous row joins and existing floating grammar, raw and input owners; focused controls pass 5 / 54 per SDK |
| Integrated `ff2f14caa` preflight | **573 Node policies; 26 Go roots / 300 children; five-architecture gate PASS** | Frozen source/tools; no translator production changes |
| Legacy floating cleanup at `a1f3cdb6a` | Current SDK **168 roots / 5,173 children**; minimum **167 / 5,173 PASS**, zero failures/skips | Existing floating, precision, sqrt, raw FP16 and row owners; one Go1.27 source owner is absent on the minimum SDK |
| Final integrated `c1e4206f2` preflight | **573 Node policies; 26 Go roots / 300 children; five-architecture gate PASS** | Code identical to qualified cleanup; documentation differs |
| Complete remaining EVEX binary join at `70e1aad95` | Both endpoint SDKs **203 roots / 4,794 children PASS**, zero failures/skips | 31 members / 186 rows / 765 witnesses; old 16 groups retain every field and output hash |
| Baseline GP feature projection at `da3e6029b` | Both endpoint SDKs **47 roots / 2,183 children PASS**, zero failures/skips | Existing grammar and all prior benchmark inputs retained; 573 Node policies and 26 / 300 Go policy checks pass |
| Actual native-assembly backend guard at `537be28b8` | Each endpoint SDK **89 roots / 5,188 children PASS**, zero failures/skips | All five source families, three public APIs, actual LLVM22 native/portable objects and retained foreign-target REDs |
| Complete CLI and policy qualification at `537be28b8` | Simple CLI **38 / 39**; rich CLI **401 / 2,198**; Go policy **26 / 300**; Node **573 PASS** | Zero failures/skips; one Linux-only rich CLI source root is absent on Darwin, not skipped; root five-architecture gate also passes |
| Native-target test contracts at `3e03de05d` | Each endpoint SDK **29 roots / 1,733 children PASS**, zero failures/skips | All old cases retained; native positives name their actual target and unspecified-target negatives remain required |
| Declaration demand at `72dd692a2` | Current **70 / 3,141**; minimum **68 / 2,799 PASS**, zero failures/skips | All baseline case IDs retained; new declaration and helper-closure controls; two current-only roots are absent on minimum |
| Declaration race and preflight | Current race **6 / 72**; Go policy **26 / 300**; Node **573 PASS** | Frozen source/tools; integrated `fe97b6d53` differs only in documentation and passes the five-architecture gate |
| Complete SDK objects at `fe97b6d53` | **41/55 PASS; 14 FAIL**, zero N/A or unattempted | All original Go1.27.2 source hashes and all 14 failure diagnostics match the preceding baseline; raw command exits 1 |
| BF16 used-memory regression at `1b34ab922` | Both SDKs **10 roots / 13 children PASS** | Original six rows retained; three full-memory rows added; both Go encoders and original object owners retained |
| LOCK target contracts at `f97cb7e42` | Both SDKs **26 / 5,374 PASS**, zero failures/skips | Original native forms retained; backend rejection is separate from numerical/FLAGS ownership |
| Normalized backend/WASM scope at `c5f1492a1` | Current **58 / 1,245**; minimum **57 / 1,245 PASS** | New five roots / 60 children include 21 actual WASM objects; one Go1.21-only source root is absent on minimum |
| Actual LLVM consumer fixtures at `c71c7cdd0` | Both SDKs **4 / 16 PASS** | Four original Unix failures reproduced first; Windows test binary cross-compiles, without claiming local Windows execution |
| Complete LOCK/CMPXCHG host scope at `5d84ab6a3` | Both SDKs **58 / 8,075 PASS**, zero failures/skips | Original 63-root run retains five real native-execution skips; Linux owners remain required in CI; original FP15 controls plus 24 backend negatives |
| Integer min/max E4 reuse at `5d6de821e` | Each SDK **34 / 743**; full rules **46 / 223 PASS** | Production net -39; existing three benchmark inputs retained, with no acceleration claim |
| Bit-count/conflict ownership at `d56a1b125` | Each SDK new **2 / 40**; full rules **47 / 231 PASS** | 48 rows / 198 witnesses; translator, grammar and initializer unchanged |
| Uniform/byte-shift formats at `d2364cd20` | Each SDK new **1 / 64**, old groups **18 / 906**, related semantics **109 / 10,991 PASS** | New 188 rows / 488 witnesses; real baseline RED; rules **53 / 449** and policies pass; native inapplicability remains separate |
| Complete rotate family at `41ac31285` | Each SDK new **1 / 28**; full rules **55 / 463 PASS** | All eight members, 48 rows / 216 witnesses; unique addition 24 / 108; old 19 groups and 241 generated batches unchanged |
| Complete duplicate moves at `fe12b2865` | Each SDK new **1 / 12**; full rules **58 / 477 PASS** | All six members, 27 rows / 72 witnesses; old 20 groups unchanged; translator and initializer unchanged |
| Immediate word/dword shuffle at `a0238ebc3` | Each SDK new **1 / 12**; full rules **61 / 491 PASS** | Seven source spellings, 34 rows / 95 witnesses; old 21 groups and 251 generated batches unchanged; separate PSHUFL/PSHUFD rows retained |
| Complete reciprocal14 at `fb1fdb122` | Each SDK **12 / 333 PASS**; full rules **56 / 463** | 32 rows / 132 witnesses and original 304 shapes; affected host scope 27 / 495 PASS plus one old scalar execution owner N/A; native RCP14 execution owner still missing |
| Original reciprocal14 addresses at `a6a85d927` | Each SDK **16 / 783**; affected old owners **57 / 1,412 PASS** | 66 original statements have objects, two require address context; final isolated unit **1 / 52** supplements byte-identical inherited evidence; private boundary checks confer no raw capability |
| Reciprocal14 native-owner preparation at `59abb07f8` | Each SDK **15 / 504 PASS** | Actual source/trace/object and C compile/link evidence only; tagged Linux execution remains pending hardware |
| SDK race partition at `cd2413c1b` | Current SDK **34 / 35**; minimum **18 / 27 PASS** | Complete SDK frontend race roots, existing exhaustive partition runner, unchanged timeouts; current-head cloud still pending |
| Complete signed/zero widening at `8e9910c0a` | Each SDK new object owner **1 / 36**; full rules **65 / 511 PASS** | 108 rows / 288 witnesses; all old 23 views identical; no translator or initializer changes |
| Complete broadcasts at `99408cca8` | Each SDK new object owner **1 / 32**; full rules **68 / 545 PASS** | 111 rows / 234 witnesses; all old 24 views identical; generic GP/K source initialization and fixed-prefix encoding shared |
| Complete single-source narrowing at `6e2da3369` | Each SDK family **5 / 88**, full rules **70 / 562**, PACK **14 / 842 PASS** | 108 rows / 324 witnesses including 54 terminal traps; old 25 views identical; all five architecture gates pass with audited probe classification changes |
| Numeric conversions at `bcad63db0` | Each SDK objects **1 / 112**, full rules **73 / 587 PASS** | 42 names / 162 ordinary rows / 876 witnesses; old 26 views identical; translator unchanged |
| Immediate integer compare at `7c3ee63d7` | Each SDK objects **1 / 32**, full rules **75 / 614 PASS** | Eight names / 48 rows / 240 witnesses; old 27 views identical; translator unchanged |
| VPTESTM/NM at `b633be333` | Each SDK family **5 / 68**, full rules **77 / 634 PASS** | Eight names / 48 rows / 240 witnesses; old 28 views identical; live/dead memory and original context classification audited |
| GFNI at `40c876851` | Each SDK objects **1 / 12**, full rules **80 / 660 PASS** | Three names / 24 rows / 84 witnesses; old 29 views identical; translator unchanged |
| Leading-immediate vectors at `9fab6a57e` | Each SDK objects **1 / 36**, full rules **82 / 684 PASS** | Eleven names / 66 rows / 270 witnesses; old 30 views identical; translator, producer and encoding primitive unchanged |
| Align/shuffle and shared renderer at `9452a9f8a` | Each SDK objects **1 / 12**, full rules **85 / 720 PASS** | 27 rows / 90 witnesses; old 31 views identical; same-coverage renderer net -37 lines |
| Compact/expand at `3d331a0da` | Each SDK objects and displacement **2 / 28**, full rules **87 / 964 PASS** | 72 rows / 216 witnesses; old 32 views identical; eight additional actual-Go disp8 anchors; same-coverage renderer net -34 lines |
| Five FP control families at `6cee2aae5` | Each SDK objects and displacement/SAE **2 / 52**, full rules **89 / 1,006 PASS** | 80 rows / 390 witnesses; old 33 views identical; translator, producer and encoder unchanged |
| Legacy scalar RMW at `05b87b318` | Each SDK family **6 / 139**, full rules **92 / 1,029 PASS** | 169 source rows; actual selected/shadow/mode mapping; amd64 167 / 278 and 386 124 / 202 through three-API O0/O2 objects; old 34 groups unchanged |
| Scalar shift/rotate at `dfad1a73b` | Each SDK family **6 / 138**, full rules **94 / 1,046 PASS** | 120 optab + eighteen double-shift records; amd64 138 / 356 and 386 100 / 264; original runtime source and prior groups unchanged |

The joint native scope executes 45 logical kernels over all 256 byte values
with separate and overlapping buffers, 36 original shifted-literal kernels
over the same values, and three layout kernels over seven inputs. It checks
**32,277 original Go observations and 193,662 LLVM comparisons**, through three
APIs at O0/O2. Logical outputs compare both complete 96-byte buffers and canaries.
Actual SSE2/AVX2 is mandatory. This is not full SHA1/P256 execution or AVX512
hardware acceptance. Earlier byte-state, compare, GP/X and pointer receipts
remain scoped to their original revisions.

The unfiltered SDK command correctly exits nonzero for its 14 failures. Its
103-second run includes driver build and is not a controlled performance A/B.
Complete P256 retains all 14 TEXTs, with ten public definitions and four private
helpers in the observed CFG. Complete SHA1 retains both TEXTs and 48,115 bytes.
Neither file is cropped or substituted with an instruction-only fixture.

## Independent official rows

| Family group | Official rows | Concrete witnesses |
|---|---:|---:|
| Add/subtract and D/Q low product | 159 | 480 |
| Integer min/max | 132 | 432 |
| Average, products and multiply-add | 90 | 258 |
| Integer and PS/PD logical | 132 | 504 |
| Unpack | 108 | 360 |
| Saturating pack | 36 | 114 |
| Per-lane shifts | 64 | 236 |
| Conditional sign | 9 | 18 |
| Packed absolute value | 33 | 108 |
| Byte-indexed shuffle | 9 | 24 |
| Integer EQ/GT compare | 72 | 168 |
| Mask blend | 36 | 144 |
| Horizontal integer | 18 | 36 |
| FMA3 | 360 | 2,028 |
| AVX whole-vector moves | 144 | 306 |
| Ordinary AVX floating binary | 132 | 684 |
| Remaining EVEX binary families | 186 | 765 |
| Bit-count and conflict | 48 | 198 |
| Uniform and per-128-bit byte shifts | 188 | 488 |
| Complete rotate view (24 rows already in EVEX binary group) | 48 | 216 |
| Duplicate-element moves | 27 | 72 |
| Immediate word/dword shuffle | 34 | 95 |
| Packed/scalar reciprocal14 | 32 | 132 |
| Signed/zero widening | 108 | 288 |
| Scalar/block/GP/K-data broadcasts | 111 | 234 |
| Single-source narrowing | 108 | 324 |
| FP/integer numerical conversions | 162 | 876 |
| Immediate integer comparisons | 48 | 240 |
| VPTESTM/NM | 48 | 240 |
| GFNI multiply and affine/inverse | 24 | 84 |
| Leading-immediate vector group | 66 | 270 |
| Align and floating shuffle | 27 | 90 |
| Compact and expand | 72 | 216 |
| Five immediate FP control families | 80 | 390 |
| Local candidate unique union | **2,927** | **11,010** |

The local union contains 2,793 AVX rows of 3,568 (78.3%) and 134 legacy rows. The
complete rotate view overlaps 24 existing rows / 108 witnesses and is deduplicated.
The 11,010 witnesses include 102 terminal trap controls, not 11,010
normal executions. These joins retain independent ordered Go rows and actual Go bytes, then compile
initialized functions through three APIs with LLVM22. Two PADDD aliases are
separate. High banks force narrow EVEX; MMX has separate region/EMMS owners.
This is neither all x86 rows nor an architecture percentage. Context and zero
unsupported classifications never establish complete translation.

The ABS join adds only a unary format adapter and reuses the existing ABS/sign
initializer, original Go byte verifier and bounded three-API object driver.
Translator production code and native semantic fixtures are unchanged. The
follow-up `65ef2b9bf` only clarifies syntax metadata and strengthens its inventory
assertion: both endpoint SDKs pass its focused 4 roots / 43 children. All 865
previous rows / 2,774 witnesses and both PADDD aliases retain identical IDs,
instructions, bytes and axes on both SDKs. That follow-up also passes all 573
Node policies and the exact Go preflight's 26 roots / 299 children. These are
local object and existing portable-value receipts, not new original-native
execution or current-head CI acceptance.

FMA3 retains all 60 official members and every ordered row. Its 2,028 witnesses
include scalar N4/N8 tuples, 324 broadcasts, 720 register-only rounding controls
and 612 zeroing forms. Arithmetic always reads the old destination, even with
`.Z`. Existing 178 amd64 and 94 386 axis source/declaration pairs remain
byte-identical; both SDKs retain every previous 898-row / 2,882-witness field.
The shared format/encoding extension adds 39 net harness lines and the existing
initializer adds nine. No new translator rule or runtime matrix is introduced.
An ignored Go1.20 overlay snapshot probe hit a vet virtual-file defect; its
failed receipt is retained. Only that diagnostic comparison uses `-vet=off`;
the complete family and CI gates retain their original default vet behavior.

Move witnesses are **276 ordinary initialized forms and 30 terminal #UD
controls**; that join's intermediate union contained 6,635 ordinary forms and
30 faults. Each fault
function, through each of three APIs, must independently contain one UD2 and
unreachable with no input reads before LLVM object compilation. Both SDKs retain
90 per-function/API body receipts. This is not successful normal-value execution.
All 144 rows have original Go memory witnesses. A further 102 load-register
alternatives select earlier store rows; their selection mappings reference
already-counted regcopy bytes instead of inventing another row hit.

The original 188/64/30 move kernel directories retain all 45 observed source and
declaration hashes, and the previous 1,258 rows / 4,910 witnesses remain identical.
The initial 115-root run retained three real skips. A compile-only owner was
incorrectly using an execution selector; it now compiles all 90 aligned cases
without requiring Rosetta. A new frame object owner preserves the original four
target objects independently of host execution. The two remaining host-native
roots are explicitly outside the final local scope and retain their required
Linux runtime owners; their partial objects are not reported as runtime passes.

Ordinary floating ADD/SUB/MUL/DIV/MIN/MAX reuses the existing binary initializer
and row driver for all 24 PS/PD/SS/SD members. The 132 rows include 192 rounding,
24 SAE and 108 broadcast witnesses. Register-only SAE preserves EVEX vector
length; embedded rounding replaces those bits. Scalar mandatory prefixes do not
widen integer grammars. The shared harness grows by 18 lines; independent tests
retain actual Go byte anchors and missing-source/merge controls. All previous
1,402 rows / 5,216 witnesses and aliases remain identical on both SDKs. Original
F32/F64 MIN/MAX execution over NaNs, signed zero, MXCSR and faults remains an
explicit gap; successful objects do not close it.

The cleanup deletes sixteen legacy MUL/DIV/MIN/MAX bodies, four redundant
precision delegations and their dead allowlist entries after the typed owners.
Live typed owners and the shared FP16 portable path remain byte-identical.
The software-alias runtime owner remains required on Linux/amd64 and is outside
the final local scope. Its initial local skip is retained. A separate 26-line
result-routing reduction was withdrawn after minimum-SDK paired benchmarks
showed regression; no acceleration is claimed for the accepted cleanup.

## Remaining work, in order

1. Finish current Plan9asm and LLGo cloud rounds and repair actual failures.
   Verify the published LLGo fixture/tool repair on actual cloud runners.
   Finish combined Plan9asm publication gates, publish complete
   family batches, then require exact-head CI/coverage. Requalify the LLGo pin
   against that public source. Keep external review, the official dependency
   requirement and native AVX512 hardware acceptance open.
   Retain the digest optimization's full byte identity and mutable-source
   checks during combined integration; keep its encoder measurements separate
   from complete SDK translation and cloud job timing.
2. Complete amd64 official-row ownership, 386 applicability and independent
   semantics. Extend existing family/object/native owners instead of creating
   another opcode registry or enumerating unrelated source/target products.
   The remaining AVX inventory has 775 rows outside the local qualified
   union. Continue through remaining shared physical-format groups and full
   semantic families.
   Independently inventoried legacy machine-candidate rows are 1,306 on the
   current SDK and 1,305 on minimum, with 112 separate `ymovtab` fallback
   records. The 301 locally joined amd64 legacy rows are a separate denominator. Preserve
   mode refusals, exact duplicates and first-match/shadow mappings; static
   inventory is not missing implementation or initialized-object proof.
   GP RMW's 169-source-row closure is qualified locally. Next prepare complete
   scalar shift/rotate and double-shift formats, preserving count-dependent
   FLAGS, CL aliases, mode selection and first-match mappings.
   Six MMX rows retain their separate unresolved machine-state contract.
   Static inventories are not successful object or execution receipts.
3. Keep the 14 implicit-entry, GC-scan and system/runtime/test-hook failures
   individually diagnosed in [issue47](https://github.com/xgo-dev/plan9asm/issues/47).
   They remain failed integration work, not automatic exclusions. Object
   success does not prove runtime ABI.
4. Preserve other architecture gates while finishing amd64. The 271 ARM64
   decoder forms, remaining WASM control/SIMD and native AVX512 acceptance remain
   explicit gaps before claiming complete architecture coverage.
   ARM64-to-WASM portable IR also has an actual `llvm.returnaddress` object
   compilation failure. BF16 optimized dead-result active-memory fault behavior
   still needs an independent runtime owner; masked-load IR alone is insufficient.

External discovery, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run inventory
locally or restore those jobs here. Keep raw failed receipts under ignored
`_out/` with their original revision and outcome.
