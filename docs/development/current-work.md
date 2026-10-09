# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). This checkpoint is not a
completion certificate. Inspect actual source, report inputs and terminal owners;
retain original failures and never transfer a component pass to a combined head.

## Contribution and execution boundaries

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to cpunion.
Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40) stays Draft.
Do not update its upstream-connected head until current-head fork CI and review
pass. Root alone publishes assembly-ledger updates.

Published `c09eff1e`, run `37711752851`, finished failed naturally: four
priority shards failed and two reached the deadline. Aggregate/completion checks
failed; skipped downstream suites are not passes. Batch repairs before pushing.

The frozen preceding integration head is `ab5302a1`, on
`codex/pr40-next-batch-20261009`. Continue the new combined batch on
`codex/pr40-carrier-literal-sleep-integration-20261009`. Preserve the preceding
integration and component worktrees; do not mutate a checkout with live consumers.

No local background index/discovery/external corpus. Inventory stays on the
designated source-only scanner; external compilation/execution belongs on a
separate disposable validator, never the inventory server. Local first-party
fixtures and corpus-driver unit tests remain allowed.

## Independently closed preceding integration

At the exact frozen `ab5302a1` head:

- Complete ordinary root: all 32 partitions close, with 5,217 roots and
  97,101 children passing, 39 explicitly named host-inapplicable root skips,
  and zero failure/incomplete nodes. This diagnostic used four workers;
  the original CI two-worker timing and required cross-runtime counterparts
  remain separate gates.
- Complete corpus-driver ordinary: all 32 partitions close, with 637 roots
  and 2,309 children passing, zero failure/skip/incomplete. The successor uses
  the original two workers, `-p=2`, GOMAXPROCS=2 and 45-minute budgets.
  Preserve the preceding four-worker run's three real fixture timeouts.
- All 16 other main-module packages: 552 roots and 3,351 children pass,
  zero failure/skip/incomplete.
- Nested CLI ordinary: plan9asm passes 23 roots/14 children;
  plan9asmll passes 229 roots/1,721 children.
- Official five-arch classification and the complete required ARM64 decoder
  gate pass separately. Context forms and decoder-only Go rejection do not
  establish complete LLVM object or runtime coverage.

These scopes have independently read source/SDK/tool/binary identities,
complete raw EOF, actual inventories and closed owners. Different module-owner
worker plans are not one coherent full-module certificate. Complete current-head
race, historical framework, strict SDK/benchmark and native CI remain pending.

The head includes complete typed instruction/caller/frame repairs, downloaded
original source/object joins, tool-fixture isolation, bounded per-validation CPP
indexes, ARM64 syscall/virtual-NZCV/carry grammar and live SDK DATA/TEXT
observation. Observation is not numerical-read or native-entry authority.

## Current component batch

The additional two-scalar ARM64 carrier component is integrated from
`585c67ce`, `16d3a828`, `7a0e0247`. Its original component head independently
closes 63 roots/947 children in both ordinary and race, actual Go-1.20 framework
28 roots/650 children, and required Linux original-Go/three-API LLVM O0/O2
runtime (760 checks per lane, one root/six children). All nine wider native-store
result combinations and the exact-width control retain declared values, native
bytes and canaries. The second parameter initializes only its four bytes,
not the following gap. SDK, pointers, R2/R3 and Or32/And32 are not granted.
Original producer-P/consumer-RED, reused-PID reader failure and Docker setup
failure remain retained; new independently closed readers do not rewrite them.

Two components are still awaiting complete handoff before integration:

- SDK readonly literal initializers: `df06fc81` plus test-lifecycle successor
  `b574efa3`. Actual original/mapped/archive DATA, original Go objects and ELF
  bytes are checked. Private integer/float/string and zero-tail bytes do not
  grant pointers, mutation, DUPok, relocation or native-vector permissions.
  Initializer observation alone is not SDK numerical consumption or CRC32 success.
- ARM64 nanosleep private input: `170c90a8`, based on `ab5302a1`.
  Exact Linux SVC0/R8=101/R1=NULL consumes a fully initialized private
  16-byte timespec, without native-frame or result/NZCV permission.
  Ordinary/race each close 81 roots/1,253 children and actual Go-1.20 closes
  nine roots/279 children. Actual Linux kernel/LLVM execution and the complete
  13-file runtime SDK counterpart remain independently pending.

## Required next gates and remaining rule classes

1. Complete and independently read each component handoff; cherry-pick only
   its ordered commits into the persistent successor. Freeze the combined head
   before testing. Do not relabel earlier certificates.
2. Run combined affected ordinary/race, exhaustive root with the original
   two-worker plan, all other packages, both nested CLIs, Go-1.20 framework,
   official encoder/decoder, strict stdlib/benchmark and required native CI.
   Root budgets stay 45 minutes ordinary and 60 minutes race. Bounded exhaustive
   disjoint partitions retain full inventories and one collector.
3. Repair SDK numerical reads, dynamic reflect frames, native entry/TLS/
   continuations, full original result carriers and complete body effects
   by typed rule class with independent regressions. Do not use library-name
   instruction exceptions or infer authority from spelling, hashes or flags.
4. Run the discovering samples and all 64 external shards on the designated
   validators, then automatically publish matching complete schema-11 reports.
   Push one verified batch to the fork and let failure-first CI finish.

The frozen strict five-target SDK baseline `2c943968` attempts all 184 files:
167 pass and 17 fail, with zero skip/N/A/unattempted. Thus strict benchmark and
stdlib matrix still fail. Successful isolated bodies do not certify full SDK
files. Never bless a coverage fingerprint to hide failures or increase budgets
to conceal repeated work.

Dispose LLVM modules before Contexts. Serialize heavy Docker/full-suite gates
on a resource-constrained host. Preserve selected lazy Go tool binaries inside
their recorded cache paths for readback; do not delete an entire cache merely
because its test process has closed. Clean only exact owned, unreferenced data
after all consumers and independent readers have closed.

## Inventory and ledger

Committed assembly ledger: three historical Snappy passes, 6,064 pending,
zero complete shards; complete=false and verified=false. Only the automatic
publisher can promote complete schema-11 reports with matching frozen source,
scan, tool, file/target/profile and consumer proofs. First-party instruction,
observation or runtime tests never hand-upgrade a module.

Committed inventory retains 215 continuous ranges, cutoff 2025-10-01 and head
2026-09-22T22:43:08Z. The cutoff is reached; history_complete instead means the
entire index to its epoch. Newer remote records are not synchronized. Inspect
actual owner, manifest and journal before importing; do not duplicate the writer
or hot-replace its binary. cgo records stay outside this repository.
Sharded status, batch lookup and publication stream records instead of retaining
the whole ledger payload. Legacy monolithic imports remain explicit.

Keep scan/coverage funnel tables in the PR body, derived from validated reports.
Partial reports, Context and provisional exclusions are not passes. CI/review/
coverage completion requires every required current-head gate to close.
