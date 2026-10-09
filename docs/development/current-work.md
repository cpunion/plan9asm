# Current work: PR 40 verification repairs

Read [validation](validation.md), [instruction families](instructions.md), and
[discovery verification](discovery-verification.md). This checkpoint is a
handoff, not a completion certificate. Inspect the actual frozen HEAD, complete
raw reports, inventories and terminal owners before accepting a result.

## Contribution boundary

Stage repairs in fork Draft [PR 4](https://github.com/cpunion/plan9asm/pull/4),
targeting `codex/expand-ecosystem-corpus-20260913`. Push only to cpunion.
Do not update the upstream-connected head until fork current-head CI and review
pass; never push to or merge upstream [PR 40](https://github.com/xgo-dev/plan9asm/pull/40).
Root alone publishes assembly-ledger updates.

Published `c09eff1e`, run `37711752851`, finished failed: four priority shards
failed and two reached their deadlines. Downstream skips and failed aggregate
checks are not passes. Batch verified repairs before another push.

## Frozen evidence and development branch

The complete ordinary module run at `2ae09eb7` naturally closed **failed**.
All eight root parts, corpus/rest owners and aggregate finished; no timeout or
cleanup failure obscures its source/SDK/tool-stable result. Its 54 failed root
tests, 34 failed corpus-driver roots and four failed gotoolprofile roots remain
regressions. Do not overwrite this checkout or its original reports.

The next integration branch is `codex/pr40-ci-family-successor-20261009`, based
on `cddbc85e`. Latest upstream main `e2164ac0` is already its ancestor. This is
still development integration, not a verified combined head. Keep component
worktrees frozen; inspect actual commits before starting gates. Earlier passes
establish only their exact component revisions:

- `4617d12d`: complete explicit IMUL source/caller family, zero-operand
  diagnostics, actual ARM64 lazy-tool routes and backend physical-source prefix.
- `8c60cde3`: complete x86 aliases, scalar positive-zero rewrites, frontend
  physical-line/label roles and mode32 operand-class diagnostic metadata.
- `e50f590a`: exact MOVL/Q address-to-LEAL/Q diagnostic joins, including I386's
  distinct EXTERN/STATIC role. Three affected roots pass; no ISA/ABI permission
  is obtained from a diagnostic spelling.
- `7b05cb4c`: an explicit trailing backend-diagnostic route binds complete
  ARM64 TEXT attributes, frame/argument values and the exact unaligned-frame
  error. Original 34-node RED becomes 34-node GREEN. The complete ordinary
  oracle package passes separately; its default diagnostic route is unchanged.
- `e34c78c6`: shared target CPU command-environment construction fixes foreign
  inherited CPU keys without sanitizing output or weakening profile validation.
  Seven focused roots and the original four instrumentation roots pass.
  The separate 34-root finite rerun ends with 33 passes and one incomplete
  generated-header root. Its six-minute per-root deadline is not the original
  fixture's fifteen-minute or full module's forty-five-minute budget. Preserve
  that failed scope without calling it a semantic failure or a full-suite pass.
- `8061bd59`: explicit wasm unknown-word lexical roles preserve complete
  original statement/suffix binding. Component ordinary/race/Go-1.20 oracle
  scopes pass; this is not wasm runtime or LLVM object coverage.
- `348e27b9`: ARM64 B/BL aliases, canonical P/W suffixes,
  complete historical four-class/current five-class directories, and actual
  selected-SDK bare From versus register/special-operand lexical roles. A bare
  From symbol prints zero offset; it is not a resolved label-PC permission.
  Complete ordinary/race/actual Go-1.20 component scopes pass.
- `dc2e4918`: private x86 whole128 observer joins exact same-Void-source/private
  caller/normalized-ordinal read16 evidence. Its complete 26-root scope passes
  ordinary/race/actual Go-1.20, and shared aliases pass ordinary/race separately.
  Public/direct/nonVoid, aligned, scalar/half, native-store and unknown-extent
  boundaries remain unchanged. Only its six reviewed production/fixture
  commits are integrated here; shared oracle prerequisites were already present.
- `69253f0e`: complete x86 alias operand joins retain Go's TYPE_FCONST value,
  canonical seventeen-digit printing and signed zero. The complete shared
  16-root scope passes ordinary/race (498 nodes each); actual Go-1.20 selects
  11 roots and passes 450 nodes. Its five build-tag exclusions are not skips.
  Actual logged Go commands number 151/151/108, all independently closed.
  No instruction, source N/A or ABI permission is added.
- `fd1b9dc9`: a literal NOSPLIT TEXT and complete straight-line deltasp join
  binds I386 PARAM += delta+4 and AUTO += delta to the original operands,
  aliases and full backend class tails. Ordinary/race pass 134 nodes each;
  actual Go-1.20 selects three portable roots and passes 84 nodes. Unknown
  flow, CPP, SP writes, symbolic frames and overflow remain rejected. Its
  ordered test/fix commits are integrated as `4fd28774` and `07a1d672`.

The combined rerun at `cddbc85e` naturally ends with 52 of the original 54
failed roots passing. Two roots still fail on four diagnostic-join children:
alias FLOAT formatting and mode32 FP rebasing. Retain that failure. A new
combined gate at `07a1d672` passes all original 54 roots plus seven new roots:
61 roots and 2,401 children, 2,462 pass with zero fail/skip/incomplete. Main's
independent fresh readback closes full raw EOF, source/tools/binary and all
144 logged actual-Go children. This is a finite scope, not full-module or CI.

Generated-header planning now has a conservative condition-free fast path.
Its query-only RED at `bf8ae5b2` retains 113 passes and 13 failures; the
successor passes all 126 nodes. Full source/include/header guards, CPU-token
fallback, Go-only partners and instrumentation expanded-output scopes remain.
The actual consumer at `1d9f6b99` passes 27 Builtin nodes across eleven targets;
seven separate query-only controls pass on Linux/amd64. Planning queries fall from
123 to 11; all eleven final headers and LLVM objects remain required. Main's
fresh readback confirms the complete reports and frozen inputs. Initial full
Go-list exports and physical temporary LLVM objects were not retained: do not
claim a fresh physical-object audit or external-library success. The earlier
canonical-target helper failure remains preserved. The five reviewed commits
are integrated through `83d0274b`, not yet whole-module verified.
The actual Go-1.20 query-model check at `1d9f6b99` also passes all 126 nodes,
including Main's fresh readback. Its earlier wrapper setup failure remains
preserved with all tests unattempted; it was not a CPP failure. The corrected
reader binds complete command-package build metadata and Go-1.20's actual
worktree/VCS rules, rather than granting a generic missing-VCS exception.

Successful numerical-consumer fixtures now report actual source Close errors
instead of discarding them. The source-bound RED at `6f22c04a` exposes the
directory-identity error; `772d2504` passes ordinary and race, twelve nodes each,
including the preserved replacement-directory canary. Main independently
rechecks both complete inventories. These test-only commits are integrated as
`c308c511` and `e4e22527`; they confer no source/ABI permissions or ledger pass.

ARM64 MOV memory-row coverage now compares the selected Go table's complete
attributes and physical order, exact unsigned aliases and independent typed
shapes. Its same-count mutation RED has one pass and seven expected failures.
The test-only successor at `4328896d` passes 306 ordinary and 306 race nodes;
actual Go-1.20 passes 261 nodes using its ten-field/twenty-row schema, not skips.
Main independently rechecks all three complete inventories and frozen inputs.
The ordered commits are integrated as `84e923a6` and `a14f9f33`. This projection
does not prove overlapping encoder-row selection, LLVM objects or ABI/runtime
behavior; the inner Go-env command has no independent command-receipt proof.

## Next work, in order

1. Freeze the integrated source after inspecting its actual clean HEAD.
   Component passes never transfer to the new combined HEAD. Preserve every
   original failure and keep the component source, tools and reports frozen.
2. Run all original failed roots, complete
   affected families, the exhaustive original two-worker module plan, nested
   CLIs, vet/build, official encoder/decoder, strict SDK/benchmark and required
   native CI. Let all owners finish, including failures; recheck complete raw
   EOF, inventories, source/tools/binaries and owned processes independently.
3. Repair remaining failures in separate persistent successors. Do not rewrite
   a running snapshot or turn incomplete/killed scopes into success.
4. Continue the SDK numerical source/object seam, dynamic reflect frames,
   native entry/TLS/continuations and full result/body contracts by typed rule
   class. The old go-fuse LoadUint16 failure remains a real ARM64 result-tail
   backing gap: parameter-padding and tail-call proofs do not authorize its
   wider result store. Never clip the store or invent a padding slot. Then
   run discovering libraries and all 64 external shards on the
   designated validators and publish only matching complete schema-11 reports.

The strict five-target SDK baseline still attempts all 184 files: 167 pass and
17 fail, with zero skip/N/A/unattempted. Official five-arch classification has
zero unsupported/parse errors but includes Context forms; classification is
not complete LLVM, ABI or runtime success. SDK numerical capture remains
query-only: its genuine CRC consumer failure is not integrated or promoted.

## Resource and ledger rules

No local background index/discovery/external corpus. Inventory stays on the
designated source-only scanner; external compilation/execution belongs on
separate disposable validators. Local first-party fixtures and corpus-driver
unit tests remain allowed. cgo inventory stays outside this repository.

Use task-owned caches for new verification. A selected shared lazy tool vanished
and disk exhaustion caused real reader/setup failures; retain those failures.
Check available disk before compilation. Do not delete selected tools or cache
entries until every consumer and independent reader closes. A prior cache size
or manifest does not authenticate its current bytes. Dispose LLVM modules before
Contexts and serialize heavy full-suite/container gates when resources are tight.

The committed inventory reaches the requested cutoff, but newer remote records
are not synchronized. Inspect its actual writer, manifest and journal before
importing; never duplicate a writer or replace its live binary. Status, lookup
and publication stream sharded records rather than loading the ledger payload.

The assembly ledger is incomplete and unverified. Only automatic publication
of complete, matching source/scan/tool/file/target/profile/consumer proofs may
promote entries. First-party passes, Context and provisional exclusions are not
module passes. Keep scan/coverage funnel tables in the PR body, not this guide.
