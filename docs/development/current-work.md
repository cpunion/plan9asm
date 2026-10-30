# PR40 closeout

Updated 2026-10-30. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates. Read
[architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).

## Published CI and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft at
`d6cb9245cd`. Its [CI round](https://github.com/xgo-dev/plan9asm/actions/runs/41055685183)
first ended with 315 successful and ten failed jobs. Seven Go 1.21 source jobs
lost their runners; one CLI collector lost an artifact download, producing two
dependent collector failures. Recovery completed with 325/325 successful jobs:
307 original successful jobs remain valid and 18 jobs actually reran. Independent
acceptance joins all 286 source artifacts, 240 partitions, Windows/race raw
outcomes and 22 coverage profiles: 88.1% statements and 93.44% patch coverage.
The initial failures remain part of the evidence.

Historical `70ced1b565` completed its own
[round](https://github.com/xgo-dev/plan9asm/actions/runs/41032133785) with 325/325
successful jobs. Independent acceptance joins 286 source artifacts, 240
partitions, Windows/race package logs and 22 coverage profiles: 88.1% total and
92.52% patch coverage. These results do not certify another head. Both review
threads are resolved; no formal approval has been recorded.

The local follow-up at `aff58904d` adds x87 MOVE, literal SIMD and scalar
source joins.
It requires its own publication, cloud CI, coverage and review. The preceding
head's recovery is accepted; publish only through the allowed fork.

## Qualified local batches

Public `d6cb9245cd` combines copied LLVM memory buffers, canonical signed frame
metadata, one ordered native-header parser and ten-byte x87 input checks.
The integrated functional snapshot passes 25,039 current-Go nodes with nine
explicit host skips and 24,485 minimum-Go nodes with eight. The skips do not
establish native execution. Headers retain kind, flags, field presence, numeric
bounds and alignment without granting special or missing entry roles authority.

The ten-byte repair closes 72 reproduced incorrect acceptances. All 32 original
FBLD/FBSTP statements remain; only 24 reviewed classification fields change.
Extended80/BCD numerical precision remains Context. Source profiles are not
eight executed compiler versions, and known addresses do not establish precision.

At `d6cb9245cd`, complete current-Go CLI runs pass 77 plan9asm and 2,599
plan9asmll nodes without failures/skips. CLI modules require at least Go 1.24;
these results do not claim Go 1.20 CLI execution. The unchanged 29-case,
two-SDK, three-block ABBA comparison against `70ced1b565` retains 696 samples.
No case median regresses by five percent; one minimum-Go block ratio
exceeds that threshold. This measures register-only Translate, excluding Parse
and live SDK capture, not whole-SDK throughput or generated-code acceleration.

The unpublished x87 MOVE join at `d20cfaab5` retains all 45 original rows across
five tables, of which 27 overlap the existing x87 view. Its 18 new rows share
the existing memory/register transport. Both SDKs pass 210 family nodes and
3,159 prior controls. A reproduced malformed-class-arity panic is rejected
before indexing; the prior 32 JSON snapshot files remain byte-exact.

The literal SIMD join at `df01da10e` shares Go's zero-terminated literal opcode,
ModRM and optional unsigned imm8 transport across complete `yxshuf` and `yaes`
tables. Its 33 rows include six existing rows, giving 27 new identities.
Both SDKs pass 851 family nodes, all 30 LLVM22 format objects and 582 existing
initialized/effect/tuple controls. All 34 old JSON snapshot files (17 per SDK)
remain byte-exact.
The original failing object-adapter run and malformed-source RED runs remain
failed evidence. The accepted adapter invokes the actual typed emitters.

The scalar join at `aff58904d` adds all 24 multiply, immediate-multiply and
MOVBE source rows through the existing scalar transport. Signed immediate
selection, W16/LQ32 literals and mode restrictions retain actual-Go observations.
Both SDKs pass 1,006 family nodes, 48 initialized three-API/five-target O0/O2
object calls and 2,995 prior typed/effect/identity nodes. All 36 previous JSON
snapshot files (18 per SDK) remain byte-exact. TDD reproduces 33 malformed-escape
panics and 18 invalid byte-width acceptances before early bounds checks close
them. These are source-formatter fixes, with no new numerical permissions.

After integration at `189be9c80`, the complete rule package passes 15,104 nodes
across 147 roots on each endpoint SDK, without failures/skips. Node policy passes 573
tests and Go policy 327 nodes. All ten five-architecture reports remain
byte-exact against `d6cb9245cd`. The minimum-Go report retains its original
unmasked `VADDPD.Z X30,X1,X0` refusal; it is not zero unsupported on every SDK.

## Inventory and semantic boundaries

The independently joined source inventory now contains 54 overlapping family
views and 4,448 records: 3,568 AVX plus 880 legacy records, comprising 862 optab
and 18 ymovtab rows. The legacy denominator remains 1,418 current-SDK /
1,417 minimum-SDK rows. The 538 / 537 unjoined rows are inventory gaps, not
unsupported-instruction totals or native execution counts.

Typed forms keep operation, encoding, register carrier, memory tuple, input
roles, old destinations and machine-state effects distinct. ROUNDSS/ROUNDSD
retain 4/8-byte memory reads even though their carrier is 128 bits. Literal
bytes, pointer representations and source-row ownership grant no numerical,
storage, raw-continuation or native-entry permission. Void and dead results
retain faulting input obligations; nontrapping dead register scratch remains
allowed without manufacturing defined bytes.

SDK capture retains original source/tool/object identity and live verification.
Mutation, closure, aliases and changed metadata revoke the observation.
Do not cache mutable verification or treat parsed header fields as authority.
Shared source formatting has not established a public Translate speedup.

The strict 55-source run at `d6cb9245cd` has 41 passes and 14 failures, all attempted,
with no newly passing files. Its six advancing diagnostics do not repair those
failures. Track individual ordinary-source and special runtime contracts in
[issue47](https://github.com/xgo-dev/plan9asm/issues/47).

## Remaining work

1. Publish the qualified follow-up as one batch, then require its own complete
   CI, raw/artifact joins, coverage and fresh review. The preceding head's
   accepted results do not certify this follow-up.
2. Continue complete legacy transport groups with original source order and
   modes. Reuse typed effects and bounded object batches; do not enumerate
   individual mnemonics or count encoding-only witnesses as execution.
3. Retain explicit native386, AVX512/AVX512ER, RCP14/MXCSR, MMX-state,
   BF16 active-memory fault and floating MIN/MAX execution gaps. Preserve
   ARM64, ARM and WASM object/control gates and all 14 strict-SDK failures.
4. Integrate an accepted public translator into
   [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) with normal public-module
   verification and fresh compile/link/run owners. A fork pin does not resolve
   the P1 official upstream dependency requirement. Coordinate one heavy local
   Go/LLVM/Node invocation at a time.
