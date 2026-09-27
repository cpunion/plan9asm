# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). Repair CI before starting
another module-index inventory scan.

## Contribution and worktrees

- PR 40 remains OPEN/DRAFT. Latest live inspection: remote head `4cf5ade`,
  base `7cc8c0f`, 74 successful checks and 20 failures (19 discovery shards
  plus their aggregate). Local repairs have not been pushed.
- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`. Never push to
  `origin` or `xgo-dev`. Batch the repairs before pushing; keep the PR draft.
- Develop on `codex/pr40-arm64-raw-20260926`. Upstream main `7cc8c0f` is
  already an ancestor. Inspect status, worktrees and processes first.
- `codex/pr40-shard25-replay-20260926` retains frozen `8134b3e` diagnostics.
  All its runners have finished; shard 29's final report was audited.
- A separate persistent verification worktree is frozen at `1c42a95`.
  Root `go test ./... -count=1 -timeout=20m` passed in 1,040 seconds; both
  nested CLI suites passed. Official five-architecture classification and
  strict benchmark passed: 184/184 files, no N/A, 25 target-seconds plus six
  seconds for driver build. Shards 4, 6, 10, 16 and 26 passed; all its runners
  finished. Preserve the reports. They predate the affine proof work below.
- `codex/pr40-evidence-20260927` separately records validated progress for
  that snapshot: 649 passed, 152 source N/A, two invalid-source skips, zero
  failures and 3,980 pending. The updater also reads back its own output.
  Evidence commit `16c4685`: five complete shards, 13,882 translations and
  43 target N/A; no partial shards. The complete/verified flags remain false.
  This is incomplete historical evidence, not current-development success;
  do not import it into a changed source snapshot as current proof.
- The affine verification worktree completed the full root suite at
  `198307f` in 835 seconds, `487fed6` in 753 seconds, `8eab02c` in
  824 seconds and `a740c74` in 822 seconds. Both nested CLI suites passed
  at `a740c74`; those runners finished. Its strict benchmark passed
  184/184 files, zero N/A, 60 target-seconds plus four seconds build.
  Official five-architecture classification also passed again at
  `487fed6`. Check the actual frozen revision and running processes before
  changing that worktree.
- Use Go 1.27.1 and LLVM 22 for external modules. Put the actual Go binary in
  PATH; setting only GOTOOLCHAIN can select a different child compiler.
  Required cross execution uses checksum-pinned QEMU 10.2.3.

## Latest committed repairs

- `cc307d5`: reject Go command/compiler/assembler version mismatches as
  infrastructure failures, never source N/A. The shard script pins its child
  Go binary to the recorded GOROOT and checks all three versions.
- `ba157c3`: pin discovery and aggregate CI jobs to Go 1.27.1.
- `f151495`: track every bounded ARM64 constant-pool pointer alias separately.
  MOV and 64-bit ADD/SUB immediate, shifted and extended-register forms
  propagate inclusive offset ranges. Killing an original does not kill its
  copies. Loads must fit at both range endpoints. Escapes, truncation,
  address-dependent flags, unknown indexes and changing-offset joins fail.
- `0f08122`: retry transient module-download failures as well as dependency
  build failures. At most three attempts share the original deadline. A
  synthetic local proxy establishes red/green recovery, persistent failure,
  no retry on 404, cancellation and workspace cleanup. Checksum, compiler,
  resource and toolchain failures are not retried.
- `596b147`: follow actual CFG edges when proving integer index bounds.
  An adjacent CMP must dominate B.cond; CBZ/CBNZ zero edges are also modeled.
  A W comparison does not bound an X register. Bypassed guards, changed flags,
  signed-negative possibilities and converging unconstrained edges fail.
- `c2567c6`: fold proven private x86 threaded TEXT helpers into one CFG. Keep
  GP/vector/flag/FP state across direct and indirect jumps. CLI proof rejects
  Go calls/address uses, linkname/other-file references, incomplete or escaping
  tables, and paths bypassing table-base initialization. Initializers and
  continuations stay in one LLVM module; wasm cannot use this mechanism.
  Also honor GLOBL RODATA instead of making every global constant, and retain
  file-local linkage for `<>` data. The former caused an actual runtime crash.
- `43a576e`: fix the whole VPBROADCASTB/W/D/Q family's overlapping X/Y/Z views
  and inactive masked memory reads. Its 108 source/width/mask cases compile
  on three OS targets and execute on Darwin/Rosetta and required Linux tests.
- `c9122cf` through `e9f131a`: add bounded modular-affine pool analysis,
  guarded CMN ranges, exact AND results, transient address cancellation,
  independent origin offsets, TBZ/TBNZ path constraints and pre-projection
  intersections. Preserve exactly one relocation origin: two address aliases
  added together must be rejected. Unknown effects/cycles and exhausted proof
  budgets remain failures. See `arm64_raw_pool_affine*.go` and
  `arm64_raw_pool_symbolic.go`; do not treat an ADR as an ordinary constant.
- `9acb690`: retain CMP/CMN/ADDS/SUBS provenance through known NZCV-preserving
  instructions; avoid speculative residual recursion on almost-full ranges.
- `487fed6`: prove single-entry counter loops, eliminate only proved-impossible
  edges, and validate the full scalar/pair pre/post-indexed load footprints.
  Positive unit countdowns retain an induction bound; a proved one-iteration
  loop does not invent further pointer updates. Seed proofs make a prior body
  execution through an enclosing loop opaque. Invariant-origin queries use
  separate cached proofs and reject changing recurrences. Predecessor order is
  deterministic. Overlap, overrun, wrong-step, zero-entry and side-entry tests
  fail closed. See `arm64_raw_pool_{loop,invariant,memory}.go`.
- `4ed0872`: trace flags across long integer-multiply and SIMD schedules, with
  an explicit preserving-family whitelist. A TST/unknown effect still stops
  the proof. The runtime oracle includes 48 instructions between SUBS and B.NE.
- `8eab02c`: multiple ADRs may reacquire the same pool at a branch join. Affine
  expressions retain a relocation count and reject a sum of two origins,
  including distinct ADRs. Resolve stable origins exposed after transient-index
  cancellation without collecting unrelated loop predicates. Join/overrun
  regressions and three-OS/native/Linux-QEMU runtime oracles passed.
- `3ed6437`, `c043b91`, `b41c426`: avoid cyclic residual speculation, retry
  inconclusive caches with a fresh proof budget, and try direct guards before
  residual decomposition. Compact tautologies/duplicate predicates and retain
  already-proved mask constants for later comparisons. The real unsigned
  parser's length proof fell from exhausting 16,384 steps to 334 steps.
- `c0652bb`: retain nonconstant AND intervals and prove disjoint-bit ORR/EOR
  relationships across immediate/register/all shifted-register forms. Numeric
  bit proofs must not interpret a relocated pool offset as physical address
  bits. Overlap, truncation and unrelated-mask regressions remain conservative.
- `22ffd2d`: certify unit-step ordered loops, both directions, all eight
  signed/unsigned strict/inclusive conditions and reversed operands. Prove
  entry ordering, invariant limits, flag provenance and no arithmetic wrap.
- `4fe7c89`: comparisons constrain values at their defining program point,
  including operands reused before the branch. Pending historical predicates
  cannot constrain replacement values. Intersect mask/guard bounds before
  applying negative displacements. Bypasses and newer flags still reject.
- `99fbba8`: infer carried address/count relationships from certified loop
  deltas and prove the whole invariant at external entries. Only this entry
  proof enables full affine arithmetic in the unguarded invariant walker;
  speculative residuals keep the cheap mode. Increasing/decreasing post-index
  runtime loops preserve the relationship instead of multiplying independent
  ranges. The focused Go 1.27/1.20, three-OS LLVM 22, Darwin-native,
  required Linux/QEMU and vet gates passed for every batch above.
- `a9000c0`: rewind ordinary 64-bit scalar/pair stack reloads to the exact
  dominating save only when every other queried GP value is preserved.
  Overlaps, possible aliasing stores, calls, SP changes, width mismatches,
  joins and unknown effects reject; direct nonoverlapping SP stores are safe.
  Rewinding uses the saved source value, never its later replacement.
  An enclosing-loop entry barrier may preserve only registers that the
  certified body never writes. Focused Go 1.27/1.20, three-OS LLVM 22,
  Darwin-native, required Linux/QEMU and vet tests passed.

The pointer/guard batches passed all focused pool tests on Go 1.27.1,
focused Go 1.20 compatibility, LLVM 22 ARM64 objects for Darwin/Linux/Windows,
Darwin native runtime and the required Linux/QEMU runtime counterparts.
The existing cross container uses Go 1.27.0 for these root runtime tests.
Both nested CLI test suites and vet passed again after the guarded-index
commit. Corpus unit tests and focused race tests passed for download retry.
Logs in the development worktree use `raw-pool-alias*`,
`raw-pool-guard*` and `discovery-download-retry*`.
The continuation batch additionally passed amd64 native-Go/LLVM runtime
oracles, required Linux amd64 and 386/QEMU execution, five x86 object targets,
Go 1.20 compatibility, both CLI suites and vet. Canonical KnoxDB's four
previously failing files passed all 12 three-OS object compilations. A separate
Linux scalar oracle passed 40 scenarios per Uint8/16/32/64 decoder (160 total),
covering all 16 selectors and mixed helper transitions. Diagnostic artifacts
are under `_out/knox-continuation-*`; these are not final shard/ledger evidence.
The affine batch passed all focused pool tests, Go 1.20 compatibility, vet,
three-OS LLVM 22 objects and Darwin/Linux-QEMU runtime oracles (80 input pairs).
TBZ/TBNZ guard tests cover all 64 bits and both edge directions. At `e9f131a`,
official classification, the ARM64 decoder-corpus gate and the strict benchmark
passed again: 184/184 files, zero N/A, 31 target-seconds plus two seconds build.
These classification and compilation gates do not establish every instruction's
runtime semantics. Keep the final full-suite and external-corpus gates separate.
The loop batches passed focused Go 1.27/1.20 tests, three-OS LLVM 22 objects,
Darwin native and required Linux/QEMU oracles. At `4ed0872`, strict benchmark
passed 184/184 files, zero N/A, 30 target-seconds plus seven seconds build.

## External evidence and remaining real failures

A passing fixture or one target replay is not a passing module or shard.
The latest three-OS SIMD replay (`_out/simd-stack.json`) passed translation
of `parseIntsNEON` and `parseUintsNEON`, but still fails the full files at
`formatIntsNEON` and
`parseIntsSVE2`. Do not upgrade its assembly ledger.

- **SIMD**: `github.com/sebishogun/simd@v1.21.1` and its mirror have 45/47
  applicable ARM64 files passing in earlier three-OS diagnostics. The two
  remaining bytes files need further proof work. Signed NEON now proves both
  vector loops run once and the tail counter stays in 1..7, including pair
  writeback. Unsigned NEON's two vector loops are also proved one-iteration,
  and the ascending signed-comparison tail now has a certified remaining
  interval of 1..19. The saved SP+8 index (instruction 181 to 323) now proves
  the carried pointer/count invariant is exactly 144 bytes relative to the
  pool. Both integer parsers translate. Formatting first fails at instruction
  59, a two-digit table read. Its UMULH/shift reciprocal division and MSUB
  remainder need a general exact unsigned-division proof, plus historical
  right-shift guard bounds; do not special-case the divisor 100.
  Inspect `_out/format-ints-neon-disasm.log` and `pool-format-neon.log`.
  Repeated ADRs retain relocation cardinality. The ignored function probe's
  scratch mask now has explicit parentheses; its earlier missing parentheses
  caused a diagnostic-only false return-escape failure.
  Signed SVE first fails at instruction 23, a vector-length-dependent address
  operation; it needs actual SVE length relationships, not a fixed-host-length
  assumption. Ignored `pool-function-*` probes and disassemblies retain these
  diagnostics; their temporary traced walker must be regenerated after edits.
  The apparent branch
  word `0x540be400` is numeric pool data, not invalid source. Implement in
  `arm64_raw_pool_*.go`; do not relax the proof merely to relocate the pool.
- **KnoxDB**: all three complete v0.2.9 candidates (`blockwatch.cc/knoxdb`,
  `github.com/blockwatch-cc/knoxdb`, `github.com/os2357/knx`) passed at `1c42a95`:
  each has 29 applicable files and 87 object translations. Shards 6/16/26
  passed. Empty Go declarations do not imply callable helpers: private
  entry proof is essential. Never apply signature widening to ordinary calls.
- **Native-layout/JIT**: GopherJRE, GoJIT and both case-distinct Sharkie module
  paths still fail. Their source observes exact code offsets or transfers
  registers/stack through generated machine code. Widening raw branch
  boundaries alone is insufficient. There is no authorized generic
  native-layout skip. A newer GoJIT version was downloaded diagnostically but
  has not been imported or verified as a replacement.
- Exact invalid-source skips (including both go-highway versions), historical
  gVisor/Skywire superseded skips and the pinned fiber/ai private AMX exception
  remain distinct from passes. Their manifests and executable proof tests are
  authoritative. Do not extend their scope without evidence/authorization.

The frozen `8134b3e` diagnostics are stored under
`_out/ci-repair-8134-shardN/shard-N.json`. Completed shards 1, 13, 20, 22,
23, 24, 25 and 29 have zero failures and individually audited reports. Shards
0, 6, 9, 14, 15, 16, 17, 26 and 28 retain the real failure classes above
(shard 16 also predates the dev9 invalid-source fix). Shard 4 completed
168 candidates = 123 passed + 35 source N/A + one invalid-source skip +
nine download TLS-timeout failures; its integrity audit passed, not its
coverage gate. Shard 10 completed 173 candidates with 133 passed, 29 N/A and
11 proxy TLS-timeout failures; its large spanneranalyzer p0/p6 candidates
passed. Shard 29 completed 151 candidates = 127 passed + 23 source N/A + one
invalid-source skip, with 2,052 object translations and three target N/A files.

A separate clean `cc307d5` shard 16 report has 132 passed + 30 N/A + one
go-highway invalid-source skip + one KnoxDB failure. It is not compatible
with the older frozen reports. Keep all of these as diagnostic evidence.

## Next actions and completion gates

1. Preserve the validated 4/6/10/16/26 evidence in its same-source tree.
   Shard 4 passed 168 candidates: 130 passed, 37 N/A, one invalid-source skip,
   4,152 translations and 25 target N/A. Never relabel network failures as N/A.
   Shard 10 passed 173 candidates: 142 passed, 31 N/A, 2,889 translations and
   six target N/A. Both retry runs and their ledger updates are complete.
2. Continue real semantic fixes above with red/green and runtime tests.
   The address-proof work does not yet resolve the two complete SIMD files.
3. Run full root, both nested CLIs, vet/build, official five-architecture
   coverage, ARM64/stdlib corpus, cross runtime and strict benchmark after
   the final implementation. Preserve evidence by exact source snapshot.
4. Freeze source/tools/ledger, run all 32 shards on identical provenance,
   then use the validated updater to refresh the assembly ledger. Even doc
   commits change provenance; do not combine old reports into the final run.
5. Batch-push the allowed fork, update PR body with validated scan/coverage
   funnels, and inspect current-head CI/review/coverage before making ready.

The development branch's assembly evidence remains stale and incomplete: 4,783 candidates,
4,624 pending, 126 passed, 32 N/A and one failed at its older revision. Do not
hand-edit it to match diagnostic counts. Reports/binaries belong in ignored
`_out/`; clean only owned generated files and candidate caches, never shared
module caches or unrelated containers. Old progress narratives remain in Git
history instead of accumulating in this checkpoint.
