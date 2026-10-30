# PR40 closeout

Updated 2026-10-30. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates.
Use [architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).
Original failures, source identities and host exclusions remain in their reports.

## Published CI and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft. Public
`70ced1b565` has its own [CI round](https://github.com/xgo-dev/plan9asm/actions/runs/41032133785),
which completed naturally with 325/325 successful jobs. Independent acceptance
joins all 286 source artifacts, 240 partitions, Windows/race package logs and
22 coverage profiles: 88.1% total and 92.52% patch coverage. Both review threads
are resolved; no formal approval has been recorded. These results certify this
head's CI, not the unpublished follow-up or unavailable native hardware.

The preceding public `67c4446bcc` completed 325/325 jobs. Those results and its
coverage belong to that head. The original `798cf56ee`
[round](https://github.com/xgo-dev/plan9asm/actions/runs/40999892704) ended with
191 successful and 134 failed jobs; repaired local tests do not change those
failures into passes. Older completed rounds remain separate historical evidence.

The local follow-up includes signed metadata (`612b6d1f0`), x87 source rows
(`924458b6b`), typed headers (`4581004bf`) and ten-byte input checks (`6805b9362`).
These newer commits require their own cloud CI and accepted LLGo dependency pin.

## Local qualification

Public `70ced1b565` replaces temporary LLVM IR files with copied memory buffers.
Unique buffer identities preserve independent default ModuleID/source filenames;
explicit IR source filenames remain intact. The fresh 148-owner scope passes
12,091 nodes per endpoint SDK, with three explicit host-inapplicable runtime
skips. Both SDKs pass the five-architecture gates; Node policy passes 573 tests
and Go policy 327 nodes. These results belong to `70ced1b565`.

CLI evidence retains 2,670 unchanged inherited nodes plus six fresh boundary
nodes at `70ced1b565`; it is not a complete CLI rerun there. The CLI modules'
minimum is Go 1.24, so their results do not establish Go 1.20 CLI execution.
Minimum-Go PCALIGN retains its exact observed assembler defect as a distinct
outcome, never a successful Go object; all three LLVM target objects remain
required for that owner.

The signed-locals repair preserves canonical uint64 representations of the
writer's int32 locals. Its frozen boundary checks pass 73 nodes with current
Go and 49 with minimum Go. Negative locals remain metadata, with no new frame
size, entry, bounds or local-backing permission.

The x87 source/object closure has 1,056 fresh passing nodes per SDK. Its
1,750-node initialized/state/object scope is inherited only across the small
source-identity output hook, not rerun after it. The prior 30 JSON identities
remain byte-exact. Source formatting and LLVM objects are not numerical or
native x87 execution evidence.

The ten-byte family closes 72 reproduced incorrect acceptances. Its five focused
roots pass 177 nodes per SDK; unknown extended80/BCD precision remains Context.
At integrated `6805b9362`, 398 current-Go roots pass 25,039 nodes with nine
explicit host skips; minimum Go passes 24,485 nodes across 354 roots with eight
host skips. Exact raw test-name/state unions and source/tool hashes are checked.
These are bounded regression results, not full-SDK or skipped-native acceptance.

At `4581004bf`, the shared ordered native-header parser passes 71 focused nodes
with current Go and 67 with minimum Go, without failures or skips. The broader
frame/SDK/Windows/private-leaf scope passes 1,118 nodes across 118 roots with
one explicit runtime skip requiring an AMD64 host. Minimum Go passes 568 nodes
across 75 roots without skips. The other 43 roots have current-SDK build constraints,
not minimum-SDK passes or skips. Header kind, flags, numeric bounds, field
presence and alignment layout are retained; special or missing entry roles
remain Context. Native flag admission remains closed to its existing roles.

The original strict 55-source run at `612b6d1f0` still reports 41 passes and
14 failures, with no unattempted sources or newly passing files. Three runtime
TEXT diagnostics advance to the native-header boundary. Header qualification
has not yet established a new complete strict55 outcome. Track the failures
individually in [issue47](https://github.com/xgo-dev/plan9asm/issues/47).

## Source inventory and semantic boundaries

The joined inventory has 51 overlapping family views and 4,379 source records:
3,568 AVX rows plus 811 legacy records (793 optab and 18 ymovtab). The legacy
denominator remains 1,418 current-SDK / 1,417 minimum-SDK records, including
112 ymovtab records. Its 607 / 606 unjoined records are inventory gaps, not
unsupported-instruction totals. Snapshot-file counts are not family-view counts.

The x87 join retains 79 original row identities in each x86 mode, 158 canonical
witnesses and 316 mode instances. Eight first-match shadows retain their own
identity, giving 166 actual-Go byte cases per mode. This is source-table closure,
not native extended80, packed BCD or arbitrary caller-state support. The safety
repair retains all 32 original FBLD/FBSTP statements across eight source profiles;
only 24 reviewed classification fields change. Source profiles are not eight
executed compiler versions. Encoding or pointer knowledge grants no precision.

Typed families preserve source roles through grammar, effects and lowering.
MULX keeps implicit DX and both outputs; count, scan, compare and memory RMW
retain complete input obligations even for Void or dead results. Nontrapping
dead register scratch remains allowed without manufacturing numerical bytes.
VSIB index NUM, data/merge bytes, masks and pointer bases remain independent.
Masking and SAE do not justify eager reads or a larger memory tuple.

ADDSUB/ROUND retain exact packed or scalar tuples, copied upper bytes and
machine-state effects. Physical legacy memory uses its original opcode;
logical paths retain their mathematical emitters. A scalar 4/8-byte access does
not become a 16-byte carrier read or acquire a new alignment guarantee.

Raw normalization replays its sealed source before deriving another lowering
view. Original bytes, operands, mode, prefixes and literal tails remain intact.
Bare uninitialized inputs retain Context alongside independent initialized
three-API/O0/O2 objects. FS/GS reads retain their physical segment and actual
numerical offsets. Static XADD/XCHG require actual module-owned writable numeric
storage and sufficient extent; pointer values do not become NUM. Stack positives
use actual source-bound frames, not rewritten raw or AUTO permissions.

SDK capture uses bounded directory membership, relative headers and complete
compiler/assembler object joins. It preserves namespace/index identity, every
DUPok body and actual initialized or zero-filled storage facts without choosing
a winning body or granting entry permission. Mutation, closure, aliases and
changed source/profile metadata revoke the original observation. Do not cache
mutable source verification or treat parsed header attributes as authority.

## Performance and remaining work

The unchanged 29-case, two-SDK, three-block ABBA comparison of `70ced1b565`
against `67c4446bcc` retains all 696 samples. No case median reaches a five
percent regression; all case medians are below baseline. This measures complete
register-only Translate, excluding source Parse and live SDK capture. It proves
neither whole-SDK throughput nor generated-code acceleration. Earlier failed
attempts and historical performance comparisons remain separate evidence.

1. Finish the follow-up's policies, CLI, strict55 and efficiency checks. Publish
   only through the allowed fork, then require its own complete CI, coverage and
   fresh review. Preserve original failures and host skips.
2. Continue the 607 / 606 legacy inventory using original prefixes, modes and
   selection order. Keep row ownership, initialized objects and native execution
   separate, with at most 32 complete functions per source/object batch.
3. Keep ten-byte safety distinct from extended80/BCD numerical semantics.
   Retain native386, AVX512/AVX512ER, RCP14/MXCSR, MMX-state,
   BF16 active-memory fault and floating MIN/MAX execution gaps.
4. Preserve ARM64, ARM and WASM object/control boundaries and all 14 strict-SDK
   failures. Special runtime entries, callbacks and library compatibility need
   their own source contracts and independent compile/link/run owners.
5. Integrate an accepted public translator into
   [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) with normal public-module
   verification. A fork pin does not resolve the P1 official upstream dependency
   requirement. Coordinate one heavy local Go/LLVM/Node invocation at a time.
