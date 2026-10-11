# Current work: separate external compatibility from PR 40

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

## Remaining delivery gates

The migration PR is draft. Its historical assembly ledger remains incomplete;
this does not change translator coverage or certify any external module.
Verify the source split, scheduler/CI regressions, shared package and both CLIs.
Then publish the cleanup to the allowed fork and update the existing PR 40
contribution. Do not merge or close upstream PRs.

Do not modify the separate frozen `fe553bff` module run or fork run 38116495799.
Their results belong to their own revisions. Preserve their terminal evidence,
but do not interpret them as verification of the split revision.

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
