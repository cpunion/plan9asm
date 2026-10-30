# PR40 closeout

Updated 2026-10-30. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates.
Start with [architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).
Original qualification, failures and host exclusions remain in their reports.

## Published and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft. Published
`3551c62a5` has its [own round](https://github.com/xgo-dev/plan9asm/actions/runs/40925324365);
acceptance requires all 325 jobs, original package logs, coverage and review.
The local follow-up `c6bedd6f2` adds the 127-row batch below and has no public
CI qualification yet. Let the published round finish before another push.

Previous `137945fdf` completed its [own round](https://github.com/xgo-dev/plan9asm/actions/runs/40904358535)
with 292 successful and 33 failed jobs: two source-test owners across eleven
profiles, plus eleven dependent collectors. The tests mutated private access
certificates but expected normalization to accept the edited sealed snapshot.
The tests-only repair retains rejection and adds fresh original-source capture
and LLVM-object positives. Original failed reports remain failed evidence.

The preceding batch `fad14e2e1` completed its [own round](https://github.com/xgo-dev/plan9asm/actions/runs/40872759334)
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
| Family views | 50; overlapping views retain explicit identity |
| AVX source rows | 3,568 / 3,568 original table rows |
| Mixed-mode source records | 4,300 = 3,568 AVX + 732 legacy/fallback records |
| Previous canonical witness union | 14,560; separate from all runtime forms |
| New 127-row batch | 208 AMD64 witnesses; independently observed 104-witness 386 subset |
| Source/object batching | At most 32 complete functions per original-Go/LLVM batch |
| Inherited identity | All 26 preceding JSON snapshots byte-exact |

Complete AVX table ownership preserves original SDK rows, first-match choices,
shadow alternatives, source modes and terminal forms. It does not establish
all-ISA support, every named syntax alternative or native numerical execution.
Legacy has its own denominator: 1,306 current-SDK / 1,305 minimum-SDK machine
candidates, plus 112 separate `ymovtab` records. The 732 joined records comprise
714 `optab` and 18 `ymovtab` entries; 686 current / 685 minimum records remain
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

The next closure adds 96 MMX/SSE and 31 CMP/TEST rows without overlap in the
previous source-record union. Shared carrier, direction, prefix and immediate
transports derive bytes from the original tables. Their 58 named 386 refusals
are observed with the actual Go assembler. The default packed formatter keeps
its MMX exclusion; the explicit carrier projection grants no native inputs.
All initialized witnesses retain original Go bytes and three-API, five-target
O0/O2 objects. The formatter uses 238 new cold lines and a small shared scalar
extension; the translator adds only three lines of input-proof routing.

The CMP/TEST repair requires the existing complete proof for a faulting memory
read even when FLAGS die. Nontrapping dead register scratch remains allowed
without manufacturing numerical bytes. Known-pointer reads retain their exact
volatile widths after optimization. Across eight source profiles, 23 AMD64
forms / 70 original statements and two 386 forms / two statements now require
context. All 576 original statement observations retain their identities;
only 48 reviewed support/context/fingerprint baseline fields change. Other
architectures, unsupported entries, runtime and encoder inventories are intact.

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

The complete cold rule package passes 128 roots / 12,083 children per SDK.
The final 127-row family scope passes 1,520 nodes across 29 roots / 1,492
children, with one existing MMX host-inapplicable runtime skip. Both complete
five-architecture gates pass; Node policy passes 573 tests and Go policy passes
327 nodes without skips. All 26 preceding identity snapshots remain byte-exact.
The preceding BSF/BSR repair changed twelve bare probe forms from IR
acceptance to Context: all eight source profiles retain the same 48 original
statements, now rejected for undefined inputs. Initialized counterparts retain
three-API, five-target objects. The reviewed baseline changes only the support,
context and fingerprint fields; unsupported and runtime counts do not change.
Retain the original failing reports and rerun the normative five-architecture
checks and policies on the integrated head. Component results certify neither
that head's CI nor its unexecuted hardware.

SDK-primary inventories now reuse the existing complete bounded directory
snapshot. Ordinary packages retain their flat 256-member limit; SDK file
bytes retain the 16 MiB budget. Capture, verification and frame cloning retain
the same physical membership observation. Member mutation, closure, aliases
and shared graph limits remain checked. Current Go passes 148 roots / 1,673
children; minimum Go selects 72 roots / 1,373 children, all passing without
skips. The other 76 roots are current-SDK build-tag owners, not minimum passes.
The strict 55-source benchmark still reports 41 passes and fourteen failures.
Five diagnostics progress from directory capacity to an unsupported relative
SDK include; generic Void ABI barriers remain. This is no full-SDK acceptance.

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

The final `3551c62a5` comparison uses the unchanged 29-case harness, three ABBA
blocks and both SDKs against `5616f67d8`: no median or three-block sustained
regression reaches five percent. Timing directions and allocations remain mixed.
These measure complete register-only translation, excluding Parse and live SDK
capture. They establish neither whole-SDK nor native generated-code acceleration
and do not qualify the later local head.

## Local GP and SDK repairs

The GP memory-read repair, hardware-segment backend guard, six-member counter
family and static bit-string effects are locally committed. Their precise
241-root regression union passes 104,095 nodes per endpoint SDK, with four
existing host-inapplicable runtime skips retained separately. These results
are not a new public-head CI or native x86 execution certificate.

SDK capture now observes bounded relative headers through the shared source
path implementation. Compiler object limits follow actual definition namespaces,
exact index sentinels and physical payload blocks within the original byte
budget. Same-member Nonpkg DUPok definitions keep every original index. An
empty compiler BSS declaration may accompany a same-size ASM initializer;
collision checking retains both records and grants no winning-body, numerical
or storage permission. Two initialized non-DUPok bodies still fail.

The complete static-storage package passes 1,768 nodes with Go 1.27.3 and 480
with Go 1.20.14, without skips. Current-Go actual runtime capture, verification
and source-revocation tests pass for both amd64 and 386. These are complete
source/object observations, not execution of the translated runtime.

The unchanged strict 55-file benchmark still reports 41 passes and fourteen
failures, without new failing files. Six diagnostics advance to subsequent
TEXT, frame, ABI or literal-initialization checks. Its 169-second complete run
is slower than the historical 116-second comparison; SDK throughput remains
unaccepted. The local current-Go profile package also retains one Rosetta
preflight failure on an ARM64 host; no program executed in that failed owner.

## Remaining closeout

1. The published Plan9asm and LLGo pin have completed their CI rounds. Publish
   the later local repairs only through the allowed fork after final gates;
   require fresh complete CI, coverage and review at each new public head.
2. Finish final GP classification review across five architectures and eight
   source profiles, the policy suites and the unchanged complete-translation
   benchmark. Do not replace initialized object or execution evidence with
   bare instruction Context outcomes. Preserve the original REDs and skips.
3. Continue legacy families using original literal prefixes, modes and selection
   order. Keep source-row ownership distinct from complete semantic acceptance.
   Do not add another opcode registry or cache mutable source verification.
4. Integrate the accepted public translator into
   [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) with normal public-module
   verification and fresh compile/link/run owners. A fork pin does not resolve
   the P1 official upstream dependency requirement.
5. Preserve native386, AVX512/AVX512ER, RCP14/MXCSR and other explicit hardware
   gaps. MMX state, BF16 optimized active-memory faults, floating MIN/MAX and
   unexecuted native owners require their own acceptance evidence.
6. Preserve ARM64, ARM, WASM and their known object/control gaps. Classify the
   fourteen strict-SDK failures in [issue47](https://github.com/xgo-dev/plan9asm/issues/47)
   individually: the repaired include and DATA parsing paths expose further
   TEXT, literal-initialization and generic ABI gaps, alongside the separately
   tracked special runtime contracts. None of these failures is a pass.
   Broader callbacks and library compatibility retain independent execution
   owners. Coordinate one heavy local Go/LLVM/Node invocation at a time.
