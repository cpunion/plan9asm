# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). This checkpoint is not a
completion certificate. Inspect actual owners, clean revisions, complete raw
reports and frozen tools before accepting any result.

## Contribution and active owners

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not update the upstream-connected head until fork current-head CI and review
pass; never push to, approve, merge or close upstream PR 40.

Published `c09eff1e`, run `37711752851`, remains the last completed fork run:
four priority shards failed and two reached their deadlines. Downstream skips
and failed aggregate checks are not passes. Batch fixes before pushing again.

Development is on `codex/pr40-result-effects-20261009`, integrated through
`55b71af0`. Independent owners retain these immutable snapshots:

- `codex/pr40-result-effects-verify-20261010`, `0529fa9c`: complete ordinary
  main-module run under `_out/result-effects-module/`. The root inventory has
  5,385 roots and 32 two-worker partitions. Subsequent DATA, shift and pure-GP
  commits are not in this snapshot. Let the entire module owner finish,
  including failures; inspect the final report, not only progress messages.
- `codex/pr40-cpp-register-matrix-20261009`, `3e82acad`: fixed external sample
  under `_out/broad-external-sample-20261009/`. `selection.json` fixes 21 exact
  candidates across 19 existing 64-way shards before outcomes. The current
  collector writes `reports-final/` and `run-remaining.log`.
  An earlier shell-variable error ran shard 5 successfully, then rejected CLI
  arguments before further candidate execution. Original failed reports remain;
  a stopped duplicate retry remains incomplete. Only the independently
  validated original shard-5 report was copied verbatim into the final set.
  Do not mix failed/incomplete journals into final aggregation.

Root alone publishes the assembly ledger. Never rebase/edit/import into a
running snapshot or supply replacement tools.

## Integrated rules and verification

Earlier integration repaired diagnostic joins, raw/AUTO replay, complete
GP/X/MMX FP widths and physical CPP register aliases. Alias filtering permits
only an exact emitted bare-symbol to register rewrite; result-name filtering
still requires complete source identity and the result CFG.

The current batch adds:

- Original native side-effecting opcodes for all seven legacy whole128 moves,
  shared by both dispatchers, preserving dead-load faults and native alignment.
  Each declares its own SSE/SSE2 features. Logical FP/AUTO storage acquires no
  native backing/alignment from this change.
- Go source DATA/GLOBL alignment on all five architectures, checked against
  actual linked Go objects, both LLVM emitters and runtime addresses. Private
  synthesized literal pools retain their separate rules.
- Complete CLMUL grammar, selected-qword byte dependencies and correct wider
  alias writes. The actual-Go format matrix covers 1,653 accepted witnesses.
- All 24 packed logical/byte/per-lane shift names share byte dependencies.
  Independent tests cover every imm8 and width/count axis; 30,628 actual-Go
  accepted witnesses and 336 mode32 high-Z rejections pass. Floating immediates
  are rejected; mode32 truncation follows Go without rewriting source.
- A lowering-only numerical result check pairs exact native whole128 loads
  with reaching incoming-pointer identity. It supplies no bounds, nonaliasing,
  alignment, initialized Go pointee or raw/caller-entry permission.
- One immutable per-file directory describes the emitted DATA/GLOBL numerical
  allocation, including zero-filled holes. Complete resolved objects reject
  relocations, unresolved extents and function/signature collisions across
  aliases. Typed consumers require full native widths, including legacy m128
  shift counts. The temporary directory cannot escape into CFG state or public
  queries. Isolated transfer clones restore it only inside the same callback.
  Building once avoids quadratic global/signature scans per TEXT.
- Closed nontrapping pure-GP operations propagate unknown scratch bytes into
  unknown dependent outputs/flags. Later branches, addresses, stores or results
  still require defined inputs. Failed ordinary transfers use isolated clones.
  No GP entry initialization or pointer-as-number permission is added; ordinary
  GP MOV remains outside this closed fallback.

Logs remain under `_out/legacy-move-native/`, `_out/scalar-result-data/` and
`_out/result-effects-combined/`. Combined families at `01010e38` pass in
100.901 seconds; combined DATA/pure-GP/polynomial tests pass in 34.155 seconds
before `55b71af0` commits the tested callback adjustment. These are finite
scopes, not a full-suite pass. Earlier misspelled VEX fixture failures remain;
corrected explicit Go names establish the actual legacy-count RED and GREEN.

The polynomial composition compares identical original Go assembly and an
independent reference over 1,156 cases, then all three LLVM APIs at O0/O2:
pointer reads, aligned constants, CLMUL, shifts, loops, joins and scalar returns.
Missing-proof mutations remain errors. Pure-GP tests cover 235 amd64 and 169
mode32 programs; 227 non-BMI2 runtime programs compare original Go and LLVM.
Exact Go 1.20 and focused race scopes pass at their component revisions.
BMI2 and true native alignment require their Linux/amd64 runtime counterparts;
local Rosetta transport is not a native execution certificate.

## Next work

1. Finish and independently read both running owners, preserving failures and
   incomplete reports. Rerun current-head combined gates after successors;
   old component passes do not transfer automatically.
2. Validate the discovering CRC library on a clean integrated disposable
   validator. Its failure involved native reads, DATA, CLMUL, shifts and an
   unused undefined DI update. Independent rules now compose; this is not yet
   an external CRC pass or a current standard-library CRC pass.
3. Complete mode32 static-SB address roles: Go ignores base annotations but not
   indexes. An SP/BP annotation must not become virtual-frame storage. Preserve
   dynamic-index bounds, segment/raw/pointer and amd64 rejection boundaries.
4. Investigate repeated Go tool/source/profile planning queries in external
   verification. Reuse only within a validated operation; retain mutation,
   tool-identity, cleanup and profile-scope checks. Do not change a running
   validator or remove applicable instrumentation scopes.
5. Run nested CLIs, vet/build, strict SDK/benchmark, required native/cross
   runtime and current-head fork CI. Refresh upstream before final promotion.

The five-arch official classification gate at `01010e38` passes with zero
unsupported/parse errors but includes Context forms, not 100% semantics or ABI.
The last strict SDK benchmark attempted 184 files: 167 pass, 17 fail, no
skip/N/A/unattempted. Rerun it; changes do not retroactively change that report.

## Inventory, ledger and resources

No local background index/discovery/external corpus. Inventory stays on the
source-only scanner; compilation/execution uses separate disposable validators.
Local first-party tests are allowed. cgo inventory stays outside this repo.

At `3e82acad`, automatic assembly evidence records five complete candidate
passes (xxhash/v2, websocket and three snappy modules), one CRC failure and
6,061 pending candidates. Reports: `_out/cpp-register-matrix/discovery-reports/`.
The ledger is incomplete/unverified, not evidence for a changed translator.

The rebuilt streaming path-plus-major remerge retains 21,257,856 contiguous
index entries, 1,051,454 unique paths, 6,700 assembly versions and 84,225 files
in independent staging. It restored 126 remote-only major-line records and one
historical local-only record. Discovery is not imported/compiled coverage;
the committed assembly denominator is 6,067. Do not rescan old ranges or import
outputs normalized by the obsolete path-only tool.

Publish only automatic matching schema-11/source/scan/tool/file/target/profile
evidence. Context, source N/A, first-party tests and provisional exclusions are
not module passes. Keep funnel tables in the PR body. Inspect owners/disk before
launching heavy gates or cleanup. Release LLVM modules before contexts and
remove only validated owned caches after all consumers close; never selected
shared tools. Never rewrite historical failures into successes.
