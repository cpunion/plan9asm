# PR40 closeout

Updated 2026-10-31. Finish amd64 correctness, shared typed families and
translation efficiency while preserving all five architecture gates. Read
[architecture](architecture.md), [family development](instructions.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).

## Published CI and local revisions

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft. The
last accepted `d905e3682` [CI round](https://github.com/xgo-dev/plan9asm/actions/runs/41082550177)
completed naturally on attempt one: 325/325 successful jobs, no reruns or job
skips. Fresh acceptance joins all 286 source artifacts, 240 partitions,
Windows/race raw outcomes for 22 packages each and 22 profiles: 88.2% statements
and 93.40% patch coverage. Test and hardware skips remain separate from passes.
Both existing threads are resolved; no formal approval is recorded.

The newer `2e85591e3` [round](https://github.com/xgo-dev/plan9asm/actions/runs/41110443099)
encountered account cache-egress exhaustion and Go installation timeouts before
tests ran. Preserve that round's failures. Shared Go setup disables remote
Actions caches; writable job-local caches and exact compiler verification remain.
The repair requires its own complete CI rather than inheriting earlier results.
Published repair `9e48d6747` completed its first
[round](https://github.com/xgo-dev/plan9asm/actions/runs/41124480505) naturally
with 325/325 successful jobs and no reruns. Complete raw/profile/Codecov
acceptance remains separate from the job conclusion.

The historical
`d6cb9245cd` [CI round](https://github.com/xgo-dev/plan9asm/actions/runs/41055685183)
first ended with 315 successful and ten failed jobs. Seven Go 1.21 source jobs
lost their runners; one CLI collector lost an artifact download, producing two
dependent collector failures. Recovery completed with 325/325 successful jobs:
307 original successful jobs remain valid and 18 jobs actually reran. Independent
acceptance joins all 286 source artifacts, 240 partitions, Windows/race raw
outcomes and 22 coverage profiles: 88.1% statements and 93.44% patch coverage.
The initial failures remain part of the evidence.

Historical `70ced1b565` completed its own
[round](https://github.com/xgo-dev/plan9asm/actions/runs/41032133785) with 325/325
successful jobs, 88.1% total and 92.52% patch coverage. At D6, both review
threads were resolved and no formal approval was recorded. Historical results
do not certify another head.

Functional `aff58904d` adds x87 MOVE, literal SIMD and scalar source joins,
published at `d905e3682`.
Each later publication requires its own cloud CI, coverage and review.

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

The x87 MOVE join at `d20cfaab5` retains all 45 original rows across
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

The implicit integer MUL/IMUL/DIV/IDIV join at `5335247d3` retains 16 original
rows, including three existing IMUL rows and 13 new identities. Both SDKs pass
623 family nodes, 30 initialized three-API/five-target O0/O2 object calls and
4,213 prior typed/trap-IR/identity nodes. All 38 previous snapshots are byte-exact.
One shared validator now checks every scalar schema's original and selected
roles before offset arithmetic, closing 24 reproduced invalid acceptances and
two panics. Twelve shared-table state/NOP rows remain separate inventory gaps;
no numerical or native permission follows from their table membership.

At `5335247d3`, the complete rule package passes 15,550 nodes across 151 roots
per SDK, without failures/skips. At `af5b97326`, Node policy passes 573 tests
and Go policy 327 nodes. All ten five-architecture reports remain
byte-exact against `d6cb9245cd`. The minimum-Go report retains its original
unmasked `VADDPD.Z X30,X1,X0` refusal; it is not zero unsupported on every SDK.

The `af5b97326` FP dispatch cleanup deletes 24 repeated projection lines while
retaining all five map checks, their order and original instruction inputs.
Before and after, current/minimum SDKs pass exactly 168/167 roots plus 5,173
children, without failures/skips. Controlled 29-case, two-SDK, three-block ABBA
retains all 696 measurements: across-case median ratios 1.00428/1.00011,
with no case median reaching 1.05. Minimum Go retains seven individual block
ratios above 1.05, including 1.49643. Allocations are approximately unchanged;
no acceleration is established. The earlier pilot overlapped local artifact
work and remains separate from the controlled trial.

The extension MOV join at worker `809e7edf9`, integrated at `1732b74ac`,
retains all 14 source rows: six overlap prior views and eight are new.
It shares the scalar schemas and prefix/ModRM transport, without changing
numerical lowering. Both SDKs pass 649 family nodes, 15,861 complete rule nodes
and 317 prior typed/input/physical-register controls. All 40 previous snapshots
remain byte-exact. The two Darwin runtime roots skip on this host; their fresh
Linux/amd64 Docker counterparts each pass 11 nodes without skips, including
all three APIs and O0/O2 numerical and fault comparisons against original Go.
These are emulated executions, not native hardware acceptance. MOVSWW/MOVZWW
retain word reads, and mode32 MOVLQZX retains its actual 32-bit write.

The ordinary MOV component at `2ec2e7fa0`, integrated locally at `4b59b7399`,
retains all four ordered MOVB/W/L/Q tables: 32 current and 33 minimum-SDK rows.
GP, MMX/XMM and address roles remain separate. The minimum SDK's extra signed
MOVQ immediate row and L/Q address-to-LEA frontend rewrites retain actual Go
observations. Canonical LEA source IDs are references, not additional MOV rows.

Complete-family tests reproduced a shared named-AUTO bias error in original
LEAQ and the new MOVQ path: original Go writes through the local address, while
LLVM previously returned the old input. Both SDKs reproduce and repair that
actual execution failure. Address formation now shares the validated original
frame projection with emission, without gaining a DATA-read certificate.
Failed access metadata cannot select emission. FP names, Unicode and original
Raw/operand replay remain checked; W/L writes revoke pointer identity and do
not manufacture numerical address bytes, while Q retains its proved origin.

Both SDKs pass 141 existing semantic roots / 8,260 nodes without failures or
skips. The complete 157-root rule package passes 16,600 / 16,623 nodes, and the
17-root family passes 1,003 / 1,027. Six Linux/amd64 Docker runtime roots pass
22 nodes per SDK, including original Go versus all three APIs at O0/O2.
The three host-inapplicable roots per SDK remain explicit skips. Runtime was
executed at `7b36254b4`; only one test changed at `2ec2e7fa0`, with all other
5,442 tracked files byte-exact. All 42 prior snapshots remain byte-exact.

Controlled comparisons retain the original 29 cases / 696 measurements plus
24 separate MOV-dense measurements using identical accepted inputs. Original
across-case median ratios are 0.99949 / 0.99750; no case median reaches 1.05.
One current-Go VPADDSW.Z block remains at 1.05696, with case median 1.01569.
MOV-dense medians are 0.98188 / 0.99104; allocations are approximately unchanged.
These descriptive Translate results establish no SDK or generated-code speedup.
An earlier 12-sample incomplete overlay attempt remains excluded, not erased.

## Inventory and semantic boundaries

At `2ec2e7fa0`, the independently joined source inventory has 57 overlapping
family views and 44 identity snapshots:

| SDK | AVX | optab | ymovtab | Total joined | Legacy joined / denominator | Unjoined legacy |
|---|---:|---:|---:|---:|---:|---:|
| Current | 3,568 | 915 | 18 | 4,501 | 933 / 1,418 | 485 |
| Minimum | 3,568 | 916 | 18 | 4,502 | 934 / 1,417 | 483 |

The denominator excludes exactly 13 Go Zpseudo rows and four Zbyte data rows;
their identities remain retained separately. The ymovtab sentinel also remains
retained separately from its 112 encoding rows. Unjoined rows are inventory
gaps, not unsupported-instruction totals or native execution counts.

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

1. Require each publication's complete CI, raw/artifact joins, coverage and
   fresh review. Earlier accepted results do not certify a later head.
2. Continue complete legacy transport groups with original source order and
   modes. Reuse typed effects and bounded object batches; do not enumerate
   individual mnemonics or count encoding-only witnesses as execution.
3. Retain explicit native386, AVX512/AVX512ER, RCP14/MXCSR, MMX-state,
   BF16 active-memory fault and floating MIN/MAX execution gaps. Preserve
   ARM64, ARM and WASM object/control gates and all 14 strict-SDK failures.
4. [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) now publishes `f1c1571a1`,
   rebased onto main `277e5b280` and pinning accepted translator `d905e3682`.
   Normal public-module verification, fresh host checks and Linux Docker
   compile/link/run probes pass. Its fresh cloud CI is accepted: 13 workflows,
   140 successful jobs and two exact PR publisher skips, with 97.06% patch
   coverage against 95.84%. Platform/test skips remain separate from passes.
   A fork pin does not resolve the P1 tagged upstream release requirement.
   Coordinate one heavy local Go/LLVM/Node invocation at a time.
