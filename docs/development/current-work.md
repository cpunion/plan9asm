# PR40 closeout

Updated 2026-10-31. Finish amd64 correctness, typed instruction families and
translation efficiency before expanding another architecture. Retain the
386, ARM, ARM64 and WASM gates. Read [architecture](architecture.md),
[family development](instructions.md), [convergence](x86-convergence.md)
and [validation](validation.md). External scanning and ledgers belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm); do not run
local inventory processes.

## Last accepted publication

[PR40](https://github.com/xgo-dev/plan9asm/pull/40) remains draft.
Head `0310af39b` completes its own
[CI round](https://github.com/xgo-dev/plan9asm/actions/runs/41190856800)
naturally: **325/325 successful jobs**, no reruns or cancellations.
Independent checks close all 286 fresh source artifacts, 240 partitions,
23 groups and all 22 Windows/race packages with 168 raw streams each.
Host-inapplicable test skips remain separate from passes.

Official function coverage is 120,283 / 136,464 = **88.1%**.
The 22-profile raw block union is 120,383 / 136,564 = 88.1513%.
Exactly 100 covered initializer statements in 82 blocks outside function
declarations explain the denominator difference. An independent Go AST
observer reproduces the official algorithm; the original failed equality
check remains recorded. Codecov patch coverage is **94.60% >= 87.83%**.

The fifteen-kernel native scalar-move owner actually runs on current
Go 1.27.2 and minimum Go 1.20.14 Linux CI: ten RUN/PASS nodes per SDK,
with six exact summaries of 15 kernels, 1,920 numerical comparisons,
180 faults and two tripwires, without failures or skips.
These observations do not qualify every native instruction or ABI entry.

Both historical review threads are resolved; no formal approval is recorded.
Earlier `e04799ae8` remains a failed round: 302 successful and 23 failed jobs
from two reproduced fixture defects and their dependent collectors.
The `0310af39b` repairs have fresh evidence; they do not relabel that old run.
[Earlier component history](https://github.com/cpunion/plan9asm/blob/0310af39bb51abbe3d3bd0b97c8ebc680fd6305a/docs/development/current-work.md)
retains prior qualifications and failures.

## Current integrated batch

The code snapshot is `10de80934`, combining the following qualified components.
A later publication needs its own complete gates and CI; an older accepted
head does not certify these changes.

- GP data/FLAGS/CPU source grammar retains 100 current / 99 minimum rows,
  including 62 existing identities. It shares the scalar transport, complete
  operand schemas, prefix handling and opcode spans. Captured enum/name
  contradictions and BSWAP padding are rejected; valid implicit zero cells
  remain accepted. Go-accepted CPU-invalid forms are source observations only.
- MOVQL and MOVLQZX use opposite directional 32-bit move forms and actual
  source mode. Shared effects and local dispatch copies preserve original
  Op/Raw/operand identity. MOVQL reuses MOVL lowering; 386 MOVLQZX preserves
  a complete ordinary word while amd64 keeps its existing extension path.
  Existing raw NUM, pointer, caller, frame and native-memory boundaries remain.
- Remaining media/x87 source grammar retains 31 rows across twelve complete
  tables: nine existing identities and 22 new ones. Old conversion/horizontal
  public views and the new view share complete encoding validation. The 95
  original malformed-metadata counterexamples still exercise the old consumers.
  New and old paths share carrier construction; MMX admission remains unchanged.

Both SDKs pass the new alias grammar/effect tests (125 nodes and thirty
five-target, three-API O0/O2 object modules each). Portable execution passes
27 nodes and eighteen backend/link/execute leaves per SDK, checking 20,736
numerical/memory observations. An independent seed makes ZF checks orthogonal
to the copied value. The earlier correlated-ZF oracle is retained separately.
This is host execution of translated IR, not original x86 execution, complete
native FLAGS acceptance or 386 native-memory execution.

The shared old semantic set passes 141 roots / 8,264 nodes per SDK.
The old alias/object join passes nine roots / 244 nodes on current Go and
seven / 166 on minimum; two Go1.27-only source owners are not compiled on
minimum and are not counted as skips or passes.

Media/x87 focused tests pass 126 nodes per SDK; the complete rule package
passes 18,618 / 18,635 across 176 roots. Existing hooks pass 107 nodes and
input restrictions 604 per SDK. All results have zero failures/skips.
The old 54 identity snapshots remain byte-exact; two new snapshots retain
the new view. One snapshot wrapper failed its post-command disk guard after
fourteen genuine passing roots. Exact source/tool/provider identity retains
those executions; the remaining three commands ran separately. The original
resource failure remains failed and is not described as a successful wrapper.

## Source inventory and limits

The current integrated source union has 63 overlapping views and 56 snapshots:

| SDK | AVX | optab | ymovtab | Joined | Legacy / denominator | Unjoined legacy |
|---|---:|---:|---:|---:|---:|---:|
| Current | 3,568 | 1,172 | 18 | 4,758 | 1,190 / 1,418 | 228 |
| Minimum | 3,568 | 1,171 | 18 | 4,757 | 1,189 / 1,417 | 228 |

These are independently joined source identities, not translated or executed
instruction counts. The denominator excludes exactly thirteen Zpseudo and
four Zbyte rows; their identities remain retained. The ymovtab sentinel
remains separate from its 112 encoding rows. Preserve overlapping rows,
first-match selection and actual mode refusals.

Source grammar grants no NUM, pointer extent, machine state or source lease.
The newly observed MMX conversions and status forms still need their separate
initialized/object/semantic qualifications. Context, hardware exclusions and
unexecuted forms never count as passing instruction coverage.

## Next work

1. Run final combined policies, all five architectures on both SDKs, full
   applicable build/vet and the controlled Translate comparison against
   accepted `0310af39b`. Retain all original 29 cases, MOV-dense cases,
   accepted-input guards, three ABBA blocks and actual outliers. Do not infer
   SDK or generated-program speed from register-only microbenchmarks.
2. Finish the remaining 228 source identities as format groups:
   forty stack/control optab rows, 94 state/system/memory optab rows and
   94 ymovtab rows. The last group has six encoding cases; reuse the eighteen
   existing double-shift rows. TLS zero payloads require real target/shared/
   relocation observations, and ymov fallback does not inherit optab suffix
   filters. Retained TR padding can contain an unconsumed 0xff.
3. Continue initialized forms and semantic families alongside the inventory.
   Keep element, carrier, active/result and memory widths, operand direction,
   masks, alias writes, FLAGS and fault/order effects distinct. Reuse typed
   forms across consumers and remove replaced selectors in the same batch.
4. Preserve explicit native386, AVX512/AVX512ER, reciprocal/MXCSR,
   extended80/BCD, MMX-state, BF16 fault and floating MIN/MAX execution gaps.
   The last strict SDK integration at D6 remains 41/55 PASS / 14 FAIL, all
   attempted; track individual source and runtime causes in
   [issue47](https://github.com/xgo-dev/plan9asm/issues/47).
5. [LLGo2694](https://github.com/xgo-dev/llgo/pull/2694) remains
   `f1c1571a1`, based on `277e5b280`, pinning accepted D905. Its own CI
   passes 140 jobs plus two conditional publisher skips, with 97.06% patch
   against 95.84%. The unresolved P1 requires a tagged upstream translator
   release; a fork replacement does not satisfy it.

Use current Go 1.27, a complete immutable Go 1.20.14 SDK and LLVM 22 only.
Keep one heavy local Go/LLVM/Node owner. Reuse existing compiler caches;
do not clear the global cache or accumulate a cold cache for every attempt.
Dispose owned modules/contexts and temporary executables promptly. Keep
failed reports and source/tool provenance under ignored `_out/`.
