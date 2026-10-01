# Current work: PR 40 verification repairs

Read [validation](validation.md) and
[discovery verification](discovery-verification.md) before continuing.
Use `git worktree list` to find persistent development and frozen evidence
trees. Never edit a tree while its tests or corpus verification are running.

## Authority and promotion

Upstream PR 40 remains Draft. Repair batches use fork Draft PR 4, targeting
`codex/expand-ecosystem-corpus-20260913`, not fork main. Push only to `cpunion`;
do not update the upstream-connected branch until current-head fork CI,
review and exclusion-policy gates pass. The last inspected fork run,
36662878534 at `3aeb6b7c`, failed discovery shard 39 and its aggregate.
Later local repairs are not pushed. Keep funnel tables in the PR body.

## Immutable historical evidence

The imported scan has continuous ranges from `2025-10-01T00:00:00Z` through
`2026-09-22T22:43:08Z`. This completes that bounded historical inventory,
not the later incremental interval or the whole Module Index.
Standalone inventory and cgo records stay outside this repository.

Source `754cbebb` finished all 64 discovery shards but failed five. Audited
publisher `3c32ecc3` preserves those outcomes: complete accounting is not
verified success. Ten exact-version download failures have a checksum-verified
read-only proxy checkpoint. The other failure was a Go compiler/GC crash in
spidermonkeywasm2go; one successful retry does not establish a fix.
Keep that source, ledger, tools and reports unchanged. New source needs fresh
reports; never copy pass flags or combine different source revisions.

## Integrated development

The integration branch is `codex/pr40-guarded-integration-20261001`. Locate
it using Git, not a personal filesystem path. Its important repairs include:

- Exact zero-size asmdecl warning filtering, frontend registration inventory,
  strict ordinary source-selection proof and mandatory proof for every PASS.
- Declaration-backed Go ABI0 frame bridging, private-frame alias checks,
  source LR/SP return guards and real source RET-width semantics. Unknown
  native branches, escaped frames and missing entry contracts fail closed.
- One ARM64 CFG/control route for every public translation entry point;
  compile-only form probes do not invent an ordinary caller return contract.
- Shared typed raw ARM64 decoders: SVE, floating, integer/crypto SIMD and
  state/memory effects. Z/V/ZA registers are not GP registers; real stores
  taint saved continuation memory. Unknown exceptions/calls remain conservative.
- Shared complete paired-atomic grammar/effects, including Go's physical
  zero-offset named-SP forms. Operand probes preserve real source SP/LR or
  explicitly terminate; they never manufacture an ordinary return proof.
- Declaration-backed ARM64 ABIInternal aggregates, caller-save availability,
  narrow-bit transport and all seven VMOV duplication arrangements, with
  independent Go/LLVM runtime oracles. Hidden native registers remain explicit.
- ARM kuser/native-continuation support and complete barrel-shift/multiply
  flag semantics. An explicit address-only machine-entry shim captures physical
  state and bounded source stack storage; asyncPreempt is not yet supported.
- Compile-only exact Go package checks via `go list -export`, rather than
  executable linking; explicit target CPU macros and actual Go package-role
  experiment macro registration. Source-required profile planning and offline
  replay and production consumers bind source-required profiles to the same
  Go export/vet, actual package selection, CPP graph and LLVM objects. Schema 10,
  progress schema 2 and assembly-ledger v2 retain all four scope dimensions.
  Fresh reports remain required; implementation alone validates no old N/A.
- Architecture-aware asmdecl access widths: scalar broadcast reads use their
  actual memory width, and ARM MOVW is not an x86 two-byte access. Real Go
  declarations and LLVM 22 objects cover the affected forms; historical
  external-module exclusions still require an exact-source replay.
- ABI0 TEXT metadata accepts the logical data end or its Go register-size
  alignment, without creating FP fields in padding. Explicit x86 FP offsets
  remain authoritative even for uniquely named results; never relocate them
  by name. Padded float32 sqrt and stale-offset source counterexamples are
  covered by Go/LLVM object, numeric and rejection regressions.
  The shared x86 LEA family takes addresses of bound typed storage, not FP
  values or invented zeros; Go-rejected immediate FP spellings stay rejected.
- Caller-owned internal LLVM contexts for Go binding, with module-before-context
  disposal. Concurrent feature observations use canonical keys, deep-cloned
  results and actual-driver/subtool byte and route rechecks. Ordinary production
  CPP capture binds actual assembly/header bytes and Go include search to the
  exact module ZIP, with guards around package checks and translation. Its
  feature-profile/branch proofs retain actual driver, child-tool and consumer
  identity. Unknown includes, unconsumed profiles and missing proofs fail.
- Raw x86 near returns retain native width and imm16 cleanup. Zero-cleanup
  forms need a stack-unobserving leaf and bounded static/typed FP accesses;
  byte-exact naked TEXT cannot silently omit Go's prologue or FP transport.
  Exact slot width, signature field/type binding and nonoverlap are checked.
  Generic partial FP writes still need a separate high-byte preservation fix.
- Unbound ARM/ARM64 FP addresses fail instead of becoming zero pointers.
  Unknown compiler diagnostics and assembler failure footers cannot establish
  source N/A. Compact ledger source rejections retain a diagnostic digest and
  portable source positions; old reason-only skips cannot acquire invented
  evidence.

Checkpoint `865b3fa6` finished the full root suite with two failing top-level
tests: declaration-less ARM64 RetArg fallback and a raw x86 RET with nonzero
caller-stack cleanup. Corpus/scanner units passed. The RetArg fixture now
derives its complete ABIInternal contract from an actual Go declaration;
raw-return repairs remain under independent source-stack review. Earlier
checkpoint `794b98b5` had 37 failures, and `3900b033` passed the 31 unchanged
failure titles. These are exact-revision results, not current-head success.

Checkpoint `69e9f37a` passed production `go build ./...`, the full corpus unit
suite, feature-cache race tests and `go vet ./...`. No current-head full root,
strict standard-library or fresh external-corpus success is claimed.

The full root run at `ac3681dc` failed 48 top-level tests after strengthened
raw x86 return contracts; corpus and other root subpackages passed. Those
failures remain open and cannot be hidden by a later focused ABI regression
pass or by historical corpus results.

Later checkpoint `a50f0870` passed bounded root regression tests, the full
corpus unit suite, vet/build and all five official classification gates.
The raw-return predecessor also passed actual Linux amd64/386 Go/LLVM numeric
oracles with Go 1.27.1, LLVM 22.1.8 and pinned QEMU 10.2.3. This does not execute
external packages or establish the complete cross-runtime matrix.
Source-diagnostic repair `b853ac1c` retains true opaque-error/footer RED logs;
its full corpus, vet/build, focused Go 1.20 and ledger-witness roundtrip gates
passed before commit. Exact-slot companion `9415db0e` is integrated from an
independently frozen focused/Go 1.20/five-target-object green batch. Its source
still needs current-head exhaustive and external-corpus verification.

ARM source-frame companions `276e74f8`, `327df1ba` and `d2e373a8` are not yet
integrated. Their source SP/continuation guards expose 16 real ARM top-level
failures and a strict stdlib/benchmark failure at MD5's named local `end-4(SP)`.
Do not relabel them as green: ARM currently lacks Go NAME_AUTO local backing,
private outgoing ABI0 slots and the required callee/effect bridges.

Checkpoint `cd2b46a4` passed all five current official classification gates,
including ARM carry-dependent single-form probes with an explicit source CMP.
All 40 Go 1.20–1.27 source-table audits passed earlier: those are enumeration
checks, not eight toolchain runtime runs. ARM32 QEMU 7.2 diagnostic oracles
do not replace the required pinned QEMU 10.2.3 cross gate.

## Active independent work

1. Discovery feature profiles: freeze and verify the integrated production
   consumer before replaying all 64 shards. Actual Go registrations, driver/
   child-tool bytes and routes, package roles, CPP graphs and LLVM outputs bind
   each ordinary file/target/profile-ID/custom-tag scope. Old schema-9 evidence
   cannot be relabeled. Reevaluate historical source N/A; missing proof or
   infrastructure errors must not become a source skip. The ordinary matrix
   is non-test and cgo-disabled: test-only and cgo-enabled roles need their own
   explicit source/tool/consumer scopes, not an invented blanket Go exclusion.
2. ARM native entry/returns: extend the explicit physical shim only with closed
   continuation/effect proofs. Go accepts RET register operands that are not
   ordinary caller returns; preserve actual Go/runtime counterexamples.
3. ARM64 private register helpers: fold only complete same-file, register-only
   leaf call graphs into their caller CFG with a proved source continuation.
   Memory/frame escapes, unknown entries and hidden closure/native registers
   require separate contracts. Never guess scalar signatures from body shape.
4. Remaining root regressions: distinguish genuinely unsafe source returns
   from stale IR-shape assertions; fix complete typed effect families and retain
   unsafe-source negatives. Framed tail jumps must not acquire an invented
   epilogue. Keep genuine failures until independently demonstrated fixes pass.
   Raw x86 near returns must retain native return-width and imm16 stack effects;
   only a proved zero-cleanup, stack-unobserving leaf is an ordinary return.
   Include effective-address constants and subregister views in that proof.
   ARM framed returns must not assume their saved LR slot stayed intact, and
   ARM64 private helpers must not virtualize hidden platform registers or
   escape through excluded siblings and native-layout observers.

## Completion sequence

1. Finish bounded repairs with true red/green evidence, affected-target LLVM 22
   objects and runtime oracles. Commit promptly; batch pushes.
2. Integrate reviewed fixes, replace this checkpoint, commit and freeze source.
3. Run official/form gates, strict stdlib and benchmark, exhaustive root/nested
   tests, vet/build and required cross-runtime gates.
4. Replay all 64 discovery shards using bounded parallel batches and verified
   read-only proxies. Own and remove candidate caches; never clean global or
   live-writer caches. Audit before publishing matching assembly evidence.
5. Every selected version needs a tested pass or explained supported skip.
   Pending/failed prevent completion; historical passes are not current proof.
6. Push to fork PR 4, finish CI/review/coverage, then promote to PR 40.

Native-byte-layout/JIT exclusions remain provisional and require review.
Final-link/runtime checks belong in the separate compatibility repository.
Its schema-4 vendor proof is captured before Go oracles and rechecked after
execution; local workflows are not deployed. Old schema-3 diagnostics cannot
be promoted. llgo's old pinned plan9asm dependency still has a genuine failure;
a local replacement is development evidence, not passing pinned CI.
