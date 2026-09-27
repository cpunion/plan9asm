# Current work: PR 40 CI repair

Read [instruction development](instructions.md), [validation](validation.md)
and [report provenance](discovery-verification.md). Fix CI before another
index scan. Older detailed checkpoints and instruction-family notes remain
in Git history; red/green logs remain in ignored `_out/`.

## Authority and current CI

- PR 40 is OPEN/DRAFT. Push only to
  `cpunion:codex/expand-ecosystem-corpus-20260913`. Never push to `origin`
  or `xgo-dev`, merge/approve/close the upstream PR, or mutate its settings/runs.
- Pushed head is `b93666f5`, including upstream `main@7cc8c0f`.
  CI run `36311131855` is still active. **The user requires this run to finish
  once: do not push, cancel or rerun it while active.** Inspect live jobs first;
  intermediate failures are not its final result.
- A read-only watcher writes `_out/ci-b936-watch.log` in the instruction tree.
  The PR body has the audited scan/assembly funnel and CI counts.
- Inspect status, worktrees, remotes and processes before editing. Never change
  a frozen corpus runner, its tools or scan ledger during verification.

## Worktree roles and evidence

Identify persistent trees by branch, not by machine-specific paths:

- `codex/pr40-arm64-raw-20260926`: main instruction development. At
  `66ef3967`, semantic source equals `2265cb14`; later evidence commits only
  update the derived assembly ledger. Preserve the SIMD snapshot before
  integrating the diagnostic fixes below.
- `codex/pr40-ci-hotfix-20260927`: frozen at `2265cb14`. Reports are in
  `_out/simd-replay-2265cb14/`, with sibling `*-shard15.log` and
  `*-shard17.log`. Shard 15 passed; shard 17 is finishing at this checkpoint.
  Both exact SIMD module candidates passed. Do not modify this tree while any
  verification process remains active.
- `codex/pr40-network-classify-20260927`: new fixes `0674b925`,
  `855675f2`, `abb64a5d`. After committing this checkpoint and rebinding the
  pending ledger, freeze it for a complete shard 21 replay. Its source differs
  from the SIMD checkpoint; old reports cannot become new-source passes.
- `codex/pr40-ci-evidence-b93666f5`: exact pushed-source CI artifacts in
  `_out/ci-b936-reports`. Evidence `1965527d` retains 21 complete reports:
  13 passing and eight failing shards. Complete/verified remain false.

Historical evidence, including `b434d36e` and `16c4685`, is not disposable.
Missing reports remain pending. Publish through the validated updater, never
hand-edit pass flags or combine different provenance.

## Verified instruction batch (not pushed)

The local ARM64 batch fixes SIMD pool analysis without weakening unknown-
instruction, address-observation, unsigned-wrap or bounds checks:

- Typed contiguous SVE load footprints across all 16 architectural VLs.
- Full indexed ZDUP forms within the first 512 bits, including Q lanes;
  out-of-current-VL results are zero, not LLVM poison.
- Independent NZCV/GP effects for broadcast, CPY/MOV, counts and scalable
  address instructions, using their typed decoders.
- Constant CMP/CMN bounds, AND/BIC affine relations, exact masked intervals,
  NEG aliases, power-of-two residues and integral carried strides.
- Certified single-iteration/sequential loops, unchanged relational guards,
  masked-difference guards and aligned disjoint ORR/EOR relations.

Later commits include `7c07c82b` (sequential loops), `0186a071` (scalable
effects), and `c433e78b` (aligned logical relations). The preceding checkpoint
describes earlier family commits. Retain whole-family and negative tests.
Old-source overlays become stale when their referenced tree moves; preserve
red logs and reconstruct old implementations from Git before replaying them.

At frozen source `2265cb14`:

- Full root suite and all root subpackages pass in 1,059.348 seconds:
  `_out/root-tests-simd-2265cb14.log` in the hotfix tree.
- Both nested CLI suites pass: `_out/simd-cli-plan9asm.log` and
  `_out/simd-cli-plan9asmll.log`.
- Full required Linux/QEMU runtime selection passes in 636.849 seconds:
  `_out/cross-runtime-2265cb14.log` in the instruction tree.
- ARM64 standard-library matrix passes 21 configurations / 950 IR objects;
  the ARM64 Plan 9 corpus gate passes. Logs use `_out/simd-arm64-*`.
- Latest strict benchmark: 184/184 files, zero N/A, 23 target-seconds plus
  two build-seconds. The official five-arch observed-form gate passes. This
  does not mean every decoder encoding has runtime coverage.
- The entire SIMD SVE byte file compiles for Linux, Darwin and Windows.
  Its SVE2 parse/format functions pass 3,120 scenarios over all 16 VLs:
  `_out/simd-aligned.json` and `_out/simd-sve-oracle-run.log`.
  The shared harness prints an old NEON label internally; wrappers and final
  output identify SVE2. Whole-module replays pass both
  `github.com/sebishogun/simd@v1.21.1` and its mirror, each with 120
  applicable files, six targets and 360 translations, zero target N/A.

Use actual Go 1.27.1 in PATH for external jobs and LLVM 22 only. The existing
Linux cross container uses Go 1.27.0 for root runtime oracles and pinned QEMU
10.2.3. Missing tools fail. Root compatibility remains Go 1.20.

## New corpus-diagnostic fixes (not pushed)

- `0674b925`: exclude only exact matching Go assembler EOF source lines,
  never the whole mixed log. Network/resource/toolchain errors remain failures.
  Checksum mismatches, cancellation, missing executables, permission/resource
  failures and OOM cannot become source N/A. Terminal failures take precedence
  over concurrent transient network diagnostics.
- `855675f2`: progress collection and final verification reject N/A evidence
  containing infrastructure failures before ledger publication.
- `abb64a5d`: recognize HTTP request `EOF` as well as `unexpected EOF`.
  CI shard 21 lost two downloads to plain EOF from the checksum-proxy endpoint.
  A real local HTTP server reproduces response closure: old code makes one
  attempt; the fix recovers or fails after at most three. Production checksum
  verification is not disabled or bypassed.

Red logs, full corpus tests, race, Go 1.20, vet and Windows test-object
compilation use `_out/mixed-diagnostics-*` and `_out/proxy-eof-*` in the
diagnostic tree. Final full corpus suite passes in 23.050 seconds, race in
54.446 seconds; all five diagnostic functions have 100% statement coverage.
That is not current-head Codecov approval. A diagnostic-only audit checked
11,020 existing N/A reasons without finding the mixed-error pattern; it does
not promote their old source/tool provenance.

## Remaining failures and next actions

1. Let CI finish once; download and audit new artifacts in the exact-source
   evidence tree. Keep Draft. Verify the shard 21 EOF fix with a full replay.
2. Finish shard 17; publish both SIMD reports using
   `scripts/update-assembly-ledger.sh` in a same-source evidence/development
   tree. An in-progress shard is not passed.
3. Preserve the SIMD snapshot, then integrate the three diagnostic source
   commits and this checkpoint into the main development branch. Refresh its
   ledger for the new semantic fingerprint and freeze the final source for
   verification. Documentation changes also invalidate old fingerprints;
   derived assembly-ledger changes alone do not.
4. Native-layout/JIT failures remain real: GopherJRE takes an out-of-group
   RIP-relative continuation and exchanges a custom register/stack contract;
   GoJIT scans bytes for `0xDEADBE00` and enters code after it; Sharkie switches
   stacks and fabricates code-relative return addresses. Widening raw branch
   bounds or dropping a marker is not compatibility. No generic native-layout
   exception is authorized. A user question about a separately counted exact
   exception is pending; absent approval, retain failures and implement a
   compatible mechanism.
5. Both case-distinct ContainerFS paths at the 2019 version fail on obsolete
   gVisor dependencies, with 503/429 mixed into 404s. Proxy latest resolves to
   the same version. No newer same-project replacement is verified. Do not
   disguise transport errors as source N/A or invent a supersession.
6. Only after the run finishes and relevant failures are repaired, batch-push
   the allowed fork and inspect CI, review and patch coverage. Distinguish
   compilation, runtime execution and inventory claims.

The remote inventory was not refreshed during CI repair. Scan records,
assembly evidence and the separately stored cgo inventory remain distinct.
Reports/binaries stay ignored. Remove only owned temporary downloads/objects,
never shared module caches, unrelated containers or frozen evidence.
