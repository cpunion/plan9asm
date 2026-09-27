# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). Repair CI before starting
another index scan. Older detailed checkpoints remain in Git history.

## Contribution and frozen snapshots

- PR 40 is OPEN/DRAFT. Push only to the allowed fork,
  `cpunion:codex/expand-ecosystem-corpus-20260913`, never `origin` or
  `xgo-dev`. Upstream `main@7cc8c0f` is already an ancestor. Inspect
  remotes, worktrees, processes and live checks before continuing.
- Pushed head: `b93666f5`, CI run `36311131855`. The latest observed
  checkpoint on 2026-09-27 has 69 successful jobs, six failures, eight
  running and nine queued. This is not complete green CI. Its exact-source local full root
  suite passed in 627.880 seconds, including all root subpackages.
- The user explicitly requires this CI run to finish once. Do not push,
  cancel or rerun it; accumulate verified local commits meanwhile.
- Develop on `codex/pr40-arm64-raw-20260926`. The persistent
  `codex/pr40-ci-hotfix-20260927` tree passed full root suites at `8a3cb6c9`
  (680 seconds) and `d9477276` (691 seconds), including all root subpackages.
  Logs are `_out/root-tests-compare-loops-8a3cb6c9.log` and
  `_out/root-tests-strides-d9477276.log`. Both CLI suites previously passed
  at `9dfb793b` (1.004/2.473 seconds). The full root suite at that snapshot
  also passed in 680.377 seconds, with all root subpackages; its log is
  `_out/root-tests-mask-anchor-9dfb793b.log`. Inspect
  revision/process before changing a frozen tree. The `95ab1bc6`
  guard-intersection fix passed the full root suite at `03182350` in
  712.316 seconds, including all root subpackages. Its log is
  `_out/root-tests-guard-03182350.log`. The sequential-loop checkpoint
  `5805df44` also passed the full root suite (745.763 seconds) and both CLI
  suites (0.834/2.546 seconds). Logs use `_out/root-tests-sequential-5805df44.log`
  and `_out/sequential-cli-*`.
- `codex/pr40-ci-evidence-b93666f5` retains current CI artifacts separately
  in `_out/ci-b936-reports`. Evidence `e6e3fafe` has fifteen complete reports:
  shards 1/2/3/4/6/7/8/10/13 pass; 0/5/9/11/14/15 fail. There are 1,817 passed
  versions, 385 source N/A, three invalid-source skips, six failed and
  2,572 pending; 37,054 translations. Complete/verified remain false. Do not combine this
  source with earlier reports or promote it to the newer development source.
- `codex/pr40-ci-evidence-0d96d26` preserves 12 validated previous-CI
  reports: four complete and eight partial. Evidence `b434d36e` has 1,046
  passed, 219 source N/A, two invalid-source skips, five failed and 3,511
  pending; 20,510 translations. Complete and verified remain false.
- Historical evidence `16c4685` at source `1c42a95` preserves five passing
  shards (4/6/10/16/26), 649 passed candidates and 13,882 translations.
  Older `8134b3e` diagnostics and persistent verification trees are not
  disposable. Never import these reports as changed-source proof.
- Use actual Go 1.27.1 in PATH and LLVM 22. Required cross execution uses
  pinned QEMU 10.2.3. The existing Linux cross container uses Go 1.27.0 for
  root runtime oracles; external corpus jobs require Go 1.27.1.

## Pushed CI repairs

All 11 failures preceding `0d96d26` passed on that source: ten standard-
library lanes and Windows. Subsequent external failures remain separate.

- Unresolved ARM64 displacement macros require header context; translation
  rejects unresolved values instead of silently substituting zero.
- Proved private runtime-sized syscall frames have actual dynamic storage,
  overflow checks and relocation of live stack aliases. Calls, escapes,
  unknown effects, numeric observations and multiple extensions reject.
- Windows fixtures handle executable suffixes, CRLF and native fake Go
  binaries. Its preceding root suite passed with 89.6% statement coverage;
  both CLI suites and Codecov patch coverage passed on that earlier source.
- `a7e98cc` fixes single-character stack annotations such as `n-8(SP)`.
  Both failing files in each exact gomonkey candidate now compile:
  `gopkg.in/agiledragon/gomonkey.v2@v2.14.3` and
  `github.com/johnnyting/gomonkey/v2@v2.14.4`. These target diagnostics
  are not whole-module/shard passes. Logs use `_out/gomonkey-*`.
- Typed SVE pool analysis models scalable aliases, scalar counts and whole
  Z/P loads across all 16 architectural vector lengths. GP-write and NZCV
  effects remain separate; unknown instructions/mode changes fail.
- Previous toolchain-matching, transient-download retry, Windows atomic-
  report publication, FFR-oracle and Go 1.20 metadata-bootstrap repairs
  passed affected CI jobs. Infrastructure errors never become source N/A.

## New verified local batch

- `39371906`: contiguous scalar-base SVE pool-load footprints, including
  signed/unsigned widening, immediate/register offsets and actual memory
  width. Reserved modes, short pools, stores, first-fault and gather forms
  do not acquire this proof. All 16 scalar arrangements, signed offsets,
  three-OS objects and 1,024 Linux/QEMU vector outputs pass.
- `5faa5234`: preserve NZCV through typed immediate/indexed/GP broadcasts.
- `15a96c14`: indexed `ZDUP` B/H/S/D/Q indices within the first 512 bits,
  matching Go and LLVM 22. Out-of-current-VL results are zero, not LLVM
  poison. Tests cover 126,976 encoding-field combinations, 124 accepted
  indexed forms and rejected boundaries. All 248 runtime cases execute at
  every VL against independent native machine code and a scalar oracle;
  three-OS objects and Go 1.20 pass.
- `4e00663c`: complete predicated CPY/MOV flag and GP-read effects using
  its typed decoder. GP/SP copies preserve NZCV but copying a pool address
  into a vector still rejects. Full family formats, read-effect negatives,
  three-OS objects, Go 1.20 and all-VL Linux/QEMU execution pass.
- `ab4c0ff6`: AND/BIC retain affine relationships only when discarded bits
  are proved constant. Register/immediate and all shifted-register modes
  are covered; truncating W forms and unknown/relocated bits never invent
  an X relationship. The old-source overlay fails the new pool fixture.
  Modular high-bit tests, three-OS objects, Darwin execution, required
  Linux/QEMU, all pool regressions, Go 1.20, vet and the official five-arch
  gate pass. Strict benchmark: 184/184 files, zero N/A, 25 target-seconds
  plus one build-second.
- `86a21fe6`: independently prove a numeric CMP/CMN register limit at its
  defining comparison, including shifts/extensions and unsigned carry/zero
  conditions. Later writes and relocated offsets cannot supply the limit.
  Full pool regression, Go 1.20, three-OS objects, Darwin/Linux runtime,
  official coverage and benchmark (184/184, 23 target-seconds) pass.
- `6e8ff7d3`: rewind the actual compared expression through a single-entry
  straight-line body, then exclude B.NE only if every external entry gives
  equality on the first iteration. This is not a general induction proof.
  Wrong steps, underflow, side entries, enclosing iterations, changed flags
  and unknown effects remain barriers. Full pool tests, three-OS objects,
  Darwin/Linux execution, Go 1.20, vet and benchmark (184/184, 24 seconds)
  pass. Retained pre-fix overlays fail the executable fixture.
- `23cdeedd`: exclude unsigned compare edges only when independently proved
  impossible. No executable control-flow rewrite or invented branch fact.
- `d9477276`: exact AND interval extrema, NEG/NEGS affine aliases and
  power-of-two counter strides with an independent residue proof. Tests
  include both counter directions, flag/CBNZ latches, mask shifts, stack
  overwrites and exhausted budgets. The frozen full suite passed in 691s.
- `75962bd3`: integral carried-address/counter stride ratios; unknown or
  non-integral recurrences reject. Full pool tests passed in 60 seconds.
- `f271511f`: one typed AND/ANDS/BIC/BICS constant-mask grammar shared by
  residue and difference proofs, including shifted masks and commuted AND.
  Preserve `n-(n&mask)` as `n&^mask`, round independently proved residues,
  center loop invariants without clamping unsigned wraparound, and discard
  induction metadata when its backedge is excluded. Full pool tests passed
  in 61 seconds. Negative address-observation/relocation tests also pass.
  Both latest batches have retained red logs, Go 1.20, three-OS LLVM 22
  objects, Darwin execution, Linux/QEMU, vet and benchmark evidence. The
  latest benchmark compiled 184/184 files, zero N/A, 24 target-seconds plus
  two build-seconds; the official five-arch gate passes.
- `95ab1bc6`: intersect masked-difference guards before applying a negative
  displacement, including a nonzero difference followed by subtracting one.
  Unit and three-OS runtime-fixture red logs reproduce the old failure;
  all pool tests (60 seconds), Go 1.20, Darwin/Linux execution, vet and the
  official five-arch gate pass. Benchmark: 184/184 files, zero N/A, 24 target-
  seconds plus two build-seconds.
  Diagnostics use `_out/pool-mask-nonzero-*`.
- `7c07c82b`: retain unchanged expressions and independently invariant
  constraints across a proved earlier loop. Changing and inactive predicates
  are discarded; counter entry cannot use its own or an enclosing latch.
  Red unit/runtime overlays reproduce the original failure. Full pool tests
  (61.387 seconds), Go 1.20, three-OS objects, Darwin/Linux execution, vet,
  official coverage and benchmark (184/184 files, zero N/A, 24 target-seconds
  plus two build-seconds) pass. Logs use `_out/pool-sequential-*`; the old
  runtime overlay is in the hotfix tree at `03182350` and becomes stale if
  that tree advances.
- `0186a071`: typed count/address families preserve NZCV while retaining
  their distinct GP effects. Vector counts have no GP input/output. Scalar
  count grammar tests cover all patterns/multipliers; the runtime fixture
  checks 132 cases at all 16 VLs. Three-OS objects, Linux/QEMU, full pool
  regression (67.713 seconds), Go 1.20, vet and official coverage pass.
  Benchmark: 184/184 files, zero N/A, 23 target-seconds plus two build-seconds.
  Logs use `_out/pool-scalable-flags-*`.
- Disjoint ORR/EOR now reuse the independent residue proof when intervals
  lose low alignment bits. Copies, masks, commuted operands, shifts and
  modular arithmetic have positive/negative tests; overlapping or unknown
  bits cannot invent a sum relationship. Red unit/runtime logs reproduce
  the original failure. Full pool tests (61.903 seconds), Go 1.20, three-OS
  objects, Darwin/Linux execution, vet, official coverage and benchmark
  (184/184, zero N/A, 23 target-seconds plus two build-seconds) pass.
  Logs use `_out/pool-aligned-*`.

Logs use ignored `_out/pool-sve-contiguous-*`, `sve-dup-indexed-*`,
`pool-sve-copy-*` and `pool-masked-logical-*`. The masked before-fix
overlays reference the hotfix tree and become stale when it advances. The red
logs retain the actual failures; reconstruct old implementations from their
Git revisions before repeating a red replay. New compare/loop diagnostics use
`_out/pool-constant-compare-*` and `_out/pool-relational-loop-*`.
Later logs use `_out/pool-stride-*`, `_out/pool-carried-stride-*`,
`_out/pool-mask-difference-*` and `_out/pool-mask-anchor-*`.

## Remaining external failures

- **SIMD** `github.com/sebishogun/simd@v1.21.1` and its mirror: full NEON
  bytes assembly compiles on three OS targets; 130 parsing and 65 formatting
  runtime cases pass. The entire SVE bytes file now also compiles on all
  three OS targets: `_out/simd-aligned.json` records 3/3 files, zero N/A.
  Its actual SVE2 parse/format functions execute successfully at all 16 VLs,
  with 130 parsing and 65 formatting scenarios per VL (3,120 total).
  `_out/simd-sve-oracle-run.log` retains that Linux/QEMU diagnostic. The
  shared C harness still prints its original NEON label, but the IR wrapper
  explicitly calls the SVE2 symbols; its final line identifies the SVE2 run.
  This is not a whole-module or shard pass; rerun complete shard 15.
  `0x540be400` is numeric pool data, not invalid source.
  In `parseIntsSVE2`, CMP at instruction 98 now reaches B.NE at 121 with
  correct NZCV provenance. Loop 74..121 subtracts invariant RDVL from X20
  until X20 equals X7. X7 is length AND (VL-1). Entry length is 8..19.
  Register-CMP guards now prove length 16..19 at VL=16; X20-X7 equals 16.
  The complete `parseIntsSVE2` pool-use proof now succeeds at tested VLs
  16/32/48/64/128/256. At higher VL the first vector loop is unreachable;
  loop 153..176 has a proved -16/-8 counter with step +8, and the later
  scalar loop has count 1..7. See `_out/pool-sequential-relational-diagnostic.log`.
  In `parseUintsSVE2`, NZCV now reaches its first loop latch through DECB;
  the disjoint low-bit ORR then preserves the later loop relation. Its pool
  proof passes all tested VLs, and whole-file translation checks all 16.
  Earlier failed diagnostics remain in `_out/simd-sequential.log`; the
  final detail is `_out/pool-aligned-unsigned-diagnostic.log`.
  Ignored `pool-function-probe_test.go` and `parse-int-sve-disasm.log`
  retain diagnostics. Do not loosen a bound because the load is supported.
- **Native-layout/JIT**: GopherJRE, GoJIT and both case-distinct Sharkie
  module paths observe code offsets or exchange registers/stack with
  generated native code. Widening branch boundaries is insufficient.
  There is no authorized generic native-layout skip.
  Current CI shard 0 fails GopherJRE's raw RIP-relative continuation address
  outside its directive group. Its JIT also expects DI/SI register and stack
  contracts; accepting the byte displacement alone is not a runtime fix.
  Current CI shard 9 fails GoJIT on the `0xDEADBE00` marker. Its Go source
  searches function bytes and jumps past the marker; this is data/native
  layout, not an invalid executable opcode exception. A user question about
  a separately counted exact native-layout exception is pending; without
  explicit approval, preserve the failure and implement compatibility.
  Current CI shard 14 also fails the mixed-case Sharkie candidate on a raw
  short JMP into ordinary assembly; 4/5 Windows/amd64 files compile. Its
  custom stack/return-PC ABI still needs a real compatibility mechanism.
- **ContainerFS**, both case-distinct paths at the 2019 version: current CI
  shards 5 and 11 fail after three retries, with 503/429 responses mixed
  into obsolete gVisor dependency 404s. Local HTTP probes also alternate
  503/404. Do not classify the network errors as source N/A. Proxy latest
  resolves to the same version; no newer replacement is verified.
- All three KnoxDB v0.2.9 candidates passed on earlier snapshots. Preserve
  private-helper reachability/ABI proof; never widen ordinary public calls.
- Pinned invalid-machine-code, superseded-mirror and private-AMX exceptions
  remain separate from passes. Their manifests and executable witnesses
  are authoritative; do not broaden them without evidence and authority.

## Next actions

1. Inspect current-head CI first; collect failed logs and complete/partial
   shard reports. Audit reports in a same-source worktree.
2. Verify the latest committed batch on a frozen source; run the full root
   suite, both nested CLI suites and remaining final gates.
3. Continue SIMD relational proof with red/green and runtime tests. Retain
   negatives for changed limits, wraparound, bypasses and unknown effects.
   Function diagnostics never upgrade a whole module's ledger.
4. Refresh derived assembly provenance with the validated updater after the
   final source/documentation checkpoint. Freeze source/tools/scan ledger
   for all 32 corpus shards; never reuse old reports as new-head passes.
5. Batch-push the allowed fork, update the PR body funnel and inspect CI,
   review and patch coverage. Keep Draft until every completion gate passes.

The inventory still contains 4,783 assembly-bearing versions; current-source
assembly evidence is pending until compatible reports exist. The remote
inventory was not refreshed during CI repair. Scan records, assembly reports
and separately stored cgo inventory are distinct. Reports/binaries stay
ignored; remove only owned temporary downloads and objects, never shared
module caches, unrelated containers or frozen evidence.
