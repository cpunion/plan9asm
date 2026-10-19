# PR 40: current checkpoint

Updated 2026-10-19. This is a handoff, not a completion certificate.
Earlier checkpoints and original failed fixtures remain in Git; ignored failed
reports must not be overwritten. Read [instruction contracts](instructions.md)
and [validation](validation.md) before changing a family.

## Scope and publication

PR40 owns translation, complete typed instruction families and independent
tests for 386, amd64, arm, arm64 and wasm. External inventories, scanners,
ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
No inventory scanner runs locally.

Push only to a cpunion/dastai fork. PR40 remains draft at `24851df2`;
staged branches do not qualify that contribution head. Latest upstream main
`e2164ac0` is an ancestor of the current candidate. Reuse existing persistent
idle worktrees; do not create another for routine work.

## Latest verified batches

| Code snapshot | Completed evidence | Limits |
|---|---|---|
| `4c6cc0964` | Complete ANDN/PDEP/PEXT repair: 36 roots / 2,699 children pass current Go, actual Go1.20 and race; identical 2,735 terminals, zero fail/skip/incomplete and unchanged source/tools | Six named/raw members share the complete operand grammar, distinct numerical/FLAGS effects and exact volatile reads; original values/aliases, five-target objects, unused-bool controls and twelve dead-read fault rows. Portable execution is not original BMI opcode execution |
| Same snapshot | Go1.20–Go1.27 audit changes exactly twelve isolated AMD64 probes per SDK from supported to Context; other families, 386 and encoder/unsupported metrics unchanged | Original probes lack their two data inputs. Initialized sequences and public objects remain independently tested; MULX's different effects are not registered by this grammar |
| `9673cc995` | Complete ADCX/ADOX input repair: 24 roots / 1,781 children pass current Go, actual Go1.20 and race; identical 1,805 terminals, zero fail/skip/incomplete, unchanged source and tools | Four named/raw members share one grammar, ordered input bytes and independent CF/OF chains; five-target objects, original portable values and exact dead volatile reads. Native original execution and complete CI remain separate |
| Same snapshot | Go1.20–Go1.27 classification audit changes exactly eight isolated AMD64 ADX probes per SDK from supported to Context; 386, other families and encoder/unsupported metrics unchanged | Those original probes lack numerical inputs and selected carry. Separate initialized whole-source probes and public objects establish successful translation |
| `5e5379ada` | Raw count/move-mask and MMX batch: 25 roots / 933 children pass current Go, actual Go1.20 and race; identical 958 terminals, zero fail/skip/incomplete | Nine scalar-count members and six move-mask names share existing effects; original Void/native-output objects, exact origin/byte controls and original count portable values pass. Four additional MMX runtime fixture families retain complete objects; native execution remains separate |
| `197521fb1` | BMI2/vector-zero/MMX batch: 33 roots / 1,386 children pass current Go, actual Go1.20 and race; identical 1,419 terminals, zero fail/skip/incomplete | Eight BMI2 members, two vector-zero members and original MMX fixture repairs; official audit and final gates remain separate |
| Same snapshot | Complete Go1.20–Go1.27 386/amd64 classification audit: zero changed forms or metrics, unchanged source/SDK/LLVM22 seals | Current Go still has zero unsupported forms; Go1.20's one pre-existing AMD64 unsupported form is unchanged. Not complete CI or benchmark qualification |
| `7cfe15222` | Complete product/accumulator repair: 28 roots / 665 children pass current Go, actual Go1.20 and race; identical 693 terminals, zero fail/skip/incomplete; seventeen focused gates pass | Four DWORD-product and six accumulator-extension members, named/raw input obligations, five-target three-API objects and actual-host integer values; original native runtime requirements remain separate |
| Same snapshot | Go1.20–Go1.27 audit changes exactly 34 AMD64 probes per SDK from supported to Context: 28 products and six implicit AX operations; 386 and all encoder metrics remain unchanged | Probe sources lack initialized inputs. Go1.20's pre-existing unsupported count is unchanged; neither these classifications nor the initialized objects certify complete CI |
| `259fbbcb8` | Cancellation-evidence RED/GREEN plus all 139 CI policy tests pass | Windows owners retain partial reports after cancellation; strict aggregation and missing-report errors are unchanged. This is a diagnostic fix, not a cancelled-test pass |
| `18bb77a09` | CI dispatch ordering and Go loader identity: complete runner package passes actual Go1.20 and current Go, each with 79 roots / 182 children; six focused race roots pass, unchanged source/tools | The earlier complete Go1.27 package timeout and failed recorder remain failed observations; real independent source/tool/library and loader execution checks remain mandatory |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 78 pass / 0 fail / 89 explicit host skips, 3,323 children, unchanged source/tools | No absent/incomplete or passing roots with skipped children; complete CI and the strict benchmark remain failed/unqualified |
| `7b97c31de` | x87 native floating moves and PUSH native reads: 69 roots / 4,988 children pass current Go, actual Go1.20 and race; identical 5,057 terminals, zero fail/skip/incomplete; all thirteen final gates pass | Complete four-member floating move and W/L/Q PUSH families, exact volatile widths, five-target objects and portable values/faults; no native FP80 or hardware exception claim |
| Same snapshot | Exact original 167-root replay: 78 pass / 0 fail / 89 explicit host skips; unchanged source/tools, zero absent/incomplete or passing roots with skipped children | Only the two original mode32 complete-source roots and x87 direct-compare object root changed fail to pass relative to `1cafc5b2f`; no regressions |
| `30a1c2679` | Mode32 logical AX returns and private FLAGS correlation: 45 roots / 2,500 children pass current Go, actual Go1.20 and race; identical 2,545 terminals, zero fail/skip/incomplete | Original two Windows386 complete-source failures pass; W/L, B/W/L subword, seven GP registers, clone/join and source/query boundaries retain independent controls |
| Same snapshot | All thirteen final gates pass, unchanged source/tools; seventy independent cancellation programs execute through three APIs, optimized/unoptimized IR and O0/O2 | No new complete 167-root replay or whole-CI claim; original Windows386 native execution remains separately required |
| `1cafc5b2f` | Raw FP16 binary and retained families: 53 roots / 3,157 children pass current Go, actual Go1.20 and race; identical 3,210 terminals, zero fail/skip/incomplete | All twelve raw members, 450 independent physical shapes on five targets through three APIs and IR/backend O0/O2; original opcode/control/mask bytes, not native execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 75 pass / 3 fail / 89 explicit host skips, unchanged source/tools | Original FP16 binary format root now passes with no regressions; complete CI and native FP16 hardware execution remain separate requirements |
| `4acf97d89` | Raw FP16 FMA and retained conversion/SQRT families: 112 roots / 5,475 children pass current Go, actual Go1.20 and race; identical 5,587 terminals, zero fail/skip/incomplete | All thirty raw FMA members, 1,134 independent physical shapes on five targets through three APIs and IR/backend O0/O2; exact opcode/control/mask bytes, not native execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 74 pass / 4 fail / 89 explicit host skips, unchanged source/tools | Original FP16 FMA portable runtime now passes with no regressions; native FP16 hardware execution and complete CI remain separate requirements |
| `bb85d27b3` | Raw FP16 SQRT and retained conversion/named-SQRT families: 82 roots / 1,934 children pass current Go, actual Go1.20 and race; identical 2,016 terminals, zero fail/skip/incomplete | Fifty independent masked FP16 shapes on five x86 targets through three APIs and IR/backend O0/O2; exact native opcode/control bytes and full 32-lane masks, not native execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 73 pass / 5 fail / 89 explicit host skips, unchanged source/tools | Original FP16 SQRT runtime root now passes, no regressions; portable X/Y/Z values, high mask bits and scalar passthrough remain distinct from hardware MXCSR/exception execution |
| `83b656042` | Six raw FP16 floating conversions: 36 roots / 400 children pass current Go, actual Go1.20 and race; identical 436 terminals, zero fail/skip/incomplete | All 148 independent physical forms on five x86 targets through three APIs and IR/backend O0/O2; object bytes retain actual width, mask and rounding controls, not native execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 72 pass / 6 fail / 89 explicit host skips, unchanged source/tools | Both original conversion runtime roots now pass with no regressions; portable numerical/guard-page execution remains separate from original native MXCSR and exception behavior |
| `ef3e886d3` | Complete static scalar RMW and native FS/GS input rules: 53 roots / 4,800 children pass current Go, actual Go1.20 and race; identical 4,853 terminals, zero fail/skip/incomplete | Eleven scalar stems, complete widths/relocations, 1,560 independent segment variants and original bare/indexed LOCK objects on five targets; not native segment execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 70 pass / 8 fail / 89 explicit host skips, unchanged source/tools | Original absolute-relocation root now passes, no regressions; original unbound sinks retain Go/emitter objects plus public Context, separate initialized DATA programs own public objects |
| `2c112ef6e` | Complete ENQCMD/ENQCMDS input, grammar and status effects: 51 roots / 1,658 children pass current Go, actual Go1.20 and race; identical 1,709 terminals, zero fail/skip/incomplete | All 1,200 independent physical address/register forms, pre-emission rejection and initialized three-API five-target O0/O2 objects; native device submission is not executed |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 69 pass / 9 fail / 89 explicit host skips, unchanged source/tools | Both original hardware-status roots now pass, without regressions. Uninitialized original formats retain Go/emitter objects and public Context; no PASID, privilege, memory-backing or successful-submission claim |
| `7b1966b70` | Complete RDRAND/RDSEED W/L/Q family: 44 roots / 1,610 children pass current Go, actual Go1.20 and race; identical 1,654 terminals, zero fail/skip/incomplete | All 128 independent physical register encodings, exact raw identity, partial-byte and missing-result controls; 52 initialized target programs retain three-API O0/O2 objects |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay remains 67 pass / 11 fail / 89 explicit host skips, unchanged source/tools | Ten original random/status children now pass; their parent still fails other families. Native randomness and complete CI remain unqualified |
| `8aa8460f4` | Complete floating-compare literal inputs: 43 roots / 481 children pass current Go, actual Go1.20 and race; identical 524 terminals, zero fail/skip/incomplete | All four PS/PD/SS/SD members, 54 independent VEX/EVEX memory shapes, all 256 selectors and 162 initialized three-API AMD64 target-object programs |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 67 pass / 11 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Original missing-vector/mask source remains Go/emitter objects plus public Context; full native comparison/MXCSR/exception emission is retained, not newly executed on this host |
| `37cd9d8a7` | Complete add/subtract literal inputs: 59 roots / 1,626 children pass current Go, actual Go1.20 and race; identical 1,685 terminals, zero fail/skip/incomplete | All sixteen wrapping/signed/unsigned saturating members, 244 independent VEX/EVEX programs and three-API AMD64 target objects; exact lane/merge/mask/tuple controls |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 66 pass / 12 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Original mixed source lacks K7 and remains Go/emitter objects plus public Context; initialized independent programs own public objects, not native execution |
| `175c84397` | Complete integer-compare literal inputs: 38 roots / 3,706 children pass current Go, actual Go1.20 and race; identical 3,744 terminals, zero fail/skip/incomplete | All eight EQ/GT B/W/D/Q members, 136 independent VEX/EVEX programs and three-API AMD64 target objects; portable values/FLAGS/canaries and host signal controls also execute |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 65 pass / 13 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Original uninitialized compare source retains Go/emitter objects and public Context; separate initialized programs own public objects, not a native AVX512 execution claim |
| `fd38dcf0e` | Complete byte-shift literal inputs: 33 roots / 291 children pass current Go, actual Go1.20 and race; identical 324 terminals, zero fail/skip/incomplete | Both directions, all X/Y/Z tuples and 256 selectors; 84 independent original programs have three-API AMD64 target objects, alongside existing portable shift execution |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 64 pass / 14 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Original byte-shift constant objects now translate unchanged, with no regressions; native AVX512 and complete CI remain separate |
| `f432afb24` | Bound logical literal inputs: 47 roots / 1,741 children pass current Go, actual Go1.20 and race; identical 1,788 terminals, zero fail/skip/incomplete | All 336 independent VEX/EVEX literal variants retain exact source, tuple and input-byte checks, with three-API objects on all three AMD64 targets |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 63 pass / 15 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Three original logical roots newly pass with no regressions; original missing-mask programs remain Go/emitter objects and public Context, not newly translated whole sources |
| `ed5d22b5b` | Physical MOVSXD/dword literal inputs: 35 roots / 820 children pass current Go, actual Go1.20 and race; identical 855 terminals, zero fail/skip/incomplete | All 52 independent register/prefix programs have three-API objects and portable zero/sign-extension values; temporary four-byte input facts do not escape |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 60 pass / 18 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Both original MOVSXD literal failures pass, with no regressions; not complete CI or native x86 execution |
| `9e3234ed2` | Bound RIP broadcasts and original raw-format contracts: 86 roots / 3,274 children pass current Go, actual Go1.20 and race; identical 3,360 terminals, zero fail/skip/incomplete | Complete 64-variant literal family, three-API objects and portable data values; missing mask/merge/register inputs remain independent Context controls |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 58 pass / 20 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Eleven former failures pass with no regressions; one oversized race process timed out, then the exact 36-root cohort passed in disjoint 7/10/19-root batches without increasing its five-minute budget |
| `9552a2759` | Architecture-correct bool-tail routing and stack-source contracts: 63 roots / 1,301 children pass current Go, actual Go1.20 and race; identical 1,364 terminals, zero fail/skip/incomplete | Complete 16-condition SETcc tail family, three-API objects and portable value execution; original invalid/unbound stack sources remain separate from initialized five-target objects and portable values |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 47 pass / 31 fail / 89 explicit host skips, unchanged source/tools and no incomplete roots | Three former failures pass, with no new failures; not complete current-head CI or native AMD64 execution |
| `3044f1483` | Decoder-bound VMOVW and raw-format input contracts: 31 roots / 1,006 children pass current Go, actual Go1.20 and race; identical 1,037 terminals, zero fail/skip/incomplete | All eight VMOVW direction/register/m16/W forms, five-target three-API IR/backend O0/O2 objects, exact byte/origin negatives and the unchanged portable guard-page/value program; not native AVX512 execution |
| Same snapshot | All thirteen final gates pass with unchanged source/SDK/LLVM22 observations; exact original 167-root replay has 44 pass / 34 fail / 89 explicit host skips | Five former failures pass with no regressions in that set; full current-head CI and strict standard-library benchmark remain unqualified |
| `6d1337fde` | Legacy logical-memory fixture contracts and complete related grammar/native-read tests: 21 roots / 3,785 children pass current Go, actual Go1.20 and race; identical 3,806 terminals, zero fail/skip/incomplete | Original 60-function segment formats remain Go/emitter objects and public Context; separate 36-function ordinary-address programs and initialized SSE1 pointer programs own public objects, not native execution |
| Same snapshot | All thirteen final gates pass, with unchanged source/SDK/LLVM22 observations | Test-only repair; complete original failure replay remains the older snapshot below until rerun |
| `0ab05d184` | ENDBR state and source-fixture contracts: 33 portable roots / 1,860 children pass current Go, actual Go1.20 and race with identical 1,893 terminals; current Go/race additionally pass two SDK-specific roots / three children | Exact named/raw numerical preservation, mandatory non-Void results and original format objects; native CET/GP/FLAGS execution remains required |
| Same snapshot | All thirteen final gates pass; exact original 167-root replay has 39 pass / 39 fail / 89 explicit host skips, no incomplete roots or passing roots with skipped children | Five previously failing roots now pass, with no regressions in that set; this is not complete CI or a whole-SDK source pass |
| `4024da958` | Complete eight-name SQRT contracts: 32 roots / 1,410 children pass current Go, actual Go1.20 and race; identical 1,442 terminals, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 seals | Includes 180 independent physical-form controls and 54 initialized raw variants through three APIs on five targets, with IR/backend O0/O2 objects; not native execution |
| Same snapshot | All thirteen final gates pass, including both-SDK build/vet and 24 scanner roots, five-arch classification, current/minimum-Go1.24 CLI checks, 132 CI policies and formatting | Eight-SDK audit changes only 42 isolated AMD64 SQRT forms per SDK to Context for missing inputs; 386 and encoder/unsupported metadata remain unchanged |
| `1d4e3ef70` | Wrapping-add format contracts: 9 roots / 1,124 children pass current Go, actual Go1.20 and race; identical 1,133 terminals, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 seals | Retains original unbound Go/emitter objects and public Context, with separate initialized arithmetic/native-read/MMX controls |
| Same snapshot | All eleven final gates pass, including both-SDK build/vet and 23 scanner roots, five-arch classification, CLI extraction and 132 CI policies | Test-only repair; not complete CI or native execution |
| `0c86e86ef` | Complete five-member SAD input family: 29 roots / 637 children pass current Go, actual Go1.20 and race; identical 666 terminals, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 seals | Includes original static-SB failures, all 256 selectors, independent encoder attributes, WIG decoding and five-target three-API O0/O2 objects; not native x86 execution |
| Same snapshot | All eleven final gates pass: both-SDK build/vet and 23 scanner roots, five-arch classification, five CLI extraction roots, 132 CI policies and formatting | Eight-SDK audit changes only 24 isolated AMD64 SAD forms per SDK to Context; 386 and encoding/unsupported metadata remain unchanged |
| `e809b3271` | Complete 32-member named/raw K-mask logical family: 27 roots / 892 children pass current Go, actual Go1.20 and race; identical 919 terminals, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 seals | Includes portable values/six FLAGS, five-target three-API objects and live private-stack/DF/ID guards; not native AVX512 execution |
| Same snapshot | All eleven final gates pass: both-SDK build/vet and 22 scanner roots, five-arch classification, five CLI extraction roots, 132 CI policies and formatting | Eight-SDK audit changes only 28 isolated AMD64 K-mask forms per SDK from supported to Context; encoding/unsupported metadata unchanged |
| `0a4a28bf5` | Omitted TEXT argument extent: 15 portable roots / 410 children pass current Go, actual Go1.20 and race with identical 425 terminal identities; current Go/race additionally pass the complete SDK `subtle/xorBytes` on three AMD64 targets | Actual Go objects and three-API LLVM22 IR/backend O0/O2 objects, not runtime or package linkage |
| Same snapshot | Both-SDK build/vet and scanner roots, five-arch classification, CLI extraction, 132 CI policies, tracked gofmt and diff checks pass | No coverage baseline changes; not complete CI |
| `c11c2130d` | COMI/UCOMI native-input continuation and status contracts: 21 roots / 911 children pass current Go, actual Go1.20 and race; identical 932 terminal identities, zero fail/skip/incomplete and unchanged source/SDK/LLVM22 seals | Three-API five-target O0/O2 objects, not native x86 execution |
| Same snapshot | Both-SDK build/vet and 21 scanner roots, five-arch classification, five CLI extraction roots, 132 CI policies, tracked gofmt and diff checks pass | No classification baseline changes; not complete CI |
| `96f91e566` | RTM/status plus complete WAITPKG/query/storage union: 49 roots / 2,029 children pass all three lanes, identical 2,078 terminals; the same final gates pass | Includes all 256 integer XABORT values; native RTM execution remains unqualified |
| `3db1fe799` | WAITPKG named/raw input and status effects: 41 roots / 1,965 children, identical 2,006 terminals pass all three lanes and final gates | Eight-SDK audit changes only isolated AMD64 UMWAIT/TPAUSE to Context for absent inputs; unsupported counts unchanged |
| `03ee09640` | Branch-token namespace: 70 roots / 384 children, identical 454 terminals pass all three lanes and final gates | Includes independent portable execution; Go's crashing local-label CALL is parser-only, never object/rejection evidence |

COMI/UCOMI lowering already retained the actual scalar native opcode, exact
4/8-byte memory width, MXCSR and exception effects. The eleven-line production
repair connects that emitter to the existing conditional native-read rule.
The complete eight-name/twelve-Go-row grammar, missing scalar bytes, strict
incoming addresses, state preservation and future-access negatives remain
independent. No declaration observer gains memory backing or numerical facts.

The original complete scalar-compare status program now compiles on all five
x86 targets without changing its source. Original uninitialized compare formats
remain actual-Go/public-Context/emitter objects; independent initialized programs
own public objects. An initial new fixture incorrectly returned full FLAGS with
unknown DF/ID; that exact program remains a negative, while a separate six-status
projection is the positive. Preserve that failed first-green report.

RTM lowering and XTEST effects share the existing four-name grammar. Only the
exact zero-operand query defines status FLAGS without numerical inputs.
Begin/end/abort gain no straight-line effect. XABORT accepts resolved integer
bytes, not floating tokens or mismatched dispatch. The combined hardware-status
owner still fails at ENQCMD; WAITPKG/XTEST qualification does not repair it.

Previous qualified code includes block-extraction native stores, GP/SIMD stores,
multiply-add/MMX format contracts, scalar-shift memory, CPU-query effects,
mask/vector data, logical floating-field addresses, private parameter copies,
integer self-comparisons, BZHI/BEXTR/BLS and inert NOP composition. Their original
native counterparts and failed complete-suite receipts remain separate.
See earlier checkpoints in Git rather than reconstructing counts by subtraction.

## Exact failure replay and online CI

The latest local diagnostic at `18bb77a09` selects exactly the original CI's
167 failed root names. One sealed Go test process ran the complete set,
instead of repeatedly linking the same large test executable.
All inventories, started/terminal identities and before/after source, SDK, LLVM
and environment observations match. This is not the canonical isolated CI runner.

| Observation | Pass | Fail | Explicit host skip | Absent / incomplete |
|---|---:|---:|---:|---:|
| Previous isolated replay, `c4c01538c` | 25 | 53 | 89 | 0 / 0 |
| Previous batched replay, `c11c2130d` | 30 | 48 | 89 | 0 / 0 |
| Previous batched replay, `1d4e3ef70` | 33 | 45 | 89 | 0 / 0 |
| Previous batched replay, `731de8a20` | 34 | 44 | 89 | 0 / 0 |
| Previous batched replay, `0ab05d184` | 39 | 39 | 89 | 0 / 0 |
| Previous batched replay, `3044f1483` | 44 | 34 | 89 | 0 / 0 |
| Previous batched replay, `9552a2759` | 47 | 31 | 89 | 0 / 0 |
| Previous batched replay, `9e3234ed2` | 58 | 20 | 89 | 0 / 0 |
| Previous batched replay, `ed5d22b5b` | 60 | 18 | 89 | 0 / 0 |
| Previous batched replay, `f432afb24` | 63 | 15 | 89 | 0 / 0 |
| Previous batched replay, `fd38dcf0e` | 64 | 14 | 89 | 0 / 0 |
| Previous batched replay, `175c84397` | 65 | 13 | 89 | 0 / 0 |
| Previous batched replay, `37cd9d8a7` | 66 | 12 | 89 | 0 / 0 |
| Previous batched replay, `8aa8460f4` | 67 | 11 | 89 | 0 / 0 |
| Previous batched replay, `7b1966b70` | 67 | 11 | 89 | 0 / 0 |
| Previous batched replay, `2c112ef6e` | 69 | 9 | 89 | 0 / 0 |
| Previous batched replay, `ef3e886d3` | 70 | 8 | 89 | 0 / 0 |
| Previous batched replay, `83b656042` | 72 | 6 | 89 | 0 / 0 |
| Previous batched replay, `bb85d27b3` | 73 | 5 | 89 | 0 / 0 |
| Previous batched replay, `4acf97d89` | 74 | 4 | 89 | 0 / 0 |
| Previous batched replay, `1cafc5b2f` | 75 | 3 | 89 | 0 / 0 |
| Previous batched replay, `7b97c31de` | 78 | 0 | 89 | 0 / 0 |
| Latest batched replay, `18bb77a09` | 78 | 0 | 89 | 0 / 0 |

The fork diagnostic `39192383617` finished naturally at `7b97c31de`: 63 jobs
pass, both ownership versions fail and complete Windows modules are blocked.
Its native x86, x87, ARM64 fallback, LOCK, unlocked RMW and value/fault priorities
pass. Both ownership versions expose the same
authenticated-read dispatch ordering regression. Commit `fb45f528a` restores
the SDK literal and numerical readers before every native family, retaining
the independent AST ordering guard rather than weakening it.

The subsequent diagnostic `39198807147` at `a1bd240cf` finished naturally:
78 jobs pass, 19 fail and one is cancelled. Both ownership versions,
all SDK priorities and all six native/cross priorities pass, but complete
Windows modules expose additional source/effect failures. Collected completed
artifacts report 88 failed roots; the cancelled CLI owner still lacks a complete
inventory and remains unknown, not passing evidence.
The CLI owner reached the job deadline and suppressed its artifact upload.
Cancellation-independent retention is repaired separately; original deadlines,
test outcomes and strict complete-module verification are not relaxed.

The current local candidate closes complete product and implicit-accumulator
input rules, preserves uninitialized format objects as public Context controls,
and adds initialized public objects. The six sign-extension names share one
grammar with lowering and exact GP effects; raw forms require their original
decoder identity even for Void. The portable integer oracle executes every bit,
preserved upper bytes and both encoding sources through three APIs and IR/backend
O0/O2. MMX comparison observers seed status outside a branch-free, uninterrupted
region and close it with real EMMS; original-Go build/link and native execution
remain distinct. CLI fallback FP words now require their original explicit TEXT
extent, without rejecting independently proved pure tail forwarders.

The complete actual-Go1.20 runner regression then exposed a separate test
contract error: external linking can change Mach-O UUIDs and the binary-content
part of Go's four-part build ID even when all build inputs are identical.
An independent minimal cgo program reproduces this without plan9asm or LLVM.
Commit `18bb77a09` validates all four hashes and compares the link action,
main-package action and main-package content. Both scopes still build and
execute their own binary, verify source ownership/bytes, and independently
observe the actual loaded LLVM library. Mutations of each input component and
malformed IDs remain negative controls; no binary or observation is reused.
At that snapshot both SDKs pass the complete 79-root / 182-child runner package,
and the six affected race roots pass with unchanged source/tools.

Preserve the earlier five-minute unsplit package timeout and the first local
recorder's incorrect `pass`/`passed` label failure. The latter's actual complete
Go1.20 command passed; a fresh reader checks its raw complete report and frozen
source/tools before continuing the other lane. Neither failed controller is
rewritten as a successful receipt. Original command budgets remain unchanged.

Raw FP16 binary arithmetic now shares complete typed input/control grammar and
the existing native emitter. Source bytes, masks and merge inputs remain required
even for Void. Scalar upper bytes come from the second source; zeroing removes
only the destination merge dependency. All 450 independent physical forms retain
native object controls. The separate 84-function portable oracle covers values,
NaN/signed-zero MIN/MAX, high mask bits, inactive access and upper aliases.
Original unbound formats retain actual Go/emitter objects and public Context;
their corrected expectations are not new whole-source translation successes.

Raw FP16 FMA reuses the complete typed FMA family and a shared native
three-operand emitter. All three inputs, including old destination under .Z,
remain required. Scalar upper bytes retain old destination knowledge; packed
Z masks retain all 32 bits. Original uninitialized formats keep actual Go and
emitter objects plus public Context; separate initialized sources own public
objects. The original portable FMA numerical/guard-page program now passes.
Preserve its real RED, retained-format failure and the first qualification's
incomplete inventory observation. That infrastructure attempt is not a test
pass; the fresh complete three-lane qualification is independently sealed.

Raw FP16 SQRT reuses the named family's typed operand and byte-effect model,
with exact decoder identity and no new caller or standalone-query permission.
Matching x86 emission retains the original native opcode; the shared scalar
FP16 emitter also serves conversions. Packed half-precision Z forms load all
32 mask bits rather than truncating to sixteen. Original missing-input formats
remain Go/emitter objects and public Context. Independently initialized public
programs own object coverage; portable numerical execution owns a separate axis.
The available scanner server has no AVX512-FP16 hardware, so neither result is
reported as native FP16 execution. Retain the failed RED and first-green format
expectations, including the unchanged scalar-binary load contract.

Raw FP16 floating conversions share typed carrier/active-byte, mask, scalar
passthrough and ER/SAE grammar. Exact original instruction identity and all
required inputs remain mandatory, including Void. Matching x86 targets retain
the actual side-effecting native opcode; the existing cross-target numerical
model is explicitly tested separately. Independent object bytes check all148
physical forms, not merely successful compilation. Dispatch conflicts and
invalid addresses fail before emission. Preserve the original uninitialized
format programs as Go/emitter objects plus public Context, and the original
scalar rounding program with uninitialized passthrough as a negative. Separate
initialized programs own public objects and portable numerical execution.

Static RMW inputs now consume actual same-module DATA identity, complete byte
extent and writable destinations. Readonly address targets remain distinct from
mutable sinks. Named FS/GS operands share existing normalization across scalar
and bit-string families, requiring every explicit GP offset and data input.
Bare segment addressing retains its original conditional native access, not
an initialized memory or declaration-query fact. The full regression found an
inherited LOCK-segment failure and an obsolete blanket segment-rejection test;
both failed receipts remain. Complete bare/indexed LOCK decoding and unlocked
BT/BTC/BTR/BTS volatile object checks now pass. Runtime segment behavior and
complete CI remain separate, unqualified requirements.

The hardware-random family defines only its exact output bytes and six status
flags; CF still independently reports availability. Its shared grammar rejects
malformed dispatch before emission, and source identity preserves unrelated
inputs, DF/ID and private storage. Unknown upper bytes, physical frames and
missing ABI results remain Context. All ten original random/status children
pass at that earlier snapshot, without changing any root outcome. The later
enqueue batch repairs both combined hardware-status roots; neither has regressed.

The original floating-compare literal root newly passes without regressions.
Its unchanged source lacks vector/mask inputs and retains complete Go/emitter
objects plus public Context. All four VCMPPS/PD/SS/SD members have 162 separate
initialized public-object programs; 54 independent memory shapes check every
predicate byte. Exact scalar/full/broadcast reads, write-only K results, source
lanes, masks and wider aliases remain independent. The existing native opcode,
MXCSR and exception effects are unchanged. The first-green state oracle wrongly
used a CFG clone that intentionally revokes transient DATA-read scope; the
corrected oracle compares the instruction-local state without that clone.
Preserve that failed receipt alongside the original production RED.

The original packed-arithmetic literal root newly passes without regressions.
Its unchanged mixed source lacks K7 and retains complete Go/emitter objects
plus public Context. All sixteen wrapping and signed/unsigned saturating add/sub
members have 244 separate initialized public-object programs. The production
adapter consumes existing typed forms and lane effects, not another opcode list
or native-read permission. Missing source lanes, merge bytes, actual masks and
full/broadcast DATA extents remain distinct. Logical and arithmetic object tests
share one bounded batch helper without changing either test inventory.

The original integer-compare literal root newly passes without regressions.
Its unchanged uninitialized vector source remains actual-Go/emitter objects
and public Context; 136 independent initialized programs own public objects.
The complete EQ/GT B/W/D/Q family reuses its typed byte effect, exact literal
tuple and decoder identity. Missing vector/mask inputs, changed DATA and source
frame/query boundaries remain checked. Two architecture-independent C signal
controls now execute on the actual host instead of borrowing an x86-only
launcher. This repairs their test routing, not native instruction execution.

The original byte-shift constant root now passes unchanged. Exact decoder-bound
EVEX memory tuples reach the existing typed shift effect; saturating counts
still require the entire memory tuple. Full raw-source, DATA and return checks
remain mandatory. No native-read, frame, declaration-query or mode32 permission
is added. All 1,536 direction/width/selector combinations retain exact extents,
destination aliases and FLAGS, alongside 84 independent target-object programs.

Three original logical-literal roots newly pass. The complete VEX/EVEX family
reuses typed logical byte effects through exact decoder-bound DATA inputs.
Actual masks and merge bytes remain required; no native-read or frame role is
added. The original valid VEX float program now translates unchanged. Original
EVEX/mixed programs lacking K7 retain complete Go/emitter objects and public
Context; separate initialized programs own public objects. Per-byte negatives
cover all 336 variants, and the independent EVEX register decoder checks
2,359,296 encodings in its own bounded cohort.

Both original physical MOVSXD literal roots newly pass. Opcode 63's default
dword, REX.W and 66+REX.W forms plus ordinary 8B dword loads share one exact
four-byte input role. Its transfer-local availability fact cannot survive the
instruction or supply pointer/storage permission. Actual DATA and complete
raw-source/return identity remain independently required. All 52 independent
programs retain original Go objects, three-API LLVM objects and portable values.

The preceding eleven repaired roots remain passing: four RIP-broadcast roots and seven raw-format
roots. Bound broadcast inputs now reach the existing typed byte effects without
inventing masks, merge bytes, native reads or entry contracts. The independent
whole-source raw-return/DATA checks still run. Unbound original encodings retain
Go/emitter objects and public Context; separate initialized programs own public
objects. A corrected Context expectation is not a new whole-source translation
success. The original valid dead mode32 XOR program still translates, rather
than being incorrectly reclassified as Context. Portable broadcast data values
now execute on the actual host instead of borrowing an x86-only test launcher.
Earlier bool-tail, stack, ENDBR, SQRT, SAD and wrapping-add repairs remain passing.
No root passes with skipped children. Host skips remain unexecuted
native requirements, never successes. Different local scheduling is disclosed;
canonical all-platform qualification is still required.

Latest fork diagnostic
[39065963275](https://github.com/cpunion/plan9asm/actions/runs/39065963275)
finished naturally at `06eedf847`: **81 jobs pass / 17 fail**. All sixteen
Windows root parts and their completion fail. Its audited raw root union is
6,998 roots: 6,568 pass, 167 fail, 263 explicit host skips, zero incomplete.
The same head has 73 native roots / 228 children passing with no skips.
None of these receipts qualifies a later staged source or the PR40 head.

## Strict standard-library benchmark

The last complete run at `18bb77a09` failed with unchanged source/tools.
All 184 original target/file pairs were attempted: 129 compile, 55 fail,
zero N/A or unattempted. Compilation is not runtime execution.

| Target | Compiled / total | Failed | Wall seconds |
|---|---:|---:|---:|
| linux/386 | 14 / 33 | 19 | 94 |
| linux/amd64 | 29 / 55 | 26 | 392 |
| linux/arm | 27 / 30 | 3 | 3 |
| linux/arm64 | 40 / 47 | 7 | 12 |
| js/wasm | 19 / 19 | 0 | 3 |

Total 504s meets 900s, but AMD64 exceeds its unchanged 300s target budget.
Compared with `bc79fd0f`, three additional 386 failures expose unbound default
scalar returns or native entry context in checktest's static TEXT, abi_test's
test-only undeclared TEXT, and the Linux bootstrap entry. Do not restore false
success by borrowing initialized register allocas or assuming a native ABI.
The measured AMD64 time was previously 709s; the different snapshots/workloads
are not a controlled performance comparison or proof of a specific speedup.
A diagnostic sample shows substantial repeated SHA256/file validation; measure
before changing implementation or claiming a speedup. Do not relax budgets.

## Immediate work

The BMI2/vector-zero/MMX code is qualified at `197521fb1`; its thirty-three
roots retain original portable executions and exact input/origin controls.
Initial dead-Void rejection assertions conflicted with the existing closed
nontrapping GP contract and remain failed observations, not implementation
defects. An overlay of the actual earlier native-read guard independently
reproduces the original 72-function failure through all three APIs.
The next batch at `5e5379ada` has its own complete twenty-five-root receipt,
not the earlier thirty-three-root certificate. Scalar-count and move-mask
register-return matrices already passed before this repair; the actual defect
was the original Void/native-output continuation. Preserve the separate failed
Go-object fixture that omitted textflag.h: it is not source rejection or proof
of the translator defect. With that header repaired, an old-guard overlay
reproduces both original translation failures on all three AMD64 targets.
The next ADX batch is qualified at `9673cc995` with its own twenty-four-root
three-lane union. Exact raw default address size is captured independently and
bound before a local numerical projection; overrides and physical segments
retain their separate Context/emitter contracts. Both ordered inputs and the
selected CF/OF chain remain required, including dead native outputs. Preserve
the original thirteen-terminal RED and the failed first GREEN whose raw guard
was inserted in the unrelated MMX loop. Original RORX physical-BP and partial
controller failures remain failed evidence as well.
The following six-member ANDN/PDEP/PEXT batch is qualified at `4c6cc0964`.
Its complete Go rows bind zcase, offsets, prefix/opcode attributes and physical
mode widths. Shared load emission also retains ADX's independent carry rules.
ANDN kills undefined AF/PF knowledge; PDEP/PEXT preserve all source FLAGS.
Ordinary named-register/logical-storage effects stay in the strict transfer;
only exact raw identity or a paired native read enters the special adapter.
Preserve the original fourteen-terminal RED, first-green unbound-format
failures and the separate raw-adapter/mode32-format failures. Their correction
is not a claim that uninitialized original programs gained public translation.

1. Continue the 88-root inventory from completed `39198807147` artifacts;
   the cancelled CLI owner's missing inventory remains unknown. Prioritize
   whole rule classes, including raw numerical continuations,
   missing typed effects and invalid format-fixture assumptions. Do not use the
   older 167-root replay as the denominator for these newly reached modules.
2. Run the final gates and batch coherent qualified repairs in the allowed
   fork. Each source snapshot retains its own three-lane and classification
   receipts; the cancellation-retention change has 139 passing policy tests.
   The PR40 contribution and completed diagnostic still have older heads.
3. Preserve the separately qualified logical integer return batch `53a93cfd4`
   and the incomplete floating MIN/MAX candidate `7009ff18f` on the prior
   source-fixture branch. Neither is merged into this CI-repair candidate.
   MIN/MAX has a first numerical/object GREEN, not complete format/raw/native
   qualification. Return to it after closing current CI failures; it addresses
   the original `math/hypot_amd64.s` MAXSD benchmark failure, not every ABI gap.
4. Rerun the unchanged complete standard-library corpus and benchmark after
   relevant source repairs. Do not convert Context into successful compilation.
   A valid encoding grants no caller frame, memory ownership or ABI transport.
5. Run complete current-head modules, native/cross-runtime CI, coverage and
   fresh review before ready. Keep known failures visible and batch pushes;
   do not launch another broad diagnostic merely to rediscover known failures.

## Reproducing and locating evidence

- `windows-raw-scalar-mmx-qualified-5e5379ad.json`: complete twenty-five-root
  three-lane union, including original raw count execution and all typed input,
  native-address, MMX and retained vector-zero/BMI2 controls.
  `raw-scalar-original-void-repaired-header-red/` restores the actual earlier
  native-read guard and retains six original text-API translation failures,
  separate from the initial missing-header Go-object failure.
- `windows-bmi2-vector-mmx-qualified-197521fb.json` and
  `bmi2-classification-audit/`: complete thirty-three-root three-lane union and
  eight-SDK zero-delta audit. `bmi2-original-native-read-red/` independently
  replays the original complete numerical/FLAGS fixture against the old guard;
  raw JSON retains all three API failures. Initial false dead-input assertions,
  the physical-BP format failure and invalid-phase controller remain failed
  evidence rather than overwritten qualification attempts.
Use [validation commands](validation.md). Freeze tracked sources and tools,
including docs, throughout a run. One heavy local Go/LLVM owner at a time;
develop in an existing idle worktree. Commit and preserve RED before repair.

Resolve worktree locations with `git worktree list`. Ignored receipts include:

- `windows-ternary-gp-qualified-4c6cc096.json`: complete thirty-six-root
  three-lane union. `ternary-gp-classification-audit/` is in the existing
  typed-string worktree and retains all sixteen sealed SDK reports.
  RED `d7f841a4`, first GREEN `e3ff46bf` and native/origin/width attempt
  `548b5f7c` remain separate failed receipts, not overwritten qualifications.
- `windows-adx-qualified-9673cc99.json` and `adx-classification-audit/`:
  complete twenty-four-root three-lane union and sealed sixteen-report audit.
  `compact-325c9759-adx-original-and-family-red-*` retains the original source
  and initialized-family failures; `e776df82` retains the failed first GREEN.
  Do not count those failed attempts or native host exclusions as passes.
- `windows-product-inputs-qualified-7cfe1522.json`: complete 28-root / 665-child
  three-lane qualification and identical 693-terminal unions. The earlier
  16-root / 546-child product-only qualification remains at `1f6694616`.
- `windows-product-final-gates-v2-7cfe1522/`: seventeen focused gates. Preserve
  the earlier missing-script/syntax attempts, `3bd2aae3` extraction failure and
  `256325d2` official-classification failure separately, not as successful gates.
- `product-accumulator-classification-audit/`: sealed sixteen-report audit
  across eight SDKs, matching prior baseline reports and unchanged tool/source
  observations before the reviewed 24-line baseline update.
- `ci391988/partial-failure-audit-*.json`: diagnostic-only downloaded Windows
  outcomes; cancellation/missing owners cannot certify a full module union.
- `cancelled-evidence-69ccb097/`: original failed retention policy and the
  complete 139-policy GREEN. Native execution and cancelled tests remain separate.
- `ci-dispatch-loader-18bb77a0/go120/`,
  `ci-dispatch-loader-continuation-18bb77a0/`: actual complete SDK runner
  reports and focused race outcomes, with preserved failed-controller evidence.
- `ci-dispatch-final-18bb77a0/`: thirteen final gates, exact original failure
  replay and complete failed five-target benchmark, with unchanged source.
- `x87-native-input-qualified-7b97c31d.json` and
  `x87-native-input-final-gates-7b97c31d/`: complete three-lane 69-root cohort
  and thirteen gates. The compact receipts retain exact started/terminal unions
  and before/after source, SDK and LLVM22 equality.
- `latest-failure-batch-7b97c31d-{summary.json,raw.jsonl,before.json,after.json}`:
  exact original 167 roots, 78 pass / 0 fail / 89 explicit host skips;
  unchanged source/tools and no incomplete roots.
- `mode32-flags-qualified-30a1c267.json` and
  `mode32-flags-final-gates-30a1c267/`: complete 45-root three-lane qualification
  and thirteen gates, with unchanged source/tools. Preserve logical-return RED
  `88602041`, FLAGS RED `1dbeae3a`, and failed first GREEN `e5610822`; the latter
  exposed the separate AUTO emitter's stale private-SP rejection.
- `compact-04ee565a-x87-floating-push-native-input-red-go127-normal-*`:
  original and independent native-input RED, with no infrastructure failure.
  These compact runner receipts live in the shared ignored receipt directory.
- `fp16-binary-qualified-1cafc5b2.json` and `fp16-binary-final-gates-1cafc5b2/`:
  complete 53-root three-lane cohort, identical 3,210-terminal unions and all
  thirteen gates. Preserve RED `a15d65d5`, first-green `32c1f749`, and the failed
  `56d63ca2` qualification (one original embedded-control format contract).
- `latest-failure-batch-1cafc5b2-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 75 pass / 3 fail / 89 explicit host skips; unchanged
  source/tools, no absent/incomplete or pass-with-skip roots. Only the original
  FP16 binary format root changed from fail to pass relative to `4acf97d89`.
- `ci-38236001451-race-root-0-original/` and
  `pr40-race-root-0-audit-20261019.json`: old online race failure at `24851df2`.
  The reentry fixture timed the whole Verify call, not the actual callback lock.
  Staged commit `7bdc112a8` checks that lock directly; this artifact audit is not
  a current-head race pass. SDK globals/AES binding already exists; inspect
  current receipts before reimplementing a failure from the old online head.
- `fp16-fma-qualified-4acf97d8.json` and `fp16-fma-final-gates-4acf97d8/`:
  complete 112-root three-lane cohort, identical 5,587-terminal unions and
  thirteen final gates. Preserve RED `57792901`, retained-format failure
  `9a8fc945` and the original incomplete inventory attempt separately.
- `latest-failure-batch-4acf97d8-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 74 pass / 4 fail / 89 explicit host skips; unchanged
  source/tools, no absent/incomplete or pass-with-skip roots.
- `fp16-sqrt-qualified-bb85d27b.json` and `fp16-sqrt-final-gates-bb85d27b/`:
  complete 82-root three-lane cohort, identical 2,016-terminal unions and
  thirteen final gates. Preserve RED `7741a3f8`, retained-format failure
  `344c1d7f` and scalar-binary expectation failure `1a118b6b` separately.
- `latest-failure-batch-bb85d27b-{summary.json,raw.jsonl,before.json,after.json}`:
  exact167-root replay, 73 pass / 5 fail / 89 explicit host skips; unchanged
  source/tools, no absent/incomplete or pass-with-skip roots.
- `fp16-conversion-qualified-83b65604.json` and
  `fp16-conversion-final-gates-83b65604/`: complete 36-root three-lane cohort,
  identical 436-terminal unions and thirteen final gates. Preserve input RED
  `3fa71bdb`, failed original-format contracts at `6852b0be`, and malformed
  dispatch/address RED `222a22768` separately.
- `latest-failure-batch-83b65604-{summary.json,raw.jsonl,before.json,after.json}`:
  exact167-root replay, 72 pass / 6 fail / 89 explicit host skips; unchanged
  source/tools, no absent/incomplete or pass-with-skip roots.
- `static-rmw-qualified-ef3e886d.json` and `static-rmw-final-gates-ef3e886d/`:
  complete 53-root three-lane cohort, identical terminal unions and thirteen
  final gates. Preserve static RED `4824e167`, inherited LOCK failure `e29ef390`,
  segment RED `c6e50807` and old bit-test expectation failure `fd9b906f`.
- `latest-failure-batch-ef3e886d-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 70 pass / 8 fail / 89 explicit host skips, 101.871s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `enqueue-qualified-2c112ef6.json` and `enqueue-final-gates-2c112ef6/`:
  complete 51-root three-lane cohort, identical terminal unions and thirteen
  final gates. Preserve RED `b9a185b5`, first GREEN `851b354f` (format-helper
  failure) and independent address-boundary RED `7395203c` separately.
- `latest-failure-batch-2c112ef6-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 69 pass / 9 fail / 89 explicit host skips, 157.926s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `hardware-random-qualified-7b1966b7.json` and
  `hardware-random-final-gates-7b1966b7/`: complete 44-root three-lane cohort,
  identical terminal unions and thirteen final gates. Preserve RED `54d2dd8d`
  and its separate first-green receipts.
- `latest-failure-batch-7b1966b7-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 67 pass / 11 fail / 89 explicit host skips, 109.125s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `raw-arithmetic-qualified-37cd9d8a.json` and
  `raw-arithmetic-final-gates-37cd9d8a/`: complete 59-root three-lane cohort,
  identical terminal unions and thirteen final gates. Preserve RED `23353639`
  (21 failed terminals) and the separate first-green receipt.
- `latest-failure-batch-37cd9d8a-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 66 pass / 12 fail / 89 explicit host skips, 101.354s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `raw-integer-compare-qualified-175c8439.json` and
  `raw-integer-compare-final-gates-175c8439/`: complete 38-root three-lane cohort,
  identical terminal unions and thirteen final gates. Preserve RED `b4e4df51`
  and the failed `765bf1e5` qualification attempts: one invalid script phase
  and two host-skipped signal controls, neither a complete qualification.
- `latest-failure-batch-175c8439-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 65 pass / 13 fail / 89 explicit host skips, 87.714s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `raw-byte-shift-qualified-fd38dcf0.json` and
  `raw-byte-shift-final-gates-fd38dcf0/`: complete 33-root three-lane cohort,
  identical terminal unions and thirteen final gates. Preserve original RED
  `983d16d2` and both failed selector fixtures (`0bffc04e`, `7c62d609`): the
  latter confused the naked decoder with normalized source including TEXT.
- `latest-failure-batch-fd38dcf0-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 64 pass / 14 fail / 89 explicit host skips, 144.318s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `raw-logical-qualified-f432afb2.json` and
  `raw-logical-final-gates-f432afb2/`: complete 47-root three-lane cohort,
  identical terminal unions and thirteen final gates. Preserve the first
  `ce298813` RED and corrected `c7b616b9` RED separately.
- `latest-failure-batch-f432afb2-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 63 pass / 15 fail / 89 explicit host skips, 146.864s;
  unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `movsxd-literal-qualified-ed5d22b5.json` and
  `movsxd-literal-final-gates-ed5d22b5/`: complete 35-root three-lane cohort,
  identical terminal identities and thirteen final gates. Preserve the original
  `b52c5426` RED and separate first-green receipts.
- `latest-failure-batch-ed5d22b5-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 60 pass / 18 fail / 89 explicit host skips, 208.505s
  test time; unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `raw-broadcast-qualified-9e3234ed.json`,
  `raw-format-contract-qualified-9e3234ed.json` and
  `raw-broadcast-final-gates-9e3234ed/`: disjoint 50/36-root cohorts, identical
  three-lane terminal unions and all thirteen final gates. Preserve the first
  raw-format failures, literal RED, missing-NOSPLIT oracle failure, earlier
  portable-runtime skip and unpartitioned race timeout as failed receipts.
- `latest-failure-batch-9e3234ed-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 58 pass / 20 fail / 89 explicit host skips, 145.433s
  test time; unchanged source/tools, no absent/incomplete or pass-with-skip roots.
- `stack-tail-routing-qualified-9552a275.json`,
  `stack-tail-final-gates-9552a275/`: 63-root three-lane qualification and
  thirteen final gates. Preserve first RED `f3ef04e5` (including fixture
  mistakes), corrected RED `baffdeac` and first GREEN `9552a275` separately.
- `latest-failure-batch-9552a275-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 47 pass / 31 fail / 89 explicit host skips, unchanged
  source/tools and no missing/incomplete roots or passing roots with skips.
- `logical-memory-family-qualified-6d1337fd.json` and
  `logical-memory-final-gates-6d1337fd/`: complete 21-root, three-lane cohort
  and thirteen final gates, retaining both original failures and real inputs.
- `raw-vmovw-family-qualified-3044f148.json`,
  `raw-vmovw-final-gates-3044f148/`: complete 31-root three-lane cohort and
  thirteen gates. Preserve RED `98859951` and failed first GREEN `9effb942`;
  the latter exposed an incorrect GP-only low-byte reader for X registers.
- `latest-failure-batch-3044f148-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, 44 pass / 34 fail / 89 explicit host skips, unchanged
  source/tools, no incomplete roots or passing roots with skipped children.
- `endbranch-family-qualified-0ab05d18.json`,
  `endbranch-final-gates-0ab05d18/`: three complete lanes and thirteen final
  gates with unchanged source/tool seals. Named/raw state RED and the separate
  missing-result RED remain failed receipts, not overwritten first greens.
- `latest-failure-batch-0ab05d18-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay with five repaired roots and no new failures.
- `wrapping-add-format-qualified-1d4e3ef7.json`,
  `wrapping-add-final-gates-1d4e3ef7/`: complete three-lane and final gates.
- `latest-failure-batch-1d4e3ef7-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay; all original root identities and source/tool seals match.

- `sqrt-family-qualified-4024da95.json`, `sqrt-final-gates-4024da95/` and
  `sqrt-classification-audit/`: complete three-lane/final gates and eight-SDK
  deltas. The failed `sqrt-final-gates-731de8a2/` remains a failed gate run.
- `latest-failure-batch-731de8a2-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root replay, including the unchanged original ABI0 SQRT success.

- `sad-qualified-0c86e86e.json`, `sad-final-gates-0c86e86e/` and
  `sad-classification-audit/`: complete three-lane/final gates and eight-SDK
  deltas. RED `bd7a7c3e5` retains 70 failed terminals. `c2487d2fb` retains
  WIG and unbound-format failures. `7df27388d` retains the same-source zero
  failure and two host-inapplicable runtime skips; that run is not qualified.
  The portable 29-root cohort excludes those two unexecuted native requirements.
- `text-extent-qualified-0a4a28bf.json` and `text-extent-final-gates-0a4a28bf/`:
  completed three-lane/final gates. Current Go has 16 roots / 413 children;
  actual Go1.20 has 15 / 410 because the original SDK fixture is Go1.27-only.
  RED `b92d370d2` retains ten failed terminals, including all three original
  `xorBytes` objects failing at TEXT extent before the production repair.
- `mask-logical-qualified-e809b327.json`, `mask-logical-final-gates-e809b327/`
  and `mask-logical-classification-audit/`: complete three-lane/final gates and
  eight-SDK form deltas. Named RED `1d5be19d4` and raw RED `e0c058745` retain
  142 and 50 failed terminals respectively; first greens are not full CI.
- `comi-native-input-qualified-c11c2130.json` and
  `comi-native-input-final-gates-c11c2130/`: complete three-lane/final-gate seals.
- `latest-failure-batch-c11c2130-{summary.json,raw.jsonl,before.json,after.json}`:
  exact 167-root diagnostic. `launch-local-failure-batch.cjs` retains explicit
  skips separately and fails if any root fails. It is not canonical CI.
- COMI RED `e8d86465b` has 27 failed terminals; `3c5e3891d` first green retains
  the new full-FLAGS fixture failure. `c11c2130d` preserves it as a negative.
- `rtm-status-qualified-96f91e56.json`, `rtm-status-final-gates-96f91e56/`;
  RTM REDs `c36a0f8f9` and `696c83919` retain 16 and five failed terminals.
- `waitpkg-qualified-3db1fe79.json`, `waitpkg-final-gates-3db1fe79/`,
  `waitpkg-classification-audit-v2/`; named/raw REDs remain at
  `76db9084a` and `fd6a63e56`. The first audit-path and intermediate compile
  failures remain failures, not completed audits.
- `latest-failures-c4c01538/raw-outcomes-audit.json`: previous isolated replay;
  `rerun-latest-ci-failures.cjs` and `inspect-previous-failures.cjs` retain its
  complete owner/source/tool and raw-output audit.
- `ci-39065963275/root-union-audit.json` and its Windows-root artifacts:
  original online denominator. Artifacts remain available from the linked run.
- `compile-benchmark-bc79fd0f/`, `benchmark-bc79fd0f/`: strict failed benchmark
  and diagnostic sample, in the frozen NOP worktree.

Receipts are ignored evidence, not tracked repository data. Their historical
successes never certify modified sources, different toolchains or execution
that has not happened.
