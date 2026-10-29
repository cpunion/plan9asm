# PR40 closeout

Updated 2026-10-29. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates.
Start with [architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).
Original qualification, failures and host exclusions remain in their reports.

## Published and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft. The published
batch `fad14e2e1` has its [own round](https://github.com/xgo-dev/plan9asm/actions/runs/40872759334)
in progress. Its formatting failure is repaired locally while the remaining
jobs finish. The public baseline `5616f67d8` completed its
[own round](https://github.com/xgo-dev/plan9asm/actions/runs/40851464196)
with 325/325 successful jobs. Those results certify that baseline. Every newly
published head needs its own complete CI, coverage and external review.
The previous `18492758d` [round](https://github.com/xgo-dev/plan9asm/actions/runs/40820775107)
passed 325/325 jobs with 88.1% statement coverage; those results belong to that
previous head. External review remains outstanding.

The published batch has the following source-table ownership:

| Local inventory | Count and meaning |
|---|---|
| Family views | 46; overlapping views retain explicit identity |
| AVX source rows | 3,568 / 3,568 original table rows |
| Mixed-mode source records | 4,097 = 3,568 AVX + 529 legacy/fallback records |
| Historical canonical witness union | 14,408; separate from all runtime forms |
| New source witnesses | 1,262, checked with actual Go and three API O0/O2 objects |
| New bounded source/declaration batches | 40; not a function or object count |
| Inherited identity | All 22 old JSON snapshots byte-exact |

Complete AVX table ownership preserves original SDK rows, first-match choices,
shadow alternatives, source modes and terminal forms. It does not establish
all-ISA support, every named syntax alternative or native numerical execution.
Legacy has its own denominator: 1,306 current-SDK / 1,305 minimum-SDK machine
candidates, plus 112 separate `ymovtab` records. The 529 joined records are not
the complete legacy inventory. Two PADDD aliases remain separate from rows.

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

ROUND RIP literals reuse the existing exact decoder, source replay, DATA and
raw-return binding. Self-contained packed literals retain public translation.
Bare scalar upper inputs remain Context; genuinely initialized raw and named
counterparts retain positive objects. Edited source, DATA or read metadata
cannot borrow the original proof. No generic raw or incoming-pointer permission
is added. The original fixture bytes and literal tails remain observable.

## Qualification boundary

The frozen combined wide scope passes both endpoint SDKs at `fad14e2e1`.
Go 1.27.3 records 172 roots / 9,734 children with 9,899 passes; Go 1.20.14
records 171 roots / 9,731 children with 9,895 passes. Each retains seven explicit host exclusions
and no failure. The extra current-SDK x87 mode32 root is absent from the minimum
SDK's actual built inventory. Both commands exited naturally with exact,
unique terminal outcomes. Six legacy members plus four typed-FP counterparts
also pass five targets / three APIs / O0/O2: one root / 20 children per SDK.
Both complete five-architecture validators pass after reviewing the 57 amd64
classification changes. The eight profiles retain their actual fingerprints;
no encoder or runtime counts are promoted. Node573 and Go26 / 301 policies pass.
The paired translation measurements cover 29 cases in three ABBA blocks per
SDK; no case has a sustained regression of at least 5%. This does not establish
a general SDK or native-runtime speedup.
Component results do not certify the next published head's CI.

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

1. Finish the current public CI round and repair its observed failures. The
   local five-architecture reports, paired translation measurements and policy
   checks are complete. Publish repairs only through the allowed fork after
   the running round finishes; require fresh complete CI, coverage and review
   at the new head.
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
