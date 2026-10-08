# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Inspect actual HEAD/status,
the frozen report inputs and terminal owners before continuing. This checkpoint
is not a current-head completion certificate; preserve historical failed evidence.

## Contribution and execution boundaries

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to cpunion.
Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft. Do not
promote its upstream-connected fork head until current-head fork CI and review
pass. Root alone publishes assembly-ledger updates.

Published `c09eff1e`, run `37711752851`, finished failed naturally: four
priority shards failed and two reached the deadline. Aggregate/completion checks
failed; skipped downstream suites are not passes. Inspect the original reports
with their frozen tools. Batch the successor before pushing.

Continue integration on `codex/pr40-next-batch-20261009`, based on `2c943968`.
The preceding integration branch and component gates remain frozen
in separate persistent worktrees. Do not add overlapping focused passes into a
full-suite total or transfer their certificates to a later combined head.

No local background index/discovery/external corpus. Inventory remains on its
designated source-only scanner; external compilation/execution belongs on a
separate disposable validator, never the inventory server. Local first-party
fixtures and corpus-driver unit tests remain allowed.

## Latest integrated batch

The preceding batch already contains the complete register/GP-SIMD, shift,
broadcast, horizontal/dword-product, raw encoding, source-CFG/result and frame
rule-class repairs. Preserve independent Go bytes, native numerical/fault/canary
and LLVM optimization axes. Logical FP bytes never grant native backing.

The following components are now integrated through `2c943968`:

- Downloaded frames: exact source/ZIP/member/selection, independently captured
  compiler/assembler objects, Windows auxiliary SEH closure, fixed-local binding
  and CLI transport. The frozen `53797b7c` family closes 176 roots and 5,654
  total nodes in ordinary/race scopes. Framework `8754ea0c` removes a Go-1.20
  incompatible test-only Chdir call; current/historical focused scopes pass.
- Numerical SB reads: live complete original declarations/options, exact
  instruction/code/PC/ordered relocation joins, typed storage owners and
  post-consumer revocation. The `1380e2fe` component ordinary/race each passes
  404 nodes; later cancellation-cause and Linux C-label fixes are integrated.
  A mutable extent is not initialization, an SB address, pointer or write grant.
- Numerical fault resources: `ab6ab6c8` ordinary/race each passes 41 nodes.
  Every generated fault child checks both core limits before protection/call.
  The lock-free signal-safe handler accepts only synchronous faults inside the
  exact protected region; unrelated faults, raised signals and normal returns
  fail. This host diagnostic does not certify Linux/386 execution.
- ARM64 same-carrier results: `e1fc500a` ordinary/race each passes 78 roots
  and 1,219 children; full storage and Go-1.20 framework scopes also close.
  Nine wider stores plus the exact-width control retain original typed value,
  native tail bytes and canary through all three APIs at LLVM O0/O2. Five-OS
  objects pass. The explicit private scalar consumer uses the actual same
  outgoing backing, complete native entry/body effects and original bounded
  lease. Its data capacity is Auto-16, excluding both frame-chain cells.
- Mixed authorities: actual `cb988eb3` regressions demonstrate ARM64 dispatch
  silently ignoring a numerical token, factory minting mixed options and a
  private numerical seal omitting the foreign ARM64 role. `bac7ec91` fixes
  the whole class; ordinary/race each passes 18 nodes. Entry, factory and
  independent post/shape guards reject mixed roles before callbacks/output and
  permanently revoke actual original owners, without nested source locks.
- Explicit SDK source observation: `1284d26f` closes ordinary/race storage
  scopes of 224 nodes and source/factory scopes of 161 nodes, with independent
  SDK object-reader replays. The default owned role still rejects SDK sources.
  Standard/module-free selection, physical installation, whole selected source,
  import graph, object flags and original capture lifetime remain separately
  sealed. No SDK numerical binding, generated-header or DATA permission passes.

Additional integrated repairs distinguish the existing 64-MiB tool-read role
from the unchanged 16-MiB source/export/ZIP limits. Original Windows objects
and shared frame fixtures bind canonical physical directories; an explicit
aliased-temporary-parent regression covers the whole source-join path.
Production physical-input checks and all shared budgets remain unchanged.

Root independently reread the component source/tool/binary/raw inventories and
terminal owners before integration. These are component certificates, not
combined-head full-suite, whole SDK, external-library or CI completion.

## Current failure classes and next work

1. Fix the actual host-tool fixture mutation. Two complete other-main-package
   ordinary runs at `2c943968` execute their owned tests successfully but fail
   the outer frozen-input check: the portable dependency fixture hard-links the
   actual LLVM shared library, then cleanup changes that host inode's ctime.
   Add a real metadata/byte-identity RED and independently copy bounded full
   bytes. Do not discard ctime checks or relabel either old overall failure.
   Close ordinary/race, complete other-main-package and both nested CLI scopes.
2. Review the frozen SDK namespace/referenced-metadata component. Hashed,
   static and DUPok definitions retain object-local identity; never choose a
   body by name+ABI. Foreign references use actual compiler SymbolDefs indices.
   Complete referenced structure is streamed without unused body/index arrays;
   default owned-reader bounds stay unchanged. Actual CRC32 now reaches ASM
   DATA `r2r1`, which needs a separately reviewed source/effect role.
3. Review private original ARM64 sleep observation and the independent Linux
   GP/NZCV oracle. Observation is not byte, stack, syscall or SDK permission.
   Original-kernel success must precede a real LLVM mismatch; a host compiler/
   vet SIGSEGV remains a precondition failure. An explicitly recorded diagnostic
   scheduler environment is not default-scheduler CI completion. Cover both
   lowering and complete source/normalized GP/flag effects before permissions.
4. Repair remaining SDK reads, dynamic reflect frames, native entry/TLS/
   continuations, result padding and complete body closure by rule class with
   independent first-party regressions, not library-name instruction exceptions.
5. Freeze the combined head, then run affected ordinary/race, exhaustive root,
   all other main-module packages, both nested CLIs, Go-1.20 framework,
   official encoder/decoder, strict stdlib/benchmark and required native CI.
   Full-root budgets stay 45m, race 60m; bounded disjoint partitions and one
   collector are allowed. Dispose modules before Contexts and wait for all
   owned processes before cleaning exact owned caches.

The independently closed strict five-target SDK baseline `2c943968`
attempts all 184 files: 167 pass and 17 fail, without skip/N/A/unattempted files.
Benchmark and standard target matrix therefore fail. Official five-arch
classification and ARM64 decoder gates pass separately, not as 100% object or
runtime coverage. Remaining classes include SDK reads, dynamic reflect frames,
native entry/TLS/continuations, padding and complete SDK body closure.
Never bless a new coverage fingerprint to hide them.

The preceding `7e00cab4` complete ordinary root run attempts all 32 partitions:
5,196 roots and 96,174 children close, with two roots and three children failing
at Windows aliases. Its 39 named skips remain skips. At `2c943968`, the repaired
Windows/frame scope closes 132 nodes in each ordinary/race lane; historical
Go-1.20 closes 73. These focused passes are not a complete root-suite pass.
The frozen `2c943968` Linux numerical matrix closes three roots and 18 children,
including the unchanged 386 fault/canary checks, with no skip/incomplete nodes.
Native ARM64 QEMU-i386 avoids nested-emulation deadlines while Go 1.27.1,
LLVM 22.1.8 and budgets stay pinned. This is not physical native x86 execution
or the whole cross-runtime matrix.

## Inventory and ledger

The committed assembly ledger still has three historical Snappy passes and
6,064 pending, no complete shard, complete=false and verified=false. Only the
automatic publisher may promote complete schema-11 reports with matching
frozen source, scan, tool, file/target/profile and consumer proofs. First-party
instruction tests, observation-only sources and runtime examples never
hand-upgrade a module.

Committed inventory retains 215 continuous ranges, the 2025-10-01 cutoff and
head 2026-09-22T22:43:08Z. The cutoff is reached; history_complete instead means
the entire official index to its epoch. The bounded standalone two-worker
incremental service has newer unsynchronized records. Inspect its actual owner,
manifest and journal; do not hot-replace its binary, duplicate the writer or
merge a live unvalidated snapshot. Streaming inventory repair is separate from
the running service. cgo records stay outside this repository.
Sharded status, batch lookup and publication stream records rather than
retaining the complete ledger payload; legacy monolithic imports remain explicit.

Recheck discovering samples and all 64 shards after systematic repairs.
Partial reports, Context and provisional exclusions are not passes. Keep scan
and coverage funnel tables in the PR body, not this checkpoint. Do not claim
CI/review/coverage completion until all required current-head gates close.
