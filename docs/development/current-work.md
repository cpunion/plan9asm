# PR40 closeout

Updated 2026-10-28. Finish **amd64 correctness, shared typed families and
translation efficiency first**, preserving all five architecture gates.
Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). Historical receipts remain in git history;
this page records the current candidate and its limits.

## Publication and review

PR40 remains draft. Publication `9fede619c` completed its
[full round](https://github.com/xgo-dev/plan9asm/actions/runs/40567452192)
with 236 successful and 89 failed checks. Repeated SDK failures and derived
completion failures are not separate instruction defects. The repair batch
keeps the original failed reports and requires new-head CI acceptance.
The preceding publication `01955e42e` passes its
[complete CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40536640591)
on attempt 1: **325/325 Go checks plus Codecov**, with **88.0%** main statement
coverage. Its complete named inventory matches the preceding 325-job round.
Host-specific skips remain recorded; native AVX512 acceptance and external
approval remain outstanding. Both inline threads are resolved. The preceding
`1c54c530e` round also passes; its original infrastructure failures are retained.

That publication adds ABS, FMA3, AVX moves and ordinary floating binary row
joins and removes 121 unreachable legacy floating lines. It preserves all SDKs
and test owners while creating current/minimum jobs first; creation order does
not guarantee external runner scheduling priority.

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

[LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694), head `74c767490`, is
rebased onto `854634f22` and pins the downloaded public
`v0.0.0-20261027190827-9fede619cf05`. Adapter tests pass 41 roots / 131 children,
assembly-build checks 21 / 23, and complete CI policies 160/160, with zero
failures/skips. Earlier rebase layout controls retain their original receipts. Current-head cloud
CI completes with 13 successful workflows and 130 successful jobs; only the two
conditional release/submission jobs are skipped. The exact public module ZIP,
all 5,037 included file blobs and the bundled license match the pinned source.
Earlier pins and their preparation timeout retain their original outcomes;
they do not qualify this pin. Its P1 dependency-provenance requirement remains
open until an official upstream version replaces the temporary fork pin.
Never push or merge upstream directly.

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
| Local candidate total | **1,720** | **6,665** |

These joins retain independent ordered Go rows and actual Go bytes, then compile
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
controls**, so the local total contains 6,635 ordinary forms and 30 faults. Each fault
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

1. Finish the five CI repair classes, publish through the allowed fork and
   finish new-head CI/coverage. Repin LLGo to that exact public revision and qualify it. Obtain
   external review and retain draft status meanwhile.
   Next local reductions are integer min/max reuse of the existing E4 pipeline
   and the next independent unary row join. Their static proposals are not passes.
2. Complete amd64 official-row ownership, 386 applicability and independent
   semantics. Extend existing family/object/native owners instead of creating
   another opcode registry or enumerating unrelated source/target products.
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
