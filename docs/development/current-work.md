# PR40 closeout

Updated 2026-10-29. Finish **amd64 correctness, shared typed families and
translation efficiency first**, preserving all five architecture gates.
Start with [architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).
This page describes the active work; earlier qualification history remains in
git and the original ignored reports.

## Published revision

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft at
`18492758d`. Its [complete CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40820775107)
passes **325/325 jobs**, with **88.1%** statement coverage and **95.07%**
Codecov patch coverage against an 87.83% target. Every expected job name and
actual job/check ID is accounted for at this exact head.

Fresh Windows and race aggregates each contain all 22 root-module packages.
Windows retains 278 root skips; race retains 299 root / seven child skips.
Zero skipped jobs does not mean zero skipped tests. These skips do not qualify
native386, AVX512 or unexecuted semantics. External approval remains outstanding.

| Published inventory | Count and meaning |
|---|---|
| Family views | 42, including explicit overlapping views |
| AVX source-row ownership | 3,147 / 3,568 (88.20%) |
| Mixed-mode source records | 3,676, including 529 separate legacy/fallback records |
| Observed canonical witness union | 12,856 |
| Source candidates | 12,861; not the observed-object denominator |
| Source/declaration batches | 462 official plus four existing runtime batches |
| Separate PADDD aliases | Two; not extra encoder rows |

Ownership retains first-match, shadow and terminal observations; it does not
assert every alternative is selectable by named Go syntax or executes normally.
The AVX percentage is neither an all-amd64 nor a five-architecture percentage.
Legacy machine-candidate rows have a separate inventory: 1,306 on current Go,
1,305 on minimum Go, plus 112 separate `ymovtab` fallback records.

## Latest complete families

K-data owns 59 names / 87 rows / 91 witnesses in both source modes. Its single
normalized mask-move form serves grammar, effects, result analysis and lowering.
The mode32 GP-Q repair preserves 64-bit K/K and memory transfers. Each endpoint
SDK independently executes 3,888 integer checks, 432 owned memory-fault processes
and 72 complete-page canary checks. Portable host execution is separate from
native386 and AVX512 execution.

Insert/extract owns all 82 original records: 79 selected rows / 194 ordinary
witnesses and three fully shadowed alternatives. Twelve original-Go memory `.Z`
forms are architectural #UD, not numerical results. They emit UD2/unreachable
before instruction inputs, while preserving obligations of the source prefix
and every reachable ordinary return. Signature contexts do not add physical
forms. No general pointer, source or unsealed raw-entry authority follows.

The remaining-move join owns 101 records / 119 candidates: 105 selected ordinary
witnesses and 14 shadow alternatives. Mode32's 54 low-bank rows / 58 witnesses
are subsets of the AMD64 union. Forty-seven high-input cases retain independent
Go acceptance/refusal, physical/grammar and numerical-input boundaries.
Existing high-X15/Y15 Void-format objects remain format evidence.

MOVHLPS/MOVLHPS and their modern counterparts share one existing typed form
through effects, result analysis and lowering. Modern writes clear bytes 16–63;
legacy writes preserve them. Eight modern named/raw kernels fail with the
controlled old lowerer. Final independent byte/canary execution passes two roots
/ 92 children on each SDK, through three APIs and three pipelines. The move
translator changes by net minus three lines; cold row machinery adds 348 lines.

The final move-family scope is 24 roots / 496 children on each SDK: current Go
inherits 22 unchanged owners and freshly runs two final owners; minimum freshly
runs the complete scope. Full rules pass 107 / 4,328 on both SDKs. All eighteen
old41 JSON snapshots retain every original field, byte, source and declaration.
Five current-SDK architecture reports, Node573 and Go26 / 301 pass. Minimum's
five-architecture script was not rerun by that component.

## Local follow-up

The common terminal-prefix repair is committed separately as `6ffc40ab5` and
integrated locally as `ed6ac1c34`; it is not in the published revision above.
Eight existing terminal mechanisms retain prior source obligations before
their final trap. Ordinary reachable returns still require their own result
writes. The shared repair changes four production files by net plus 16 lines;
it adds neither another opcode table nor storage permission.

Both endpoint SDKs pass 35 roots / 1,088 children, retaining one real local
host-inapplicable native root. A separate official Linux/amd64
[fork staging run](https://github.com/cpunion/plan9asm/actions/runs/40828444104)
executes that root: one root / 25 children pass without skips. Sixteen original
Go cases and 144 C cases observe the required SIGSEGV/SIGILL ordering. This
qualifies this fault contract, not all SIMD semantics or AVX512 hardware. The
single-use staging workflow does not belong in PR40.

The qualified cold join is committed as `c17154be8` and integrated locally as
`a2638a7b6`. It covers the complete 32-name / 76-row conversion family.
It reuses four shared transports. Original Go selection separates 287 ordinary,
three terminal and eight partially shadowed candidates. Both endpoint SDKs
pass 30 modules / 60 LLVM22 O0/O2 objects, format controls and the complete
109-root / 5,088-child rule scope. All twenty old42 snapshots remain identical.
Both SDKs pass the five-architecture reports. Node573 and Go26 / 301 policies
also pass on the final frozen source.
An interrupted minimum-SDK link ran zero tests because of disk exhaustion;
its retained setup failure is followed by the successful four-root / 372-child
retry. These local results do not certify a future published CI revision.

The local union is **43 views / 3,752 mixed-mode source records**, including
**3,223 / 3,568 AVX rows (90.33%)** and the same 529 legacy/fallback records.
Observed canonical keys total **13,146 = 12,856 + 287 ordinary + three terminal**;
the eight partially shadowed candidates add no key. Source/declaration batches
total 472 official plus four runtime; two aliases remain separate.

## Architecture and efficiency

Scale implementation with legal format groups and semantic families, not a
product of source opcodes, target opcodes and formats. LLVM chooses target
instructions from portable IR. Preserve distinct operand roles, active/carrier/
memory/result widths, encoding, masks, alias writes, FLAGS and fault behavior.
Typed forms carry those facts through effects, result analysis and lowering.
Source binding and memory authority remain separate from a legal grammar.

Existing cleanups share packed physical formats and GP byte/FLAGS effects,
delete unreachable legacy lowerers, and reuse original source/tool observations
within a bounded translation. Module declarations are selected from actual
references. Do not cache mutable source verification or introduce another opcode
registry. Parser/allocation improvements and encoder measurements are not
complete SDK translation or generated-program speedups.

A prepared raw-prefix cleanup would replace 44 identical FS/GS/address-size
scanners with one observer, net minus 816 lines. It is only statically reviewed:
baseline decoder tests, candidate qualification and paired measurements remain.
A previous result-routing reduction was withdrawn after minimum-SDK benchmarks
regressed. Do not restore it or claim an unmeasured acceleration.

## Remaining work, in order

1. Publish a qualified batch with the integrated conversion76, terminal-prefix
   repair and raw-prefix cleanup. Require the new head's complete CI,
   coverage and review; keep all original failures, setup errors and skips.
2. Finish the other 345 AVX rows in eleven semantic groups using seven transport
   groups. The existing-helper tests-first draft is unrun. Floating arithmetic,
   VSIB, comparison, horizontal/blend/dot/SAD, BMI, crypto, four-register, string,
   bitshuffle-to-K, vector-to-FLAGS and state operations keep distinct semantics.
   Complete original row ownership is separate from runtime acceptance.
3. Qualify the raw-prefix cleanup and measure complete translation separately.
   Preserve original bytes, source replay, mode/segment/address-size behavior,
   negative boundaries and resource disposal. Do not reduce test scope to obtain
   a faster result.
4. Integrate a qualified public translator into
   [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) through the allowed fork.
   Its existing `b4381f2de` round is complete: 140 successful jobs and two
   conditional publication skips across thirteen workflows, patch coverage
   97.06%. It still pins `bebc7092d`; a new pin requires normal public module
   verification and fresh actual compile/link/run owners. The official upstream
   dependency review requirement remains unresolved by a fork pin.
5. Preserve other architectures and explicit gaps: 271 ARM64 decoder forms,
   remaining WASM control/SIMD, ARM64-to-WASM `llvm.returnaddress` object failure,
   and native AVX512/RCP14/MXCSR acceptance. Six MMX rows retain a separate
   machine-state contract. F32/F64 MIN/MAX NaN, signed-zero, MXCSR and fault
   execution and BF16 optimized active-memory faults need independent owners.
6. Keep fourteen strict-SDK entry, GC-scan and runtime/test-hook failures tracked
   in [issue47](https://github.com/xgo-dev/plan9asm/issues/47). Runtime callbacks,
   system-stack, arbitrary function-value, general ABI0, full SQLite and purego
   contracts are separate integration work; object success does not prove them.

External scanning, ledgers and package compatibility belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run local
inventory or restore those jobs here. Push only allowed forks, preserve draft
status and coordinate one heavy local Go/LLVM/Node invocation at a time.
