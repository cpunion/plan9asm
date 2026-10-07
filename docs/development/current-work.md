# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). Locate persistent worktrees
with Git; never edit a live source, tool or ledger snapshot.

## Contribution and CI

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to `cpunion`.
Do not promote the upstream-connected branch until current-head fork CI and
exclusions review pass. Upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40)
stays Draft.

Published head `47ea649d` and completed run `36942816488` still fail the six
priority discovery shards (5, 29, 36, 38, 39, 42) and aggregate. Policy passed;
downstream skipped jobs did not pass. This local batch is not yet pushed.

## Continue here

Development branch:
`codex/pr40-x86-immediate-shuffle-grammar-20261007`.
Read its actual HEAD/status. Root alone writes ledger. An independent agent
is developing bounded multiple-SCC composition in a separate worktree.

Latest whole-family repair covers PSHUFW, PSHUFL/PSHUFD, PSHUFHW/PSHUFLW and
VPSHUFD/VPSHUFHW/VPSHUFLW:

- One complete Go-row grammar supplies selectors, native memory shapes and
  caller byte availability. Legacy preserves upper aliases; VEX/EVEX clears
  them. Partial input availability follows only selected bytes; unknown mask
  or merge bytes never acquire values or pointer permission.
- Full E4NF reads remain observable with dead outputs and zero masks; broadcast
  reads m32. Legacy m128 retains its original native SSE2 opcode/alignment.
  Virtual FP/SP/BP and unbound address-size roles remain Context; caller-region
  identity does not supply legacy alignment.
- Independent review found float/unresolved class confusion, mode32 constant
  classification and raw PSHUFW signed-rendering gaps. Actual Go-backed RED
  reproduced them. Every source consumer now excludes non-integer metadata;
  mode32 classification does not mutate the original operand.
- Initial caller/effect/runtime RED failed 55 nodes. Expanded diagnostics
  finish with 922 passes, no Fail/Skip: shuffle 24 roots/483 children and
  existing YMM/MMX 13 roots/402 children. Original Go covers all 4,096 named
  imm8 witnesses, 512 physical MMX byte cases, complete encoding rows and
  mode32 wrap/rejection controls. Three APIs compile five LLVM22 targets.
- Forty portable whole-byte programs, 24 full-read programs and eight legacy
  live/dead memory programs run without/with an IR O2 prepass. The legacy
  reference executes original Go opcode bytes, with canaries and flags.
  These are not backend O0/O2 axes or native AVX512/MMX-state certificates.
  Rosetta alignment observations remain diagnostic.
- RED/green logs are ignored diagnostic receipts, not current-head formal
  corpus certificates. Freeze a clean integration successor, run its fresh
  current/old formal scopes and official five-arch gate, then replay the exact
  original Weaviate candidate. Do not publish old reports into the ledger.

## Integrated ancestor and remaining source proofs

Frozen `7f684cc5` integrates the earlier ARM64 scratch/input and transitive
same-package call proofs, CPP companion/header consumer closure, full ordinary
Go-body/SDK-type frame observation, compact original CPP resource index, and
complete eight-name mask/vector B/W/D/Q caller and alias effects.

Its independently read current scopes pass 3,421 nodes (full CLI 1,804,
ARM64 1,264, CI policy 7, mask family 346); old framework scopes pass 1,610.
Race, complete source/tool/dependency/environment guards, vet/build, nested
CLI checks and genuine producer/process-group closure passed. Its official
five-architecture classification/fingerprint gate passed, not complete encoder
or execution coverage. CPP resource microbenchmark passed, not the full stdlib
benchmark.

Original Weaviate replay still fails ten AMD64 files with cyclic joins or
missing shuffle caller effects; companion CPP mismatch and mask/vector effect
errors no longer appear. Partial file/object success is not a passing version.
Retain failed reports and rederive every successor receipt from its own source.

The source-only cyclic audit authenticates a fresh exact Weaviate download and
retains seven function/SCC projections. The older failed report lacks per-file
hashes; a fresh download cannot invent them. Whole-function unique-backedge
proof is too narrow for these functions. Compose only independently proved
bounded SCCs first; dynamic indexed/cursor/sentinel loops still need their own
machine-width relational and pointer-region proof. Never enable unbounded
cycle hull merges, increase caps or discard original return/stack checks.

## Running corpus, ledger and inventory

Frozen `529019e1` owns `_out/priority-shard{5,29,36}-bitfield-replay`.
Shard 5 closed after 96 explicit candidates and remains partial; shard 29
closed after all 106. Both failed. Shard 36 is still running at this checkpoint.
Read actual owner/process handles and `closure.json`; the old status utility's
missing `owner.result.json` check is not producer closure evidence.
Obtain terminal producer and independent-reader results before promotion.
Historical reports cannot certify later implementation or documentation heads.

Committed assembly ledger remains three independently published Snappy passes,
6,064 pending, no complete shard, `complete=false, verified=false`.
Only the automatic publisher may update exact matching ledger records.

Inventory retains 215 continuous ranges from 2025-10-01 through
2026-09-22T22:43:08Z and 6,067 assembly-bearing latest-per-path candidates.
Newer standalone inventory is unsynchronized; cgo stays outside this repository.
Scan/coverage funnel tables belong in the PR body, not this checkpoint.

## Completion gates still open

Repair/replay original libraries: native SP/caller backing, imported/generated
headers, instrumentation and foreign-type closure, greater-than-64 profile
planning, native FFI/JIT entries and remaining dynamic loop effects.
Unknown errors are not source N/A; proposed exclusions still need review.

Latest strict stdlib benchmark `f8ff1d9f` failed: 184 attempted, 172 passed,
12 failed; wasm 19/19. Dynamic reflect/caller frames, runtime continuations/TLS,
ARM64 result padding and SDK body closure remain.

Full root/nested CLI tests, official gates, strict benchmark, all 64 discovery
shards, cross-runtime jobs, exclusions review and current-head CI are incomplete.
Clean only exact owned caches after all consumers close; preserve unowned
orphans/shared/live caches. LLVM22 only, no wasm blockaddress or tool Skip.
