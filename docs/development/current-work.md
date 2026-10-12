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
1061.421 seconds. Its full local profile suite is still running. Fork-only
[38273415720](https://github.com/cpunion/plan9asm/actions/runs/38273415720)
diagnoses every actual Windows CLI/profile inventory root in four disjoint
groups per package; raw inventories/outcomes remain artifacts, not a replacement
for the complete contribution gates or Windows process-ownership implementation.

The next metadata repair shares ARM64/x86 object-source spelling validation.
Four local RED controls reproduced ARM64 accepting `$GOROOT` dot/repeated-slash
aliases after Join cleaned them. Check spelling before joining; normalize only
Windows separators, then retain each consumer's captured file/hash checks.
Go 1.27 metadata and actual SDK/object position regressions passed, as did the
Go 1.20 metadata counterpart. The new x86 slash-spelling control still needs
native Windows confirmation. Reports are `object-source-alias-red.log` and
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

1. Rerun affected complete gates and both CLIs, publish to the existing
   allowed-fork contribution ref, then inspect fresh CI. No upstream manual
   cancellation/rerun is authorized. Native Windows executable/coverage fixes
   require their real host jobs; diagnostics alone do not resolve bugs.
2. Validate the common Windows junction repair through the complete CLI, actual
   SDK routes and CPP rejection controls. A standalone path test is not complete
   compilation/linking or source-proof validation.
3. Windows' complete main-module suite exhausts its thirty-minute deadline.
   Partition without losing actual inventory. Do not enable the Unix-only
   exhaustive runner without real descendant ownership and Windows tool/source
   scope. The sixteen-part Linux race schedule completed every owner; the
   remaining callback-timer repair still needs the full current-head race gate.
4. Darwin stdlib compile replay at `ba02841e` clears inventory parsing but fails
   at CRC32's CPU-feature SB comparison/result initialization in the simple
   `plan9asm` CLI. Its richer `plan9asmll` counterpart has a separate actual SDK
   numerical-source binding. Do not turn the missing binding into initialized
   bytes or a skip. Other runtime/frame contracts remain; missing generated CPP
   headers stay unresolved, never ignored/no-op macros.
5. The 21 benchmark failures still require runtime/reflect entry, TLS,
   hidden-register, caller-storage or source contracts. Dynamic reflect frames
   cannot be inferred from `func()`. ARM64 result-tail bytes are observable by
   assembly callers. Maps AES mutable-global storage remains separate from its
   repaired shuffle effects. Do not force CGO_ENABLED=0 in ordinary loading:
   that changes selected runtime/cgo source and hides a different contract.
6. Required LLVM 22 tools fail if missing. Official opcode/form classification,
   LLVM object success and executed semantics remain separate claims.

Earlier complete CLI receipts at `0527ded5` were 0.998 seconds (`plan9asm`) and
1028.540 seconds (`plan9asmll`). Earlier repairs and their failed controls remain
in Git history and ignored reports, not current-head completion evidence.
Use [instruction-family development](instructions.md) and [validation](validation.md).
Keep this handoff concise, not a second ledger.
