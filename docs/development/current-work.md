# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md),
and [report provenance](discovery-verification.md). Do not resume discovery
while the existing CI failures remain.

## Contribution and source

- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`. The user
  requests complete fixes before a batch push. Remote head remains
  `4cf5ade`; PR 40 is open and draft. Run `gh pr view` for current checks.
  Last inspection: 74 successful checks, 19 failed discovery shards and a
  failed aggregate. No later local batch has been pushed.
- Upstream main `7cc8c0f` was fetched again and is already an ancestor.
  Never rewrite a frozen runner or push upstream.
- Develop on `codex/pr40-arm64-raw-20260926`, implementation through
  `e0d4b4d`. Check worktrees, status and running processes before editing.
  Use current Go 1.27 and LLVM 22; focused root compatibility uses Go 1.20.
- Frozen runners are `codex/pr40-fp16-20260926` and
  `codex/pr40-ecosystem-fixes-20260925`. Inspect each HEAD and its logs
  before advancing. Reports and generated tools live under ignored `_out/`.

## Completed local batches and gates

- Earlier commits cover complete raw SVE memory, arithmetic, conversions,
  predicates, DOT, wide add/sub, extra shifts, ADR, and aliasing families.
  Consult Git and focused tests; do not restart completed families.
- `acf334b`: all 79 ordinary unsigned load formats, including gather,
  widening, PN multi-vector and quad forms, share typed ordinary-memory
  lowering. Negative/reserved encodings and runtime address effects pass.
- `6a7ad09`: FDIV/FDIVR/FSCALE share raw/named typed grammar. Full root suite
  passed in 704 seconds; official five-arch classification gate passed.
- `f70af81`: complete raw saturating add/sub family. Runtime TDD also exposed
  incorrect signed-byte/halfword saturation with unsigned large immediates.
  Raw and named lowering now widen and clamp correctly.
- `630cdef`: complete ZFCPY/ZFDUP floating-immediate decoding, including
  P0–P15/M. Native/scalar/translated runtime checks cover all 256 constants
  and H/S/D at six SVE lengths, batched into fifteen functions.
- Full root at `630cdef` passed in 644 seconds:
  `_out/full-root-float-copy.log` in the full-root runner. Its official gate
  and strict five-arch benchmark passed: 184/184 files, 26 target-seconds.
  Official observed-form classification is not all-encoder semantic coverage.
- Full required cross-runtime gate at `f70af81` passed in 463 seconds:
  `_out/cross-runtime-saturating.log` in the ecosystem runner. Newer pool
  and floating-copy batches have focused cross-runtime passes and need the
  next combined gate. Use checksum-pinned QEMU 10.2.3, not QEMU 8.2.
- All recent batches passed focused Go 1.20/1.27 tests, CLI tests, vet and
  affected Darwin/Linux/Windows LLVM 22 object checks. Logs include
  `sve-saturating-add-sub-*`, `sve-float-copy-*`,
  `raw-pool-register-effects-*` and `raw-pool-contract-identity-*`.

## Constant-pool proof: preserve its safety boundary

- `eafa6a8`, `f7cb75f`, `e0d4b4d` establish typed instruction effects and
  explicit return contracts, not a blanket caller-saved assumption.
- Only R0–R17 can be terminal scratch kills. Explicit parameter/result
  contracts are required. Normal ABI results and every frame-result fallback
  register are excluded. Unknown signatures and custom ArgRegs stay strict.
- Proven independent loads, ordinary pair loads, shifts and FMOV can overwrite
  an old pointer. Known vector-only effects reuse validated SVE decoders.
  ADD/SUB Xn,Xn,#0 is an identity; W truncation, flags and nonzero offsets
  are deliberately not equivalent.
- Calls, indirect branches, escaping addresses, exclusive-monitor accesses,
  unknown effects and reachable invalid words still fail. Read-only memory
  effects use an explicit whitelist, not an LD mnemonic prefix.
- Remaining simd pools need proper offset/alias/range and additional SVE
  operand-effect proofs. Do not simply permit arbitrary pointer arithmetic,
  memory operands, unknown words or native-code address escapes.

## External replay and ledger evidence

- `simd@v1.21.1` ARM64 replay at `630cdef` passes 35/47 files on each of
  Darwin/Linux/Windows (105/141); see `_out/simd-float-copy.{json,log}` in
  the full-root runner. The latest development diagnostic reaches 40/47 on
  Darwin only, with seven remaining failures:
  `_out/simd-contract-pool-diagnostic.{json,log}` in the active tree.
  This dirty diagnostic is not a whole-candidate or ledger pass.
- Remaining simd first failures: jsonQuote NEON/SVE, atan2 NEON, exp SVE,
  their fast variants, and partitionFloat32SVE2. Constants resemble branches
  or other instructions because the pool proof still rejects their uses.
- Historical complete shard 17 at `41279bb`: 159 selected = 126 passed +
  32 source N/A + one failed simd. The validated snapshot is committed via
  `e54160f` (merged at `7b13d39`): 4,783 total, 4,624 pending, 126 passed,
  32 N/A, one failed, incomplete and unverified. It is stale against new code.
- Its report is retained in the ecosystem runner:
  `_out/ci-repair-41279bb-shard17/shard-17.json`.
  Historical shard 25 at `bc4c7cb` passed 111, with 26 source N/A and zero
  failed; retain `_out/ci-repair-223-shard25/shard-25.json` and the evidence
  branch `codex/pr40-shard25-evidence-preserved-20260926`.
  Never combine those different source snapshots.
- Update assembly evidence only with the validated updater at the exact
  tested source. Documentation changes also alter corpus fingerprints.
  Never hand-promote dirty, partial or stale reports to current passes.

## Remaining CI failures and next actions

- Finish the combined gates on a clean snapshot; continue simd pool proofs
  with red/green and runtime tests, then rerun the discovering shard.
- Knoxdb/forks need cross-TEXT ABI0 shared-frame/register semantics.
  GopherJRE/sharkie/gojit depend on native code layout/addresses; GopherJRE
  also hardcodes its JIT entry offset. Loose label/frame checks are unsafe.
  Go-highway dev9 retains three absent amd64 RIP pools.
- gmgo/gmsm, puter and fiber/ai have remaining raw-word issues, including
  private Apple instructions and genuine upstream defects.
  `initLijing/gmsm@v0.15.6` has byte-swapped SM4 WORDs: LLVM 22 rejects
  `0x09c961ce`, while `0xce61c909` decodes as SM4EKEY.
  No user authorization to isolate proven external defects as
  `blocked_external` has been received. Keep them failed; do not patch the
  exact module silently, invent constants or relax CI.
- Once fixes are complete, rerun all 32 shards with identical frozen
  source/tool/scan provenance, publish validated ledger evidence, then push
  only to the allowed fork. Keep the PR draft until current-head tests, CI,
  review and required coverage pass.
- Clean owned generated IR/candidate caches and containers after completion;
  retain reports and logs. Several completed IR directories were moved to
  the system trash; they are regenerable. Do not delete global module caches
  or unrelated containers. No new inventory scans were started.
