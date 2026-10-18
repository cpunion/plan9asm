# PR 40: current checkpoint

Updated 2026-10-18. This is a handoff, not a completion certificate.
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

The latest local diagnostic at `8aa8460f4` selects exactly the original CI's
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
| Latest batched replay, `8aa8460f4` | 67 | 11 | 89 | 0 / 0 |

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

The last complete run at `bc79fd0f` failed with unchanged source/tools.
All 184 original target/file pairs were attempted: 132 compile, 52 fail,
zero N/A or unattempted. Compilation is not runtime execution.

| Target | Compiled / total | Failed | Wall seconds |
|---|---:|---:|---:|
| linux/386 | 17 / 33 | 16 | 73 |
| linux/amd64 | 29 / 55 | 26 | 709 |
| linux/arm | 27 / 30 | 3 | 5 |
| linux/arm64 | 40 / 47 | 7 | 12 |
| js/wasm | 19 / 19 | 0 | 5 |

Total 804s meets 900s, but AMD64 exceeds its unchanged 300s target budget.
A diagnostic sample shows substantial repeated SHA256/file validation; measure
before changing implementation or claiming a speedup. Do not relax budgets.

## Immediate work

1. ENDBR's three-lane cohort and final gates are complete at `0ab05d184`.
   Preserve its real side-effecting CET instruction, exact raw-source identity
   and missing-input/result negatives. The required Linux native wrapper has
   not run locally. SQRT's native rounding/passthrough/mask/MXCSR requirement
   likewise remains separate from its completed object-only qualification.
2. The logical-memory and VMOVW/raw-format batches are qualified and included
   in the latest exact replay. Repair the other
   observed classes:
   raw FP16 native-read composition; raw literal and unbound-format
   contracts; MMX closure; hardware enqueue/random/mask status; custom ABI,
   runtime386 `cpuidProbe` entry. Use the actual diagnostic,
   not an earlier successful preflight, to choose each next class.
   The floating-compare literal batch is qualified at `8aa8460f4`. The exact eleven
   failures are five FP16, two hardware-status roots, two runtime386 source
   fixtures, one absolute relocation and one x87 compare root. Repair each
   complete rule class with original Go/emitter controls and independently
   initialized public programs; do not weaken source/data or entry proofs.
3. Rerun the unchanged complete standard-library corpus and benchmark after
   relevant source repairs. Do not convert Context into successful compilation.
   A valid encoding grants no caller frame, memory ownership or ABI transport.
4. Run complete current-head modules, native/cross-runtime CI, coverage and
   fresh review before ready. Keep known failures visible and batch pushes;
   do not launch another broad diagnostic merely to rediscover known failures.

## Reproducing and locating evidence

Use [validation commands](validation.md). Freeze tracked sources and tools,
including docs, throughout a run. One heavy local Go/LLVM owner at a time;
develop in an existing idle worktree. Commit and preserve RED before repair.

Resolve worktree locations with `git worktree list`. Ignored receipts include:

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
