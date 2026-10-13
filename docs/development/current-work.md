# Current work: repair PR 40

Continue from `codex/pr40-external-corpus-split-20261011`, not older family
worktrees. Inspect status, worktrees and remotes. Push only to an allowed fork;
keep the contribution draft until current-head CI, review and coverage pass.
Let active frozen CI runs finish before publishing another batch.

## Scope and completed CI checkpoint

External scanning, ledgers and compatibility batches moved to Draft
[llgo-compat/plan9asm#1](https://github.com/llgo-compat/plan9asm/pull/1).
That repository owns inventory imports, cursors, external failures and ledger
updates. Do not restore those workloads or promote historical outcomes here.
The translator retains instruction families, shared source/tool observations,
independent regressions, official coverage and benchmarks.

Published `24851df2` completed [run 38236001451](https://github.com/xgo-dev/plan9asm/actions/runs/38236001451)
with 143 successful and 15 failed jobs, down from 29 failures and one cancellation.
Ordinary module and coverage aggregates passed, as did the Windows `plan9asm`
CLI. Race root 0 failed one reentry test; its aggregate correctly failed too.
Benchmark compilation remained 164/185:
386 29/33, amd64 49/56, arm 27/30, arm64 40/47 and wasm 19/19.
These are that frozen run's results, not certification of the following repairs.

## Latest local repair batch

Windows ownership and dependency scope now have native Go 1.20/1.27 evidence.
The test-only [38332958829](https://github.com/cpunion/plan9asm/actions/runs/38332958829)
reproduced the unsupported process-tree boundary. Suspended leaders enter a
private non-breakaway Job before resume; actual thread ownership is rechecked.
Accounting zero did not guarantee signaled process handles, reproduced by
[38345922514](https://github.com/cpunion/plan9asm/actions/runs/38345922514).
`9524c3d6` waits for the Job completion-port exit notification instead.
[38346249492](https://github.com/cpunion/plan9asm/actions/runs/38346249492)
passed both SDKs, including 25 repeated real-tree regressions each.

Inherited `-tags=byollvm` has an exact Windows-only explicit-flag handoff;
undeclared/conflicting/hidden flags still fail before execution. Complete native
runner packages passed [38346896685](https://github.com/cpunion/plan9asm/actions/runs/38346896685).
At frozen `582934e7`, complete local runner/ownership suites passed in
640.906/1.987 seconds. Actual Windows module dependencies subsequently exposed
mixed native separators, Node CRLF and the Unix-only DLL loader probe.
`f69465e7`, `7b222c49` and `4e3f8dee` retain exact path/packet identity while
observing the Windows imported function's real owning DLL. Full native runner
and dependency tests passed both SDKs in
[38348037017](https://github.com/cpunion/plan9asm/actions/runs/38348037017).
Local dependency/environment regressions passed at `7b222c49` in 263.070 seconds;
that receipt predates the new Windows/Unix loader source split.

Actual Windows complete-module partition/coverage and sticky-failure tests now
pass both SDKs in [38354976736](https://github.com/cpunion/plan9asm/actions/runs/38354976736).
The first fixture incorrectly demanded full `passed` status from a partial root
owner; its correction requires `completed_owned`, retaining `passed` only for
the aggregate. Native complete runner suites took 144.827/168.959 seconds
(Go 1.27/1.20), followed by 25 real process-tree repetitions each. Missing root
parts still fail; a real root failure does not suppress later packages or publish
uploadable coverage. Tests are integrated as `ab72f4c0`/`df29b71e`.
Local Unix dependency regressions passed at `25173f37` in 184.667 seconds.
Formal Windows workflow partitioning is integrated in `3a48e00d`. Complete
native [38380889962](https://github.com/cpunion/plan9asm/actions/runs/38380889962)
finished with 17 successful and 10 failed jobs. Both complete CLIs and their
coverage aggregates passed; all eight main-root parts and the profile owner
failed, so the main-module aggregate correctly failed. Root failures expose Windows lazy file identity,
ordinary x86 source-position separators, a fixture's missing host executable
suffix and the fixed-local Windows GNU target join. Preserve its raw failures
under `windows-complete-38380889962`; they are not coverage passes. Keep the
fork diagnostic workflow out of the contribution branch.
`4838d597` captures Windows file identity from an open handle, normalizes only
actual object-source separators and adds the Windows GNU fixed-local join.
Its focused Go 1.27 source regressions passed in 68.195 seconds; Go 1.20 portable
checks and the Windows helper cross-build passed. Native confirmation is pending.
The following test-oracle batch fixes host executable suffixes, SysV byte-code
calls, logical LLVM aggregate transport, Bash paths and bounded wasm startup
environments. Three masked-memory families reuse native Windows/Unix guard pages.
Local portable/wasm/source regressions passed in 46.326 seconds and the Linux
container's three guard-page families passed in 13.772 seconds. The broader
container runtime run remains failed: original MADD/PACK/timestamp and scalar
comparison observations disagree on AF under emulation. Preserve the failed
receipt and require native x86 confirmation; do not weaken flag checks.
Native [38386789839](https://github.com/cpunion/plan9asm/actions/runs/38386789839)
now passes all fourteen repaired x86 runtime roots in 11.512 seconds, including
those four flag oracles. Windows ownership/sourcepath pass both Go versions;
the complete-module jobs remain queued. The next batch replaces three Unix-only
INSERT/PACK/UNPACK fault drivers with direct owned processes and shared guard
pages. It adds positive full-extent reads, retains NULL/partial-read faults,
canary/setup rejection and checked zero core limits. Linux container regression
passed in 29.438 seconds; Go 1.20 header/policy checks passed in 0.883 seconds.
These new fault drivers still need native Windows execution.
`798ad021`, `339523d3` and `c4ff4f3b` now cover all six formerly Unix-only
INSERT/PACK/UNPACK/SHIFT/MMX/scalar-conversion fault families. Direct owned
children preserve exact entry/completion receipts, real fault status, canaries,
both MMX guard boundaries and the original SSE precision comparison. Ordinary
exits, joined cleanup errors and cancellation do not establish a native fault.
Go's Windows #UD diagnostic also requires the tested original assembly frame;
the modern sigpanic format has independent RED/GREEN controls. The expanded
Linux container suite passed in 57.314 seconds and scalar precision/control
follow-up passed in 14.167 seconds. Native Linux precision trapping is a separate
required axis, not certified by container emulation. Go 1.20 header/diagnostic
controls passed in 0.848 seconds. Root vet and the Go 1.27 five-architecture
official gate passed at `c4ff4f3b`; classification does not certify execution.
These three commits are staged locally on the fork probe branch but not pushed.
Run 38386789839 finished with 19 successful and 9 failed jobs. Both complete
CLI execution/coverage aggregates and the main
profile/rest owners pass. All eight main-root parts completed their 5,840-root
inventory: 5,541 passed, 36 failed, 263 host-inapplicable skips and zero incomplete.
The skips are not runtime passes. Preserve all eight raw partitions under
`windows-complete-38386789839`. Failures are seven fault-oracle roots repaired by
the staged commits and twenty-nine ARM64 source/consumer roots reaching the
same byte-carrier call/source-PC path join. The main aggregate correctly rejects
those failed partitions; retain it as failed until the actual roots pass.
The byte-carrier repair decodes Go's Windows object-source slash spelling
through the shared canonical-path reader, without an SDK alias. Exact file,
source line, instruction text, PC range and the original position record remain
required. Go 1.27 actual source/entry/five-target object and consumer regressions
passed in 195.235 seconds; Go 1.20 spelling/boundary controls passed in 0.700
seconds. Native Windows confirmation remains pending. Logs are
`byte-carrier-position-go127-green.log` and `byte-carrier-position-go120-green.log`.
The frozen local Go 1.27.2/LLVM 22 benchmark replay, with ordinary host source
selection, remains failed: 163/184 files, 21 failures, zero N/A/unattempted, build
8 seconds and target total 49 seconds. Counts are 29/33, 48/55, 27/30, 40/47 and
19/19 for 386, amd64, arm, arm64 and wasm. Do not compare its host-dependent
inventory to the earlier CI's 185-file denominator as a pass regression.
Reports are in the fork probe worktree under
`_out/register-result-repair/benchmark-portability-c4ff4f3b/`.
Besides the native runtime/frame contracts below, its chacha20poly1305 fixed-local
capture lacks an explicit GOENV profile; keep that failure visible, not excluded.
The default-GOENV follow-up is RED/GREEN through actual Linux fixed-local Go
export and LLVM compilation. Before package loading, an implicit nonexistent
configuration becomes explicit `off`; existing settings and user-selected paths
remain selected. CGO still follows the actual compiler default. All affected
Linux environment/fixed-local regressions passed in 87.001 seconds; host profile
controls and LLVM CLI vet passed. A fresh complete benchmark is still required.
`65ad1a65`/`73ddf781` share the real SDK numerical-source reload/binder across
both frontends, preserve actual CGO defaults and retain eager owned-directory
identity. `79f2a2e7` checks the public file/package entrypoints, real LLVM 22
objects and rejected-profile no-overwrite behavior. The integrated complete
simple CLI passed in 99.508 seconds; richer source/environment regression passed
in 37.526 seconds before the following core extension. The earlier Linux CRC32
failure stays preserved as `simple-cli-crc32-linux-0428ee36.log`, not current
head's result. Non-Linux x86 source roles and the complete stdlib gate remain
unresolved. Logs for current integration are `sdk-*-integrated-go127.log` and
`sdk-publication-integrated-go127.log`.
`9cd81fe3` adds an independent typed SDK-global observation and complete AES/
logical numerical-read consumers with original source/PC/relocation joins.
Actual SDK storage/profile guards passed in 8.441 seconds; the complete vector
family's four APIs/five x86 backend/optimization objects passed in 26.899 seconds.
Complete `gostaticstorage` passed Go 1.27/1.20 in 78.373/0.481 seconds; the shared
frontend passed Go 1.20 in 0.436 seconds, and root vet passed. That initial
batch did not close maps' entry, caller loops or private literal contracts.
`c1ab80ac`/`1de48c7d` now join the genuine scalar-register ABIInternal declarations
and exact empty current ASM DATA member without borrowing another member's
private table permission. Their broad SDK/source/frame guards passed in
482.918 seconds and ordinary consumer guards in 169.135 seconds. The private
namespace fixture regression stays failed in its earlier receipt.
`66103829` supplies the complete selected-package DATA observation through both
frontends. Integrated public CRC32/CGO1 and complete original nosimd maps CLI,
metadata and LLVM 22 O0/O2 tests passed in 125.158 seconds; root vet passed.
Full memhash amd64/386 private addresses, indexed reads and caller/pointer/loop
contracts remain unresolved. The frozen `05a664d7` benchmark replay now passes
164/184 files, with 20 failures and zero N/A/unattempted: 29/33, 49/55, 27/30,
40/47 and 19/19 for 386, amd64, arm, arm64 and wasm. Build time is two seconds;
the target total is 82 seconds, within the unchanged 300/900-second budgets.
This is one additional complete original nosimd file, not a complete benchmark
pass. Reports remain in `benchmark-sdk-05a664d7/` under the ignored report root.
`a1eb6dc2` independently binds actual Go TLS direct-NOP source rewrites; no
native thread, TLS address, g or entry permission is created. Integrated
three-API objects, unknown-state/callback and owned-command guards passed in
60.226 seconds; shared frontend tests passed in 0.384 seconds. All three module
vets and tracked Go formatting passed before the scalar-entry follow-up; final
root vet passed after it. Native Ubuntu TLS byte parity is required in the fork
probe workflow, alongside all seven repaired memory/precision fault roots.
The batch is published only on the allowed fork probe ref as `094fac7b`.
[38407134879](https://github.com/cpunion/plan9asm/actions/runs/38407134879)
passes the complete selected native Linux oracle suite in 25.032 seconds,
including all seven repaired fault/precision roots and the actual original-Go/
LLVM TLS byte comparison in 5.570 seconds. Both Windows ownership SDK lanes
also pass; the complete Windows modules are still running. Preserve the native
artifact as `native-x86-38407134879/native.log`; selected native roots do not
certify the complete main module, race gate or upstream PR head.
At frozen `c36bb835`, both Go 1.27 and Go 1.20 official five-architecture gates
pass. The four SDK/TLS source callback/reentry/close roots also pass the native
Darwin race check in 110.627 seconds; this is source lifecycle evidence, not
native x86 instruction execution or the complete race gate. The test-only
`fff0968c` adds all 201 independently read Go x86 register names to the TLS
destination-bank audit. Its complete simple CLI passes in 128.133 seconds.
Logs are `official-go*-c36bb835.log`, `sdk-source-reentry-race-c36bb835.log`
and `simple-cli-complete-bank-batch.log`.
`01f75845`/`c9098d62` repair the SDK-global source inventory boundary without
omitting nested packages, testdata or unselected original files. Its finite
4,096-file/64-MiB role budget applies to capture, Verify and offline validation;
ordinary and pinned CPU sources keep 256 files/16 MiB. New unique dependency
members are checked before retention with linear accounting, not repeated
whole-map sums. Complete Go 1.27 storage passed in 143.572 seconds and the
Go 1.20 counterpart in 1.511 seconds; integrated five-target original runtime
and typed-global/boundary regressions pass. The genuine runtime fixture retains
1,204 original files plus its selected dependencies and only observes the
pointer-free `worldsema` definition. The actual amd64 benchmark remains 49/55:
its runtime files now reach the separate pointer/linkname storage boundary,
not the old misleading missing-module error. Preserve that failed replay in
the SDK snapshot worktree's `sdk-storage-snapshot/benchmark-8f65e3de/`.
Run 38407134879 has now finished: 26 successful jobs and two failed jobs.
Both complete Windows CLIs and coverage aggregates pass. All eight main-root
parts completed their 5,867-root inventory: 5,603 passed, one failed, 263
host-inapplicable skips and zero incomplete. The main aggregate correctly fails
because `TestCrossLinuxRuntimeMatrixX86PackedZeroMemoryStoreFault` rejects the
actual Windows Go fatal header's one newline where its fixture demanded two.
The original exception is #UD (`0xc000001d`) in the exact tested ASM frame;
LLVM's native #UD counterpart already passes. Preserve all eight partitions
under `windows-complete-38407134879-all/`, not as successful root coverage.
`1b01c408` adds the independent LF/CRLF RED controls and accepts either legal
fatal-header separation, while rejecting incomplete signal fields, invalid
source positions and bare CR. Go 1.27's complete packed-zero object/diagnostic
family passes in 4.639 seconds; Go 1.20 diagnostics pass in 0.824 seconds.
Fresh native Windows confirmation is required before promoting that root.
Ownership logs are
`windows-owned-process-*-0428ee36.log`, under the report directory below.

- `7bdc112a`: Windows C text-mode CRLF now retains the exact observation
  fields; bare/embedded CR and all incomplete packets still fail. Reentry checks
  the Frame lock at the actual external callback instead of imposing ten seconds
  on complete source replay. Go 1.27 focused tests passed in 53.327 seconds,
  Go 1.20 protocol checks and vet passed, and Linux/amd64 container race passed
  all three lifecycle cases in 845.246 seconds. This is not native ISA evidence.
- Windows native diagnostics reproduced the junction failure with
  `winsymlink=1`, independent of LLVM installation. The new `internal/sourcepath`
  uses the opened file's normalized Windows final path and rechecks file identity;
  Unix retains EvalSymlinks. Its native junction RED/GREEN runs are
  [38261044890](https://github.com/cpunion/plan9asm/actions/runs/38261044890) and
  [38261402328](https://github.com/cpunion/plan9asm/actions/runs/38261402328).
  All source/tool consumers now use the shared resolver; no lexical GOROOT
  fallback remains. Focused main-module/source/CPP/tool tests passed with
  Go 1.27 and Go 1.20, and the Windows helper cross-build passed with Go 1.20.
  Complete local LLVM CLI passed in 1062.169 seconds at `0d90f0cd`.
  Native Windows source-observation integration passed in
  [38262857492](https://github.com/cpunion/plan9asm/actions/runs/38262857492),
  but its complete CLI failed and exhausted thirty minutes. Remaining failures
  include portable selection paths, physical SDK versus junction roots,
  slash-spelled replacement roots, ARM64 scratch and imported type recapture.
  Diagnostic-only workflows stay on the fork probe branch, outside the PR.
- `f18f3e97`: unresolved GLOBL sizes retain `SizeRaw` and fail translation
  with Context; no guessed 64-byte or zero-byte storage remains. Resolved
  integer sizes pass actual-Go and LLVM 22 objects on all five architectures.
- `26475c2a`: stdlib inventory expands literal includes using the selected SDK
  and target macros. All ten Go 1.27 target scans have zero parse/unsupported
  failures. This is inventory, not complete stdlib compilation. Official gates
  pass on Go 1.20 and 1.27; eight wasm baselines add only the previously omitted
  FUNCDATA form. Go 1.20's historical unsupported form remains explicit.
- The next Windows repair gives all five virtual `go/build` contexts portable
  joins and computes active CPP tool identity relative to the resolved SDK.
  Independent local RED/GREEN tests retain missing, changed and escaping source
  rejection. Module replacement roots normalize separators only, not dot
  segments. Native Windows confirmation remains required. Expanded ARM64 source
  metadata and same-source type-recapture tests diagnose the other failures.
- `e7bfc99b` / `ba02841e`: native Windows
  [38268632096](https://github.com/cpunion/plan9asm/actions/runs/38268632096)
  passed portable selection and ordinary/test declaration groups. Its ARM64
  metadata regression reproduced rejection of Go's actual `C:/...` source
  spelling; `objabi.AbsFile` explicitly emits forward slashes on Windows.
  Normalize only those separators before the existing canonical-file/digest
  checks. Local Go 1.27 scratch source/SDK/call-result tests passed in 266.515
  seconds; Go 1.20 metadata checks passed. Native metadata, SDK body and actual
  CLI scratch closure then passed in
  [38270866927](https://github.com/cpunion/plan9asm/actions/runs/38270866927)
  with `b7f3309f`'s separator repair.
  The complete native diagnostic run also passed the CPP/generated-header group
  (32 top-level tests, 592.455 seconds). Instrumentation retained four failures:
  the CLI test binary lacked `.exe`, and short/physical invocation names changed
  actual same-module exports. A local failing command-directory regression now
  covers producer/consumer physical invocation consistency; do not relax export
  equality. `d4db4086` binds the physical invocation and uses the host executable
  suffix. The native follow-up exposed an older fixture-only GOFLAGS leak:
  test source replay copied mode features but not the observed command context,
  retaining the runner's `-tags=byollvm`. Five fixture callers now share the
  captured context and explicit GOENV=off; a local ambient-tag RED/GREEN test
  preserves strict production environment rejection. Native instrumentation
  verification is still required; its CLI step was not reached in that run.

Native instrumentation then passed
[38271832444](https://github.com/cpunion/plan9asm/actions/runs/38271832444)
at `c94f30eb`: the physical type-source fixture and all four previously failing
CLI roots passed, including imported SDK/same-module and foreign original-source
recapture. Complete local LLVM CLI also passed at that frozen revision in
1061.421 seconds; its complete local profile suite passed in 926.033 seconds. Fork-only
[38273415720](https://github.com/cpunion/plan9asm/actions/runs/38273415720)
diagnoses every actual Windows CLI/profile inventory root in four disjoint
groups per package; raw inventories/outcomes remain artifacts, not a replacement
for the complete contribution gates or Windows process-ownership implementation.
That native run finished without missing/skipped roots: CLI 259/260 and profile
261/262 passed. Remaining roots were the actual CRC32 numerical source reload
and the quoted-path instrumentation export. The latter normalizes native archive
paths but missed Go's slash-spelled object source names; exact observed roots
now normalize both spellings without relaxing the nm symbol grammar. Local
Go 1.27 actual export and Go 1.20 grammar controls pass. The CRC32 reload error
now includes the actual package/type/error fields; this is diagnosis, not a fix.
Its complete local explicit/default-profile regression passed in 84.189 seconds.
Native [38317575302](https://github.com/cpunion/plan9asm/actions/runs/38317575302)
passed the quoted-path export and both ARM64/x86 object-source metadata controls.
CRC32's actual error was `std/hash/crc32` with an illegal internal import:
different SDK-root/package-directory spellings changed cmd/go's namespace.
The reload now uses one physical SDK root and package directory. Its comparison
projection resolves every selected source without rewriting original package
snapshots; byte-identical foreign files retain distinct identities. The new
local alias-identity regression is RED/GREEN, and the complete CRC32 explicit/
default-profile test passed in 81.877 seconds. Native
[38318470610](https://github.com/cpunion/plan9asm/actions/runs/38318470610)
confirmed the namespace repair and reached the next CRC32 failure: the header
import-config writer rejected Windows export filenames containing backslashes.
The writer now distinguishes import-path grammar from literal export filenames,
matching the compiler's first-equals-sign split. Newline/NUL injection, invalid
imports and missing exports remain rejected. Portable path controls are
RED/GREEN on Go 1.27 and pass on Go 1.20; root vet and the complete local CRC32
explicit/default-profile regression pass (91.653 seconds). Native
[38319672694](https://github.com/cpunion/plan9asm/actions/runs/38319672694)
passed the complete CRC32 explicit/default-profile test in 142.000 seconds at
`01accd6f`, including rejection controls and actual LLVM object outputs.
Follow-up [38320341678](https://github.com/cpunion/plan9asm/actions/runs/38320341678)
completed all nine diagnostic jobs successfully at the same code revision.
Raw inventory/outcome unions contain CLI 261/261 and profile 263/263 passed
roots, with no missing/duplicate roots or unterminated children. Profile's 2024
root/child executions have no skips; CLI's 2129 include one explicit native
ARM64-only child skip on Windows/amd64. Its complete-argument execution
counterpart passed on native Darwin/arm64 at `cf18ec55` (1.583 seconds, no skips),
including all three target objects. Do not count the Windows skip as execution.
The original Windows CMPXCHG Go/C observation and rejection controls passed;
741 conversion forms also passed three Go targets, three LLVM APIs and O0/O2,
as did the readonly DATA three-target objects. These are native diagnostics,
not the full main-module/coverage/race or benchmark contribution gates.

The frozen `3b4c246f` local complete CLI and profile suites passed in
1168.046 and 1041.168 seconds respectively; these predate the header-config fix.
At `01accd6f`, all internal-package tests and the Go 1.20/1.27 five-architecture
official coverage gates passed. The Go 1.20 historical unsupported form remains
explicit; classification is not runtime execution. Native diagnostic timings
explain the CLI's global deadline failure: the prior complete list totals
2440.65 seconds across roots, exceeding its 1800-second direct-suite budget.
Four exhaustive diagnostic groups take 489.56–713.05 seconds each. Formal
partitioning still needs complete-inventory/coverage aggregation; do not extend
the deadline or promote diagnostic-only receipts into the contribution gate.
At `cf18ec55`, the complete SDK ASM DATA test family passed in 89.244 seconds,
as did `go vet ./...` and all thirteen workflow scheduling regressions.

The next metadata repair shares ARM64/x86 object-source spelling validation.
Four local RED controls reproduced ARM64 accepting `$GOROOT` dot/repeated-slash
aliases after Join cleaned them. Check spelling before joining; normalize only
Windows separators, then retain each consumer's captured file/hash checks.
Go 1.27 metadata and actual SDK/object position regressions passed, as did the
Go 1.20 metadata counterpart. Both metadata controls also passed in native
run 38318470610. Reports are `object-source-alias-red.log` and
`object-source-path-*.log`; this is not a new complete-head pass.

The complete local Go 1.27 `gotoolprofile` suite passed in 1031.912 seconds on
the frozen `ba02841e` checkout. Later type-invocation and fixture repairs have
focused receipts, not a new complete-module certification.
- `43d64b2a`: downloaded artifacts showed module aggregates differed from root
  owners in PATH because only root jobs installed QEMU. Every Linux role now
  installs the same pinned QEMU before Node; exact environment/source comparisons
  remain mandatory. The fixture inventory uses portable `grep -Fx`, not undeclared
  ripgrep. Windows CLI coverage uses explicit Bash and requires a nonempty profile.
  The generic-alias oracle sets `gotypesalias=1` for Go 1.24–1.26 defaults.
- `2efcb98f`: original-Go Windows oracles use `.exe`; the C counterpart uses the
  shared host-executable helper. An actual cross-built PE regression reproduced
  the missing suffix. The SDK mutation suite reuses its unchanged observed source.
- `177d651c`: only race grows from eight to sixteen root scheduling parts.
  Each owner has one two-worker wave, retaining the original 32 hash shards,
  command/job budgets and denominator. Actual sixteen-part aggregation tests
  retain missing/foreign/duplicate/shrunken-inventory guards.
- `ebfc63af`: subtool directory errors identify GOROOT/GOTOOLDIR/GOCACHE while
  preserving their cause. This is diagnostic work, not a Windows fix.
- `75db11da`: each mutated bind already performs full live verification, so the
  test no longer duplicates it. It checks source revocation after each rejection
  and still finishes with a successful unchanged original binding.

At the frozen code bytes committed as `75db11da`, focused Go 1.27 root/profile/
runner suites passed in 80.655/3.029/140.601 seconds; Go 1.20 counterparts passed.
Thirteen scheduling tests, actionlint, shell syntax and root vet passed. Actual
Go 1.24.13, 1.25.14 and 1.26.9 generic-alias/policy regressions also passed.
Linux/amd64 container race verification of the SDK mutation root passed in
802.933 seconds, including all seventeen rejections and the positive control.
The prior 1200-second failure remains failed. This establishes race/source
lifecycle behavior, not native x86 ISA execution under Rosetta.

Reports remain under ignored `_out/register-result-repair/`: `ci-2a04-*`,
`ci-environment-2a04-*`, `original-windows-executable-*`,
`ci-followup-16-*`, `race-scheduling-16-*`, `subtool-path-origin-role-red.log`
and `sdk-binding-reuse-linux-race-v2.log`. Receipts bind frozen source and
tools; later edits do not inherit a complete-head pass.

## Remaining work, in priority order

1. Confirm the single remaining Windows packed-store diagnostic failure in a
   fresh native run, then complete the main-module aggregate. Partitioning
   retains inventory and the thirty-minute budgets; both complete CLIs,
   process ownership, source-path repairs and the other fault drivers now pass.
   The latest 5,867 main-root outcomes are retained above. Selected native
   regressions are not complete-module certification.
2. Rerun affected complete gates and both CLIs, publish to the existing
   allowed-fork contribution ref, then inspect fresh CI. No upstream manual
   cancellation/rerun is authorized. The sixteen-part Linux race schedule
   completed every owner; the
   remaining callback-timer repair still needs the full current-head race gate.
3. Darwin stdlib compile replay at `ba02841e` clears inventory parsing but fails
   at CRC32's CPU-feature SB comparison/result initialization in the simple
   `plan9asm` CLI. Its richer `plan9asmll` counterpart has a separate actual SDK
   numerical-source binding. Do not turn the missing binding into initialized
   bytes or a skip. Other runtime/frame contracts remain; missing generated CPP
   headers stay unresolved, never ignored/no-op macros.
4. The 20 benchmark failures still require runtime/reflect entry, TLS,
   hidden-register, caller-storage or source contracts. SDK runtime's repaired
   complete inventory supplies neither pointer/linkname global authority nor
   a physical native-entry contract. Dynamic reflect frames
   cannot be inferred from `func()`. ARM64 result-tail bytes are observable by
   assembly callers. Maps AES mutable-global storage remains separate from its
   repaired shuffle effects. Do not force CGO_ENABLED=0 in ordinary loading:
   that changes selected runtime/cgo source and hides a different contract.
5. Required LLVM 22 tools fail if missing. Official opcode/form classification,
   LLVM object success and executed semantics remain separate claims.

Earlier complete CLI receipts at `0527ded5` were 0.998 seconds (`plan9asm`) and
1028.540 seconds (`plan9asmll`). Earlier repairs and their failed controls remain
in Git history and ignored reports, not current-head completion evidence.
Use [instruction-family development](instructions.md) and [validation](validation.md).
Keep this handoff concise, not a second ledger.
