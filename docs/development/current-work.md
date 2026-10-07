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

Development branch: codex/pr40-instruction-family-integration-20261007.
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
- Unresolved x86 memory displacements remain in the parsed inventory, but
  complete translation rejects them before raw decoding. Descriptive bare
  AUTO names retain Go's zero-offset syntax, not a native memory permission.
- The complete CMOV condition/width family shares typed grammar and byte
  effects. Numerical FP bit views preserve NaN payloads and signed zero.
  Every potential memory source is read even when the predicate is false.
  Fixed-local AUTO certificates and emission consume Go's actual frame-biased
  address role, including shared GP/SIMD consumers; they cannot certify a
  positive named offset that is physically outside the frame.
- Corpus alias fixtures now distinguish registered ADDQ read effects, public
  all-path result initialization, and Go's actual unknown-name diagnostic.
- Instrumentation header queries inherit their actual planning context.
  Cancelled parents cannot launch another query or hide cancellation behind
  an independent source error. This does not increase the candidate budget.

The legacy logical m128 repair is integrated: logical FP bytes are not native
backing/alignment. Those public forms remain Context without the native
contract; all twelve legacy operations retain their original opcode faults.
PS keeps an SSE1 carrier, while PD/integer keeps SSE2. Scalar ANDNL/ANDNQ
feature identities must not attach BMI to packed ANDNPS/ANDNPD. Rosetta
cannot certify Intel alignment traps. A required Linux/amd64 fault axis stays
separate from portable numerical/object observations.

## Verification next

Frozen 190fa86a completed all 32 root partitions: 4,922 passed, two failed,
39 host-inapplicable roots skipped, none incomplete. Both stale destination
fixtures now have focused repairs, not a passing integrated full-root
certificate. Its separate corpus-driver inventory completed all 32 partitions:
562 of 565 roots passed and three failed. Two alias failures are repaired;
the imported-types header-planning timeout remains independently reproduced.
Other root-module packages, both nested CLIs, workflow policy, vet and build
passed on that frozen source. Do not relabel those reports as current-head.

The CMOV successor's complete affected scopes pass on current and historical
SDKs, including fixed-local and caller authority guards. Cancellation scopes
likewise pass on both SDKs. Each certificate belongs to its own clean commit;
overlapping focused scopes cannot be added into a full-suite coverage count.
Scalar insert/extract byte effects and uncached full-file hash efficiency are
being repaired independently. The mode32 VPEXTRQ native width requires its
own original-encoder check; an object pass alone is not a numerical proof.

Obtain terminal family owners and independent physical readbacks. Freeze the
clean combined successor and run the complete
root inventory in 32 partitions with bounded two-worker collection, other
root-module packages, nested CLIs, focused historical compatibility, official
five-arch classification and strict benchmark. Required native Linux/cross
counterparts run in fork CI. Missing tools, unknown errors and timeouts fail.

The mode32 extreme-SP witness still observes an LLVM22 O0 allocator limitation:
its multi-gigabyte virtual frame is not a successful object/runtime case.
Preserve actual Go acceptance and the backend error, never source N/A.

The latest strict stdlib baseline fails 14 of 184 attempted files. Unresolved
source macros are now retained without parse loss, but PINSRD/AES private
result effects still stop the actual 386 memhash object. Dynamic
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
inventory is unsynchronized; cgo stays outside this repository. A read-only
remote status check found 304 continuous ranges through 2026-09-30T02:32:07Z
and the requested history cutoff reached. No scanner owner was live; its
previous head service was OOM-killed. Inspect memory bounds, ownership and the
actual ledger before continuing incremental work; do not restart blindly or
claim this queue has been imported/tested.

Inspect remote owners and frozen inputs before replay. Recheck the original
discovering samples and all 64 required shards after rule-class repairs.
Partial files, Context and provisional exclusions are not passing versions.
Funnel tables belong in the PR body, derived from validated reports.
Clean only exact owned caches after all consumers close. Never hand-upgrade
reports or claim CI/review/coverage completion early.
