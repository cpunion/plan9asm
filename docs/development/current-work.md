# PR40 closeout

Updated 2026-10-29. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates.
Start with [architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).
Original qualification, failures and host exclusions remain in their reports.

## Published and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft. The published
batch `fad14e2e1` completed its [own round](https://github.com/xgo-dev/plan9asm/actions/runs/40872759334)
with 280 successful and 45 failed jobs. The failures comprise formatting,
eleven jobs each for private MULX context, bare raw ADDSUB and legacy ROUND
memory, plus eleven derived collectors. Local repairs require their own CI.
The public baseline `5616f67d8` completed its
[own round](https://github.com/xgo-dev/plan9asm/actions/runs/40851464196)
with 325/325 successful jobs. Those results certify that baseline. Every newly
published head needs its own complete CI, coverage and external review.
The previous `18492758d` [round](https://github.com/xgo-dev/plan9asm/actions/runs/40820775107)
passed 325/325 jobs with 88.1% statement coverage; those results belong to that
previous head. External review remains outstanding.

The local follow-up has the following source-table ownership:

| Local inventory | Count and meaning |
|---|---|
| Family views | 48; overlapping views retain explicit identity |
| AVX source rows | 3,568 / 3,568 original table rows |
| Mixed-mode source records | 4,173 = 3,568 AVX + 605 legacy/fallback records |
| Historical canonical witness union | 14,560; separate from all runtime forms |
| New source witnesses | 1,414, checked with actual Go and three API O0/O2 objects |
| New bounded source/declaration batches | 44; not a function or object count |
| Inherited identity | All 22 old JSON snapshots byte-exact |

Complete AVX table ownership preserves original SDK rows, first-match choices,
shadow alternatives, source modes and terminal forms. It does not establish
all-ISA support, every named syntax alternative or native numerical execution.
Legacy has its own denominator: 1,306 current-SDK / 1,305 minimum-SDK machine
candidates, plus 112 separate `ymovtab` records. The 605 joined records comprise
587 `optab` and 18 `ymovtab` entries; 813 current / 812 minimum records remain
unjoined. These are inventory gaps, not counts of unsupported translations.
Two PADDD aliases remain separate from rows.

The new CMOV join adds all 16 conditions across W/L/Q: 48 original rows and
96 canonical register/memory witnesses. The 386 subset has 32 rows / 64
witnesses; the 16 Q rows retain actual frontend refusals. The original 22
identity snapshots remain byte-exact. No CMOV translator or FLAGS rule changes.

The remaining non-CMOV `yml_rl` closure adds 28 original legacy rows through
shared scalar prefix/ModRM transport: 56 AMD64 witnesses and the independently
observed 34-witness 386 subset. The eleven REX.W rows retain actual Go source
refusals. All 24 preceding identity snapshots remain byte-exact.
LAR/LSL selector reads stay m16; extension source and result widths remain
distinct. BSF/BSR now share their existing strict result step with the paired
exact-width volatile reader. Continued execution supplies only that input's
numerical bytes; it supplies no reusable memory, bounds or caller permission.
Zero-input destination knowledge and portable ZF-only observation stay intact.
Void and unused bool cannot hide missing scan inputs. Native numerical
execution remains separate from initialized source and LLVM object evidence.

## Current class repairs

The remaining formats share typed source, result, mask, carrier and tuple roles.
GP/BMI objects retain actual source selection and mode boundaries. MULX now
preserves implicit DX and both output dependencies through observable results;
memory, raw and physical-frame roles keep their existing proof requirements.

Gather, scatter and prefetch share one VSIB grammar through lowering and result
analysis. Index NUM, data/merge bytes, mask state and the original pointer base
stay independent. Active gather loads remain volatile so a dead result cannot
erase its fault. Mode32 high data/mask banks retain separate numerical rejection;
legacy Void-format objects do not certify hardware execution or native entry.
The paired access grants no contiguous extent or later-read authority.

Every data gather/scatter now enters complete input validation even with a
dead Void result. Eight source profiles retain exactly 128 classification
changes over 517 original statements: 209 unclosed effects, 272 undefined
inputs and 36 SP/AUTO contexts. No prefetch, unsupported entry or other
architecture changes. These probe wrappers have no initialized prefix or
bounded original TEXT. Their Context outcomes do not erase initialized objects.

The ten-member / twenty-row approximation family reuses the existing native FP
helper, removing 227 production lines. Masking, SAE and MXCSR remain owned by
the original instruction; source operands are not eagerly read before masking.
This is a code-size change, not a measured performance or AVX512ER runtime gain.

ADDSUB and ROUND share typed grammar, effects, result analysis and lowering.
Their existing mathematical emitters are unchanged. Observable VEX reads retain
the original packed or scalar tuple when results die under optimization;
scalar 4/8-byte access widths do not become 16-byte X carriers. Legacy reads do
not borrow that VEX native-read premise or a new alignment/fault guarantee.
Go scalar ROUND uses Yu8 and rejects ordinary mode32 four-operand syntax;
normalized raw physical forms keep their own identity and NUM restrictions.

Physical legacy memory now executes the original SSE ADDSUB/ROUND opcode with
its exact memory tuple, immediate, copied upper bytes and machine-state effects.
This supplies numerical bytes only on continuation, with no extent or later-read
permission. Its SSE3/SSE4.1 carrier, matching target and source inputs remain
explicit. Logical FP and register paths retain their mathematical emitters.
Initialized raw ADDSUB uses the existing exact source identity and whole-source
replay; the bare original bytes retain public Context and private format objects.

Repeated raw normalization checks the existing sealed checkpoint before making
a new lowering view. Text fallback starts from that source rather than an
already rewritten address view. Default AMD64 and explicit AMD64 remain the
same decoding mode without changing retained source spelling; 386 stays distinct.

ROUND RIP literals reuse the existing exact decoder, source replay, DATA and
raw-return binding. Self-contained packed literals retain public translation.
Bare scalar upper inputs remain Context; genuinely initialized raw and named
counterparts retain positive objects. Edited source, DATA or read metadata
cannot borrow the original proof. No generic raw or incoming-pointer permission
is added. The original fixture bytes and literal tails remain observable.

## Qualification boundary

Go 1.27.3 and Go 1.20.14 use LLVM22. Each passes the local repair scope of
35 roots / 1,268 children without skips. The complete original ROUND runtime
assembly and opaque five-pointer frame also pass a separate root / 30 children
through three targets and three APIs at O0/O2. The original assembly, native
guard and C driver remain byte-identical. This is source/object evidence;
actual native execution retains its existing independent runtime owner.

The complete cold rule package passes 119 roots / 10,699 children per SDK.
The original candidate five-architecture validators fail on the new VSIB
classification until its independently audited 24 baseline fields are updated.
All original failures remain retained. Run the merged broad scope, normative
five-architecture checks, Node and Go policies, whole-source regression and
paired translation measurements before publishing. Component results certify
neither the next head's CI nor its unexecuted hardware.

The terminal-prefix class retains all prior input obligations before UD2 and
all reachable ordinary return obligations. Its official Linux/amd64
[staging run](https://github.com/cpunion/plan9asm/actions/runs/40828444104)
passed one root / 25 children without skips. Portable fault/canary execution,
LLVM object success and actual native execution remain separate evidence.
Original semantic REDs, fixture/setup failures and ordinary skips are retained.

The qualified raw-prefix cleanup shares 44 identical scanners, removing 816
production lines. Paired decoder measurements show no stable speedup and no
allocation change. A prior result-routing candidate was withdrawn after a
minimum-SDK regression. Complete translation needs its own paired measurement;
neither a smaller helper nor a microbenchmark proves faster SDK compilation.

## Remaining closeout

1. Finish the merged qualification and the current public CI round. Publish
   repairs only through the allowed fork after that round finishes; require
   fresh complete CI, coverage and review at the new head.
2. Continue legacy families using original literal prefixes, modes and selection
   order. Keep source-row ownership distinct from complete semantic acceptance.
   Do not add another opcode registry or cache mutable source verification.
3. Integrate the accepted public translator into
   [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) with normal public-module
   verification and fresh compile/link/run owners. A fork pin does not resolve
   the P1 official upstream dependency requirement.
4. Preserve native386, AVX512/AVX512ER, RCP14/MXCSR and other explicit hardware
   gaps. MMX state, BF16 optimized active-memory faults, floating MIN/MAX and
   unexecuted native owners require their own acceptance evidence.
5. Preserve ARM64, ARM, WASM and their known object/control gaps. The fourteen
   strict-SDK failures in [issue47](https://github.com/xgo-dev/plan9asm/issues/47),
   general ABI0/runtime callbacks and broader library compatibility remain
   separate work. Coordinate one heavy local Go/LLVM/Node invocation at a time.
