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
`codex/pr40-next-batch-20261009`. Preserve preceding integration and component
worktrees; do not mutate a checkout with live consumers.

The original two-worker combined root run at `ae0eee75` closes failed at its
45-minute outer deadline: 31/32 shards pass, the last remains incomplete.
The original owners are gone; strict independent readback rejects the timed-out
owner instead of granting a complete pass. Retain that failure. The preceding
root-part/parsed-grammar batch and its focused gates remain frozen at `eb8a8779`.
The frozen `8f095690` successor includes the GNU label and VMOV lifecycle
repairs below. The new development successor is
`codex/pr40-oracle-module-integration-20261009`, which also includes the complete
strict assembler-oracle repair. Freeze it before its own complete gates;
component passes do not transfer to this head.

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

SDK readonly literal initializers are integrated from the seven-commit chain
ending at `b574efa3`. Root independently replays the exact ordinary/race union:
123 roots and 1,088 children pass per lane, plus Go-1.20's five roots/65 children
and actual Go27/Go20 vet/build. Every required original/mapped/archive DATA/TEXT,
Go object and linked ELF initializer reader closes. Private integer/float/string
and zero-tail bytes do not grant pointers, mutation, DUPok, relocation or
native-vector permission. Initializer observation is not SDK numerical
consumption or CRC32 success. Earlier callback-timer, Go20 metadata and exact
Darwin linker-warning harness failures remain separately retained.

ARM64 nanosleep is integrated from the 13-commit chain ending at `19b60584`:

- Private input: `170c90a8`, based on `ab5302a1`.
  Exact Linux SVC0/R8=101/R1=NULL consumes a fully initialized private
  16-byte timespec, without native-frame or result/NZCV permission.
  Ordinary/race each close 81 roots/1,253 children and actual Go-1.20 closes
  nine roots/279 children. Actual Linux original-Go kernel axes pass, but the
  six LLVM runtime children initially fail compiling the C driver's unquoted
  ELF package/path label. The test-only `19b60584` quotes that unchanged symbol;
  its actual Linux rerun and root's independent reader close one root and nine
  children without fail/skip/incomplete. Each LLVM lane checks the intercepted
  input equation, three short real sleeps and a real EINTR with NULL rem.
  The initial Linux failure remains retained. The complete actual runtime SDK
  counterpart selects 11 assembly files (13 archive members include the two
  compiler members): seven compile and four fail, with zero N/A. Whole-file
  sys_linux_arm64.s now advances past usleep to walltime's native saved-frame
  context failure; asm_arm64.s, preempt_arm64.s and tls_arm64.s also fail.
  This partial result and every changed combined-head gate remain separate.

The parsed-op diagnostics component, `codex/pr40-parsed-op-diagnostics-20261009`,
retains actual RED macro/target-report cases and 30 Go-accepted quoted-DATA
parser failures. Its `01d44f6f` source checkpoint derives hints from the complete
actual parsed input, clears them per file, and preserves whole quoted tokens
when splitting semicolon statements. Focused ordinary/race and Go-1.20 parser
gates pass; 30 DATA fixtures retain exact padded bytes through three APIs and
LLVM 22 object compilation on all five targets. Nested diagnostic ordinary/race
passes separately. These are not complete-module or external-ledger passes.

The CI root-parts executor component retains its original RED fixtures. Its
`f0177658` checkpoint adds eight root jobs, once-only corpus/rest jobs, fresh
aggregate inventories and strict raw/coverage union checks. Its bounded rerun
and root's independent reader close six roots/38 children, the existing CI
policy root, Node policy, actionlint and shell checks. Earlier setup and exact
Go temporary-prefix-map nonce wrapper failures remain retained; the new wrapper
normalizes only the actual direct-child uint32 nonce and preserves raw input.
These commits are integrated into the successor; complete legacy runner,
race/Go-1.20 results are recorded below; other changed combined-head gates
remain required.

At exact combined head `8f095690`, the complete runner package now closes
ordinary and race with 59 roots/155 children each, zero failure/skip/incomplete,
and actual Go-1.20 framework with ten roots/six children. Root's fresh independent
readback verifies the complete actual inventories, raw EOF and closed owners.
The 35 Node policy tests, actionlint, shell checks and runner build/vet also
close. This is the runner package, not the complete root/module, native CI,
SDK benchmark or an external-library pass. The original timeout remains failed.

The actual eight-part module run at that head remains failed: part zero closes
with 648 roots passing and seven named host skips, but part one is interrupted
by its owner's one-second process-query timeout. The remaining six parts,
corpus/rest owners and aggregate were not attempted. Read-only process and
source checks close the observed failed generation, not the module certificate.
Repair the bounded observer and rerun the complete plan on the next frozen
integration head; never reuse this partial pass or enlarge test deadlines.

At the same frozen head, the original five-architecture official inventory
script closes naturally, and root independently replays its unchanged Python
checker against all five retained complete reports. Source, SDK, selected tools
and report generations remain stable. All original fingerprints match and
unsupported/parse-error counts are zero, but 260 forms still require context.
This inventory check is not complete LLVM object, ABI or runtime coverage.

The GNU C-label component retains its actual Linux compiler RED at `f6e7ea58`.
The test-only `ebfce45f` keeps the exact LLVM symbol while distinguishing
Clang's raw symbol from GNU's assembler quoting layer. Root's independent Linux
readback closes one root/14 children: ordinary, slash-qualified and nested-path
symbols compile, link and execute with both actual compilers; old GNU rejection
controls remain. The discovering CMOV Linux/ARM64 GCC path is still a separate
required runtime gate, not certified by these tiny AMD64 fixtures.

VMOV's original Go runtime command now uses the shared deadline and owned
process-group lifecycle. `54a7918c` first fails its newly added AST regression;
`bdea351d` closes affected ordinary tests and the complete focused race/Go-1.20
scope (13 roots/202 children each, no skip/failure/incomplete). The original
506-form matrix, four LLVM targets, three APIs and actual host Go/LLVM runtime
checks remain intact. This lifecycle repair grants no new instruction or ABI
permission, and does not replace the required Linux runtime counterpart.

The assembler-oracle component and SDK development retain these distinct scopes:

- The ordered assembler-oracle chain ending at `1fd11b98` is integrated into
  the development successor. Original real-tool REDs remain preserved. Its
  shared strict observer binds exit status, complete separate streams and
  original statement positions; accepted cases require complete actual objects.
  Whole identifiers follow Go's lexer rather than mnemonic-name membership.
  Component ordinary/race each close 24 roots and 140 children, zero skips.
  Independent readers reconstruct split output records and recheck every
  instrumented owner; the root scope has no individual inner-owner records.
  The original complete decoder script closes with 703 accepted objects and
  91 concrete source refusals across all 794 cases, zero unknown failures;
  all 800 recorded query/case owners are absent and the baseline is unchanged.
  This is not combined-head, LLVM/runtime, SDK or external-module success.
  Its ignored launcher lacks a total owner timer and reviewed cleanup guard;
  the actual source/object result does not certify that harness contract.
  A separate real Go-1.20 run retains three portable roots passing and the two
  fixed-current-SDK producer roots failing the object-schema pin. The successor
  integrates their explicit `go1.27` build-role split without changing either
  test body; shared helpers and all portable checks stay untagged, never Skip.
  Fresh split-head current/historical inventories and complete gates remain
  required; the preceding component results cannot establish them.
- SDK literal-read models remain query-only. At `85d36df1`, the complete
  statement/listing grammar scope closes ordinary and race with 20 roots and
  346 children each; the actual Go-1.20 framework closes 18 roots/238 children.
  Original fixture and canonical-number failures remain retained. The real
  whole-source join at `136336a2` still fails after its complete original CRC
  source, initializer and same-SDK CPU prerequisites pass. Root independently
  checks all six original/mapped/archive DATA/TEXT readers. Public numerical
  capture remains Context and borrowed owners/commands are unchanged. Connect
  this actual source/tape/listing/object seam before numerical consumption;
  do not integrate the retained failing consumer or claim CRC runtime success.

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
