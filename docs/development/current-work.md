# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Inspect actual HEAD/status
and terminal owners before continuing; never edit a live frozen snapshot.

## Contribution boundary

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting codex/expand-ecosystem-corpus-20260913. Push only to cpunion.
Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft;
do not promote its fork head until current-head fork CI and review pass.

Published repair head 47ea649d, completed run 36942816488, still fails priority
discovery shards 5, 29, 36, 38, 39, 42 and aggregate. Skipped downstream jobs are
not passes. Refresh the remote state before publishing the next batch.

## Continue here

Development branch: codex/pr40-closed-ci-successor-20261007.
Latest upstream main is already an ancestor. Root alone writes ledger.
This successor integrates these test-first rule-class repairs:

- GP/SIMD transfer grammar, all-path result initialization, bounded independent
  SCC composition, escaped CFG labels, raw near-RET numerical LEA adaptation,
  and physical raw segment/x87 handoff.
- Stale fixtures now invoke the typed dispatcher, distinguish complete GP/X
  writes from partial MOVSS, and assert exact volatile read width.
- Inactive SIB scale is compared on copies only; active physical operands and
  segment metadata remain strict. No production decoder permission changed.
- Twelve horizontal integer operations share grammar, caller effects, native
  legacy m128 access and complete 64-byte alias/flag/canary tests.
- Four dword-product operations share grammar, byte effects and native access.
  The first implementation had a real compressed-disp8 defect; its successor
  binds full tuples 16/32/64 and broadcast 8 to actual Go bytes. Backend O0/O2
  and IR default<O2> are explicitly separate axes.
- Scalar FP results consume shared logical-byte dependencies. Masks, every
  input byte and every return remain checked; no caller/memory permission is
  created. Bare and grouped SB stores require the complete result proof and
  cannot substitute AX fallback for unwritten results.
- GP/X native stores retain their exact volatile width even when overwritten.
  An independent test-first regression covers five moves, two native address
  representations, three APIs, five targets and real IR optimization.

The legacy logical m128 repair is integrated: logical FP bytes are not native
backing/alignment. Those public forms remain Context without the native
contract; all twelve legacy operations retain their original opcode faults.
PS keeps an SSE1 carrier, while PD/integer keeps SSE2. Scalar ANDNL/ANDNQ
feature identities must not attach BMI to packed ANDNPS/ANDNPD. Rosetta
cannot certify Intel alignment traps. A required Linux/amd64 fault axis stays
separate from portable numerical/object observations.

## Verification next

Frozen ea4a5a31 completed all 32 root partitions: 4,859 passed, ten failed,
39 host-inapplicable roots skipped, none incomplete. All ten now have focused
repairs, not a passing successor full-root certificate. Current/old scalar
diagnostics each pass 50 roots/3,790 nodes; their separate caller-logical scope
passes five roots/823 nodes. These overlap and prove only their own snapshots.

Obtain terminal family owners and independent physical readbacks. Freeze the
clean combined successor and run the complete
root inventory in 32 partitions with bounded two-worker collection, other
root-module packages, nested CLIs, focused historical compatibility, official
five-arch classification and strict benchmark. Required native Linux/cross
counterparts run in fork CI. Missing tools, unknown errors and timeouts fail.

The mode32 extreme-SP witness still observes an LLVM22 O0 allocator limitation:
its multi-gigabyte virtual frame is not a successful object/runtime case.
Preserve actual Go acceptance and the backend error, never source N/A.

The latest strict stdlib baseline fails 12 of 184 attempted files. Dynamic
reflect/caller frames, native continuations/TLS, ARM64 result padding and SDK
body closure remain separate contracts. A microbenchmark or official
classification pass cannot certify these.

## Corpus and ledger

Local background corpus owners were terminated at the user's request. Do not
replace them with local inventory scans or external corpus batches. Historical
partial/failed reports remain evidence, never promoted to passes.
Inventory belongs on its remote scanner; external compilation/execution belongs
on separate disposable validation executors, not that scanner. Local first-party
fixtures and corpus-driver unit tests remain allowed.

The committed ledger still has three published Snappy passes, 6,064 pending,
no complete shard, complete=false and verified=false. Only the automatic
publisher may change records with exact matching frozen provenance. Focused
instruction tests do not advance this ledger.
The committed inventory has 215 continuous ranges through
2026-09-22T22:43:08Z, with the 2025-10-01 history cutoff. Newer standalone
inventory is unsynchronized; cgo stays outside this repository.

Inspect remote owners and frozen inputs before replay. Recheck the original
discovering samples and all 64 required shards after rule-class repairs.
Partial files, Context and provisional exclusions are not passing versions.
Funnel tables belong in the PR body, derived from validated reports.
Clean only exact owned caches after all consumers close. Never hand-upgrade
reports or claim CI/review/coverage completion early.
