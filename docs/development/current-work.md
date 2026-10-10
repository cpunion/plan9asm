# Current work: PR 40 systematic coverage

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Component tests, opcode
classification and historical reports are not current-head CI/runtime success.
The earlier chronological checkpoint remains in Git history.

## Authority and integration

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote its upstream-connected branch before current-head CI/review pass.
Never push to, approve, merge or close upstream PR 40.

Published `c09eff1e` still has four failed priority shards (5/29/38/39), two
canceled (36/42), and incomplete downstream gates. No newer component result
changes that outcome. Develop on `codex/pr40-result-effects-followup-20261010`,
including the RNG and fixture repairs below. Root alone publishes the assembly ledger.
Never change tracked files, tools or ledger in a running frozen snapshot.

Earlier x86 typed conversions/compares/moves, ARM VFP/NEON, Wasm comparisons/
minmax, original-source roles and CPP/tag closure remain integrated. This batch
adds complete ARM64 FPCR/FPSR and NZCV effects, native scalar binary operations
and min/max, and cross-family syntax-before-target preflight. Add each new
native validator to the shared list, not an early-return hook in one API.

Scalar FADD/FSUB/FMUL/FNMUL/FDIV passes 28,214 original-Go state rows and 451,424
LLVM comparisons on each of Darwin and Linux/QEMU. The complete family includes
ten names, Go two-/three-operand rows, raw H/S/D, first-source sparse NEP merges,
dead status effects and all four consumers. FPCR/FPSR and NZCV also pass their
four-consumer/four-axis required Linux oracles; no general system permission
follows. Keep source syntax, object, runtime and ABI claims separate.

At `e1762415`, 23 merged roots pass current Go and Go 1.20, vet passes, and
five-architecture official classification passes with zero unsupported forms
or parse errors. Context is not execution credit. The later min/max merged
ten-root syntax/object gate passes; full runtime completion remains pending.

SDK observation now retains every direct member of each actual imported SDK
package instead of unrelated descendant package trees. Selected descendants,
literal includes and shadow guards retain separate roles; ordinary modules
remain recursive. Independent REDs and current/Go 1.20/race/Linux negatives
pass without increasing any bound. Original crypto examples are rerunning.

## Frozen consumers and evidence

Paths below are relative to the named worktree. Preserve logs and binaries;
clean only exact owned caches after every consumer closes.

- `codex/pr40-arm64-scalar-minmax-native-20261011`, `1f29b927`:
  `_out/scalar-minmax/darwin-full-measured.log` and `linux.log` run all
  28 batches, original Go and three public APIs at four optimization axes.
  Typed-Go objects are a separate fourth consumer, not runtime credit.
  Finite current/Go 1.20 gates pass. Keep the earlier 25-minute timeout failed.
  Measured backend O2 dominates; do not drop rows to improve timing.
- `codex/pr40-arm64-rndr-named-effects-20261011`, `464c8b8b`:
  MRS/MOVD RNDR/RNDRRS and WORD now share grammar, NZCV effects and +rand.
  Current/Go 1.20 formats, objects, boundary/optimization negatives, ownership,
  extraction and vet pass. `_out/rndr-named/linux-runtime.log` runs original
  Go/QEMU allowed-state invariants across 64 forms and all four consumers.
  It does not compare independent random values or claim forced RNG failure.
  Preserve platform/g entry boundaries. Integrated; RED is `93b6c4a6`.
- `codex/pr40-original-budget-20261011`, `37eadcf8`:
  `_out/original-budget/` contains SDK graph RED/GREEN/negative gates.
  `_out/crypto-recheck/` reruns exact shard-39 versions runZeroInc/excrypto
  v0.43.3, tdv/crypto v0.0.0-20260609182332-5f2de1a9f1e2 and x/crypto v0.57.0.
  Stamped tools, fresh events/reports and disposable validator stay frozen.
- `codex/pr40-refill-sample-20261010`, `02c9ebbb`:
  `_out/refill-sample/` continues the identical fixed 53-candidate selection
  across 37 shards with four refillable workers and 60-minute phase bounds.
  Read partial reports for progress; terminal candidates do not complete shards.
  SDK-bound, ABI/asmdecl and effect failures remain failures. No ledger credit
  follows from live logs or this older revision.
- `codex/pr40-instrumentation-planning-reuse-20261011`, `b57e99be` plus
  Go-only fixture change: `_out/instrumentation-planning/` preserves repeated
  three-minute REDs. Complete 32-CPU-query alias gate passes. The combined
  modes/emission/Go-only gate now passes in 968.633 seconds. Go-only retains
  all 18 queries and nine objects. Six-minute fixture budgets do not change
  production/benchmark limits. Both fixture commits are now integrated.
- `codex/pr40-full-family-verify-20261010`, `044bb668`:
  `_out/full-family/reports/` is terminal: 18/20 packages pass. Root has
  5,617 pass, one obsolete helper-assertion failure, 39 host-runtime exclusions;
  corpus driver has 686 pass and three fixture-timeout failures.
  Newer repairs do not relabel this failed full report. Its closed 2.5-GiB
  owned cache is reclaimed. Both nested CLIs remain separate gates.

Closed component evidence remains under `_out/fp-control-effects/`,
`_out/nzcv-transport/` and the scalar-binary owner's `_out/`.

## Next completion gates

1. Close min/max and RNG native runs; finish merged gates. Do not confuse
   component and combined evidence.
2. Finish the broad sample and repaired crypto repeat. Each counterexample
   needs an independent complete family/rule regression and original recheck.
3. Audit remaining native FP operators by whole families. Scalar ARM64 FMA
   remains an unclosed pure-intrinsic/status class; it has not been rewritten.
4. Run current-source full module, nested CLIs, race/coverage, strict SDK and
   required cross-runtime gates before promotion. Batch verified fork pushes.
5. Eleven Linux AMD64 native AVX512 runtime roots still lack a compatible
   executor. Local execution and QEMU cannot supply that hardware; portable,
   Rosetta and object success cannot satisfy the required gate.
6. Latest strict SDK evidence is 184 attempts, 163 passes, 21 failures in
   frame/entry/ABIInternal/TLS/global/alignment contracts. Time budgets pass,
   correctness does not. Unbound storage or zero-valued TLS is not a repair.

## Inventory and ledger

The inventory server scans source only; do not compile external code there or
start local background index scans. cgo inventory stays outside this repository.
External compilation uses disposable isolated validators.

Committed inventory remains 18,758,534 index records, 965,347 paths and 6,067
assembly versions. Independent staging has 21,257,856 contiguous records,
1,051,454 paths and 6,700 assembly versions after path-plus-major remerge;
it is not imported or tested coverage. Do not rescan committed ranges.

Historical `3e82acad` assembly evidence has five passes, one CRC failure and
6,061 pending. Later samples cannot be relabeled current-head ledger passes.
Only the automatic writer may publish matching schema-11 report, progress-v3,
assembly-ledger-v3 and complete file/target/profile/source/tool evidence.
Never mix revisions, hand-edit status, commit obsolete gzip/run ledgers or load
the full inventory just to display progress. Keep funnel tables in the PR body.
