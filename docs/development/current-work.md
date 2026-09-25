# Current work: ecosystem coverage

This is a replaceable handoff, not an append-only log. Previous checkpoints
remain in Git history. Scan and coverage funnel tables belong in the
[draft PR body](https://github.com/xgo-dev/plan9asm/pull/40), not here.

## Active checkout

- Contribution branch: `codex/expand-ecosystem-corpus-20260913` for PR 40.
  Publish only to the allowed `cpunion` fork; never push or merge upstream.
- The contribution branch is frozen at `2025e3b` while shard reports run.
  Its recent batch covers ARM64 anonymous WORD data islands, Unicode macro
  names, complete x86 VEX packed moves/gathers, VPERMQ/VPERMPD and VPSIGN.
- A separate integration branch applies the PR's final tree to upstream
  `main` at `7cc8c0f` (including native ARM64 source lowering and CI retry
  changes). It is not the pushed PR head. Keep the frozen worktree unchanged
  until its reports finish, then verify a new single-provenance revision.
- A separate development worktree has TDD-tested tool changes: a temporary Go
  build cache shared across a coordinated `all` run, Git transport errors
  retained as retryable failures, and isolated noninteractive Git settings
  for public module fetches. It also lowers the complete ARM64 raw RNDR/RNDRRS
  family with its NZCV status, feature and LLVM 22 intrinsic. The discovering
  HopOS assembly file now translates and compiles 1/1 in an isolated replay.
  Its Go/LLVM object fixture passes locally; the required QEMU runtime
  conformance is added but awaits a current-head cross-runtime CI run.
  Do not cherry-pick these while frozen reports run: even tool changes
  invalidate their provenance.
- A separate evidence worktree stores audited assembly-ledger checkpoints;
  run `git worktree list` to locate both persistent worktrees.
- Keep the PR draft. The remote CI still describes an older pushed revision;
  inspect the live PR and jobs before claiming any current-head result.

## Validation checkpoint

- At code revision `2025e3b`, the full root suite (`GOMAXPROCS=2 go test
  -p=2 ./... -count=1 -timeout=45m`), both nested command suites, the
  official five-arch Go 1.27 assembler tables, the ARM64 x/arch gate, and
  standard-library cross-target corpus passed. The development worktree's
  tool changes passed focused and complete corpus-tool package tests, but
  still need the full matrix after integration.
- The validated scan ledger has 154 contiguous index ranges, 12,665,430
  entries, 811,704 unique module paths, 802,019 scanned exact versions,
  4,783 assembly-bearing versions and 13,631 retryable failures. History
  currently reaches 2026-03-18 and is incomplete. Source-frozen shard reports
  are under `_out/discovered-library-corpus-6db9a2f/`; trust their embedded
  `2025e3b` provenance, not the older directory name. The evidence worktree
  holds audited checkpoints; read its generated
  `_out/assembly-ledger-validated.json` for current passed, N/A, failed,
  pending and completed-shard counts. Sync partial reports periodically and
  each completed shard. Use the updater only in that worktree,
  never mutate the frozen contribution checkout during a run. A local ignored
  supervisor queues the remaining shards only when the frozen source is clean
  and free disk exceeds its safety threshold. Running shard-owned build caches
  can occupy tens of GiB; they are cleaned when each shard exits. The new
  shared-cache change will reduce this duplication on the required replay.
- The older pushed revision's `build` job failed the ARM64 x/arch coverage
  fingerprint baseline (756 supported forms expected, 766 observed), not its
  job timeout. Later local commits update that baseline; the current local
  `scripts/check-arm64-plan9-corpus.sh` gate passes with 770 supported, 76
  context and 262 unsupported forms. The current branch also raises root
  test/coverage timeouts from 30 to 45 minutes, still needing new CI
  confirmation. Old discovery shards had real failures, including raw x86
  PC-relative addresses and VEX byte sequences; they cannot be relabeled N/A.
- On the integration branch, the upstream native ARM64 source backend now
  accepts the parser's physical `RSP` spelling while still rejecting Go's
  pseudo `SP`. Its LLVM 22 object test fails when the required compiler is
  absent. Focused native tests, root build/vet, nested command tests, and the
  official Go and ARM64 x/arch coverage gates pass; the full root suite and
  all discovered-corpus shards still need current-head replay.
- Do not resume inventory writers while freezing the ledger for corpus
  verification; check the standalone host before making any process claim.

## Current failure classes

- `gmgo@v0.1.1` has ARM64 SM4 `WORD` values that LLVM 22 rejects even with
  SM4 enabled; byte-reversed values decode to the instructions in its source
  comments. Keep the source bytes failed rather than count them as support.
- `puter@v1.2.3` has ARM64 `WORD` values rejected by LLVM 22's decoder with
  NEON enabled. The current shard confirms the failure; do not classify it as
  successful translation or source N/A.
- `knoxdb@v0.2.9` tail-jumps from functions with a 56-byte ABI0 frame to
  zero-signature helper `TEXT`s, including register-indirect dispatch. Helpers
  share registers and write the original caller's `value+48(FP)` result. This
  needs a cross-`TEXT` shared-frame model; merely allowing that FP offset
  would compile wrong behavior. The four currently rejected files are honest
  failures; other passing files in this family still need runtime validation.
- `GopherJRE` uses a raw x86 RIP-relative `LEAQ` to capture a post-indirect-
  jump continuation in a JIT bridge. Go's own objdump confirms the target.
  Moving the raw displacement into LLVM unchanged would be unsafe; map it to
  a proven label and preserve the dynamic control-flow semantics.
- `gojit` embeds `LONG $0xDEADBE00`/ARM64 `WORD` markers after indirect JIT
  jumps, then inspects function bytes to recover a continuation address.
  Plain raw-instruction decoding and layout-changing LLVM lowering cannot
  claim this passed.
- `outfix` declares `GLOBL` character tables but emits `BYTE`s after a
  `RET`, not `DATA` initializers. Go's objdump places those bytes in `.text`;
  they are not a missing opcode family and must not be silently skipped.
- `containerfs` exceeded the target's 60-minute deadline during a Go package
  build. This is infrastructure failure to retry, not instruction coverage.
- `sharkie` uses a raw x86 short jump over a call stub, then manufactures a
  return address with a layout-dependent `LEAQ`/`ADDQ`. Preserve the raw
  cross-directive target and runtime layout before calling it supported.
- `gvisor-module-go126` uses an ARM64 exception-vector tail jump into a
  zero-argument Go-declared helper that reads the vector number from its
  caller's frame. The development branch now infers this borrowed ABI0 frame
  only when every local tail entry writes all required slots immediately
  before branching. A native ARM64 runtime fixture passes, and the real
  `entry_arm64.s` translates and compiles 1/1 with LLVM 22 after also proving
  the complete logical-operation width family crossed by its raw ADR. The
  frozen gVisor-fork reports remain failed until full current-head replay.
- `HopOS/metal/v2` exposed raw RNDR's status flags. The development worktree
  has a focused complete-family fix and an isolated 1/1 LLVM 22 object replay;
  the original frozen report remains failed until a new-provenance rerun.
- `github.laiyagushi.com/sebishogun/simd@v1.21.1` contains Clang 22 output
  copied verbatim into Go assembly as large `QUAD/LONG/BYTE` sequences. Its
  own generator preserves internal PC-relative branches by moving each whole
  function body as one byte blob; splitting and re-lowering instructions can
  invalidate those offsets. The discovering module has 64 failing files out
  of 73, including internal PC-relative references and VEX/EVEX decode gaps.
- `go-highway@v0.0.0-dev9` has distinct unhandled ARM64 families: AdvSIMD
  BF16 dot and FP16 FMA, scalar conversion/vector extract, and SME tile/matrix
  raw encodings. Its frozen report has 16 failing files out of 31; handle
  complete instruction families and formats, not one raw word at a time.
  This needs a sound whole-body layout strategy, not a permissive opcode skip.

## Next actions

1. Finish and audit all 32 shards. Each runner removes its candidate workspaces
   and shard-owned build cache on exit. Monitor disk space before increasing
   concurrency; keep reports and source frozen. Sync their validated partial
   outcomes into the separate evidence worktree at each checkpoint.
2. Diagnose new failures against Go assembler/compiler behavior. Build red
   tests and runtime oracles for full instruction or ABI families. Do not
   paper over shared-frame tail jumps or invalid raw encodings.
3. Once one source revision is ready, integrate verified development commits,
   rerun all affected gates and all 32 shards with one frozen provenance, and
   keep ledger statuses aligned with those reports. The shared-cache tool
   change itself requires a fresh replay.
4. Push a substantial batch only to the fork, update the PR body from
   validated counts, then inspect current-head CI, review and coverage.
   Ready-for-review requires all completion gates.

See [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md) for commands and invariants.
