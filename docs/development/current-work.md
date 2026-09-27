# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). Repair CI before starting
another index scan. Older detailed checkpoints remain in Git history.

## Contribution and frozen snapshots

- PR 40 is OPEN/DRAFT. Push only to the allowed fork,
  `cpunion:codex/expand-ecosystem-corpus-20260913`, never `origin` or
  `xgo-dev`. Upstream `main@7cc8c0f` is already an ancestor. Inspect
  remotes, worktrees, processes and live checks before continuing.
- Pushed head: `b93666f5`, CI run `36311131855`. At 11:45 UTC on
  2026-09-27 it had 67 successful jobs, three failures, eight running and
  14 queued. This is not complete green CI. Its exact-source local full root
  suite passed in 627.880 seconds, including all root subpackages.
- The user explicitly requires this CI run to finish once. Do not push,
  cancel or rerun it; accumulate verified local commits meanwhile.
- Develop on `codex/pr40-arm64-raw-20260926`. The persistent
  `codex/pr40-ci-hotfix-20260927` tree passed full root suites at `8a3cb6c9`
  (680 seconds) and `d9477276` (691 seconds), including all root subpackages.
  Logs are `_out/root-tests-compare-loops-8a3cb6c9.log` and
  `_out/root-tests-strides-d9477276.log`. Both CLI suites previously passed
  at `a880ec19`. Inspect revision/process before changing a frozen tree;
  the next full gate must include `75962bd3` and `f271511f`.
- `codex/pr40-ci-evidence-b93666f5` retains current CI artifacts separately
  in `_out/ci-b936-reports`. Evidence `cbb63c68` has ten complete reports:
  shards 1/2/3/4/6/7/8 pass; 5/9/11 fail. There are 1,220 passed versions,
  255 source N/A, two invalid-source skips, three failed and 3,303 pending;
  24,627 translations. Complete/verified remain false. Do not combine this
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
  runtime cases pass. SVE bytes still needs relational loop proof.
  `0x540be400` is numeric pool data, not invalid source.
  In `parseIntsSVE2`, CMP at instruction 98 now reaches B.NE at 121 with
  correct NZCV provenance. Loop 74..121 subtracts invariant RDVL from X20
  until X20 equals X7. X7 is length AND (VL-1). Entry length is 8..19.
  Register-CMP guards now prove length 16..19 at VL=16; X20-X7 equals 16.
  The complete pool-use proof succeeds at VL=16. At higher VL the first
  vector loop is proved unreachable, and loop 153..176 has an independently
  proved -16/-8 counter with step +8. Its carried pointer range is 12..132;
  reads 155/157 are bounded. The whole function still fails for VL>=32:
  inspect later scalar paths/loops and proof budgets. Do not claim a module
  pass. `_out/pool-mask-anchor-final-diagnostic.log` has this last probe.
  Ignored `pool-function-probe_test.go` and `parse-int-sve-disasm.log`
  retain diagnostics. Do not loosen a bound because the load is supported.
- **Native-layout/JIT**: GopherJRE, GoJIT and both case-distinct Sharkie
  module paths observe code offsets or exchange registers/stack with
  generated native code. Widening branch boundaries is insufficient.
  There is no authorized generic native-layout skip.
  Current CI shard 9 fails GoJIT on the `0xDEADBE00` marker. Its Go source
  searches function bytes and jumps past the marker; this is data/native
  layout, not an invalid executable opcode exception. A user question about
  a separately counted exact native-layout exception is pending; without
  explicit approval, preserve the failure and implement compatibility.
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
