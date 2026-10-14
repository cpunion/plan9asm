# Current work: repair PR 40

Continue from `codex/pr40-external-corpus-split-20261011`, not older family
worktrees. Inspect status, worktrees and remotes. Push only to an allowed fork;
keep the contribution draft until current-head CI, review and coverage pass.
Let active frozen CI runs finish before publishing another batch.

## Current checkpoint: 2026-10-14

Fork [38477743604](https://github.com/cpunion/plan9asm/actions/runs/38477743604)
tests frozen `7c342c8b` (contribution `2d696307`, plus exact fork priorities).
It naturally finishes failed: nine priority jobs pass, one fails, and the full
Windows barrier correctly skips. All 48 selected native Linux x86 roots pass
in 95.188 seconds, with zero skips, missing or duplicate terminals; both actual
instruction-byte and genuine Go ABI0 bool CMP axes execute. Both Windows SDK
ownership lanes and the three new shared path roots also pass. The original
default-profile root and canonical-path counterpart now pass natively in
21.22/28.63 seconds, retaining the actual undeclared-helper rejection.
The remaining failure is CRC32's effective-defaults public CLI child: Go rejects
`internal/byteorder` because physical GOROOT is paired with an unresolved cwd.
Its explicit profile and direct compilation pass. Preserve the failed priority
artifact; this run does not certify complete Windows modules or the benchmark.

`9bf01156` pairs that same actual working directory's opened physical spelling
with the environment before both initial package loads. The shared resolver
does not change package selection, internal-import rules or source permissions.
Actual aliased SDK inventory reproduces the defect independently; outside-SDK
internal imports still fail. Complete simple CLI passes 31 roots/26 children in
389.095 seconds; the affected LLVM CLI passes 18 roots/34 children in 203.537
seconds, with exact fresh inventories and no skips. Shared Go 1.27/1.20, race,
all three vet gates and formatting pass. Parent integration separately passes
the actual CRC32 public/default profiles in 168.107 seconds, both directory
controls in 2.238 seconds and Go 1.20 shared controls in 0.629 seconds. Native
Windows confirmation of this paired-directory repair remains required.

Fork [38469815846](https://github.com/cpunion/plan9asm/actions/runs/38469815846)
tests frozen `dd35c7c1` (contribution `7251abd3`, plus fork-only priorities).
It naturally finishes failed with 41/43 successful jobs. The main root's exact
5,974-root inventory has 5,711 passes and 263 explicit platform skips; all 21
other main-module packages pass. Simple CLI passes all 29 roots. Raw terminal
unions, fresh inventories, source/tool identities and coverage merges agree,
without missing, duplicate or incomplete outcomes. LLVM CLI has 262/263 passing
roots: `TestRunOneTargetSDKFixedLocalPreservesActualDefaultProfile` fails both
profiles at the strict original command-directory identity guard. Its aggregate
correctly fails and does not publish coverage. All ten priority/native jobs pass;
the installer retry mechanisms do not convert a test failure into success.

`91b8d174` independently reproduces that command-path class and canonicalizes
the actual selected caches, SDK root, temporary directories and existing GOENV
file before package loading. GOPATH components retain their original order and
physical targets; explicit missing configurations and actual CGO/config defaults
stay visible. Strict source guards and private-helper/frame permissions do not
change. Final-byte affected tests pass all 19 LLVM CLI roots in 168.883 seconds
and five simple SDK roots in 191.320 seconds; Go 1.20, race and vet also pass.
Complete chacha still contains all three original TEXTs and rejects its undeclared
private helper without output. Native Windows confirmation is still required.

`5427fb76` completes the existing B/W/L/Q CMP grammar for independently proved
caller-address participants, retaining numerical, memory and hidden-entry
boundaries. `11e3dff5` strengthens its real Go ABI0 counterpart to a one-byte
bool return with seven untouched adjacent bytes. The two Linux AMD64 roots
pass under emulation in 32.725 seconds, not native execution. Parent integration
passes all seven portable/object roots in 4.359 seconds. Both Go 1.27 and the
complete official Go 1.20 five-architecture classification gates pass at this
frozen parent; Go 1.20 retains its explicit historical AMD64 unsupported form.
The following fork batch must inventory and execute both native CMP axes and
the actual failed/default-profile and canonical-path roots before full modules.

The integrated frozen `2d696307` replays both actual CLI path roots successfully
in 16.557 seconds, the complete Go 1.20 shared frontend in 0.499 seconds and both
root/LLVM CLI vet gates. Fork `7c342c8b` stages that source and explicit native
inventories in [38477743604](https://github.com/cpunion/plan9asm/actions/runs/38477743604);
let the run finish before publishing another batch. The fresh five-target
benchmark at `2d696307` remains failed: 163/184 files, 21 failures, zero N/A or
unattempted; build three seconds and target total 113 seconds, within unchanged
300/900-second budgets. Its counts remain 29/33, 48/55, 27/30, 40/47 and 19/19.

The SDK literal-minimum frontend factory remains unmerged on
`codex/pr40-sdk-numeric-maps-entry-20261013`. Its atomic numerical-source
revocation must be rechecked after the final caller-context callback, including
Verify, Readback and all three translation publication guards. Preserve its
failed old-source infrastructure receipts and complete current-parent lifecycle
verification before integrating it. Neither that factory nor the numeric-only
LEA effect follow-up is included in the frozen fork run above.

Fork [38441921522](https://github.com/cpunion/plan9asm/actions/runs/38441921522)
tests frozen `19e74032` (contribution source `b3755a3e`, plus fork-only workflow).
The run completed with 28/28 successful jobs. Both Windows ownership lanes and
the selected native x86 suite pass. The three
previous actual failed roots execute and pass before the complete modules:
the 386 instruction fixture, sink return/teardown protocol and LLVM CLI platform
query. Complete Windows inventories and strict aggregates pass: main
5,906 roots = 5,643 passed + 263 explicit platform skips; simple CLI 29/29 and
LLVM CLI 262/262 roots pass. The latter retains one native ARM64-only child skip.
These skips are not executed semantics. This verifies the frozen fork run, not
the full contribution's Linux/race/stdlib/benchmark gates. PR 40's published
head is still `24851df2`, not these staged fixes.

The following fork [38456568476](https://github.com/cpunion/plan9asm/actions/runs/38456568476)
tests frozen `05367452` (contribution `d0f0701a`, plus the fork-only workflow).
It naturally finishes failed: both Windows ownership SDK lanes and the native
Linux suite pass, as does the LLVM priority inventory. All 40 selected native
roots pass in 26.663 seconds without skips, including canonical bool instruction
bytes and the genuine Go ABI0 byte-representation counterpart. The simple SDK
priority cohort times out at eight minutes after both CRC32 source roots pass
and file/package publication succeeds; remaining publication controls and maps
are incomplete, not passed. Complete Windows modules correctly do not start.
Split these slow priority roots into independent lanes with the original
per-command budgets, keeping exact inventories and the complete-module barrier.

Fork [38466545308](https://github.com/cpunion/plan9asm/actions/runs/38466545308)
tests frozen `192cacc9` (contribution `37fe7732`, plus fork-only priorities).
It naturally finishes failed: seven jobs pass, three fail installation and the
complete Windows barrier correctly does not start. All 46 selected native x86
roots pass in 87.864 seconds with no skips, failures, duplicates or omissions.
All six independently inventoried simple/source/frame SDK lanes also pass.
Three jobs fail before any tests:
two MSYS2 database upgrades and one pinned-package dependency download hit the
mirror's ten-second low-speed limit. Preserve them as infrastructure failures,
not source rejections or test passes. The following installer repair retries
bootstrap at most three times and the identical LLVM 22 package transaction at
most three times. The last bootstrap failure is fatal; exhausted transactions
return the final exit code, and the exact version check precedes build flags.
Local actual-script fault injection covers early success, retry success, full
exhaustion and LLVM 23 rejection. This is not Windows installation evidence;
the completed failed run is unchanged and the repair needs the next native batch.

At frozen `b3755a3e`, the complete local LLVM CLI passes in 862.415 seconds and
the complete new SDK literal-alignment producer's race cohort passes in 440.062
seconds. `4818b84c` selects the actual Darwin SDK for independent LLVM 22 C
oracles; the earlier 899.925-second missing-SDK failure remains failed.
The official complete Go 1.20.14 SDK also passes all six existing alignment
roots (five architectures, no skips) in 1.922 seconds. Its original wasm assets
are required; the downloaded-module SDK's missing-asset failure remains failed.

`b3755a3e` observes minimum alignment of uniquely retained original SDK literals
through the complete original archive graph and ordinary Go linker. It grants
neither numeric addresses nor read/call permission. `9e7879f2` preserves the
SDK vendor namespace/import aliases and compacts consecutive CPP position
markers within unchanged budgets. The complete untouched chacha source retains
all three TEXTs; this source observation is not translation or execution success.
An independent alignment regression also reproduces `go list runtime .` using
the vendor directory's unrelated namespace. Both capture and live verification
now query the original logical package path. The genuine empty-main link still
rejects unretained literals; the focused regression passes in 7.542 seconds.
These later source repairs are not in the completed frozen CI run.

At `df5348d7`, the combined SDK source/literal/bool/result class passes Go 1.27
in 203.695 seconds and its Go 1.20 counterpart in 13.322 seconds. Both official
five-architecture classification gates and root vet pass. Go 1.20 retains its
explicit historical unsupported form; do not describe that baseline as zero
unsupported. [Canonical bool results](x86-register-bool-result.md) require a
separate initialized/canonical byte must-fact through every public lowerer.
The standalone original instruction-byte and genuine Go ABI0 representation
oracles passed under Linux emulation; native x86 execution is still required.

The SDK-primary frame observer now consumes the original selected config/CGO
profile and complete native TEXT/frame inventory, without granting owned-main
frame permissions. Its GOENV identity uses the shared eager handle capture on
Windows. Full original chacha helper and split-stack functions remain present;
their caller/entry/frame contracts are not bypassed. The integrated observer
and new native bool roots still need the following frozen validation/CI batch.

The complete `e6bf0bf6` CLI replay found a genuine namespace convergence
regression: simple CLI 25/29 passed in 36.141 seconds; LLVM CLI 269/270 passed
in 813.834 seconds. All five failed roots compare complete numerical/literal
SDK selections. DATA queried the logical package but numerical capture/verify
still queried `.`; the raw `Match` fields therefore genuinely differed.
`21121e24` uses one logical selector for SDK capture, live verification and
dependency graphs, keeping ordinary owned queries at `.` and retaining strict
raw-field comparisons. Independent RED/GREEN and actual complete CRC32
three-API/O0/O2 object tests pass in 64.683 seconds. Do not erase `Match` or
rewrite either failed complete receipt.

At `515cd1d2`, complete simple CLI passes in 195.264 seconds, the SDK numerical/
literal/profile join class in 265.576 seconds, and portable Go 1.20 controls in
2.694 seconds. All five actual previously failed CLI roots also pass their
priority inventory. The complete LLVM CLI replay then passes in 956.491 seconds:
all 270 root tests and 1,872 subtests pass, with no skips, incomplete outcomes or
inventory omissions. Both Darwin native C oracles actually execute. Source and
tool identities remain unchanged throughout this frozen `515cd1d2` replay; it
does not certify later commits or the published PR head.
The SDK-primary direct inventory now binds all root members and child-directory
identities without treating child generators as selected source. Existing
ordinary flat inventory, byte/member bounds and relative-include rejection stay
unchanged.

`c6ed942b` adds an explicit live SDK literal-minimum handoff for masked address
predicates. It grants only authenticated low-bit facts, not numeric addresses,
memory extent, native entry or caller permissions. OR propagates these facts
without making high-bit FLAGS known; SETcc/Jcc/CMOV/LAHF/PUSHF rejection controls
retain that distinction. The integrated literal/numerical/bool family passes in
278.822 seconds. Complete original maps source still needs a separate typed
caller-address contract; do not promote this predicate repair to a file pass.

`99a1146e` closes that separate typed caller-address predicate class. Its
must-domain is distinct from literal/OR identities and numerical initialization;
all B/W/L/Q TEST views, typed overwrites and joins retain the original FLAGS
contract and unchanged budgets. Complete unmodified maps source passes both
TEXTs/all DATA through three APIs, three OS profiles and LLVM 22 O0/O2 objects
in 203.337 seconds. Mutable AES keys remain external loads and actual literal
alignment agrees with the observed floor. The independent instruction-byte and
genuine Go ABI0 caller axes pass under Linux emulation, not native SDK execution.
The frontends still need their explicit typed-demand minimum handoff, so this
does not establish benchmark completion. Parent focused integration passes in
5.404 seconds; retain the separate 975.640-second Go 1.27 family receipt.

`c761245d` hands authenticated SDK-primary ABI0 frames to the existing fixed-local
checker without granting owned-main, helper, split-stack or caller permissions.
Actual MD5 frame capture passes; its unrelated complete body remains Context at
SHRQ. Complete chacha retains its undeclared private helper instead of cropping
it. `d63f7b76` reproduces ordinary Verify and SDK Bind caller-context callbacks
under a frame lock, then samples them outside locks through a bounded private
command context. Published bindings retain the original lifetime and permanent
revocation. Parent frame/context/C-driver integration passes in 125.987 seconds
and the real CLI default/saved-profile cohort in 13.041 seconds.

`79b7ab3f` fixes source TEXT ordinal versus Go's physical namespace-bank order.
Exact symbol/ABI/namespace/index associations retain original physical records,
code, PCs and relocations; SDK native listings consume the source association.
The complete actual bytealg metadata now verifies, while whole equal translation
still rejects its unresolved helper tail, DX closure and no-DATA source handoff.
The parent combined mapping/frame/caller/minimum cohort passes Go 1.27 in 47.337
seconds; Go 1.20 mapping and caller/context controls pass in 0.323/5.913 seconds.
These receipts do not replace the still-failed complete benchmark or broad host
cohort. Native instruction/caller confirmation needs the following fork batch.

The following frozen `05aea382` five-target compile benchmark remains failed:
163/184 assembly files compile, with 21 failures and zero N/A/unattempted.
Counts are 29/33, 48/55, 27/30, 40/47 and 19/19; build takes ten seconds and
target total 120 seconds, within unchanged 300/900-second budgets. Compared
with `3f437f3a`, the additional failed file is `internal/bytealg/equal_amd64.s`:
the canonical bool result check now reaches an unresolved whole-source helper
branch, and the SDK fallback lacks its exact cross-namespace/native ABI TEXT
join. Keep canonical bool requirements and this actual failed receipt intact;
neither an opcode classification pass nor a local CLI test pass closes it.
The current Go 1.27 and official Go 1.20 five-architecture classification gates
also pass at this frozen head; the older AMD64 baseline still has one explicit
unsupported form.

An independently reproduced native C-driver SDK omission affects numerical
fault observation and two core-limit helpers. Explicit native-host commands now
share the actual Darwin SDK selection while retaining LLVM 22 and owned bounded
contexts; cross-target object commands are unchanged. Complete repaired
fault/core-limit/native-entry/header/link controls pass Go 1.27 in 33.764 seconds
and Go 1.20 in 10.482 seconds. Neither this repair nor the SDK frame factory
certifies a broad Go 1.27 cohort that failed after 2,097.742 seconds: that receipt
also retains missing-Rosetta diagnostics and subsequent deadline exhaustion.
Split complete inventories instead of increasing the timeout or treating those
infrastructure failures as successful source exclusions.

Both official five-architecture gates also pass at frozen `5b269f5e` with
current Go 1.27 and the complete official Go 1.20.14 SDK. No baseline is changed;
the historical AMD64 unsupported form remains explicit.
The disjoint `c9fb679b` replay passes all 29 numerical-source roots in 272.663
seconds and all 63 SDK DATA/literal roots in 661.541 seconds. Its ordinary
numerical cohort remains failed: 50/51 pass, while an old cancellation fixture
misses the inner frame callback after lifetime sampling moves into a helper.
The test-only follow-up follows the complete exact observer route and checks
actual Err/Deadline/Done cancellation, lock-free callbacks, outside-route
negative controls and permanent revocation. Its seven-root Go 1.27 cohort passes
in 120.713 seconds, race in 316.756 seconds and portable Go 1.20 controls in
3.224 seconds. Production permissions are unchanged. Preserve the broad failed
receipt and rerun its exact inventory; these mixed snapshots are not a complete
277-root current-head certification. The exact failed 51-root inventory is then
replayed at `91fabe98`, source-identical to `37fe7732`: all 51 roots and 717
children pass in 912.491 seconds with zero skips, missing, duplicate or foreign
roots. This closes that cohort, not the remaining complete SDK inventory.

`15f31a2e` extends the exact source/object TEXT association to the numerical
native consumer. Original/expanded whole tuples and live owner revocation stay
strict; physical records are never reordered and no tail/closure/entry role is
added. Its independent complete class passes 37 roots/216 children in 533.639
seconds, race in 102.722 seconds and Go 1.20 controls in 1.021 seconds. Parent
integration runs all five exact roots in 12.081 seconds, portable controls in
0.845 seconds and vet successfully. Whole bytealg still has one actual failed
file: equal's same-file helper tail, hidden DX closure and no-DATA handoff remain
unresolved; do not promote metadata or this consumer repair to a whole-file pass.

Inventory runs only on the server. Streaming ledger status and a complete
five-minute incremental window have been verified before restoring catch-up;
the trusted contiguous head now advances through 2026-10-08T16:22:43Z.
The requested historical cutoff remains 2025-10-01; subsequent hourly windows
publish only after their complete scan. Real source failures remain retryable,
not pass/skip. No local index/ledger scan is needed for translator work.

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
Fresh native [38419176385](https://github.com/cpunion/plan9asm/actions/runs/38419176385)
now passes that actual original-Go/LLVM fault root on both Go 1.20 and 1.27;
their complete packed-store suites take 2.841/2.123 seconds. Ownership, source
paths and repeated real process-tree checks also pass both SDKs. The run has
finished with all 28 jobs successful, including the three complete Windows
modules and their coverage aggregates. All eight main-root parts retain the
5,868-root inventory: 5,605 passed, zero failed, 263 host-inapplicable skips and
zero incomplete. The skips are not execution passes. Raw partitions and the
independent completion artifact remain in `windows-complete-38419176385-all/`
and `windows-complete-38419176385-completion/`; retain the older run as failed.
The intermediate fork-only ownership step in 38418570567 failed before tests
because it omitted the explicit `byollvm` tag; `3b796b07` corrects that step,
without changing production source or treating the infrastructure error as a pass.

`7011ea62` completes Go's named MOVQ aliases, MOVD and MOVDQ2Q, through shared
source-only dispatch. Original opcode/text and caller/frame certificate identity
remain unchanged; encoded MMX instructions retain their separate native-state
boundary. Independent complete Go MOVQ rows, source/raw rejection controls and
three-API objects pass. Its integrated root regressions pass in 16.439 seconds;
the complete local simple CLI passes in 171.127 seconds. The new native alias
root and counted-loop/fixed-local consumers still need the next fork batch;
38419176385 predates this alias implementation and cannot certify it.
`c832dad9` completes counted provenance, fixed-local widths, caller shapes and
carry initialization; affected Go 1.27/1.20 suites pass in 847.549/419.507 seconds.
`73e582f5`/`17659d69` bind bounded private SDK literal address/index reads and the
same source aliases, retaining original identities and all source functions.
Maps still reaches Context at linked address-bit observations, not an invented
alignment guarantee. Their isolated broad/final guards pass; integrated current
head validation is required before publishing the next batch.
The SDK numerical source now independently observes Darwin/amd64 and Windows
x86 profiles rather than borrowing Linux metadata. Exact default PIE/shared
object flags, original CPU NonpkgDefs and complete compiler directive prologues
retain their typed source/object joins; none supplies native imports or calls.
The original unchanged CRC32 source first failed on both platforms, then passed
LLVM 22 objects in 49.739 seconds. Integrated core/source/literal/alias guards
pass in 361.220 seconds and the complete simple CLI in 194.169 seconds. The
first complete storage run retained one stale fixture failure because it still
called now-supported Windows an unknown target; the genuinely unsupported
FreeBSD case replaces that input while cross-target report/profile mismatches
remain rejected. The final complete storage/frontend suites pass Go 1.27 in
111.924/0.529 seconds and Go 1.20 in 1.400/0.368 seconds; root vet passes too.
The complete Darwin corpus now advances past original CRC32 to maps' independent
linked-address-bit boundary. Preserve `stdlib-darwin-sdk-platform-final.log` as
failed, not as complete corpus or benchmark success.
The integrated batch is published only on the allowed fork probe as `0ba4cf94`.
[38429537721](https://github.com/cpunion/plan9asm/actions/runs/38429537721)
passes the selected native Linux suite in 29.409 seconds. Both independently
inventoried MOVQ alias roots execute: the direct root takes 1.290 seconds and
the counted-provenance root 1.960 seconds, with 9,216 original-Go calls and
independent LLVM O0/O2 references. Native TLS parity also passes in 5.690 seconds.
Complete Windows simple CLI execution/coverage passes. The LLVM CLI retains
one failed root: its old cost-query fixture still labels Darwin unsupported.
`c90920ed` independently reproduces that failure locally, adds Darwin/Windows
positive controls and keeps FreeBSD and unavailable Darwin/386 negative controls;
the complete focused SDK-numerical group passes in 0.608 seconds. This repair
is not in that frozen run. It finished with 23 successful jobs and five failed
jobs: three actual failed roots and two correctly failed module aggregates.
The complete main-root inventory is 5,896: 5,631 passed, two failed, 263
host-inapplicable skips and zero incomplete. Raw partitions remain under
`windows-complete-38429537721-all/`; do not promote their skipped or failed roots.
`3f3b7dbe` corrects the old 386 fixture's Go-rejected `MOVQ M0,AX` to the actual
MOVL dword transfer and a dword result. Actual Go accepts MOVL and rejects MOVQ
and both source aliases; the affected complete family passes Go 1.27/1.20 in
33.503/28.205 seconds. `33027aed` reproduces sink-return misclassification during
slow process teardown and supplies immediate continuation markers through both
original-Go and LLVM C drivers. Five complete sink/protocol repetitions pass
in 39.605 seconds; Go 1.20 and Go 1.27 race counterparts pass in 8.947/10.092
seconds. The observation window and direct-child exit checks are unchanged.
Complete native confirmation of all three root repairs passes the priority gates
and complete-module validation in 38441921522, as recorded above.
The complete local LLVM CLI retains a separate infrastructure failure after
899.925 seconds: two native ARM64 C oracles select the Homebrew clang default
configuration's absent SDK. Explicit CGO flags and SDKROOT do not fix those
direct C commands. Keep that receipt failed; the shared actual-SDK repair and
subsequent complete pass are recorded in the current checkpoint above. Never
fall back from LLVM 22 or modify global tool configuration.
The fresh five-target benchmark at `3f437f3a` remains 164/184, 20 failed and zero
N/A/unattempted. Build is nine seconds and target total 123 seconds, within the
unchanged 300/900-second budgets. Its updated failures still expose actual SDK
linked literal addresses, runtime/frame contracts and the fixed-local SDK role;
`benchmark-platform-3f437f3a/` is a failed report, not current-head completion.
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

1. Finish the frozen repaired LLVM CLI replay after the five real namespace
   failures pass priority execution. Integrate only closed instruction/proof
   families, rerun their affected joins, then batch the next fork push. Include
   all five prior CLI roots, actual new native bool roots and SDK/profile
   controls in explicit inventories before complete Windows modules. Retain
   strict aggregates, platform classifications and thirty-minute budgets.
2. Rerun affected complete gates and both CLIs, publish to the existing
   allowed-fork contribution ref, then inspect fresh CI. No upstream manual
   cancellation/rerun is authorized. The sixteen-part Linux race schedule
   completed every owner; the
   remaining callback-timer repair still needs the full current-head race gate.
3. Rerun the complete stdlib compilation gate with the shared SDK numerical
   binder now integrated in both frontends. The old `ba02841e` CRC32 failure is
   not current-head evidence; focused original CRC32 source/objects and the
   complete simple CLI now pass. Other runtime/frame contracts remain; missing
   generated CPP headers stay unresolved, never ignored/no-op macros.
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
