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

The latest local diagnostic at `0ab05d184` selects exactly the original CI's
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
| Latest batched replay, `0ab05d184` | 39 | 39 | 89 | 0 / 0 |

The five newly passing original roots cover ENDBR, the fuzzer bit-test/vector
program, packed-compare frame extents, LOCK carry rejection and the SDK equal
source's independent Context boundary. Preserve each original failing source;
separate format-only controls from initialized public programs. A corrected
Context expectation is not a new whole-source translation success.
Earlier SQRT, static-SB SAD and wrapping-add repairs remain passing.
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
2. Repair the latest 39 local failure roots: legacy logical-memory format
   fixtures with unbound segment/pointer inputs;
   raw FP16/VMOVW native-read composition; raw literal and unbound-format
   contracts; MMX closure; hardware enqueue/random/mask status; custom ABI,
   source-stack and runtime386 `cpuidProbe` entry. Use the actual diagnostic,
   not an earlier successful preflight, to choose each next class.
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
