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

Development branch: codex/pr40-final-batch-20261008.
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
Scalar insert/extract byte effects, uncached full-file hash efficiency and
imported-type header planning have focused repairs. The encoder-table gate
now executes its historical-SDK path instead of returning early. AES and
scalar caller effects are integrated, including an independent Go-byte/FIPS
runtime fixture and a missing-input negative control. Their current and
historical SDK scopes pass; this is not a current-head full-suite certificate.
The mode32 VPEXTRQ native width requires its own original-encoder check; an
object pass alone is not a numerical proof.

Both x86 result observers now transport incoming opaque pointer identities
without marking their bytes numerical or granting initialized pointee/native
storage. MOVD consumes its existing MOVQ word-alias grammar through a local
dispatch copy. The focused objects, numerical/canary runtime and strict negative
controls pass at their own frozen source. Original crc32 now passes its former
MOVQ import check but still fails at a CPU-feature SB memory read; it is not a
passing library.

Normal sharded-ledger operations now stream: startup/status, active-batch
checkpoint lookup, bounded retries, publication and sharded report transfer.
They still validate every record through EOF, counts, range continuity and
generation readback. Batch lookup retains only active paths and their exact
repository case aliases; publication and transfer retain current module groups,
not ledger-wide payloads or version maps. Retry selection obeys its limit and
persists a fair queue marker independently of the two index cursors.
Superseded-version checks likewise retain only requested replacement keys.
Explicit monolithic legacy JSON/gzip imports and full-report reference helpers
remain separate compatibility paths, not bounded streaming interfaces.
The frozen streaming integration f53ef7dd passes the complete Discovery package
(235 root/subtests) and the affected superseded/progress controls (49), separately
in ordinary and race runs, plus both packages' vet/build and committed-ledger
readback. Its initial canonical-path oracle failure is retained; explicit
successful/corrupt aliases now use the actual physical source path. These tests
do not certify a later successor or advance assembly-ledger statuses.

The official AES baseline now records its unbound SP/BP memory operands as
Context on all eight SDKs. Independent before/after form reports isolate the
same three changed shapes, while actual Go bytes and certified native-memory
positive controls retain support. Encoder/grammar/runtime counts and historical
exceptions are unchanged; this is a storage-proof correction, not an ISA or
non-target-architecture exclusion.

The next fork batch adds an exact original standard-prefix module versus
stock-SDK namespace refusal path. Keep retained complete portable command output,
independently recomputed hashes, exact module/source/profile/argv roles, and the
valid sibling's five-target object obligations. Generic ambiguity, unknown
tools, extra diagnostics and missing evidence remain failures. This refusal is
source N/A for its exact package/profile, not an instruction pass or whole-module
skip. Its first one-target producer passed; the expanded five-target producer
timed out during native linking and remains incomplete, not a passing scope.
Obtain the final component's actual terminal status before reporting success.

The focused historical scalar list hit its 20-minute pre-test link limit, with
zero attempted tests; plain tests never started. Keep its incomplete/failed
receipt and unresolved owned-child state, not a compatibility pass. Do not
borrow the current-SDK component's 352 passes or repeat unlimited local links.

Obtain terminal family owners and independent physical readbacks. Freeze the
clean combined successor and run the complete
root inventory in 32 partitions with bounded two-worker collection, other
root-module packages, nested CLIs, focused historical compatibility, official
five-arch classification and strict benchmark. Required native Linux/cross
counterparts run in fork CI. Missing tools, unknown errors and timeouts fail.

The mode32 extreme-SP witness still observes an LLVM22 O0 allocator limitation:
its multi-gigabyte virtual frame is not a successful object/runtime case.
Preserve actual Go acceptance and the backend error, never source N/A.

The strict stdlib run on 55ecfe9d attempted all 184 files: 167 passed and 17
failed, with no N/A or unattempted files. All five targets stayed within the
normal 300/900-second budgets. The failures include private scalar result proofs
with symbol/native memory inputs, dynamic reflect/caller frames, native
continuations/TLS, ARM64 result padding and SDK body closure. Preserve this
failed baseline and repair the shared rule classes; performance or official
classification success cannot certify these files.

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
inventory is unsynchronized; cgo stays outside this repository. The requested
history cutoff has been reached. The earlier head service was OOM-killed
before publication. The standalone streaming inventory successor has since
resumed one bounded, two-worker head controller; its completed windows extend
the head cursor without redoing historical ranges. Status, batch
lookup, publication and transfer stream the standalone ledger. Inspect its
current owner, manifest and journal for the actual cursor; never hot-replace
its binary or start a second writer. These new records are still an
unsynchronized, untested queue, not assembly passes. The cgo inventory remains
outside this repository.

Inspect remote owners and frozen inputs before replay. Recheck the original
discovering samples and all 64 required shards after rule-class repairs.
Partial files, Context and provisional exclusions are not passing versions.
Funnel tables belong in the PR body, derived from validated reports.
Clean only exact owned caches after all consumers close. Never hand-upgrade
reports or claim CI/review/coverage completion early.
