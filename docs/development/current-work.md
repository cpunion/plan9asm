# Current work: PR 40 checkpoint

Checkpoint: 2026-10-14 08:01 UTC. Contribution source: `15e683d0`.
This is a handoff, not a ledger or a claim that current-head CI/stdlib passes.

## Scope and authority

This repository owns translation, source/tool observations, independent tests,
official forms and benchmarks. External scanning/ledgers belong to the draft
[llgo-compat/plan9asm contribution](https://github.com/llgo-compat/plan9asm/pull/1).
Do not run local index/module/ledger scans or copy their gates back here.
Local edits/commits are allowed. Push only to `cpunion` or `dastai` forks.
Do not push/merge or mutate upstream CI/tags/settings. Do not cancel active
frozen jobs; explicitly authorized allowed-fork retries remain allowed.
Keep the contribution draft until current-head required gates and review pass.
Use Go 1.27.2 and LLVM 22; retain root Go 1.20 compatibility.
Read [instruction rules](instructions.md) and [validation](validation.md).
Missing tools fail; source Context, object success and execution are distinct.
Freeze source/tools/docs during tests. Old failures remain failed; no budget
increases, denominator shrink, source cropping, guessed bytes/alignment or exemptions.

## Where to continue

Inspect status/worktrees/remotes before work; reuse the persistent idle tree.
Do not edit the primary or probe while their frozen jobs run.

| Worktree basename | Branch / purpose |
| --- | --- |
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | `codex/pr40-external-corpus-split-20261011`; integrated source `15e683d0`; SDK composition FAILED |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | `codex/pr40-windows-owned-process-probe-20261013`; staged fork-only priorities, `e8466405`; not pushed |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | `codex/pr40-sdk-numeric-maps-entry-20261013`; verified metadata `d9033ff4`, context `4daf1d73`, idle |
| `plan9asm-wt-pr40-sdk-tls-20261013` | `codex/pr40-sdk-tls-20261013`; verified graph `4bdac0fb`, fixing two SDK counterexamples |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | `codex/pr40-sdk-storage-snapshot-20261013`; verified input protection `b17f25f8`, idle |

## Latest two frozen fork runs

[38507506824](https://github.com/cpunion/plan9asm/actions/runs/38507506824)
is COMPLETED SUCCESS, 45/45 jobs (API updated 04:29:21 UTC).
Frozen fork `af4b6a3f22d0947008674dfad25041c79f2b20a5` carries contribution
`5a5a477aebfa447d83220a9e898055c96af6c0cc` plus fork-only priorities.
Native x86: 49/49 PASS in 91.126s, six inventories/eleven required roots,
zero skipped/missing/duplicate terminals. NUM LEA's native axis is included.
Windows main: 16 parts/32 shards, 5,995 roots = 5,732 PASS + 263 explicit
platform SKIP; 135,898 child PASS, zero child FAIL/SKIP; 21 other packages PASS.
Simple CLI: four parts/eight shards, 31 roots/26 children PASS.
LLVM CLI: four parts/eight shards, 278 roots PASS; 1,906 child PASS plus one
explicit native-ARM64-only SKIP. These skips are not executed semantics.
All scope/before/after source heads match `af4b6a3f`.

[38502949565](https://github.com/cpunion/plan9asm/actions/runs/38502949565)
is COMPLETED SUCCESS, 45/45 jobs, frozen `e544c00193c1e2287467beae88726c4761127d0f`.
Contribution `94bd05f1897bf882c744ba29a610ac56371677ac` plus fork-only priorities.
Native x86 48/48 PASS; Windows main 5,981 roots = 5,718 PASS + 263 platform SKIP;
21 other packages PASS; simple CLI 31/31; LLVM CLI 274/274 with its explicit
native-ARM64-only child skip. These runs do not certify later local source.
The reconnect's TLS-handshake timeout was a watch failure, not a CI test failure.

## Verified local factory integration

`ba474753` is integrated as `7752fc1f`; take only that cohort, not its old merges.
Both CLIs now use the shared typed-demand literal-minimum binder.
Ordinary CRC32 gets no extra Go link; minimum capture/cache uses the same live
readonly owner/context/profile. Close/replacement permanently revoke old
numerical/literal publication tokens; final callbacks cannot leak Options/IR/modules.
Original mutable AES keys stay external loads. Literal floors authorize only
guaranteed masked low bits, not arbitrary addresses or caller/native state.
NUM LEA is already integrated through ancestor `12b3daf0`: actual files are
`amd64_lea_byte_effects.go`, `amd64_lea_effects.go`, `x86_result_parameter_roles.go`.
NUM dependencies require initialized bytes; caller/literal pointers remain opaque.
No descriptor grants memory/extent/frame/entry authority or rewrites original Op/Raw.
Paired initial-load cwd `9bf01156` and caller CMP provenance remain separate joins.

Agent factory source passed all eleven lifecycle race partitions and six late
API axes, 39 exact source/native roots, complete original CRC32 and maps objects,
both affected CLIs, Go20 portable gates, vet and formatting. Not whole-module CI.
Ordered 22-file seal: `2bb0df72f6f9426b4decea1ebf58a82c55333e1170cb69f2e83fed1fdd72cc58`.
Parent `7752fc1f`: fresh shared 21 roots PASS218.862s; rich eight PASS252.029s;
Go20 shared/storage/query PASS0.488/1.292/0.723s; simple six/ten children
PASS315.306s; final callback six/fifteen children PASS158.238s. Fresh inventories
have zero failures, skips, missing or duplicate roots. All three Go1.27 module
vets, formatting and diff checks pass. Full current-head race/CI remain pending.
386 ABI0 bool, no-DATA owner and unresolved native/helper contracts remain distinct.

## Verified continuation and output publication

`f3e67767` is integrated as `f1db0fa8`. Continuation checks the actual selected
package's complete Go/s/S/h siblings, including unselected variants and separate
names for the same physical file; original request/member identities stay live.
Whole and chunk outputs share one same-filesystem stage outside source roots.
Every final source verification and owner Close precedes public modifications.
Publication failures attempt rollback; failed restoration preserves old backups.
Post-commit temporary cleanup errors explicitly say outputs were committed.
Multiple-file publication is not an atomic filesystem transaction.

The exact 70-root/171-child class passes in 235.258s; its disjoint race union
passes in 301.843s. Go1.20/1.27 caller controls pass 170 roots/5,220 children in
235.816/266.923s; Go1.20 region22/sourcepath also pass. Parent independently
replays all 70/171 in 222.617s, then the eight SDK roots/29 children in 241.311s,
with strict inventories and vet. Before/after parent head remains `f1db0fa8`.
Native Windows, full current-head modules/race and fresh benchmark remain pending.
Alternate-extension consumed includes and module metadata can still collide
with outputs: actual six-child RED is preserved, with a separate repair underway.

## Verified metadata-only SDK handoff

`d9033ff4` is integrated as `71ae5584`. `CaptureMetadata` returns the existing
live SDK observer and an actual selected complete member, without numerical,
readonly, literal-minimum, Options or native/frame permission. Empty DATA and
comment-only selected members retain their distinct complete-source roles.
The selected inventory is bounded by the existing archive-member limit; this
canonical source observation does not certify the caller's original alias node.

The frozen agent cohort passes eight roots/39 children in ordinary, race and
Go20 runs, plus the complete three-OS empty-DATA source/object fixture in normal
and race runs. Parent `71ae5584` independently passes all eight/39 and the one
authentic root/three OS children in 7.130s, with zero failures/skips. Its official
Go1.27 and Go1.20 five-architecture classification gates pass. Go1.20's existing
one AMD64 unsupported form is unchanged, not claimed fixed. Before/after heads
match `71ae55849d342fac8581d8a22c1697dd9a71914d`; diff is clean.
These gates do not certify whole-module CI or native execution at this head.

## Latest verified cohorts and failed composition

`4bdac0fb` is integrated as `d6f2048e`: one whole-source x86 direct-tail graph,
not a separate helper ABI with reinitialized registers. Go27 59 roots/867 children
PASS10.211s; Go20 portable58/862 PASS10.113s; Go27 race59/867 PASS18.364s.
CLI six/100 ordinary/race PASS0.867/4.138s. Five x86-target objects and complete
original-Go/LLVM O0/O2 execution pass; local Linux execution remains emulated,
not native certification. Original old-source execution genuinely fails (got0,
want7). Go20 CLI cannot parse its existing Go1.24 module requirement; that
attempt stays an infrastructure failure, not a compatibility pass.

`4daf1d73` is integrated as `9361399a`: six context lifetime routes share a
bounded terminal sampler. No user callback follows the final deadline on a
potential-success path; closed Done retains actual errors and never yields nil.
Agent and parent ordinary/race/Go20 each pass12 roots/76 children. Parent times
0.644/1.876/0.553s. Source identity, deadline and generation roles stay separate.

`b17f25f8` is integrated as `15e683d0`: same-pass source reads/search misses,
actual selected embeds, metadata, aliases and every publication member remain
protected. Reserved source extensions include case variants, without changing
Go's source selection. Ordinary and disjoint race unions each pass121 roots/
559 children, zero fail/skip/missing/duplicate, in352.026/428.953s. Complete
eight-root reader tests pass on Go27 and actual Go20; vet and stable-head checks
pass. Previous failed receipts remain unchanged.

Parent `15e683d0` authentic metadata root/three OS children PASS7.416s, but its
complete eight-root SDK composition is FAILED250.155s: seven roots pass, and
`TestCompileOneX86SDKBytealgExistingObjects` fails exactly two children.
`count_amd64.s` needs source-bound PC-relative branches; `indexbyte_amd64.s`
needs an independently proved result-slot address role. Fix these rule classes
before publication; do not restore the old incorrect separate-helper state or
turn Context into object success. The failed parent head remains stable.

Fork priorities retain13 SDK lanes and the original budgets. Mandatory native
execution adds AMD64 and386 direct-tail roots plus a required i386 compiler;
the new51-root profile is not yet executed. All29 policy regressions and
actionlint pass. Native audit controls reject missing inventories, substitutions,
skips and duplicates; synthetic auditor fixtures are not execution receipts.

## Stdlib benchmark: current complete failed receipt

Frozen `7752fc1f`: 164/184 files, 20 FAIL, zero N/A/unattempted.
386 29/33; amd64 49/55; arm 27/30; arm64 40/47; wasm 19/19.
Complete original AMD64 maps now passes; 386 still requires its independent
literal/ABI0 contract. Build3s, target total165s, within unchanged300s-per-target
and900s-total budgets. Reports remain FAILED; object success is not execution.
The earlier `2d696307` 163/184,21-failure receipt stays unchanged.

| Failed-file mechanism family | Count | Contract still needed / new evidence |
| --- | ---: | --- |
| 386 maps linked literal predicates | 1 | Independent 32-bit literal minimum/ABI0 consumer; AMD64 complete source passes the current benchmark |
| bytealg same-file tail / hidden closure | 2 | exact root-to-helper state, DX/R26 closure; AMD64 no-DATA handoff remains separate |
| chacha private helper | 1 | genuine undeclared/native helper and split-stack/caller role; keep all three TEXTs |
| reflect dynamic FP frame | 4 | caller-frame extent/layout; `func()` and local stack size are insufficient |
| startup/TLS/native state | 6 | actual OS/arch/shared/thread/native entry and pointer/linkname globals |
| asynchronous preempt state | 3 | original FLAGS/CPSR/NZCV and mixed register-target state, not guessed initialization |
| VDSO/syscall caller stack | 2 | exact FP/saved-caller memory and disjoint syscall proof |
| ARM64 atomic result tail | 1 | observed eight-byte write vs four-byte result/padding contract |

## External server checkpoint (no local scan)

Trusted observation completed 2026-10-14 07:50:36 UTC:
`incremental_since=2026-10-10T05:22:43Z`; `history_before=2025-10-01`;
`history_complete=false`. The next hourly window is unfinished; never advance
its cursor from partial progress.
Index21,867,676 entries; unique_modules1,066,010;
scanned_records1,107,433; matched_records6,768; assembly_files85,732;
failures16,823; latest_records996,045. Cgo:1,101,609 records/23,234 modules,
2,401 failures/1,882 pending. These are server statistics, not translator passes.
The earlier542-range checkpoint was independently checked continuous; its first
and last endpoints equal the derived history/incremental cursors. Periodic
exports can lag the completed manifest/journal; never interpret that as rollback.
Source failures remain retryable, never pass/skip. Later values require readback.

## Owners and next gates

- Parent: all three new cohorts integrated; complete SDK composition found two
  real failures. Retain failed receipt, then replay the repaired joins and fresh
  benchmark before the next allowed-fork batch. Full CI remains required.
- TLS agent owns the only heavy slot: PC-relative source CFG first, result-slot
  address and alias/escape contract separately. Original source is never cropped.
- Numeric/frontend agents' completed source is committed; no active tests or
  external/local discovery jobs. Capacity interruptions lost no work or receipts.
One heavy source/SDK/object owner at a time; small bounded pure controls may overlap.
Preserve native execution vs QEMU/emulated diagnostics, all original fault effects,
canonical bool bytes (not arbitrary2/3 truncation) and unknown GP upper bytes.

## Receipts and history

Parent ignored `_out/register-result-repair/`: `factory-7752fc1f/`,
`publication-f1db0fa8/`, `metadata-71ae5584/`, `context-9361399a/`,
`combined-15e683d0/`,
`benchmark-2d696307/`, `ci-e544c001-{completion,llvm-parts}/`,
`ci-af4b6a3f-{main-aggregate,main-parts,simple-parts,llvm-aggregate,llvm-parts}/`,
`ci-af4b6a3f-native/native-x86-oracles/native.log` and all four independent
`ci-af4b6a3f-*-audit.json` files.
Numeric tree: `_out/sdk-numeric-maps/literal-minimum-partition-receipt.md`;
TLS tree: `_out/direct-tail-state-b7e29095/`; SDK tree:
`_out/register-result-repair/continuation-reference-closure/`.
Raw heads/tools/inventories and complete terminal unions are authoritative.
Use `git show <frozen-sha>:docs/development/current-work.md` for older history;
do not append old timelines here or promote failed/incomplete/emulated receipts.
