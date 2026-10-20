# PR 40: current checkpoint

Updated 2026-10-20. Implementation handoff: source-storage repair after
`2547cc6e5`, with the independent bigmod priority owner. This is not a
current-head completion certificate. Preserve ignored failed evidence and
read [instruction contracts](instructions.md) and [validation](validation.md).

## Scope and publication

This repository owns instruction translation, complete typed families and
independent tests for 386, amd64, arm, arm64 and wasm. External scanners,
inventories, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
No inventory scanner runs locally.

PR40 remains draft at `24851df2`; local integration and fork staging are not
that contribution head. Push only to an allowed cpunion/dastai fork. The
diagnostic staging branch is `codex/pr40-windows-owned-process-probe-20261013`.
Recheck upstream main before publication, let each CI run finish, and update
the contribution only after qualifying the exact staged source. Never push,
merge or approve upstream. The sole formal review thread is resolved, not
reviewer approval; orphaned-IR and symlink-include regressions remain required.

## Online evidence is not yet qualified

[39379754007](https://github.com/cpunion/plan9asm/actions/runs/39379754007)
finished naturally at `a530a1688`: 89 successful and 25 failed jobs, with
none cancelled or skipped. Both complete CLI gates and all 58 priority SDK
lanes passed. The complete Windows root inventory is 7,903 roots: 7,571 pass,
60 fail and 272 explicit host skips, zero absent/incomplete, each owned once.
The skips are 237 ARM64, 23 ARM and 12 required Linux-x86 counterparts, not
missing-tool exemptions. The root and diagnostic aggregates correctly fail.
The ignored `root-union-outcomes-audit.json` joins all sixteen owners to raw
events and unchanged complete source/tool/environment observations. Its only
registered environment normalization is the actual Go build-directory nonce
in GOGCCFLAGS; PATH remains strict. Do not subtract local repairs below from
these online failures or combine different heads into a coverage percentage.

All four native reduction owners and their aggregate passed. Native conversion
owners 1 and 4 passed on their actual allocations; owners 0, 2, 3, 5, 6 and 7
failed required CPU preflight with AVX512F/VL absent and XCR0=7. Their aggregate
correctly failed. Native duplicate/reciprocal ABI0 helpers and segment-query
fixtures exposed real failures, repaired locally below. Their corrected
original programs still need execution; floating helpers and wide half-moves
also retain their independent required hardware checks.

The user explicitly deferred unavailable AVX512 execution to pending-hardware
acceptance. Preserve its failed/pending status: no skipped required owner,
emulated native execution or whole-CI pass claim. Mixed official runner
allocations are not a stable hardware guarantee.

Windows now executes after all nine prerequisites finish, even if one fails.
A separate final qualification still requires all nine and complete Windows to
succeed. Failure, cancellation, skip or missing observations cannot qualify.
Diagnostic scheduling does not weaken the final gate.

The preceding complete main-module inventory is
[39285088077](https://github.com/cpunion/plan9asm/actions/runs/39285088077) at
`47e44aa7d`: 7,379 roots = 7,063 pass + 52 fail + 264 explicit host skips,
zero absent/incomplete, each owned exactly once. Its jobs were 83 pass and
16 fail; both complete CLI modules passed. The ignored
`ci392850/root-union-outcomes-audit.json` binds that inventory to raw events.
This older inventory does not qualify later code.

## Recent focused local evidence

Every scope below has zero failed, skipped or incomplete terminals within
its recorded tests, or all commands exit zero for command-gate receipts.
Each belongs to its own source/tool-sealed snapshot. Roots/children overlap:
never add them. Objects, portable values, original execution and ABI authority
differ. Older qualifications remain in earlier Git checkpoints.

| Snapshot | Independent local scope | Limits |
|---|---|---|
| `b2d639a6e` | Complete MULX/GP/ADX format, physical storage and source contracts: 25 / 780 on current Go, actual Go1.20 and race | Zero failure/skip, unchanged complete seals; two exact historical macOS linker warnings retained. Not the complete command gate, native execution or a new strict benchmark |
| `2fe616991` | Complete MULX/GP/ADX original-format and source contracts: 24 / 714 on current Go and actual Go1.20 | Zero failure/skip and unchanged complete seals. Historical macOS linker warnings retained. The later physical-memory-source repair needs its own complete current/minimum/race replay |
| `fbdfb33b9` | Fresh sixteen-report official classification comparison after the reviewed baseline update | Zero form, encoder or metric deltas; complete seals unchanged. Not an initialized-program or later-head qualification |
| `e8c985358` | Complete original MathExp, field and bigmod CLI sources: three roots pass in 218.138 seconds | Original ASM and stock-Go archive priors, LLVM22 objects, zero failure/skip/stderr and unchanged complete seals. Bigmod alone takes 192.1 seconds; neither native execution nor a benchmark performance pass |
| `376f4f02d` | MULX typed inputs, native reads, five-target objects and original owned DATA: 6 / 427 current, 6.671 seconds | Sixty-four initialized function instances; three APIs and original/O2 IR to backend O0. Exact reads, two outputs and FLAGS obligations, not executed values/faults. Later physical-output and format repairs need the complete current/minimum/race replay |
| `2e4320f62` | Complete ADX physical boundaries and existing raw/read contracts: 12 / 583 current, 4.233 seconds | Four official rows, 98 actual Go/LLVM physical witnesses, eight RIP priors; no scalar/NUM/source permission. Zero failure/skip/stderr and unchanged complete seals |
| `c9fcc52ee` | Existing complete VEX GP register/memory/default-role decoder controls: 19 / 276 current, 1.594 seconds | Zero failure/skip/stderr and unchanged complete seals; the new public MULX input contract subsequently requires separating its old uninitialized format fixtures |
| `846526a6e` | Native listing and floating backend compatibility: 10 / 880 actual Go1.20, 18.424 seconds | Sixteen current-SDK roots are absent by build role, not skipped. Zero failure/skip, unchanged complete seals; two exact macOS linker warnings retained |
| `ecb95a59b` | Complete listing, source lifecycle and floating backend objects: 26 / 947 current, 191.722 seconds | Zero failure/skip/stderr and unchanged complete seals. Original Go headers include real split-static and FIPS producers |
| `dab65118f` | Isolated floating backend objects/profile: 4 / 570 race, 78.587 seconds | All three targets and three pipelines retained, zero failure/skip/stderr and unchanged complete seals. Does not erase the preceding combined five-minute-budget failure |
| `bcab1e58d` | Complete named/raw half, raw-only isolation and original floating diagnostics: 49 / 1,334 current, 95.756 seconds | Zero failure/skip/stderr and unchanged complete seals. Original fixed-rounding ASM/C and native objects remain separate from explicitly private mathematical execution; no native MXCSR execution claim |
| `89d5189e9` | Half, floating diagnostics and legacy scalar stores: 56 / 2,315 actual Go1.20, 94.029 seconds | Zero failure/skip and unchanged complete seals; two exact macOS linker build warnings retained. The current-SDK encoder root is absent by build role, not skipped |
| `bcab1e58d` | Complete packed precision, unsigned scalar and shared consumers: 50 / 2,820 actual Go1.20, 73.306 seconds | Historical reserved maskless-zero bytes are observed, not accepted semantics. Zero failure/skip, unchanged complete seals; two macOS linker build warnings retained |
| `32a5aadbb` | Complete permute128 and SDK physical boundaries: 30 / 1,188 current, 20.038 seconds | Two-member grammar, every selector, initialized objects, portable values/faults, actual Go/LLVM mode32 and RIP-length priors; no NUM/DATA/caller permission. Zero failure/skip, unchanged complete seals. Original AVX/AVX2 execution remains separate |
| `e22831f42` | Complete packed precision and shared consumers: 35 / 2,024 current, 34.973 seconds | Six names/eighteen Go rows, fifty-three initialized shapes, original raw/named objects, exact native opcodes and strict aliases. Zero failure/skip, unchanged complete seals; not original x86 execution or minimum/race |
| `f7e631cfa` | Complete blend/minpos: 27 / 1,012 current, 66.509 seconds | Complete original formats, initialized objects, mode32 grammar and portable mathematics; zero failure/skip and unchanged complete seals. Two original Linux native owners are wired, not locally executed |
| `53dda86fc` | Complete VNNI integration: 24 / 963 current, 68.012 seconds | Independent Go rows, every-byte inputs, exact dead-access tuples, raw replay, five-target objects and BF16 controls; zero failure/skip, unchanged complete seals. Original native VNNI execution remains separate |
| `53dda86fc` | Complete raw unsigned zero-extension: 15 / 414 current, 13.293 seconds | Full original initialized source, sixty orthogonal five-target kernels, source-identity and DF/private/GP preservation controls; zero failure/skip, unchanged complete seals. Minimum/race remain required |
| `ddf4a082a` | Complete unsigned scalar integer-to-float family: 18 / 1,091 current, 126.145 seconds | Four members, all legal ER forms, five-target objects and 76 independently Go/LLVM-checked mode32 producer displacement cases; zero failure/skip, unchanged complete seals. Minimum/race and original AVX512 execution remain separate |
| `3cc0ff5cf` | Complete raw ABS literal family: 13 / 484 on current Go | Independent 66 physical cases, actual LLVM decoding, every-byte/mask/identity negatives and three-API literal objects; minimum/race and original native execution remain separate |
| `d4d0d8519` | Five original unbound format drivers and initialized controls: 11 / 270 current | Original Go bytes and public Context retained; private format objects are not initialized public programs |
| `3f5e0bb26` | Logical-FP DIV/IDIV and existing arithmetic: 23 / 2,656 current | Exact inputs and native division effects; new AMD64/386 original-value/fault owners await CI execution |
| `153f85f78` (integrated at `c56a9f3aa`) | Current-SDK text identity and existing observers: 12 / 276 current | Actual independent Go archive kind/ABI priors; same-name mixed-kind collisions stay rejected, raw/historical ownership unchanged |
| `60b691f35` | Complete legacy scalar native stores: 9 / 991 current | Exact volatile m32/m64 stores, unknown-input boundaries and portable values/faults; original x86 execution remains separate |
| `e3e24a9af` | Complete original vector-move formats and initialized controls: 15 / 141 current | Five-target objects preserve original Go rows, including architectural faults; no source-input exemption |
| `a951f91b6` | Variable blend/minpos typed inputs and original initialized objects: 10 / 678 current | Core qualification only. Native faults, private mathematical driver and format corrections are integrated candidates, not a complete later-head qualification |
| `5529f23d3` | Complete original SDK field-multiply and MathExp CLI sources compile to LLVM22 objects | Actual stock-Go archive priors, unchanged whole assembly and complete seals; not native execution or a full benchmark |
| `59fb59445` | Complete implicit/explicit multiply: 34 / 4,296 on current Go, actual Go1.20 and race | Exact native reads, CF/OF, all-byte masks, negative authority controls and five-target three-API objects. The 35th original-x86 runtime root requires a real x86 host, not a local pass |
| `f6a73b3e3` | Native fixture source/layout controls: 6 / 42 on current Go, actual Go1.20 and race | Actual Go assembler/object and asmdecl priors, all return widths, 386/AMD64 ABI0 slots, complete SETOS source and defined AF seed; repaired original programs still need CI execution |
| `a530a1688` | Integrated FMA3: 49 / 8,431 on current/minimum/race; 33 command gates, 206 Node policies and unchanged sixteen-report classification comparison pass | Complete seals unchanged; not a later-head, full-CI, strict-benchmark or 50th original-x86 caller execution qualification |
| `afa117ae1` | Integrated named/raw FMA3: 49 / 8,431 on current Go, actual Go1.20 and race | Zero failures/skips, unchanged complete source/tool seals. Includes exhaustive decoder, caller-effect, five-target object, portable rounding and access regressions. The fiftieth original x86 caller runtime root remains separately pending suitable hardware |
| `d63bc455a` | Ordinary raw FMA3: 16 / 6,872 on current Go | Exact VEX/EVEX, 528 physical shapes, whole-source inputs/replay and private FLAGS/DF. Five-target objects, not original native floating execution; integrated minimum/race are still required |
| `0b0cd1a7f` | Original raw FMA formats, physical mode32 and initialized controls: 6 / 47 current | Original bytes and DATA remain unchanged; private format objects, public Context and initialized public objects are separately checked |
| `044ee4030` | All 33 command gates exit zero, including current/minimum build/vet, official five-arch coverage and 206 Node policies; complete seals unchanged | Integrated named FMA, portable access, format objects and audited baseline; not full CI, the strict benchmark or the unintegrated ordinary raw FMA candidate |
| `8ec12ae76` | Original FMA3 formats and initialized axes: 8 / 69 on current Go | Complete original formats remain separate from 178 amd64 / 94 mode32 orthogonal initialized entries, 144 LLVM object compilations across five targets; not a full initialized Cartesian matrix or native execution |
| `f194e6a1e` | Portable FMA access and original native controls: 7 / 378 on current Go, actual Go1.20 and race | Three APIs and three executed ARM64 pipelines each admit all 14 access classes: 98 independent references / 98 translated probes, zero Context. This preserves access/faults, not original x86 MXCSR execution |
| `cbb448b58` | Complete original SDK MathExp CLI source and LLVM object pass | Unchanged original assembly, actual stock-Go archive prior and complete source/tool seals; not a full benchmark receipt |
| `0d35d8b1e` | Complete named FMA3: 5 / 282 on current Go, actual Go1.20 and race | All 60 members, three inputs including old destination, native floating effects and initialized live/dead programs. Ordinary raw F32/F64 generation and source ownership remain separate work |
| `dc3c228b4` | Ordinary floating numerical declarations: 15 / 97 current and race, 11 / 80 actual Go1.20 | Direct/named/alias F32/F64 language and ABI0 slots; ordinary frame, caller, pointer/composite and register-entry negative controls remain strict. Four current-SDK roots are absent by minimum build role |
| `5e4cf7d0e` | All 33 command gates exit zero, including 206 Node policies; source/tools unchanged | Supersedes the failed `bdb7ed6bb` attempt for this scope, without erasing it; not full CI, the strict benchmark or initialized FMA3 execution |
| `5a3b67fb7` | Shared integer grammar and existing consumers: 25 / 4,783 on current Go, actual Go1.20 and race | Complete expr/term/factor, 55 independently Go-validated values, DATA/SB/GLOBL, five targets and three public APIs; malformed/unknown offsets cannot create images |
| `d887767b0` | Complete floating native boundaries and existing consumers: 14 / 144 current and race, 13 / 144 minimum | 88 FMA/arithmetic/SQRT members, exact original MathExp bytes/PCs, valid inline-literal NUM negatives and actual mode32 legacy identity. Current-SDK original source root is absent by minimum build role |
| `b2d198ae8` | All 33 command gates exit zero; source/tools unchanged | Current/minimum build/vet and consumers, official five-arch gate, 206 Node policies, shared current/minimum/race state. Not complete CI, the strict benchmark or the two newer RED families |
| `bea55cdaa` | GLOBL zero addends: 4 / 1,984 on current Go, actual Go1.20 and race | Five arches, original Go objects and three public APIs. A subsequent independent review exposes shared integer grammar differences; these receipts do not certify arbitrary expressions |
| `1cbe5377d` | Integrated source-defined BP, SDK directory and orphan-IR controls: 16 roots / 228 children | Current Go, including unchanged complete MD5 across three public APIs |
| `72af0cff1` | BP overwrite and historical frame controls: 9 / 925 on current Go and actual Go1.20 | Ordered old/new prologues, partial writes, aliases, joins and FP-slot reimports; no caller/frame authority inferred from BP |
| `1cbe5377d`, `bdd6a0397` | Complete 850-witness raw packed-shift object matrices for all five x86 targets: four roots / 260 children per target | Three public APIs and three pipelines. Linux/386 and Windows/386 use the earlier snapshot; the three amd64 targets use the latter. Objects are not original native execution |
| `bdd6a0397` | Original 396 shift forms compile for Linux/Darwin/Windows amd64: one root / 143 children per target | Complete original source, three public APIs and three pipelines |
| `cfa4197d3`, `0e3833aac`, `72af0cff1`, `bdd6a0397` | All nine packed-shift members pass complete portable value/fault programs | Original 396-form companion plus 36 independent word-immediate forms; not native AVX512 execution or a single-snapshot certificate |
| `08010ffe1`, `bf1ec0a5b` | Generated DATA metadata and readonly plans: 17 / 613 current; integrated shared 14 / 126 on actual Go1.20 and race | Exact float-pool/FUNCDATA ownership explains inventory, never grants ordinary initializer images or NUM permission |
| `9e187e272` | Scalar-shift entry-shape/reaching-state controls: 7 / 131; complete SDK compare: 3 / 21 | Separate current-Go scopes. Shape classification no longer borrows stale entry pointers; reaching NUM and FLAGS obligations remain unchanged |
| `7948b7ece`, `94f85c871` | Complete four-source rules and objects: current 30 / 2,118, minimum 29 / 2,118 | The current-SDK-only official encoder root is absent by build role on actual Go1.20, not skipped. Native rare-ISA execution and private portable mathematics remain distinct |
| `b2c4c110b` | Previous frozen batch: all 33 command gates exit zero | Current/minimum build/vet and scanner; CLI on current Go and its Go1.24 minimum; official five-arch gate, 190 Node policies and 16 shared roots in current/minimum/race lanes. Not 33 test roots or a current-union certificate |
| `b5e621d0b` | Explicit vector-mask moves: 24 roots / 1,390 children | Current Go; typed inputs, aliases, objects and owned guarded-access controls |
| `ad92005db` | Complete producer replay through strict numerical success: 3 / 4,142 | Current Go; edited original producers cannot inherit a successful numerical continuation |
| `8e0bf6975` | Integrated immediate-shuffle and explicit-mask regressions: 24 / 4,673 | Current Go; initialized native readers stay separate from unbound format probes |
| `58997e227` | Physical-SP fixture contracts, LEAVE mathematics and XLAT: 20 / 88 | Current Go; private SP mathematics is not native Go stack ABI execution |
| `1eec6500b` | ABS/sign input/access controls and isolated ISA profiles: 26 / 6,568; eight focused gates | Separate local scopes, not current-union or rare-ISA execution qualification |
| `cc350e9cd` | Complete scalar precision conversion and original regressions: 22 / 609 | Current Go; exact raw architecture/generation/replay and minimum SSE2/AVX/EVEX profiles; original MXCSR execution remains separate |
| `6fb23fe5d` | Floating MIN/MAX generation, replay and DATA coordinate bindings: 14 / 260 | Current Go; physical low-bank EVEX and decoder/literal identity cannot be replaced by a named spelling |

All 206 Node policies passed at `b2d198ae8`, including the independent final
diagnostic qualification. These are local policy receipts, not an online CI
result. At `bdb7ed6bb`, 21 command gates passed before policies failed: 205 of
206 policies passed, with one fake-shell child exceeding its unchanged five
second limit. Eleven remaining command gates did not run. Source/tool seals
were unchanged. Isolated and complete recorded-environment replays passed,
but do not erase this failed receipt or qualify the unrun commands. Preserve
the failed receipt and unchanged budget. The fresh `5e4cf7d0e` snapshot now
passes every one of the 33 commands and all 206 policies, with unchanged seals;
it does not certify later FMA3 work or the separate strict benchmark.

The first FMA3 race attempt did not execute tests because its shared Go build
cache lost required files. Preserve that infrastructure failure. A fresh owned
cache passes the exact five-root scope at `0d35d8b1e`; no tool fallback or budget
increase was used. Every successful scope above has unchanged complete seals.

## Active repair queue

Keep the integration branch `codex/pr40-typed-ci-batch-20261020` frozen during
qualification. Three authorized agents work separately; only the coordinator
runs heavy Go/LLVM locally. No scanner runs here. Batch publication follows
current/minimum/race, policy, classification and benchmark verification, not
the completion of a single focused scope.

- VNNI, raw unsigned widening, unsigned scalar conversion and blend/minpos
  have the complete current-Go scopes above. New shared changes still need
  integrated minimum/race, command gates and original native qualification.
- Blend/minpos has a separate required two-root Linux native scope; the
  existing eight integer owners are unchanged. All 25 focused native-policy
  tests pass. The first full 221-policy run failed two isolation tests because
  its independent owner inventory omitted the new scope. `fc3d7f7ba` adds that
  obligation; the later frozen preliminary union passes all 221 policies.
  The complete current-head command qualification remains pending.
- Permute128 now has its complete thirty-root GREEN above. Preserve the
  earlier dead-NULL observer failure and minimal LLVM22 experiment: backend
  legalization may read either sixteen-byte half first, so test-only NULL
  ownership is exactly the original 32-byte tuple. Separate actual Go/LLVM
  physical-PC/RIP RED precedes the six-line decoder/registry fix. Boundary
  observation never supplies NUM, symbolic reads, DATA or caller authority.
- Packed precision now has its complete thirty-five-root GREEN above.
  Preserve the initial thirteen-root production RED and later failed 23-root
  integration receipt. Independent fixtures retain all original unbound rows,
  explicit mode32 physical-bank Context and strict same-call memory aliases.
- Named/ordinary raw half now has the complete current/minimum scopes above.
  Preserve the valid eighteen-root RED at `48460ca6`, the separately validated
  integer-expression RED at `113697d1f`, and earlier invalid-CMP/identical-literal
  fixture attempts as distinct evidence. Whole-producer replay, physical
  generation and mandatory source/mask/merge bytes stay independent of raw-only
  FP16. Dynamic MXCSR 4–7 retains its original opcode; fixed-rounding private
  mathematics is not native F16C/MXCSR execution. Go-accepted zero-store #UD
  precedes inputs; actual-Go SAE-memory rejection is a different boundary.
- At `f6a860154`, all 221 Node policies pass in a source-frozen preliminary
  diagnostic. This is not the complete 33-command qualification. The actual
  backend object audit now distinguishes original Go KMOVQ from translated
  KMOVW and F-only scalar from F/VL FMA. All three object/profile roots pass
  at `0b744db7b`, including three targets and all three pipelines; actual CPU
  execution still belongs to the original owners.
- Complete STEXT/STEXTFIPS ownership now passes the 26-root current and
  ten-root actual-minimum scopes above. Retain the earlier 25-root failure:
  one old fixture wrongly required nosplit. Unknown/truncated non-text still
  revokes ownership. The combined 26-root race attempt at `d00d05088` fails
  after exhausting its shared five-minute budget during Windows FMA objects;
  25 roots pass. The isolated complete backend race passes at `dab65118f`.
  Rebalance the final union without increasing budgets or omitting objects.
- Full bigmod replay exposed two independent byte-observer gaps, PC519 VEX
  GP and PC524 legacy ADX. Their complete shared decoder projections now
  retain 26 GP names/28 Go rows and all four ADX names; physical observations
  enter neither scalar nor NUM reads. Complete original bigmod now passes
  beside MathExp/field above. Source roles and original execution stay separate.
- MULXL/Q now share typed implicit-DX/two-output effects and exact volatile
  native reads. Independent RED proves missing-pointer Void acceptance and
  absent initialized effects; the first six-root GREEN is above. Further
  valid RED covers physical SP/AMD64 BP outputs; the two-line source-obligation
  repair passes those controls in the 24-root attempt at `b8ffd18eb`.
  The private-format assertion failures are repaired separately: SIB address
  multiplication is not a MULX product, and a private missing DX slot can
  appear as an exact zero literal in a correctly typed zero extension.
  All nine original byte formats, targets, public Context controls and
  initialized objects pass in the 24-root current/minimum receipt above.
  A further test-only RED at `2547cc6e5` establishes 21 named physical-memory
  source failures across three APIs; actual Go/LLVM priors, raw controls,
  incoming-pointer, mode32 BP and owned-DATA positives pass. The minimal
  candidate sends physical SP/AMD64 BP base/index sources through the existing
  whole-source proof, and passes all three lanes at `b2d639a6e` above.
  Preserve Go386's ignored static-SB base annotation and
  exact absolute/native accesses; do not invent new Context negatives from
  the routing predicate alone.
- Five original GP DATA negatives were already successful before the decoder
  repair; they were invalid negatives, not permission regressions. Keep their
  original owned-DATA controls and thirteen actual missing-BX counterexamples.
  Actual Go1.20 additionally reproduces numeric versus named R_PCREL display;
  test-only compatibility binds the exact selected SDK enum and unique full
  relocation row. Do not skip historical RIP cases or call that a compiler RED.
- Fresh eight-class audit at `2e4320f62` passes all sixteen reports with
  exactly 179 supported-to-Context changes per AMD64 SDK and no 386 changes.
  The subsequent nine-class audit at `24aaae415` adds only the two complete
  MULX native-memory forms, 181 total. Complete encoder/form fields and all
  unrelated metrics remain identical. Only the audited status counters and
  fingerprints update the eight AMD64 baselines. Current Go is 2,390 supported,
  4,346 Context and zero unsupported. Preserve historical Go1.20's one existing
  unsupported form. Fresh zero-delta replay passes at `fbdfb33b9` above;
  final integrated-head comparison remains required.
- Required native hardware remains pending, never a skipped success. The
  original ABS/sign, scalar precision and FMA caller SIGILL failures do not
  establish hardware absence without their own actual-allocation CPU receipt.

The first `b2d639a6e` full command attempt passes 21 gates, then correctly fails
policies: 220 of 221 pass, with the new original bigmod root missing from the
priority inventory. Eleven later commands remain unrun; complete seals match.
An independent five-minute bigmod owner now supplements the unchanged original
MathExp/field owner. The policy RED detects that missing lane and retains exact
root, module, disjointness, substitution and budget negatives.
The next `31aa48004` attempt again passes 21 command gates before policies:
208 of 221 pass, with thirteen failures all caused by the shared SDK lane-count
constant still expecting 60 instead of 61. Source/tool seals remain unchanged;
the eleven later commands did not run. The one count is updated; exact lane
and original-root checks remain mandatory.

Newly integrated changes have not passed the full 33 command gates, final-head
classification replay or a new strict benchmark. Preserve the failed benchmark
below; do not turn local repairs into an inferred full pass.

Raw continuations retain sealed complete producer replay, not just a mutable
origin pointer. SDK literal/numerical owners precede native family hooks.
Dead components use actual preceding source facts, never future/shape-only
facts, and cannot flow back into reachable states. Keep the shared
16,384-visit / 64MiB budget and independent negative controls unchanged.

The two old physical-SP programs keep their complete source/C and all 78 Ymb
aliases, actual Go objects/bytes and public Context controls. Their unrecovered
physical SP/late FP/RET sequence is not a valid native return contract. Private
LLVM register mathematics executes those observations without granting NUM,
caller, frame or entry authority to physical SP/BP. Required Linux AMD64/IA32
LEAVE counterparts own four width/mode numerical and FLAGS checks separately;
their new CI owner is not yet an online execution certificate.

XLAT has a separate legitimately initialized public counterpart: unsigned AL8,
real BX pointer input, exact volatile m8, preserved AX upper bytes and FLAGS,
all 256 indices and dead-result fault controls. Independent native-byte access
references do not claim execution of an original Go XLAT entry. Likewise,
retained LEAW portable DF instrumentation is not unmodified native execution.

## Classification and performance references

The earlier independent classification comparison at `bdb7ed6bb` covers all
sixteen Go1.20–Go1.27 x86 reports. Every form, status, fingerprint and metric
matches the separately qualified `18a4052d3` reference; no baseline was edited.
That reference, recorded by baseline-only `ae2579cf2`, moved exactly 126 further
AMD64 input-deficient probes per SDK from supported to Context across complete
four-source, precision, mask, MIN/MAX and XLAT families; every 386 form and
unrelated encoder/unsupported/parse metric stayed unchanged.
Current-SDK AMD64 counts are 3,113 supported, 3,623 Context and zero unsupported.
These counts are rechecked at `bdb7ed6bb`, not a later-head qualification.
Context is not an initialized-program success. Retain the historical Go1.20
unsupported form;
the current five-architecture gate still requires zero unsupported/parse errors.
Re-audit complete forms and fingerprints before changing any baseline.

The subsequent `68fbaa638` audit independently checks all sixteen reports
against that qualified prior. Exactly 542 input-deficient FMA3 forms per AMD64
SDK move from supported to Context; all 386 forms and every other metric stay
unchanged. Current Go AMD64 is now 2,571 supported / 4,165 Context / zero
unsupported. Only these audited classifications and their fingerprints update
the baseline. Initialized family programs and objects remain separate evidence.

Ordinary raw FMA candidate `837d6e955` passes its physical generation, isolated
ISA, snapshot, initialized Weaviate and complete RIP controls, but its full
fifteen-root attempt fails at `ANDL $0xcd5,DX` in both x86 modes. The FMA result
step leaves the six status flags intact; the paired source-flow fallback loses
the independently defined DF, so saved FLAGS lacks bit 10. The independent
saved-FLAGS-across-FMA RED is `09a2e66ac`; the minimal private-state repair
`d63bc455a` now passes the complete sixteen-root current-Go scope above with
unchanged source/tool seals. Both the candidate failure and independent RED
remain preserved. The integrated `a530a1688` snapshot passes the full 49-root
minimum/current/race scope, all 33 command gates, unchanged complete forms and
baseline fields against `68fbaa638`, and the full MathExp CLI. Older receipts
do not qualify the later multiply/native-fixture integration.

The preceding full benchmark at `a530a1688` remains failed: all 184 target/file
pairs attempted, 130 compiled, 54 failed, no N/A/unattempted. Target wall times
are 33/274/1/7/1 seconds for 386/amd64/arm/arm64/wasm, total 316 seconds,
within unchanged 300/900 budgets. All timestamped completion rows join the
original reports; complete seals match and owned-process cleanup is clean.
These parent receipt intervals are not pure lowering CPU times or a controlled
speedup measurement. MathExp is the one removed failure versus `47e432d4e`;
do not extrapolate the later field-source fix into a new benchmark percentage.
Ordinary SHA/GCM local-frame and private-helper failures are not all
runtime/reflect special-entry ABI failures. Preserve both classes explicitly.

The newer `4d129bcd3` benchmark still fails: all 184 pairs attempted, 132
compiled, 52 failed, zero N/A/unattempted; complete source/tool/environment
seals match and all completion rows join their reports. Only `vlop_386.s` and
`fe_amd64.s` leave the previous failure set; no new file fails. Target wall
times are 56/505/5/11/2 seconds, total 579 seconds. AMD64 exceeds the unchanged
300-second target budget; a total below 900 seconds does not qualify it.
One diagnostic sample and source audit locate repeated live input hashing and
complete observation serialization; neither is a controlled speedup result.
Do not remove live verification boundaries or increase budgets to claim a pass.

Readonly classification joins all 52 failures to the original reports:
eight ordinary positive-frame ABI0 cases, nine helper/code-address closures,
four inventory-budget barriers, 21 special runtime/compiler/OS entry contracts,
nine other independent contracts and one unresolved SDK-graph cause. These
are blocking layers, not promises that a single repair compiles each full file.
The later bigmod success is not a new 184-file benchmark or speedup measurement.

The complete original SDK math/exp CLI regression passes at `cbb448b58`.
Bounded package inventory, complete 34-name generated float-pool ownership and
mixed ordinary readonly/generated-metadata accounting are repaired. The next
GLOBL identity and complete floating native boundaries now pass. The unchanged
MathExp object has 579 TEXT bytes / 101 instruction PCs. Ordinary F32/F64
numerical declaration binding passes its independent source/archive and
language/layout controls. The former `VFNMADD231SD` failure is repaired for the
complete sixty-member named family, including original-opcode floating effects.
An independent cross-target RED at `90e4fcb07` proves that IR O2 previously
deleted dead input accesses; `f194e6a1e` reuses exact volatile/guarded readers.
Ordinary raw VEX/EVEX has a separate valid RED at `cad09be57`; independent Go
and LLVM physical priors pass. The repaired generation, ISA and complete raw
continuations now pass the integrated scope without borrowing named or FP16
permissions. Keep original unbound format sources, objects and explicit
Context distinct from initialized program and execution evidence.


Do not open ordinary FrameSource, caller, ABIInternal register or aggregate
permissions; this is distinct from runtime/reflect special-entry ABI contracts.

Go-AST integer evaluation has been replaced with the bounded complete assembler
grammar. Preserve its RED history: `5934717d1` mixed five upstream assembler
panics with real failures; uniformly bounded DATA-displacement fixtures at
`1864b5e28` and expanded `ca7fd4bdc` establish normal Go rejection instead.
The latter validates 55 complete Go symbol/images per architecture. Two old
synthetic positives (`7&^3` and byte-rune `\\377`) were independently rejected
by Go and corrected in separate fixture commits. Never count compiler panics,
timeouts or unknown process outcomes as valid rejection evidence.

The packed-shift portable driver now consumes normalized TEXT and retains unused
intrinsic declarations and their generic attributes byte-for-byte. Actual native
calls/addresses, foreign attributes and ambiguous ownership still reject. All
nine portable value/fault members passed their recorded snapshots above; this
does not certify original native Go execution or the final integrated head.

## Next work and verification

1. Preserve the qualified FMA3, complete multiply and native-fixture snapshots.
   The field CLI has a valid test-only RED at `3d6b7048d`: stock Go compiles,
   but translation fails at instruction 3, `MULQ (BX)`. The shared typed fix
   closes all 25 such accesses in the unchanged file, not a package exemption.
   Add complete priority owners without shrinking existing inventories.
   Keep special-entry ABI work separate. Integrate repairs without editing
   active snapshots; preserve original inputs and all earlier failed evidence.

2. Freeze the final integrated head, including documentation and used tools.
   Complete its current-Go, actual Go1.20 and race regressions and final command
   gates; older or partial receipts do not qualify this union.
3. Rerun root/CLI build and vet, scanner and isolated supported-op extraction,
   ownership/policy gates, complete official forms/classifications and the
   five-architecture corpus. Rich CLI minimum is Go1.24, not root Go1.20.
4. Complete the strict benchmark and required native owners on suitable
   hardware. Conversion needs F/DQ/BW/VL, floating helpers F/DQ/VL and wide
   half-move transport F. Reduction/reciprocal need SSE/AVX; retained LEAW and
   private LEAVE have no AVX512 dependency. Never weaken preflight or replace
   real execution with objects, portable values, fake flags or skipped tests.
5. Publish one coherent allowed-fork batch, watch CI through completion and
   inspect every owner and aggregate. Keep PR40 draft until exact-head CI,
   review and coverage qualify; only then update the contribution head/body.

Use current Go1.27 and exact LLVM22; missing tools/backends fail, never skip.
One heavy local Go/LLVM owner at a time. Keep failed reports under ignored
`_out/`, release LLVM objects/subprocesses promptly, and do not edit snapshots
while they run. Native host exclusions require explicit required counterparts.
