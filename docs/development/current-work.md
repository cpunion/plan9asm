# PR 40: current checkpoint

Updated 2026-10-20. Implementation handoff: `044ee4030`. This is not a
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

## Online evidence remains failed

[39325946539](https://github.com/cpunion/plan9asm/actions/runs/39325946539)
finished naturally at `a46ca8ab2`: 72 jobs pass, eight fail, one dependent gate
skips, none cancelled. Full Windows modules did not run. Source-priority budget
oracle and duplicate/division fixture-package defects are repaired locally;
this does not change that run's failed status or establish current-head CI.

All four native reduction owners and their aggregate passed. Native conversion
owners 0, 5 and 7 passed on actual AVX512F/DQ/BW/VL allocations; owners 1, 2,
3, 4 and 6 failed required CPU preflight with F/VL absent and XCR0=7. Their
aggregate correctly failed. Official Linux and Windows allocations are mixed:
one capable allocation is not a stable hardware guarantee. Independent native
steps keep their own failure status while allowing unrelated owners to run.

The next workflow lets Windows execute after all nine prerequisites finish,
even if one fails. A separate final qualification requires all nine and complete
Windows to succeed; failure, cancellation, skip or missing observations cannot
qualify. This only improves diagnostic scheduling, not AVX512 availability or
native execution qualification. The online run above remains failed.

The user has explicitly deferred unavailable AVX512 execution to a separate
pending-hardware acceptance item. Preserve its failed/pending status; this is
not permission to skip the required owner, emulate native execution or report
the whole CI as passed. Continue qualifying every independent software repair.

The last complete main-module inventory is still
[39285088077](https://github.com/cpunion/plan9asm/actions/runs/39285088077) at
`47e44aa7d`: 7,379 roots = 7,063 pass + 52 fail + 264 explicit host skips,
zero absent/incomplete, each owned exactly once. Its jobs were 83 pass and
16 fail; both complete CLI modules passed. The ignored
`ci392850/root-union-outcomes-audit.json` binds that inventory to raw events.
Do not subtract local repairs from its 52 failures or combine different heads
into a new whole-suite coverage percentage.

## Recent focused local evidence

Every scope below has zero failed, skipped or incomplete terminals within
its recorded tests, or all commands exit zero for command-gate receipts.
Each belongs to its own source/tool-sealed snapshot. Roots/children overlap:
never add them. Objects, portable values, original execution and ABI authority
differ. Older qualifications remain in earlier Git checkpoints.

| Snapshot | Independent local scope | Limits |
|---|---|---|
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
saved-FLAGS-across-FMA RED is `09a2e66ac`; the minimal private-state repair is
`d63bc455a`, still awaiting validation. Preserve the failed candidate report;
do not integrate or call the whole ordinary raw family qualified yet.

The latest full benchmark diagnostic at `47e432d4e` remains failed: all 184
target/file pairs attempted, 129 compiled, 55 failed, no N/A/unattempted. The
five targets, script and 300s/900s budgets are unchanged; AMD64 took 450s and
total target wall time was 497s. All 184 timestamped completion rows join the
original reports, with unchanged source/tool seals and no owned-process cleanup
errors. These are parent receipt intervals, not pure lowering CPU times.
Every failed filename matches the earlier strict `b2d198ae8` run (AMD64 315s,
total 366s); do not claim a controlled performance comparison or a speedup.
Original MD5 and bytealg/compare compile, while MathExp's next blocker is below.
Many runtime/reflect failures need explicit special-entry ABI contracts rather
than more opcode names; keep those failures visible.

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
Ordinary raw VEX/EVEX has a separate valid RED at `cad09be57`: independent Go
and LLVM physical priors pass, but generation/ISA and initialized complete raw
source continuations fail. Finish that rule class without borrowing named or
FP16 permissions. Keep original unbound format sources, objects and explicit
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

1. Finish raw FMA3 physical generation and complete-source result rules; qualify
   original format objects and independent initialized axes. The named-family,
   portable-access and complete MathExp scopes above are qualified at their
   recorded snapshots. Keep special-entry ABI work separate.
   Integrate completed rule-class repairs without editing active snapshots;
   preserve original inputs and all earlier failed evidence.
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
