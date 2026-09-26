# Current work: PR 40 CI and ecosystem coverage

Read [instruction development](instructions.md), [validation](validation.md),
and [report provenance](discovery-verification.md). Keep progress funnels in
the [draft PR](https://github.com/xgo-dev/plan9asm/pull/40), not this checkpoint.

## Contribution and frozen runners

- Push only to `cpunion:codex/expand-ecosystem-corpus-20260913`. The user
  requests complete fixes before pushing. Remote head is still `4cf5ade`;
  upstream main `7cc8c0f` is an ancestor. PR 40 is open and draft. Old CI
  `36087649450` has 74 successful jobs, 19 failed discovery shards and one
  failed aggregate. Do not describe local fixes as a current-head CI pass.
- Active development branch: `codex/pr40-arm64-raw-20260926`, implementation
  through `e002ae5`. Focused Go 1.20/1.27, CLI, vet, LLVM 22 objects and
  required runtime checks pass for these batches. Inspect Git before editing.
- Full-root runner: `codex/pr40-fp16-20260926`. Complete suites passed at
  `b6ccde3` (602 seconds) and `a5b0605` (591 seconds). It has advanced to
  `4885851` for signed-load corpus/official/benchmark checks; inspect Git and
  running processes before advancing. `_out/simd-signed-load.json` is a
  completed diagnostic replay, not a full discovery-candidate pass.
- `codex/pr40-ecosystem-fixes-20260925` completed shard 25/32 at `bc4c7cb`:
  137 selected, 111 passed, 26 source N/A, zero failed, 1,811 translations.
  Inspect `_out/shard25-proxy-repair.log` and
  `_out/ci-repair-223-shard25/shard-25.json`; the directory name is not its
  source revision. Preserve these completed reports when advancing the runner.
- This runner also preserves completed historical reports:
  `_out/ci-repair-02a1850/shard-0.json` (159 selected, 130 passed, 28 source
  N/A, one failed GopherJRE) and `_out/current-local-shards/shard-10.json`
  (173 selected, 142 passed, 31 source N/A, zero failures, source `74b02be`).
  Never combine reports with different source fingerprints.
- The committed assembly ledger (`1780b49`, merged at `4885851`) records the
  validated `bc4c7cb` shard-25 snapshot: 4,783 candidates, 4,646 pending,
  111 passed, 26 source N/A, zero failed; incomplete and unverified. It is
  stale against later source. Publish replacement evidence with the update
  script in an evidence checkout at the exact tested source, not by manually
  promoting statuses. Documentation changes also alter the source fingerprint.

## Verified implementation batches

- Earlier work covers raw FP16/BF16, SVE floating arithmetic/conversions,
  comparisons, predicates, INDEX/count, compact, copies and integer unary
  families, plus coherent Z/V/F register aliases. Use Git history and focused
  tests for details; do not restart completed families.
- `87457ea`: complete raw structured LD2/3/4 and ST2/3/4 memory grammar.
  `f292cb3`: all raw integer DOT forms, including indexed lanes and mixed
  signedness. SVE2.3 byte-to-halfword DOT retains encoding/object tests;
  available QEMU does not execute that newer ISA.
- `1b43a4e`, `7bc1db0`: complete XAR rotations and both SPLICE forms, including
  raw full-width shifts and wrapping vector lists absent from Go's named
  grammar. Native instructions and scalar oracles check six SVE lengths.
- `65c6d9f`, `e30558b`: ordinary/quad logical reductions and all six ternary
  bitwise operations share typed specs with named lowering. Independent
  encoding fields, aliases, predicate identities and runtime results pass.
- `1f2d6d6`, `a023a58`: all predicated/indexed integer MLA/MLS forms and
  vector INC/DEC patterns/multipliers. Runtime tests cover every indexed lane,
  accumulator aliases, wraparound, all predicate-count patterns and six VLs.
- `01688b7`: ordinary and non-faulting contiguous memory offsets scale by
  actual memory footprint, including widening loads, truncating stores and
  Q elements. TDD reproduced wrong addresses; 132 ordinary and 48 non-faulting
  functions now agree with native/scalar oracles, including null inactive
  pointers, signed extension, positive/negative offsets and memory guards.
- `0734bbe`: MAD/MSB join the shared MLA/MLS grammar, removing duplicated
  lowering while retaining distinct destructive-product operand semantics.
- `a5b0605`: all 13 extra-shift operations, 52 width formats and 1,212 raw
  immediate/reverse-vector cases. `db4e655` batches assembly and shares C
  oracle loops: identical 344 runtime cases dropped from 109 to 4.7 seconds.
- `e920d8a`: all 38 LD1SB/SH/SW scalar/vector address forms share ordinary
  memory lowering. Runtime oracles cover guarded low-address memory, null
  inactive pointers, signed/unsigned extended indices, aliases and six VLs.
- `e002ae5`: all eight wide ADD/SUB operations and 24 widths, with independent
  register fields and 72 native/scalar runtime functions in 1.8 seconds.

## CI infrastructure and tool evidence

- `bc4c7cb` pins checksum-verified QEMU 10.2.3 using
  `scripts/install-ci-qemu.sh`. QEMU 8.2 misexecutes indexed H-to-D DOT at
  non-power-of-two VLs and can overwrite predicate state at maximum VL.
  Do not work around that emulator bug in translation or skip the oracle.
- Required full cross-runtime matrix passed at `4885851` (405 seconds),
  including MAD/MSB, shifts and signed loads. Wide ADD/SUB has a focused
  pass and still needs the next combined gate. Use LLVM 22 and pinned QEMU.
- The strict five-arch standard-library benchmark at `4885851` passed all
  184 files in 24 target-seconds, within 300-second/900-second budgets.
- Discovery CI uses `GOPROXY=https://proxy.golang.org,https://goproxy.cn,direct`.
  Skywire's exact version remains on the fallback proxy, authenticated by
  Go's checksum database; its candidate now passed 11 translations. Never
  disable checksum verification or turn download failures into N/A.

## Remaining failures and next actions

- `simd@v1.21.1`: earlier amd64 replay passed all 73 files. At `4885851`,
  ARM64 passes 28/47 files on each of Darwin/Linux/Windows (84/141 total).
  Remaining first failures include unsigned LD1W widening/gather, predicated
  FADD immediate, UMULH and vector ADR. LLVM 22 disassembly confirms these
  raw opcodes; do not infer them from function names. Wide ADD/SUB needs a fresh replay.
  Unlabelled constant pools also remain. Removing the first instruction
  failure does not establish a whole-file/candidate pass.
- Constant-pool relocation requires a load-only address-use proof. Unknown
  effects, escaping addresses, indirect calls and reachable invalid words
  remain failures; do not guess data from comments or caller-saved registers.
- Knoxdb/forks require a real cross-TEXT ABI0 shared-frame/register model.
  GopherJRE/sharkie/gojit observe native code addresses/layout/markers.
  GopherJRE also hardcodes a JIT entry offset; loose label/frame checks would
  silently miscompile it. Go-highway dev9 has three absent amd64 RIP pools.
- gmgo/gmsm, puter and fiber/ai retain raw-word issues, including byte-swapped
  words and private Apple instructions. Inspect the latest exact report,
  not only its first old diagnostic. Do not invent missing source constants.
  `initLijing/gmsm@v0.15.6` has byte-swapped SM4 WORDs; LLVM 22 rejects the
  original `0x09c961ce` and decodes `0xce61c909` as SM4EKEY. An optional user
  question asks whether proven upstream defects may be isolated as
  `blocked_external`. No permission or policy change has been received;
  keep failures and do not silently rewrite the exact module or relax CI.
- Finish all current gates, continue failing families with red/green tests,
  and rerun all shards in one frozen source/tool/scan snapshot. Publish
  validated assembly evidence, then push only to the allowed fork. Keep the
  PR draft until tests, CI, review and coverage satisfy completion gates.
  Do not resume discovery while CI failures remain. Clean owned temporary
  packages/objects and containers; retain reports, not generated IR caches.
