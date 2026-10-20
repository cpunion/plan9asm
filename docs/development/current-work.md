# PR 40: current checkpoint

Updated 2026-10-20. Implementation handoff: `798bff706`. This is not a
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
| `b2c4c110b` | Previous frozen batch: all 33 command gates exit zero | Current/minimum build/vet and scanner; CLI on current Go and its Go1.24 minimum; official five-arch gate, 190 Node policies and 16 shared roots in current/minimum/race lanes. Not 33 test roots or a current-union certificate |
| `b5e621d0b` | Explicit vector-mask moves: 24 roots / 1,390 children | Current Go; typed inputs, aliases, objects and owned guarded-access controls |
| `ad92005db` | Complete producer replay through strict numerical success: 3 / 4,142 | Current Go; edited original producers cannot inherit a successful numerical continuation |
| `8e0bf6975` | Integrated immediate-shuffle and explicit-mask regressions: 24 / 4,673 | Current Go; initialized native readers stay separate from unbound format probes |
| `58997e227` | Physical-SP fixture contracts, LEAVE mathematics and XLAT: 20 / 88 | Current Go; private SP mathematics is not native Go stack ABI execution |
| `1eec6500b` | ABS/sign input/access controls and isolated ISA profiles: 26 / 6,568; eight focused gates | Separate local scopes, not current-union or rare-ISA execution qualification |
| `cc350e9cd` | Complete scalar precision conversion and original regressions: 22 / 609 | Current Go; exact raw architecture/generation/replay and minimum SSE2/AVX/EVEX profiles; original MXCSR execution remains separate |
| `6fb23fe5d` | Floating MIN/MAX generation, replay and DATA coordinate bindings: 14 / 260 | Current Go; physical low-bank EVEX and decoder/literal identity cannot be replaced by a named spelling |

All 194 Node policies passed at `58997e227`; thirteen merge-sensitive policies
passed after integration at `798bff706`. These are local policy receipts, not
an online CI result. The present union still needs its own consumer checks,
final 33-command gate batch, current/minimum/race regression qualification,
classification audit and benchmark. Previous receipts cannot fill those gaps.

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

The last committed independent classification comparison is at `79dd1ca75`,
baseline-only `420233de8`: all sixteen Go1.20–Go1.27 x86 reports were inspected.
Exactly 53 AMD64 input-deficient probes per SDK changed supported to Context,
limited to thirteen ABS/sign names and SHLW/SHRW; every 386 form and unrelated
encoder/unsupported/parse metric stayed unchanged. Its current-SDK AMD64 counts
were 3,239 supported, 3,497 Context and zero unsupported. These are historical
reference counts, not a classification of `798bff706`. Context is not an
initialized-program success. Retain the historical Go1.20 unsupported form;
the current five-architecture gate still requires zero unsupported/parse errors.
Re-audit complete forms and fingerprints before changing any baseline.

The last strict standard-library benchmark at `18bb77a09` remains failed:
184 target/file pairs attempted, 129 compiled, 55 failed, no N/A. AMD64 took
392s, exceeding its unchanged 300s budget; total 504s is below 900s. No newer
complete pass or performance improvement is qualified.

## Next work and verification

1. Finish the independently assigned four-source and raw-variable families.
   Integrate completed rule-class repairs without overlapping active/frozen
   worktrees. Preserve original initialized programs and independent oracles.
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
