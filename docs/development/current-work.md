# PR40 closeout

Updated 2026-10-27. Finish **amd64 correctness, shared typed families and
translation efficiency first**, preserving all five architecture gates.
Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). Detailed historical receipts remain in git
history; this page records the current candidate and its limits.

## Publication and CI

PR40 remains draft. The next publication carries integrated code `7f7d3bfbd`;
complete current-head CI, coverage and external approval remain pending.
The preceding [round at `57ebe257d`](https://github.com/xgo-dev/plan9asm/actions/runs/40446787035)
failed complete chacha translation in the rich CLI on Linux and Windows,
with dependent completion failures. The repair now produces an actual object.
The old failed round is not relabeled successful.

Publication `55831306c` previously passed 325 CI jobs with 88.0% main statement
coverage. Those results do not certify newer code. Both inline review threads
are resolved; no external reviewer has approved. The stale official-corpus
table, README dependency description and variadic internal ABI-mode parameter
noted in review are corrected.

[LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694), head `cdca51fc4`, still
pins `v0.0.0-20261026211521-55831306c713`. Its current CI is unfinished.
The documentation-link and native-command timeout repairs passed focused
checks; a host timeout does not establish that a guest runtime hang is fixed.
Update the dependency after qualifying this publication. Official upstream
dependency provenance remains unresolved until an upstream version is available.
Do not merge or push upstream directly.

## Integrated implementation

- Native machine bits and dereference permissions are separate within existing
  typed effects. Arithmetic can consume defined native addresses; integer results
  do not gain pointer authority. Private relocated addresses, readonly storage,
  undefined bytes and source lifetime retain their checks.
- Observed private automatic frames reuse the existing call CFG, save/restore BP
  facts and invalidate stack-adjustment FLAGS. Helpers cannot observe SP/FP or
  nest calls. Exact native frame identity and Windows SEH remain required.
- Fixed-local comparison uses existing packed EQ/GT byte effects for 16 named
  B/W/L-or-D/Q forms with legacy/VEX alias rules: 12 added production lines.
  MMX and K-result storage retain separate owners.
- GP/X transfers reuse one existing grammar and defined-byte projection for
  MOVL, MOVQ, MOVD, MOVDQ2Q, VMOVD and VMOVQ: 70 added production lines,
  no new opcode registry or proof state. Go MOVD is a 64-bit MOVQ alias;
  VMOVD is 32-bit.
- Byte-definedness uses a bitset with unchanged frame bounds and resource
  limits. Frame-clone allocation falls from 3,504 to 1,040 bytes for a
  1,408-byte frame and 131,760 to 17,072 bytes for 64KiB. Complete replay
  times vary: linear -1.96%, diamond +4.73%, loop +0.03%. These are not
  whole-translation speedup claims.

The four-family packed normalization pilot removes 65 production lines; the
shared feature projection removes 11 more. A full-form function cache was
measured and rejected: more retained memory, higher allocation bytes in 18/20
cases and slower short ABBA results. It is not integrated. Continue sharing
semantic rules without an expensive cross-stage cache or another interpreter.

## Qualification

| Scope | Result | Boundary |
|---|---|---|
| Integrated Go1.27.3 / LLVM22 at `c78f24658` | 142 roots / 11,548 children PASS; zero failures/skips | Byte-state, machine-bit, private-frame, compare and GP/X union |
| [Linux execution at `c78f24658`](https://github.com/cpunion/plan9asm/actions/runs/40459995914) | 10 roots / 117 children PASS; artifacts independently audited | Original Go and three LLVM APIs at O0/O2 where specified |
| SDK assertion repair at `7f7d3bfbd` | 7 roots / 50 children PASS | Only two test files differ from `c78f24658`; production and native execution logic unchanged |
| Final Go1.27.2 CLI checks | Small CLI 2 roots / 2 children; rich CLI 9 / 8; all PASS | Both former chacha failures now check actual AMD64 ELF |
| Complete original Linux/amd64 SDK objects | **40/55**, 15 failures, zero N/A/unattempted | Same 55 Go1.27.2 source hashes; all previous 39 passes retained; P256 is new |
| Five-architecture official-corpus gate | PASS at `7f7d3bfbd`, no baseline mismatch or parse errors | Classification is not initialized-form or runtime coverage |
| Isolated minimum-SDK checks | Bitset 39 / 2,869; GP/X 28 / 1,502; compare 59 / 6,980; all PASS | Constituent snapshots, not a combined final minimum-SDK run |

The full SDK command correctly exits nonzero for 15 failures. Its 97-second
duration including driver build is one observation, not a speedup claim.
Chacha20poly1305 became an object pass in the previous 39/55 run. Complete P256
retains all 14 TEXTs, emitting ten public definitions and incorporating four
private helpers through the observed CFG. Whole-file execution is not claimed.

The obsolete blanket rejection of nonzero private helpers was separately
reproduced. Its original P256 input is consolidated into
`TestGoAMD64SDKFramedLeafWholeOriginalSource`, preserving complete TEXT
inventory and six native-frame mismatch controls. Private/native role conflicts,
uninitialized bytes, source mutation and missing SEH still reject. Failed
receipts remain failed; no skip or cropped input substitutes for success.

The native scope includes 253 machine-bit programs, 48 pointer programs,
24 integer-compare kernels and six GP/X kernels. GP/X observes 96 original Go
runs and 576 LLVM comparisons. Memory/view checks compare LLVM with C, not
original AVX512; physical-frame FLAGS checks are original-Go only. Native AVX512
remains pending hardware. Later SDK test edits are not another native receipt.

## Independent official rows

| Family group | Official rows | Concrete witnesses |
|---|---:|---:|
| Add/subtract and D/Q low product | 159 | 480 |
| Integer min/max | 132 | 432 |
| Average, products and multiply-add | 90 | 258 |
| Integer and PS/PD logical | 132 | 504 |
| Unpack | 108 | 360 |
| Saturating pack | 36 | 114 |
| Per-lane shifts | 64 | 236 |
| Conditional sign | 9 | 18 |
| Byte-indexed shuffle | 9 | 24 |
| Integer EQ/GT compare | 72 | 168 |
| Mask blend | 36 | 144 |
| Total | **847** | **2,738** |

These joins retain independent ordered Go rows and actual Go bytes, then compile
initialized functions through three APIs with LLVM22. High banks force narrow
EVEX; two PADDD alias witnesses are separate. MMX has independent region/EMMS
owners. This is neither all x86 rows nor an architecture percentage. Context
and zero unsupported classifications do not establish successful translation.

## Remaining work, in order

1. Publish the qualified batch through the allowed fork after the previous CI
   round finishes. Watch complete current-head CI and repair failures. Update
   LLGo's exact dependency and obtain review; retain draft meanwhile.
2. Restore ordinary SHA1 with bounded shared private-layout analysis, reusing
   the CFG and typed effects. Do not add an opcode interpreter, library whitelist,
   zero-initialization assumption or a larger budget.
3. Keep four implicit-entry, two GC-scan and eight system/runtime/test-hook
   failures individually diagnosed in [issue47](https://github.com/xgo-dev/plan9asm/issues/47).
   They are not automatically exclusions. Object success does not prove runtime ABI.
4. Finish amd64 row ownership, 386 applicability, semantics and measured
   simplification before extending the migration. Preserve all other arch gates.
   The 271 ARM64 decoder forms, remaining WASM control/SIMD and native AVX512
   acceptance remain explicit gaps.

External discovery, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run inventory
locally or restore those jobs here. Keep raw failed receipts under ignored
`_out/`, retaining their original revision and outcome.
