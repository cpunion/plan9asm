# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git; never modify a running verification snapshot.

## Contribution and CI

Repairs are staged in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to
`cpunion`. Do not promote the upstream-connected branch until current-head fork
CI and review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
remains Draft.

Published source `0dad3b63` failed all six priority shards in run `36922137520`.
Their complete reports were audited and imported automatically in `c057aa97`.
This is historical evidence, not success for subsequent source revisions.
Skipped downstream root, standard-library and benchmark jobs did not pass.
Preserve the failure-first scheduling and all 64 required shards.

Development branch: `codex/pr40-sm3-profile-integration-20261002`, inheriting
`codex/pr40-guarded-integration-20261001`. Freeze a clean final batch before
rebuilding stamped tools and collecting new reports. Documentation changes
also change the source fingerprint. Put progress/funnel tables in the PR body,
not here; derive them with the shared report readers.

## Immutable inventory and evidence

Imported Index ranges are continuous from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. This interval is consumed, not the whole Index or later
updates. Inspection failures remain retries. Do not import or resume scans
while corpus verification uses that ledger. Standalone inventory and cgo
records remain outside this repository.

Only matching source/ledger/tool provenance may update assembly evidence.
Do not copy old passes into a new-head snapshot or mix host tools. Run the
automatic writer against fresh reports and re-read its result; missing shards
stay pending and genuine failures remain failed. Partial reports cannot meet
the completion gate. Never hand-edit status flags or report schemas.

## Integrated mechanism repairs

The batch preserves declaration-backed ABI0 frames, bounded stack/static-read
contracts, exact data relocations, wasm logical-PC contracts without
`blockaddress`, and schema-10 ordinary profile proofs.

Recent repairs cover ARM64 private-SP stores and bounded affine frame reads;
the typed ARM status-register family; source-order CPP EOF and relative-include
binding; actual ordinary-package roles for test-named assembly; and ZIP-bound
zero-byte selected siblings. Nonempty, whitespace/comment and macro-empty
assembly still requires its normal Go/CPP/LLVM proof.

The complete raw ARM64 SM3 family has exhaustive decoding, independent LLVM-MC
encodings, alias/target object tests and independent Go/LLVM runtime oracles.
The original gmsm candidate passed its default-target production replay at
frozen `45f9e6d7`; scoped Go source rejections remain in the report. The separate
`96c1d58f` oracle observes every NZCV flag and exercises both values. Neither
report proves the subsequently integrated source without a fresh replay.

Bulk Go source errors retain their complete bounded raw diagnostic digest and
a canonical sample of 256 distinct positions. Sampling does not stop checks
for late infrastructure errors or invalid source positions. The original avo
candidate passed its default-target replay at frozen `ff488337`; source-excluded
scopes are not LLVM objects or runtime passes.

## Required next work

Run the affected focused suites, root `./...`, both nested CLIs, all five
official instruction-table gates, strict standard-library corpus, benchmark
and all 64 Discovery shards. Preserve earlier timeouts/failures; do not count
individual successful tests as a full-suite pass. Use bounded parallel batches
and remove owned caches only after every writer exits.

The official opcode classification gate is not complete standard-library
object coverage. Dynamic reflect FP frames, native runtime entries/context
switches and private ABI transport remain real contracts to implement.
Provisional native-byte-layout/JIT exceptions require review before promotion.
Raw private-SP frames in go-krypto are a separate bounded-proof investigation,
not permission to accept arbitrary raw stack effects.

ARM64 closure entries require an explicit immutable `{code pointer, uint64
capture}` carrier and the real hidden register ABI. The default API remains
conservative; ordinary descriptor environments are not capture sizes. The
dependent compiler must provide actual source/caller evidence, not merely
register the translator contract.

## Dependent llgo contribution

Fork Draft [PR 259](https://github.com/cpunion/llgo/pull/259) is rebased onto
upstream main `d98b43f90`. Its public `go.mod replace` resolves published
plan9asm head `0dad3b63`; no local path or invented pseudo-version is committed.
Publish the reviewed next plan9asm batch before updating that pin, then rerun
the actual public-pinned compiler.

The nested `test/asm` module has exact CRC32, huff0, go-hex, websocket,
modernc/libc Uint128 and purego regressions. Dedicated Linux and Darwin jobs
execute every applicable fixture and retain source/tool/dependency provenance;
root `./test/...` alone does not reach this module. Required Linux ARM64 memmove
execution uses pinned QEMU 10.2.3, not a skipped host-only test.

Frozen llgo `4bbc5cec` passed five actual Linux issue fixtures; purego still
failed at native `syscall15X`. Darwin was blocked by two original R26 closure
entries. A real typed caller/carrier producer is in development, not yet final
public-pin evidence. Purego needs a native C/g0/callback/address-table contract,
not an invented Go declaration or suppressed standard-library assembly.

The next llgo batch repairs the cold replacement installer lookup, formatting
and unnecessarily patch-specific fixture module minima. Windows include-path
diagnostics preserve the failing directory/stage without weakening symlink
guards; the actual missing-directory cause remains to be established from CI.
