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

Published `2a04a4b2` completed [run 38211857025](https://github.com/xgo-dev/plan9asm/actions/runs/38211857025)
with 120 successful, 29 failed and one cancelled job. All sixteen cross-runtime
execution owners and their aggregate passed. All ten coverage execution owners
passed, but coverage aggregation failed. Benchmark compilation remained 164/185:
386 29/33, amd64 49/56, arm 27/30, arm64 40/47 and wasm 19/19.
These are that frozen run's results, not certification of the following repairs.

## Latest local repair batch

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
2. Windows `plan9asmll` has a common bare path error during feature capture and
   SDK CPP fallback. Use directory-origin diagnostics to locate it; do not
   replace failed source/tool observations with guessed paths or skips.
3. Windows' complete main-module suite exhausts its thirty-minute deadline.
   Partition without losing actual inventory. Do not enable the Unix-only
   exhaustive runner without real descendant ownership and Windows tool/source
   scope. The cancelled Linux race root hit its 75-minute job budget; the new
   sixteen-part schedule still needs current-head CI confirmation.
4. The final macOS stdlib job reproduces two scan parse errors at unresolved
   `PUSH_REGS_HOST_TO_ABI0()` in `runtime/asm_amd64.s` and
   `runtime/sys_darwin_amd64.s`. `scanPackages` ignores includes and the token-aware
   opcode parser sees `()` as an operand. Do not restore an ignored/no-op macro
   path or bypass malformed operand validation. Real CPP expansion and its
   source context are separate from opcode coverage.
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
