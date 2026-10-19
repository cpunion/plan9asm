# PR 40: current checkpoint

Updated 2026-10-20. This is a handoff, not a completion certificate.
Earlier checkpoints remain in Git; preserve ignored failed evidence.
Read [instruction contracts](instructions.md) and [validation](validation.md)
before changing a family.

## Scope and publication

This repository owns translation, complete typed instruction families and
independent tests for 386, amd64, arm, arm64 and wasm. External inventories,
scanners, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm).
No inventory scanner runs locally.

Push only to an allowed cpunion/dastai fork. PR40 remains draft at
`24851df2`; a staging branch is not that contribution head. Recheck upstream
main before publishing; last checked `e2164ac0` is an ancestor of this batch.
Stage the coherent batch on the fork's
`codex/pr40-windows-owned-process-probe-20261013` branch, whose push starts
the diagnostic workflow. Let each run finish; do not supersede active runs.
Update the contribution only after qualifying its staged source.

## Latest online evidence

[39325946539](https://github.com/cpunion/plan9asm/actions/runs/39325946539)
finished naturally at `a46ca8ab2`: 72 jobs pass, eight fail, one dependent gate
skips, none cancelled. The full Windows modules did not run, so this run does
not replace the last complete root inventory below. Both ownership SDKs and
all SDK priorities except `source` pass. The three source-priority failures
come from the independent budget oracle missing the existing two-byte LEA
observation; production already charges it. Duplicate native fixtures fail on
a duplicate Go package clause; the same defect also affects division fixtures.
The next local batch fixes both without changing native ASM/C or budget limits.

All four native reduction owners and their aggregate pass. Native conversion
owners 0, 5 and 7 actually pass with AVX512F/DQ/BW/VL. Owners 1, 2, 3, 4 and 6
fail required CPU preflight with AVX512F/VL absent and XCR0=7; their aggregate
correctly fails. The official Linux pool is mixed, not uniformly incapable.
A successful allocation is not a stable capability guarantee. Never substitute
objects, portable values or skips for these native owners. Independent native
steps now retain strict failure status while allowing unrelated owners to run;
the next workflow has 190 passing local policy tests, not an online pass.

[39285088077](https://github.com/cpunion/plan9asm/actions/runs/39285088077)
finished naturally at `47e44aa7d`: 83 jobs pass, 16 fail, none cancelled.
The sixteen-owner main-module union is 7,379 roots: 7,063 pass, 52 fail,
264 explicit host skips, zero absent/incomplete. Every root is owned once.
Fifteen root owners and their strict aggregate fail; both complete CLI modules
pass. Compared with the preceding complete run, eighteen old failures now
pass and there are no newly failing roots. Local repairs do not rewrite
this online inventory or qualify a whole-suite percentage.

The ignored `ci392850/root-union-outcomes-audit.json` binds the raw events
and complete inventory. Earlier failed runs and local recorder failures remain
failed. Do not combine different-source receipts into a current-head pass.

The sole formal PR40 review thread is resolved, not reviewer approval.
The orphaned-IR and symlink-include regressions remain required. Complete
current-head CI, coverage and the strict benchmark are not yet qualified.

## Focused local evidence

Every success below has zero failed, skipped or incomplete terminals within
its own scope. Scopes overlap and cannot be added into a coverage percentage.
A source/tool-sealed receipt qualifies its own unchanged snapshot, not a later
merge. Objects, portable values, native execution and ABI authority are distinct.

| Snapshot | Focused evidence | Lane and limits |
|---|---|---|
| `d5ce4441e` | Actual duplicate/division fixture package ownership, original-source controls and duplicate five-target three-API objects: 7 / 34 | Current Go, actual Go1.20 and race on this exact snapshot; native opcode execution is still separate |
| `d16fb9ea3` | Complete raw legacy floating compare inputs, all selectors, immutable identities and five-target objects: 16 / 328 | Two disjoint current-Go cohorts; original complete source/C remains unchanged and still requires its native owner |
| `cabff6cd8` | Complete ABS/sign Go rows, typed roles, initialized five-target objects, portable values and WIG/EVEX boundaries: 11 / 419 | Current Go; all 13 names, 42 official rows and 114 named forms; no named MMX or original AVX512 execution claim |
| `c82b27515` | Scalar shift/rotate guarded architectural outcomes, complete Go rows, initialized objects, original source controls and immutable replay: 23 / 7,846 | Current Go; ISA-undefined producer outputs become finite only after actual inputs are proved. Ordinary queries stay conservative; original native owners remain required |
| `3b21b7720` | Whole-word private opaque-pointer cells: 9 / 249 | Current Go; original g-alias source, three-target objects and actual-host LLVM/C companion. No numerical bits, caller/frame/entry role or physical stack authority follows |
| `579c4cba1` | MOVNTI, floating helpers, conflict/funnel and compact/expand: 64 roots / 1,472 children | Current Go, actual Go1.20 and race; original native execution remains separate |
| `013181561` | ALIGN, DIV/IDIV, raw logical returns and original formats: 53 / 2,130 | Current Go, actual Go1.20 and race; no new entry or memory permission |
| `47e44aa7d` | Raw IFMA/lookup and shared decoder: 23 / 378 | Current Go, actual Go1.20 and race |
| `4f0119031` | Raw moves and physical EVEX identity: 33 / 3,322 | Current Go; actual Go1.20 independently passes the same complete scope at `ae667fabb` |
| `06c8615cf` | Complete integer min/max: 25 / 472 | Actual Go1.20, two disjoint owners on this same snapshot; current-Go scope passes at `4f0119031` |
| `5982a0612` | Complete reduction plus original format/literal regressions: 35 / 638 | Actual Go1.20; includes the strict assertion on the actual function body, not unused declarations |
| `5982a0612` | All numerical/native/raw LEA and retained-pointer controls: 42 / 670 | Race; portable DF instrumentation is explicitly separate from unmodified native execution |
| `8206232b5` | CLMUL complete physical/input matrix: 15 / 1,759; original/decoder regressions: 16 / 607 | Two overlapping current-Go scopes, not a summed denominator or original native execution |
| `5b153b532` | Complete ternary named/raw initialized matrix: 10 / 352 | Current Go; subsequent shared source-closure correction is independently tested below |
| `356457017` | Shared dead-source closure and original ternary guard-page program: 8 / 368 | Current Go; all 36 E4 forms × 5 masks × 5 boundaries × 2 IR pipelines execute on the actual host |
| `767c98550` | Shared source-closure RED regressions plus original whole-CFG root: 3 / 312 | Actual Go1.20; no synthetic shape facts authorize native addresses or read masks |
| `75c9095f7` | Genuine SDK owner precedence: 3 / 234 | Current Go; SDK literal/numerical consumers remain ahead of native family hooks |

The initial shared closure RED at `c1f23961b` has 302 failed terminals.
Its correction keeps dead states separate from reachable states, uses actual
preceding source facts, and shares the existing 16,384-visit / 64MiB budget.
Typed accesses distinguish memory effects from LEA address syntax. The
ordinary unknown scratch domain remains unchanged. The detector RED at
`1460cb8cb` fails six direct classifications; its nine public controls already
pass and must not be reported as a public-source regression.

Retained LEAW keeps old caller-address high bytes only in the existing
single-use address-bit carrier, never NUM, pointer or backing facts.
`146bb6a2a` passes six current-Go roots / 145 children. The cross-host companion
isolates exactly the checked native STD/CLD sideeffect calls while retaining
the full logical DF lifecycle, original source/C and all 300 input triples.
It is instrumentation, not unmodified x86 execution.

Required Linux native owners now separately retain actual Go bytes and
original LLVM bodies for LEAW, and complete SSE/AVX reduction and reciprocal
programs. Their fixtures/ownership pass locally; their native execution does
not. LEAW, duplicate and reciprocal steps run before AVX512-dependent steps.
The independent LEAW auditor requires all fourteen terminals, actual host/SDK
brackets and every execution marker; missing, duplicate, skipped, incomplete
or late observations fail.

## Classification baseline

The eight-SDK audit at `356457017` inspects all sixteen Go1.20–Go1.27 x86
reports. Exactly 343 input-deficient AMD64 forms per SDK, across 58 opcode
names, change from supported to Context. All 386 forms, unrelated fields and
encoder/unsupported/parse metrics remain unchanged. Complete report forms and
fingerprints equal the earlier independent observation at `8206232b5`.
Baseline-only `05782c0d2` records this audit.

The next independent audit at `79dd1ca75` compares all sixteen reports with
that baseline. Exactly 53 AMD64 input-deficient probes per SDK change from
supported to Context, limited to the thirteen ABS/sign names and SHLW/SHRW.
Every 386 form and all unrelated encoder/unsupported/parse metrics stay
unchanged. Baseline-only `420233de8` records precisely those 24 metric fields;
the ignored `closeout-classification-79dd1ca7-initial` reports retain every form.

Current Go AMD64 has 3,239 supported and 3,497 Context forms, zero unsupported.
Context is not a successful initialized-program translation. The historical
Go1.20 unsupported form remains unchanged; the current five-architecture gate
still requires zero unsupported forms and parse errors.

## Remaining work and hardware

1. Freeze the integrated batch, run minimum/race shared-rule regressions and
   the final build/vet, scanner, isolated-CLI, policy and five-arch gates.
   Publish one coherent allowed-fork batch and inspect every resulting owner
   and aggregate. Do not subtract local repairs from the old 52-root inventory.
2. The conservative source mapping at `8206232b5` has 37 corresponding repairs
   or format corrections, fourteen unclosed/runtime-unverified roots and the
   then-open ternary regression (now corrected above). This is triage, not a
   predicted new-CI result. Preserve original initialized runtime programs.
3. The current batch repairs guarded scalar shift/FLAGS outcomes, explicit
   g-alias pointer roundtrips, packed abs/sign and raw legacy compare. Remaining
   families include physical SP/BP, MMX, four-source operations, scalar precision,
   immediate/variable shuffles, explicit mask stores and raw FP/RIP consumers.
   Fix complete rule classes, then rerun original roots to discover later
   failures. Do not infer missing family support from a repaired initial load.
4. Required original AVX512 qualification needs suitable execution hardware.
   The official Linux allocations are mixed as recorded above. Windows owners are mixed;
   some report F/VL=1 and others F/VL=0, so allocation is not a stable guarantee.
   Conversion needs F/DQ/BW/VL, floating helpers F/DQ/VL, and physical wide
   half-move execution F. Do not weaken checks, replace execution with objects,
   or count skips as passes. Reduction and reciprocal require only SSE/AVX;
   retained LEAW only scalar x86, so those are not AVX512 blockers.
5. Keep PR40 draft until current-head CI, review and coverage qualify. Update
   its head/body only with the actual staged result; never merge or approve
   upstream.

## Benchmark and verification discipline

The last strict standard-library benchmark at `18bb77a09` remains failed:
all 184 target/file pairs attempted, 129 compiled, 55 failed, no N/A.
AMD64 took 392s, over the unchanged 300s target budget; total 504s is below
900s. No newer speedup or complete benchmark pass has been certified.

Use current Go1.27 and exact LLVM22; root-module minimum is actual Go1.20,
rich-CLI minimum Go1.24. One heavy local Go/LLVM owner at a time.
Commit before validation; freeze tracked source, documentation and used tools.
Develop in another idle worktree while a snapshot runs. Release LLVM objects
and owned subprocess resources promptly. Missing tools fail, never skip.
Native host exclusions require an explicit counterpart.
