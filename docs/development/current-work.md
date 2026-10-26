# PR40 closeout

Updated 2026-10-27. Finish **amd64 correctness, shared typed families and
translation efficiency first**; retain all five architectures' regression gates.
This page identifies current evidence and remaining work, not a completion
certificate. Read [architecture](architecture.md),
[x86 convergence](x86-convergence.md) and [validation](validation.md).

## Publication and CI

PR40 remains draft. The last completed full run at `3f634d936`
[completed its full CI](https://github.com/xgo-dev/plan9asm/actions/runs/40343133972)
naturally: **325 jobs, 284 successful, 41 failed**. The 24 failed test jobs share
four causes; 17 completion jobs reject their failed prerequisites.
Keep the failed receipts. Resolved review threads do not constitute approval.

The published repair batch addresses:

- Missing default cache directories in actual feature-profile fixtures.
- A preliminary fixed-local role query invoking a mutating symbol resolver
  before capturing the original source identity.
- Windows SDK command spelling being replaced by its physical path, changing
  builtin tool bytes. Feature commands retain their observed spelling. Frame
  observations separately retain command and physical SDK identities; source,
  tool-byte and live mutation checks remain strict.
- The command-ownership AST auditor mistaking `testing.B.Run` for process
  execution. It now recognizes actual lexical `*testing.T/B` bindings and
  rejects shadowed command variables. Required CI selects its regression roots.

The actual Windows
[follow-up run](https://github.com/cpunion/plan9asm/actions/runs/40360487618)
succeeded at `363a28904`: **47 roots / 842 children**, zero failures/skips.
Its source, actual inventory, unique terminals, tools and artifact digest were
independently audited. The preceding Windows run remains failed
at 43/46 roots: the frame-path defect was exposed after the profile repair.
Fork-only diagnostic workflows are not part of this contribution.

Local source `cfdcd58d5` passes 15 roots / 79 children on Go1.27.3 and
12 / 37 on Go1.20.14, with no failures/skips, including the actual SDK alias
and command/physical-root mutation controls. Other SDK and numerical observation
roles retain their own stricter entry contracts; this repair does not qualify
every alias in every observer.

The final combined source `04f11d7f6` passes 84 roots / 1,900 children on
Go1.27.3 and 80 / 1,853 on Go1.20.14, with zero failures/skips. The four absent
minimum-SDK roots each have an explicit Go1.27 build constraint. Both SDKs also
pass all 17 rule-reader roots / 48 children. Root/CLI builds, root vet, all
5,011 tracked Go files' formatting, actionlint and all five official corpus
baselines pass; Node22.23.4 passes 573 policy tests. Later documentation-only
changes preserve these tested Go/module sources. A new complete PR CI run and
current-head coverage/review remain required. Published head `bdbefa7e4` is
[running full CI](https://github.com/xgo-dev/plan9asm/actions/runs/40366498164).
Its initial CLI failures reproduce a new private-CALL routing regression:
`TestCompileOneInfersVoidABI0LocalHelperAtTargetWordSize` must retain the
existing ordinary ABI0 helper with an explicit target-word parameter extent.
The isolated zero-argument helper graph must not claim this distinct call role.
Keep the original positive and omitted-extent negative tests unchanged.

Repair `85685440d` (integrated as `9d6817224`) preserves that bounded ordinary
call through the shared ABI0 layout validator. Complete CLI passes are
335 roots / 1,703 children on Go1.25.14 and 398 / 2,186 on Go1.27.3;
104 related root tests / 2,018 children also pass. All have zero failures/skips.
Twelve new positive combinations compile LLVM22 objects and 27 negative controls
remain rejected. This is local qualification; the published CI still runs the
pre-repair source. Wait for its natural completion before the next batch push.

## Current functional evidence

| Scope | Qualified result | Limit |
|---|---|---|
| Public `bdbefa7e4` Linux cross-runtime owners | 16 owners; 562 roots / 4,586 children, zero failures/JSON skips | Fresh artifacts retain every previous case and add private CALL. Individual scopes include portable LLVM oracles; this is not native AVX512 acceptance. |
| Private helper CALL at `fc4daae9a` | [Official Linux run](https://github.com/cpunion/plan9asm/actions/runs/40358527392): 7 roots / 96 children, zero skips | 11 kernels × 8 inputs, original Go versus three APIs at O0/O2; not full AES execution. |
| Private CALL local regression at `f358a4918` | 62 roots / 1,328 children, zero skips | Objects, source/return identity, negative controls and shared graph regressions. |
| Complete Linux/amd64 standard-library objects | `fff470ab9`: 37/55; `fc4daae9a`: **38/55** | All files attempted; 17 failures, zero N/A. AES is the only newly passing file; all old passes remain. |
| Current/minimum official-row join at `1927d1cae` | Each 16 roots / 240 children, zero skips | Original Go encoding plus three-API LLVM22 objects; runtime remains separate. |
| Current public integrated regression at `48f14c4cf` | 387 roots / 22,198 children, zero skips | Historical qualification of that source, not later edits. |
| CI policy at `fff470ab9` | Node22.23.4: 573 tests, zero skips; actionlint passed | Later changes require appropriate checks. |

The standard-library comparison uses the same Go1.27.2 and LLVM22.1.8, empty
GOFLAGS and complete 55-file inventory. Target wall observations were 67 and
68 seconds; one pair is not evidence of a performance change. Its nonzero
benchmark exit remains a failure.

Private CALL reuses the existing source graph and typed effects. Each callsite
has its own return context; helpers keep original TEXT/PC identity and acquire
no invented Go signatures. Zero-frame/NOSPLIT constraints, recursion and escaped
call rejection, the 4096-node limit and existing work budgets remain. Nonzero
helper frames and caller private-frame sharing are not covered by this batch.

## Official rows and implementation size

The initialized amd64 join retains each official encoding row, including
overlapping VEX/EVEX rows. High register banks force narrow EVEX encodings;
actual original Go bytes must match before any public API/object result counts.

| Family | Official rows | Concrete format witnesses |
|---|---:|---:|
| Packed add/subtract and D/Q low product | 159 | 480 |
| Signed/unsigned integer min/max | 132 | 432 |
| Rounded average, word products, multiply-add and widened products | 90 | 258 |
| Integer and PS/PD packed logical | 132 | 504 |
| Integer and PS/PD unpack, low/high | 108 | 360 |
| Total | **621** | **2,034** |

Two PADDD alias witnesses are separate. MMX rows retain their independent
region/EMMS owners. These counts do not represent every official x86 row,
386 mode applicability, raw decoding or native execution.

The logical family reuses the same grammar, original-byte verifier and
32-function object batches. Its encoding/scope implementation grows by 19 net
lines, with no new translator lowerer. The shared arithmetic pilot shrank from
403 to 346 lines; later cleanups removed 143 shadowed arithmetic lines and
40 shadowed dword-product lines. The retired diagnostic workflow/policies
removed another 4,157 lines while retaining all 79 semantic auditor tests.

The four packed average/product families now share physical normalization in
`e6f7f4fc8` (integrated as `c76820997`): production code shrinks by 77 lines.
Go1.27.3 and Go1.20.14 each pass 123 roots / 10,776 children without skips.
All 40 parser microbenchmarks remove one allocation; per-case median time falls
15–33% across five samples. This does not measure whole-source performance;
native execution of the refactor still requires the next complete CI.

The unpack join at `eed7b44df` reuses the existing grammar and object driver:
24 names require ten scope-selection lines, with no translator or grammar rule
added. Current/minimum SDKs each pass 25 roots / 337 children without skips,
including every earlier initialized row. Its six MMX rows retain their separate
region owner; the new join does not claim additional native execution.

Earlier analysis benchmarks reduced allocations by 65–88% through bounded CFG
replay. Do not claim that as whole-source translation speed or cache mutable
source verification to hide subprocess costs.

## Remaining work, in order

1. Watch the next published batch's complete CI and fix any new failure. Retain
   draft until current-head CI, coverage and external review pass.
2. Restore the three ordinary ABI0 crypto failures: SHA1's conditional choice
   between readonly DATA and an incoming pointer; P256's private helper graph
   including actual nonzero helper frames; chacha20poly1305's helpers accessing
   a caller-owned aligned private frame. Each needs a shared rule-class repair,
   complete original sources and independent execution tests.
3. Keep four implicit-entry cases distinct: bytealg equal's hidden closure
   size, chacha8rand's ABIInternal entry, and memclr/memmove's register/void
   contracts. GC scan's two special return/indirect-call sources also remain
   unresolved. These are not automatically all full-runtime exclusions.
4. Retain eight system/runtime/test-hook failures separately: reflect's dynamic
   frame; runtime asm/sys/time, preemption and secret register erasure; syscall
   vfork; and the start-line test hook. Package-inventory limits are current
   diagnostics, not proof that increasing a limit supplies their missing ABI.
5. Complete initialized official-row ownership and semantic validation across
   the remaining amd64 families. Continue shared typed parsing/effects and
   measured code reduction without losing 386 regressions. Only then generalize
   the architecture to the remaining targets.
6. Keep the 271 ARM64 decoder forms and remaining WASM control/SIMD work explicit.
   Native AVX512 acceptance awaits capable hardware; portable/object passes
   cannot close that item.
7. Update the exact translator pin in [LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694)
   after publishing a qualified revision. Its separate draft head `2e820f5c2`
   had 131 successful checks and two release skips. Complete runtime-entry
   integration continues in [issue47](https://github.com/xgo-dev/plan9asm/issues/47).

External discovery, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run its
inventory locally or restore those jobs to translator CI.

Preserve failed receipts under ignored `_out/`. The longer historical handoff
remains in git at `3f634d936:docs/development/current-work.md`; its old source
revisions and statuses must not be used as current-head certificates.
