# PR 40: current checkpoint

Updated 2026-10-20. Implementation handoff: `5e4cf7d0e`. This is not a
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

The fresh independent classification comparison at `bdb7ed6bb` covers all
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

The fresh strict standard-library benchmark at `b2d198ae8` remains failed:
184 target/file pairs attempted, 129 compiled, 55 failed, no N/A/unattempted.
AMD64 took 315s, exceeding its unchanged 300s budget; total 366s is below 900s.
Original MD5 and bytealg/compare now compile, with no new failing files compared
with `ae2579cf2` (127 compiled / 57 failed). MathExp still fails at the native
floating boundary described below. Neither a speedup nor a complete pass is
qualified. Many remaining runtime/reflect files require explicit special-entry
ABI contracts rather than more opcode names; keep their failures visible.

The complete original SDK math/exp CLI regression is still red at `5e4cf7d0e`.
Bounded package inventory, complete 34-name generated float-pool ownership and
mixed ordinary readonly/generated-metadata accounting are repaired. The next
GLOBL identity and complete floating native boundaries now pass. The unchanged
MathExp object has 579 TEXT bytes / 101 instruction PCs. Ordinary F32/F64
numerical declaration binding now passes its independent source/archive and
language/layout controls. The next full-source failure is `VFNMADD231SD`:
the ordinary sixty-member FMA3 family has typed forms and caller effects but
is omitted from result/mandatory-effect dispatch. Its existing F32/F64
arithmetic also needs independent original-opcode/MXCSR/dead-access obligations,
not just a registration patch. Keep both defects under whole-family TDD.
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

1. Qualify complete FMA3 result/input and native floating-effect rules, then
   rerun the complete original MathExp CLI. Keep special-entry ABI work separate.
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
