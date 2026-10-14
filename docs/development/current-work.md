# Current work: PR 40 checkpoint

Checkpoint: 2026-10-14 05:22 UTC. Contribution source: `f1db0fa8`.
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
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | `codex/pr40-external-corpus-split-20261011`; parent integration, source `f1db0fa8` |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | `codex/pr40-windows-owned-process-probe-20261013`; fork-only workflow, `af4b6a3f` |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | `codex/pr40-sdk-numeric-maps-entry-20261013`; verified `ba474753`, idle |
| `plan9asm-wt-pr40-sdk-tls-20261013` | `codex/pr40-sdk-tls-20261013`; `10aa1cbd`, direct-tail TDD |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | `codex/pr40-sdk-storage-snapshot-20261013`; `b43caaa1` + uncommitted continuation/publication repair |

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

Trusted observation completed 2026-10-14 05:14:07 UTC:
`incremental_since=2026-10-09T19:22:43Z`; `history_before=2025-10-01`;
`history_complete=false`. The next hourly window is unfinished; never advance
its cursor from partial progress.
Index539 ranges/21,713,183 entries; unique_modules1,062,663;
scanned_records1,099,911; matched_records6,757; assembly_files85,385;
failures16,816; latest_records992,517. Cgo:1,094,086 records/23,185 modules,
2,397 failures/1,883 pending. These are server statistics, not translator passes.
Source failures remain retryable, never pass/skip. Later values require readback.

## Owners and next gates

- Parent: factory and publication integrations/readback finished. Integrate
  verified metadata/tail/input-protection families next, replay their exact joins,
  then batch the next allowed-fork update. Full race remains required.
- Numeric maps agent: metadata-only/no-DATA accessor and factory reuse the
  existing observer. Eight pure/race/Go20 roots pass; six-file source is frozen
  for actual three-OS complete empty-DATA and lifetime verification.
- TLS agent: `10aa1cbd` direct-tail state-loss light IR TDD RED (six shapes ×
  three APIs + negatives); design actual whole-source graph/root must-state,
  then original-Go/LLVM runtime counterexample after explicit heavy clearance.
- SDK/frontend agent: verified `f3e67767`; next consumed-input collision repair
  reuses the plain assembly reader's same-pass input report, never another parser.
One heavy source/SDK/object owner at a time; small bounded pure controls may overlap.
Preserve native execution vs QEMU/emulated diagnostics, all original fault effects,
canonical bool bytes (not arbitrary2/3 truncation) and unknown GP upper bytes.

## Receipts and history

Parent ignored `_out/register-result-repair/`: `factory-7752fc1f/`,
`publication-f1db0fa8/`,
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
