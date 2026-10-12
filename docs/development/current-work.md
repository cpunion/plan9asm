# Current work: repair PR 40

Continue from `codex/pr40-external-corpus-split-20261011`, not older family
worktrees. Inspect status, worktrees and remotes. Push only to an allowed fork;
keep the contribution draft until current-head CI, review and coverage pass.
Let active frozen CI runs finish before publishing another batch.

## Scope and published checkpoint

External scanning, ledgers and compatibility batches moved to Draft
[llgo-compat/plan9asm#1](https://github.com/llgo-compat/plan9asm/pull/1).
Its initial ledger was byte-identical to `c8724add`; historical pending/failing
outcomes were not promoted. That repository owns subsequent inventory imports,
cursors, external failures and ledger updates. Do not restore them here.

The translator retains instruction families, shared source/tool observations,
independent regressions, official coverage and benchmarks. `gotoolprofile` is
public for the separate harness. Exhaustive runner owners remain root 0,
profile 1 and other packages 2, with mandatory disjoint, complete aggregation.

The published head is `432796c7`; [run 38193906662](https://github.com/xgo-dev/plan9asm/actions/runs/38193906662)
is still running. Its standard-library and benchmark jobs have already failed;
let the remaining jobs finish before the next batch push. The last completed
full CI evaluated `346cec2f`; [run 38187769673](https://github.com/xgo-dev/plan9asm/actions/runs/38187769673)
finished with 83 successful and 65 failed jobs, including aggregates. Its benchmark compiled
164/185 files, with 21 failures; wasm passed 19/19. Neither its timing nor the
five-architecture classification gate establishes complete compilation or
runtime coverage. Previous run `38182744838` remains failed.

## Verified local repairs after publication

- `9ac3013d`: the remaining numerical-source file proxy uses the shared escaped
  URL helper. A comma-containing fixture reproduces the old failure without
  permitting network fallback. Source mutation guards remain enforced.
- `18217144`: composed source proofs have a finite 20-minute lifetime, separate
  from unchanged five-minute command stages and capped by the test deadline.
  The two formerly timing-out Linux/amd64 race tests passed in 440.81 and
  545.69 seconds, including original-Go and LLVM execution at both optimization
  levels. This repairs lifetime composition; it is not a performance speedup.
- `9b60d37f`: the embed fixture compares actual SDK selection independently.
  Historical Go's unrecognized Unicode directive separator retains an empty
  selection; a separate Unicode pattern-whitespace case must select its file.
  Source/asset mutations remain failures. Go 1.24 and 1.27 checks pass without
  skipping cases or changing the production parser.
- `3fad6ecb`: machine-readable Go output retains separate bounded stdout and
  stderr through the existing owned-command observer. Dependency-download
  diagnostics no longer corrupt JSON. Production diagnostic rejection is
  unchanged. Go 1.20/current SDK source, ownership and vet checks pass.
- `fa5d2a0b`: all seven 386 scalar arithmetic/logical families consume the
  existing complete absolute-address source classifier for Go's L-width,
  four-byte relocation rows. Addresses remain addresses, not memory loads.
  No frame, caller, raw or mode64 permission is added. Family, stack, LOCK,
  TLS, Go 1.20 and three-API/two-target object tests pass. The pinned IA32 QEMU
  oracle checks 42 value/carry cases against original Go and all three LLVM
  APIs at O0/O2 with independent numerical references.

The Windows/386 standard-library rescan now reports zero unsupported opcodes,
operand forms and parse errors across 41 assembly files. This is classification,
not a complete standard-library compilation pass. The five-architecture
official classification gate also passes without baseline edits.

Historical Windows SEH now has independently pinned Go 1.23.12, 1.24.13,
1.25.14 and 1.26.8 producer schemas, alongside 1.27.1/1.27.2. Each schema
binds all eight source inputs, builtin indexes, symbol/Aux enums, flags and
listing grammar; generic archive/DATA admission remains unchanged. Actual SDK
object readers and complete source/listing mutation checks pass on all five
minor versions. The discovering fixed-local Windows LLVM object matrices pass
on 1.23.12 and 1.27.2. Evidence uses the `windows-seh-` prefix in the same
ignored report directory; no historical or native runtime result is implied.

The subsequent repair batch also passes focused coverage/ordinary tests on
Go 1.20 and 1.27, and the CLI environment/identity matrix with TMPDIR unset.
It retains Go's exact coverage trailer grammar, prepares selected cache roots
before loading, and supplies direct scratch fixtures the ordinary environment.
Original frame fixtures isolate the harness's Windows `byollvm` build tag;
explicit caller-supplied foreign flags remain rejected. C oracle process setup
now selects binary stdout/stderr on Windows instead of normalizing captured
bytes. Windows-target IR, host byte-protocol execution and actual Linux
process/header tests pass; Windows CRT execution still needs CI.

Evidence lives under ignored `_out/register-result-repair/`, including
`numerical-local-race-lifetime-green.log`, `embed-separator-go124-green.log`,
`go-json-stream-green.log`, `scalar-relocation-family-green.log`,
`scalar-relocation-linux-runtime.log` and
`scalar-relocation-official-gate.log`. Preserve their RED counterparts.
Receipts certify only their exact frozen source and tool snapshots.

## Remaining failure work

The main working branch now also contains the composed proof-factory leases,
actual Go cache-copy quoting/source registration, Windows oracle executable
paths, and normalized Windows source-PC origin joins. The source-PC family
passes Go 1.23 and 1.27, including the broader 485-second source/object suite.
The quoted-cache snapshot passed the complete Go 1.27 nested CLI in 1282 seconds;
its later producer-format repair has a separate Go 1.24 full CLI running.
Do not edit a frozen repair worktree until its process finishes.

The Linux race replay passed the complete profile guards (751.72 seconds) and
readonly deadline/cancellation matrix (525.26 seconds), but its SDK API reentry
root failed the old two-second timer. That report remains failed. The repair
isolates each real reentry case in a bounded, owned child, requires the callback
and completion receipt, and removes the leaked-worker timeout. All 18 original
and verification Err/Deadline/Done API combinations now pass on the host and
under Linux/amd64 race checking (920.553 seconds).
Windows test-only repairs additionally preserve exact killed exits, bind file
IDs before replacement, and avoid cross-volume relative-path assumptions.
Host Go 1.20/current identity tests pass; Windows-native execution still needs
CI. IA32 artifacts retain complete ELF/identity checks without requiring Unix
execute bits from Windows. Stream tests check the OS's live-rename refusal,
complete EOF, then actual replacement/reopen rejection after close.

Latest-run failures also exposed three distinct version boundaries: DATA
headers require actual-version-selected pre-1.27 versus current grammar;
Go 1.26.9 has the exact same eight pinned SEH inputs as 1.26.8; and ARM64's
five-class diagnostic arity predates its renamed AUTO class. Retain complete
schemas, not mixed optional fields. The DATA family passes Go 1.24/current,
and actual Go 1.26.9 SEH/schema mutation and CPU-profile objects pass. All Linux
root lanes now provision pinned QEMU: historical lanes execute the new scalar
relocation runtime too. The policy regression reproduces the old version-filter
failure. The diagnostic/policy multi-SDK suite still needs its final results.

Evidence is under `_out/register-result-repair/` and the
temporary repair worktrees' `_out/ci-proof-lifetimes/` / `_out/ci-quoted-cache/`.

1. Collect actual root and nested CLI failures separately. A successful main
   artifact does not establish nested CLI success. The repairs above still
   require current-head CI; preserve the old failed receipts.
2. Windows job `114614102112` has remaining raw-machine-code test bridges that
   assume SysV registers instead of the Windows calling convention, POSIX-only
   fault helpers, and source/path/ownership failures. Its raw log is retained
   as `ci-346-windows-job.log`. Fix shared transport contracts, not individual
   expected values or broad host skips. Audit remaining race artifacts too;
   the two repaired numerical proof lifetimes do not certify every race root.
3. The 21 standard-library failures need actual runtime/reflect entry, TLS,
   hidden-register, caller-storage or SDK source contracts. Maps' immediate
   shuffle result analysis is repaired; AES mutable-global storage remains
   separate. Dynamic reflect frames cannot be inferred from `func()`.
   ARM64 result-tail bytes are observable by assembly callers: do not clip
   them, invent frames or convert missing contracts into passes.
4. Rerun affected complete gates and both CLIs before batch publication.
   Required LLVM 22 tools fail if missing. Native Linux fault/AVX512 semantics
   are not certified by Rosetta diagnostics or successful object emission.

Earlier integrated repairs cover scalar ABIInternal results, immediate
shuffles, ARM64 PCALIGN tail padding and bytealg register transport, LLVM/QEMU
provisioning, Go 1.26 wasm selection, Windows host tools/proxy URLs/LF handling,
and oracle process isolation. Detailed narratives remain in Git history.

Use [instruction-family development](instructions.md) and
[validation](validation.md). Keep this handoff concise; do not grow it into a
second historical test ledger.
