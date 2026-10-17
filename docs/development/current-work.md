# PR 40: current checkpoint

Updated 2026-10-18. Keep this document a current handoff, not an append-only
history. Earlier checkpoints and original failed fixtures remain in Git.
Read [instruction contracts](instructions.md) and [validation](validation.md)
before changing a family. Preserve existing WIP, frozen branches and failed
reports; do not discard them while simplifying this checkpoint.

## Scope and publication

PR40 owns instruction translation, complete typed families and independent
tests for 386, amd64, arm, arm64 and wasm. External inventories, scanners,
ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). No inventory
scanner runs locally. Do not expand that workload in this translator PR.

Push only to the allowed `cpunion` or `dastai` fork. PR40 remains draft at
`24851df2`; verified local or backup branches do not qualify that contribution
head. Promote one verified batch after the frozen CI finishes, without a
direct upstream push or merge. Latest upstream main `e2164ac0` is an ancestor
of the current candidate. Do not create another worktree for routine work.

## Frozen snapshots and evidence

| Snapshot | Evidence and limits |
|---|---|
| `d0505fd99` | Active branch `codex/pr40-scalar-fp-address-closing-20261018` is TDD only: three roots / 22 children reproduce 20 failed terminals, including a falsely accepted uninitialized floating result; no production repair or qualification yet |
| `d5bdb0399` | All 179 historical failed roots re-observed locally: 29 pass, 58 fail, 92 explicit host skips, zero absent/pending/incomplete. Original output hashes, all run/terminal identities and both source/SDK/tool/environment seals audit unchanged; selected aggregate remains failed, not a full/native CI pass |
| `0cb14d2fa` | Integer self-comparison and mandatory Void repair: 54 roots / 7,277 children pass current Go, actual Go1.20 and race; all 7,331 terminals match, zero fail/skip/incomplete, unchanged source/SDK/LLVM22 observations |
| Same code head | Build/vet, twelve scanner roots on both SDKs, five-arch official classification, three CLI extraction roots, 132 CI policies and tracked formatting pass; staged in allowed fork branch `codex/pr40-integer-self-compare-closing-20261018`, not the contribution head |
| `bc79fd0f` | Combined named/raw NOP repair: 64 roots / 8,606 children pass current Go, actual Go 1.20 and race; all 8,670 terminal identities match, zero fail/skip/incomplete, unchanged source/SDK/LLVM 22 observations |
| Same code head | Three additional current-Go SDK producer roots / eight children pass normal and race; they are not part of the Go1.20 denominator |
| Same code head | Build, vet, eleven scanner roots on both SDKs, five-architecture official classification, three CLI extraction roots, 132 CI policies and tracked formatting pass |
| `aec6c45a` | BZHI/BEXTR/BLS affected union: 26 roots / 620 children pass all three lanes; eight-SDK audit changes only four isolated AMD64 BZHI forms per SDK supported-to-context, no unsupported drift |
| `92acfbb2` | Paired native-input repair: 44 roots / 3,522 children pass all three lanes; the original latest failed priority selection is retained |
| `06eedf847` | Published frozen diagnostic [run 39065963275](https://github.com/cpunion/plan9asm/actions/runs/39065963275); do not cancel/restart it for a small push |
| Same published run | Audited actual native execution: 73 roots / 228 children pass, zero fail/skip/duplicate/incomplete; these receipts do not certify later BZHI/NOP repairs |
| `8ca976c6` | Last completed full main-root denominator: [run 39054965721](https://github.com/cpunion/plan9asm/actions/runs/39054965721), 16 parts / 64 shards, 6,986 roots: 6,544 pass, 179 fail, 263 explicit host-inapplicable skips, zero incomplete |

Do not subtract newer focused successes from that completed CI denominator.
Canonical complete modules, all-platform execution, the strict benchmark,
coverage and fresh review must qualify the final contribution head.
Classification, emitter-only objects, initialized public translation, portable
execution and native execution remain different claims.

## Latest repairs

All 24 fixed/immediate integer EQ/GT comparison names share the same validated
same-vector dependency rule. Self comparisons need no incoming numerical vector
bits; actual masks, different inputs, memory, raw origins and MMX keep their
separate requirements. Retain operands for emitter allocation, existing FLAGS
and upper-alias rules. Complete Go rows/attributes, all imm8 controls and banks,
three-API five-target objects and whole-byte/mask/FLAGS/canary execution pass.
Actual-host portable execution is not native AVX512 execution.

The scanner gate then found a real mandatory-descriptor bypass: the Void/no-
result shortcut ran before `requiresResultProof`. True RED `f50765b07` has
91 failed terminals, five original Go object receipts and no fixture/tool error.
Moving the shared descriptor check before that shortcut fixes immediate compares,
CRC32, implicit MUL/IMUL and the remaining registered mandatory effects without
another opcode list. The two new roots / 157 children pass at `7d0ab042d`.
Three old unbound format roots fail their obsolete public-success expectation;
`0cb14d2fa` preserves every original source row as Go/public-Context/emitter-only
object coverage. Existing initialized public owners remain independent.

The sealed Go1.20–1.27 audit changes exactly eight AMD64 self-initializing formats
Context->supported, 92 isolated missing-input formats supported->Context, and
one 386 IMULL format supported->Context per SDK. Encoding rules and every other
metric remain unchanged, with no new unsupported forms. Baseline changes retain
that distinction; a form classification is not execution of every instance.

## Strict standard-library benchmark

The complete run at `bc79fd0f` finished naturally with exit 1 and unchanged
source/SDK/LLVM 22 observations. All 184 original target/file pairs were attempted;
132 compile and 52 fail, with zero N/A or unattempted. Compilation is not runtime
execution. The result is failed, not a qualified performance pass.

| Target | Compiled / total | Failed | Wall seconds |
|---|---:|---:|---:|
| linux/386 |17/33|16|73|
| linux/amd64 |29/55|26|709|
| linux/arm |27/30|3|5|
| linux/arm64 |40/47|7|12|
| js/wasm |19/19|0|5|

Total target wall time 804s meets the 900s total budget, but AMD64 exceeds its
300s target budget. A one-second diagnostic sample captures substantial
SHA256/file-validation work; inspect `benchmark-bc79fd0f/amd64-sample.txt`
before assuming instruction lowering is the bottleneck. The sampled diagnostic
cannot certify a performance improvement. Full reports remain under ignored
`compile-benchmark-bc79fd0f` and `benchmark-bc79fd0f`.

## Immediate work

1. Repair the 52 complete standard-library failures by actual shared rule class,
   retaining their unchanged complete sources and the original 300s/target and
   900s/total budgets. Current failures include source-input effects, private
   helpers/custom entries, AUTO/native frames, TLS/implicit state and typed
   result backing. A Context rejection is not successful compilation; determine
   why a real original Go program lacks its translation contract before adding
   a new factory, exception or validation layer. Investigate repeated checking
   using the captured performance sample without removing mutation, lifetime
   or source-boundary failures. A concrete next candidate is the parser's
   conflation of omitted TEXT arguments (`$0`) with explicit zero (`$0-0`).
   Go's assembler retains ArgsSizeUnknown for the first; the private Void extent
   check rejects both. `subtle/xorBytes` reaches this error. Reproduce before
   changing it; keep explicit mismatch and native-frame/address controls.
2. Watch the published diagnostic run to its natural terminal. Its native
   owners already pass; Windows SDK priorities and complete modules still need
   their own final results. Download and inspect actual failed roots before
   another diagnostic-branch push. No full result exists merely because all
   prerequisites pass.
3. Repair only remaining observed rule classes. Preserve discovering sources,
   independent complete Go rows/attributes, initialized positives and negatives.
   Use shared typed grammar/effects rather than library-specific exceptions.
   Existing failure classes include scalar input/effect gaps, MMX-region
   composition and stale unbound-format expectations; the completed CI failure
   index is the authority, not an estimated remaining count. The complete local
   historical-failure replay now has 58 failed roots, not 179 unexamined ones;
   92 host skips remain unexecuted native counterparts, never repaired tests.
   This selected queue does not revalidate the old complete suite's passed roots.
   Its main classes are numerical/field-address effects, MMX closure, raw/native
   read composition, custom ABI/stack roles and obsolete unbound-format contracts.
   The first actual failure remains `TestGoABI0PaddedSqrtObjects`. The new floating
   field-copy tests separate address/storage transport from SQRT arithmetic.
   Current roles admit integer/bool parameters and integer results only; parameter
   backing also requires exact logical end rather than ABI0 rounded allocation.
   The uninitialized-floating-result negative additionally exposes missing cost
   routing, not a fixture error. Keep all five target Go objects, three APIs and
   padding/input-write/SP negatives; do not special-case SqrtByAddress or grant
   contiguous/native frame backing. Run complete affected Go forms before adding
   SQRT effects, and retain the unchanged original discovering program.
4. Require current-head complete CI, benchmark, coverage and fresh review before
   ready. Keep ABI/context failures visible; do not weaken the benchmark or
   treat Context classification as successful program compilation.

## Reproducing and finding receipts

Use the commands and LLVM 22-only environment in [validation](validation.md).
Run focused RED first, then the complete affected union on current Go, actual
Go 1.20 and race. One heavy local Go/LLVM owner runs at a time; develop in an
existing idle worktree while its tested source stays frozen.

Resolve branch locations with `git worktree list`. Latest ignored receipts use
`integer-self-compare-final-{core,effects,grammar,mandatory}`;
`integer-self-compare-qualified-0cb14d2f.json` records the exact disjoint union.
The eight-SDK audit is `integer-self-compare-classification-audit` at `7d0ab042d`;
later changes before qualification are only test contracts and their baseline.
The local 179-root replay is `previous-failures-d5bdb039`; its
`raw-outcomes-audit.json` retains every actual failure and skip separately.
The floating-field RED is `logical-floating-field-address-red` at `d0505fd99`.
Earlier receipts use `nop-source-state-qualified-{nop,stack,identity}` and
`nop-current-sdk-producers`; `qualified-nop-source-state-bc79fd0f.json` records
the exact union. Earlier cohorts use `bitfield-qualified-*` and
`paired-native-qualified-*`. CI artifacts are retrievable from the linked runs;
local reports are useful evidence but do not become tracked repository data.
