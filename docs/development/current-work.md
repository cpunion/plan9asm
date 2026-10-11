# Current work: consolidate and repair PR 40

External scanning, ledgers and corpus verification have moved to Draft
[llgo-compat/plan9asm#1](https://github.com/llgo-compat/plan9asm/pull/1).
Its initial ledger import was compared byte-for-byte with the pre-split
`c8724add` snapshot. Historical pending/failing outcomes were not promoted.
The original source and failed reports remain recoverable; the new repository
owns subsequent inventory imports, cursors, external failures and ledger updates.

## Translator scope

- Keep all instruction-family repairs through `c8724add`, official Go coverage,
  standard-library compilation, cross-runtime tests and benchmarks.
- `gotoolprofile` is now public so the independent harness and translator CLI
  share the same guarded observations without copying the implementation.
- Remove external commands, ledgers, exception manifests, scripts and jobs.
  Core CI depends on its policy preflight, not the external corpus.
- The exhaustive module runner assigns owner 1 to `gotoolprofile` rather than
  the moved corpus driver. Every actual package still has exactly one owner;
  root partitions and the strict complete-module aggregate remain mandatory.

## Consolidated repair checkpoint

`3dc36d71` contains the instruction repairs and external-corpus separation.
Continue from this integrated source, not an older family worktree. The latest
review's SDK, environment, tool provisioning, fixture and cache findings have
focused repairs below; broader CI gates still need current-head completion.

The first repair batch fixes the ARM memory-source ADD and the ARM64 branch
label that collided with a vector register. Runtime and portable preflight
tests share their four-architecture fixtures. Their actual Go objects and
LLVM 22 objects pass on Go 1.27.2 and Go 1.20.14; native Linux execution still
requires the cross-runtime lane. Ordinary, race and coverage jobs now install
the required IA32 compiler/libc, with an independent ordering regression.
Workflow policy tests and actionlint pass; the split CLI import formatting is
also corrected. RED evidence is retained locally, not converted into a pass.

Go 1.27.2 observation support now retains the audited format pins, separately
registers its two changed native-header source files and binds object/archive
headers to the producer's exact patch. ARM64 actual-source, CRC32 three-API
LLVM 22 objects, static-storage and marker mutation guards pass; Go 1.27.1 and
Go 1.20.14 focused compatibility checks pass. Unknown patches remain rejected.

The ordinary CLI now resolves its pure-Go defaults and actual cache paths
before package loading, then hands that same environment to source observers.
Explicit nonordinary options remain visible, not silently normalized away.
Default-invocation, explicit-profile, overlay and source-selection tests pass.
The twelve ARM64 undefined-byte cases now share only a test-owned build cache;
their source/temp/observation state remains independent. Local elapsed time
was 32.52 seconds before and 12.72 seconds after, without removing any cases.

CI run `38131270471` finished with failures. Its standard-library benchmark
compiled 163 of 185 files; its 68-second target total is not a complete-success
performance result. Keep the PR draft and preserve failed evidence.

## Remaining delivery gates

The migration PR is draft. Its historical assembly ledger remains incomplete;
this does not change translator coverage or certify any external module.
The split is already published to PR 40. Verify the new repairs with focused
RED/GREEN cases, the shared package, both CLIs and affected full gates, then
publish a verified batch to the allowed fork. Do not merge or close upstream PRs.

Historical runs, including `fe553bff` and fork run `38116495799`, certify only
their own revisions. Retain their evidence when retiring inactive worktrees;
never promote old or incomplete reports to current-head success.

The five-architecture official Go classification gate passes after the split.
The supplementary x/arch ARM64 decoder baseline remains failed: 774 supported
and 271 unsupported forms, versus the baseline's 778 and 267. An actual replay
using the published pre-cleanup `77a70d89` dependency reproduces the identical
`ff28d1d7ad89967b24ed2dfe813716569960ecc942c9b87994fb6a963c4b1be1`
fingerprint. This is a pre-existing gate failure, not a migration pass or a
reason to edit the baseline. Preserve the before/after reports under `_out/`.

Use [instruction-family development](instructions.md) and
[validation](validation.md). Keep old detailed progress narratives in Git
history instead of growing this checkpoint again.
