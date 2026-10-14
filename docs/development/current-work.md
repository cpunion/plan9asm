# Current work: PR 40 checkpoint

Checkpoint: 2026-10-14 12:31 UTC. Contribution source: `3f972336`.
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
| `plan9asm-wt-pr40-scalar-lane-native-20261011` | `codex/pr40-external-corpus-split-20261011`; integrated source `3f972336`; full SDK composition still pending |
| `plan9asm-wt-pr40-stdlib-preprocessing-20261012` | `codex/pr40-windows-owned-process-probe-20261013`; staged fork-only priorities, `1d9e009b`; not pushed |
| `plan9asm-wt-pr40-sdk-numeric-maps-entry-20261013` | `codex/pr40-sdk-numeric-maps-entry-20261013`; scoped bool/literal leaf `6e63c125`; repairing existing logical bool callers |
| `plan9asm-wt-pr40-sdk-tls-20261013` | `codex/pr40-sdk-tls-20261013`; caller leaf `92742505`; authenticated NUM node/emitter composition still dirty |
| `plan9asm-wt-pr40-simple-sdk-bindings-20261013` | `codex/pr40-sdk-storage-snapshot-20261013`; GP/X objects `b4f6d916`; fixing full bit-scan zero semantics |

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
That snapshot still had alternate-extension consumed-input collisions. Its
actual six-child RED is preserved; the later `15e683d0` cohort repairs the class.

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

Fork priorities retain13 SDK lanes and the original budgets. The source lane
now also inventories all eight plain-reader roots and the child-harness mode
root. Required i386 packages join the existing bounded LLVM22 installation,
without a second apt transaction. Mandatory native execution adds AMD64/386
direct-tail and PC/result/external follow-up roots: the new52-root profile,
seven inventories/fourteen required roots, is not yet executed. All30 policy
regressions and actionlint pass. Native audit mutation controls pass8 roots;
synthetic expanded profiles are not execution receipts.

## Latest rule-class leaves (not full SDK or CI certification)

`06ac4707` shares the existing complete24-opcode packed-unpack grammar and pure
byte dependency calculation between independent caller/result consumers.
Ordinary/race/Go20 each pass6 roots/200 children in0.579/1.939/0.550s.
Memory permissions remain separate: masks do not shrink E4NF reads, logical
bytes do not prove legacy alignment, and MMX has no SIMD-state permission.

`36cf733a` is integrated as `d32c86cf`: original Prog PC coordinates, exact
declared ordinary external terminals, and bounded logical result-slot aliases.
Final frozen source passes Go27 71 roots/1544 children13.484s, Go20 70/1539
12.603s, race71/1544 23.246s, CLI6/100 normal4.702s/race8.135s, vet/build and
both unchanged official gates. The extra native owner passes original-Go44
and LLVM264 checks with canaries in21.720s locally emulated; native CI pending.
The earlier complete count/indexbyte replay remains FAILED4.333s: CPU NUM
binding and unpack/sparse-byte composition were not closed in that snapshot.

`6874bcd3` is integrated as `d89b5c92`: sparse GP/X numerical byte propagation,
exact overwrite and pointer/address/canonical revocation, no guessed high bytes.
Its26-root/693-child scoped union passes Go27/race/Go20 in2.629/8.401/4.840s.
The six original plain-GP/X API admission failures are repaired by `7f26c84a`,
integrated as `0fff8be9`: ordinary registered GP/X views with integer results
require the existing cached complete-source proof. The frozen parent union
passes31 roots/1357 children normal2.049s/race7.942s; agent Go20 also passes.
No PEXTRD probe or guessed high bytes repair the original source.

Child harness leaves `0a1cee92`/`21960a76` are integrated as `eff931be`/`94c198f6`.
Framework child builds explicitly select the compiled parent's LLVM mode,
including empty tags in default mode; original SDK environments are untouched.
Parent frozen `d89b5c92` replays all nine reader/harness roots and five children,
including a real trimpath child: Go27 default/byollvm PASS20.467/20.064s, actual
Go20 default PASS26.955s. All inventories and before/after heads match, no skips.
These local results do not certify actual Windows or changed-head full modules.

## Current rule-class integration

`9acdf6a6` models all six BSF/BSR W/L/Q Go rows and a bounded ZF-associated
pending result. Only its nonzero edge defines formerly unknown destination
bytes; joins, raw roles, replaced flags and unknown upper bytes cannot borrow
the fact. Its9-root/412-child union passes Go27/race/Go20 in0.667/2.813/0.700s.
This conservative first leaf still rejects an initialized-prior zero-result
case in the unchanged `goHexWordOps`; the separate native opcode/FLAGS repair
is pending. Never change that original conformance expectation to hide it.

`047796dd` extracts one complete vector-test shape for lowerer and observer.
The exact comparison-byte/test/ZF-nonzero/movemask chain supplies a bounded
nonzero GP fact, not arbitrary initialized-vector nonzero knowledge. Different
widths, registers, domains, raw/MMX/K results and intervening writes revoke it.
The17-root/726-child union passes Go27/race/Go20 in0.645/2.381/0.652s.
Parent `42ea73ef` separately passes19 roots/726 children in0.734/0.606s on
Go27/Go20. These are source/IR checks, not physical AVX execution.

`b4f6d916` is integrated as `42ea73ef`: complete positive and independent
unknown-return GP/X sources, seeded original-only ABI0 callers and canaries.
Agent33 roots/1357 children pass Go27/race/Go20 in6.466/14.819/3.486s;
five-target original-Go and60 LLVM object axes pass one root/80 children3.855s.
Docker startup stalled before any root started and is an infrastructure failure.
Original executable, emulated and both required native owners remain pending.
No object or synthetic audit receipt may stand in for those executions.

`92742505` is integrated as `627cacd2`: existing complete movemask/broadcast/
vector-test grammars preserve unwritten caller addresses; BSF/BSR revoke only
their exact GP destination. Its20-root/151-child union includes actual Go/LLVM
objects and passes Go27/race/Go20 in5.721/12.191/5.523s.
`3f972336` adds independent32-byte typed-entry origin IDs. Only full-word
same-origin address subtraction cancels the opaque base into numerical bytes;
mixed origins, partial writes, stale joins and nonordinary effects stay Context.
The cached complete-source gate also bypasses the unchecked AMD64 fast-module
route. Original pointer/native/bounds/result-alias permissions remain separate.
Its frozen25-root/198-child union passes Go27/race/Go20 in5.719/10.044/5.469s,
with fresh exact inventories, no skips, equal six-file seals and vet success.

TLS dirty-v4 complete SDK2 replay remains FAILED71.688s: all six API children
ran; Count needed exact original helper-node emitter identity, Index needed
same-origin subtraction. The later complete original Count3-TEXT/3-API replay
passes translation100.488s on dirty-v5, not objects, runtime, fullSDK2 or SDK8.
The source join and emitter leaf must still be sealed/committed and replayed
with `3f972336` and the pending zero-input repair. Saved timing attributes91.60s
to three APIs and8.68s to combined setup; finer stage durations were not logged.
Repeated actual source/export verification needs measured optimization, never
deleted guards or increased budgets.
After joining `3f972336`, dirty-v8 also translates the unchanged complete
IndexByte3-TEXT source through all three APIs in116.540s. The fresh combined
two-file/six-API union and completeSDK8 are still pending; no object/runtime
or parent-head certification follows from these individual dirty-tree passes.

Numeric `6e63c125` is scoped pure/IR evidence only, not integrated yet. Its
canonicalFP and physical-I8 intermediates repair raw2/3/255 washing, but the
existing complete logical bool-frame fixture still fails `BoolCaller` across
all three APIs. The follow-up complete84-root Go27/race and83-root Go20 unions
remain FAILED, each with exactly that one root/three children. Preserve the old
fixture; close readonly same-file callee/outgoing-byte/source identity first.

Fork-only `1d9e009b` retains13 lanes and original budgets, adds12 exact predicate
source roots and both native sparse-GP/X owners. All32 policy tests/actionlint
pass. The planned54-root native profile requires eight inventories/16 mandatory
roots. Its12 auditor mutation controls pass, but expanded logs are synthetic,
not execution. The fork still lacks later source merges; do not push it yet.

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

Trusted latest manifest generated 2026-10-14 11:12:53 UTC:
`incremental_since=2026-10-10T19:22:43Z`; `history_before=2025-10-01`;
`history_complete=false`. The next hourly window is unfinished; never advance
its cursor from partial progress.
Index22,093,583 entries; unique_modules1,070,612; matched_records6,789;
scanned_records1,117,175; failures16,815; latest_records1,001,014;
cgo_records1,111,351. These are inventory records, never translator passes.
All563 manifest ranges were independently checked continuous; their sum equals
index_entries and first/last endpoints derive history/incremental cursors. Periodic
exports can lag the completed manifest/journal; never interpret that as rollback.
Source failures remain retryable, never pass/skip. Later values require readback.

## Owners and next gates

- Parent: latest leaves integrated and actual child verified; full SDK composition
  still needs replay after every known rule/admission repair. Run a fresh strict
  benchmark before the next allowed-fork batch. Full CI remains required.
- TLS agent owns the only heavy slot: authenticated NUM/original-node graph
  composition, then complete SDK failures. Original source is never cropped.
- Numeric agent: independent386 canonicalFP/physical-I8 and readonly logical
  same-file calls; complete existing bool-frame failure remains open.
- Frontend agent: complete bit-scan zero semantics and original conformance;
  native sparse GP/X execution remains pending. No local scan.
One heavy source/SDK/object owner at a time; small bounded pure controls may overlap.
Preserve native execution vs QEMU/emulated diagnostics, all original fault effects,
canonical bool bytes (not arbitrary2/3 truncation) and unknown GP upper bytes.

## Receipts and history

Parent ignored `_out/register-result-repair/`: `factory-7752fc1f/`,
`publication-f1db0fa8/`, `metadata-71ae5584/`, `context-9361399a/`,
`combined-15e683d0/`,
`unpack-{complete,final,v2}-*`, `harness-94c198f6-*`, `reader-d89b5c92-*`,
`bitscan-family/`, `caller-origin/`,
`benchmark-2d696307/`, `ci-e544c001-{completion,llvm-parts}/`,
`ci-af4b6a3f-{main-aggregate,main-parts,simple-parts,llvm-aggregate,llvm-parts}/`,
`ci-af4b6a3f-native/native-x86-oracles/native.log` and all four independent
`ci-af4b6a3f-*-audit.json` files.
Numeric tree: `_out/sdk-numeric-maps/literal-minimum-partition-receipt.md`;
TLS tree: `_out/direct-tail-state-b7e29095/`,
`_out/direct-tail-pc-result-15e683d0/RECEIPT.md` and `v2-*`; SDK tree:
`_out/register-result-repair/continuation-reference-closure/`.
Raw heads/tools/inventories and complete terminal unions are authoritative.
Use `git show <frozen-sha>:docs/development/current-work.md` for older history;
do not append old timelines here or promote failed/incomplete/emulated receipts.
