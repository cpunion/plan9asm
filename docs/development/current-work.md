# PR40 closeout

Updated 2026-10-27. Finish **amd64 correctness, shared typed families and
translation efficiency first**. Retain all five architectures' regression
gates. Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). This page distinguishes published evidence
from local work; neither is a complete architecture coverage certificate.

## Publication and CI

PR40 remains draft. The new publication carries integrated code `cb23f48a5`
and the qualification recorded below. Its complete CI round is pending;
the previous green round does not certify it. Previous publication `55831306c`
[finished its complete CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40402863391):
**325 jobs, all successful**, main statement coverage **88.0%**. The round
finished naturally, without cancellation or reruns. Its cross-runtime
completion artifact reports a passing aggregate with 16 owners and 562 selected
roots. Portable oracles in that scope do not establish native AVX512 acceptance.

Both review threads are resolved; no reviewer has approved the contribution.
The current round at `57ebe257d` has failed the rich CLI's complete chacha
checks on Linux and Windows; completion jobs consequently fail too. The local
repair below produces an actual object. Let this round finish before publishing
the next batch; do not reuse the earlier green round as its certificate.
The previous green run at `5406c2756` and earlier failed receipts remain
historical evidence. They must not certify subsequent local changes.

[LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694) now has published head
`cdca51fc4`, still pinning the exact public Plan9asm fork version
`v0.0.0-20261026211521-55831306c713`. Its new CI round is incomplete.
The preceding round finished with 127 successful jobs, two failures,
one cancellation and two release-condition skips. The new batch repairs:

- Four moved README library links: the actual link checker passed all 102 links.
- Native test execution ignoring the existing host-side RunnerTimeout:
  focused RED failures are retained; the repair passes 47 roots / 64 children
  and all 22 workflow policy tests. Windows ARM64 MinGW core tests now stream
  JSON diagnostics without changing the package inventory or timeout.
- Fedora is an existing installer shell/container test. Its preceding failure
  reached the shell assertions, then blocked in a kernel D-state flush on exit.
  No translator target, skip or enlarged timeout was introduced for it.

The timeout repair bounds the host command; it does not establish that an
underlying guest runtime hang is fixed. Watch the fresh CI rather than treating
a local host-runner test as a Windows runtime certificate. The dependency
provenance review remains unresolved until the contribution can use an official
upstream Plan9asm version.

## Local candidate and evidence

This publication's integrated code is `cb23f48a5`. It adds pointer
choices, complete framed roots with observed zero-frame private leaves,
Windows leaf observation, broadcast boundaries, implicit MUL routing and
three initialized official-family joins. It also integrates the smaller
native-address domain described below, the four-family normalization pilot and
shared FP numerical reads. Current-SDK joint qualification passes **186 roots /
12,670 children**, zero failures/skips: 157 main-package roots plus all 29
rule-reader roots. The exact union includes the changed families, frame/FP
consumers, private-leaf boundaries and the complete 90-row / 258-witness
official average/product join. Fresh complete CI remains pending. Minimum-SDK
receipts belong to the individual batches, not a combined run at this head.

| Scope | Evidence | Limit |
|---|---|---|
| Pointer/shuffle integration at `a5354ca1a` | Go1.27.3: 58 roots / 1,051 children; Go1.20.14: 55 / 998; zero skips/failures | Earlier integrated subset, not every later local change |
| Observed private leaves | Current and minimum SDK partitions passed; source/ownership and budget controls retained | A preceding single-command timeout remains failed; partitioned passes are separate receipts |
| Windows zero-frame leaf observation | Current root/profile: 13 / 161 and 6 / 109; historical Go1.23.12: 4 / 67 and 4 / 68 | Exact observed leaf only; framed-root SEH requirements remain |
| Implicit MUL/IMUL fixed-local adapter | Go1.27.3: 38 / 4,086; Go1.20.14: 36 / 4,043; zero skips/failures | Native execution is separately qualified below |
| Native execution at `511029b0d` | [Linux run](https://github.com/cpunion/plan9asm/actions/runs/40434110395): 3 roots / 27 children, zero skips/failures; original Go and three APIs at O0/O2 | 42 pointer programs (1,344 observations/pipeline), 5 address/FLAGS programs (50), 36 MUL kernels plus 2 helpers (2,880); not whole crypto files or native AVX512 |
| Integrated runtime at `cb23f48a5` | [Linux run](https://github.com/cpunion/plan9asm/actions/runs/40441342032): 4 complete-function roots / 36 children plus 16 packed-family roots / 22 children, zero failures/skips | Includes 121 TypedGP kernels with 29 direct FP formats; packed-family owners distinguish original SSE/MMX/VEX execution from portable LLVM, not native AVX512 |
| Complete Linux/amd64 SDK objects at `511029b0d` | **38/55**, 17 failures, zero N/A/unattempted | Same 55 Go1.27.2 source hashes and exact same pass set as the previous 38/55; no regression or new pass |

The integrated native-address refactor removes the four-origin enumeration cap
and redundant dynamic DATA bounds analysis while retaining private/frame,
readonly-store, source-lifetime and known constant-bound checks. Its first
frozen related run completed 84 roots / 2,491 children: 82 roots passed,
two failed on six old rejection expectations. Those failures are retained.
After preserving the original source bodies, the corrected expectations and
independent address/FLAGS driver pass 6 roots / 94 children on the current SDK,
with zero skips.
The complete minimum-SDK scope subsequently passed 66 roots / 2,131 children,
with zero skips/failures. These are composite receipts, not a single successful
final current-SDK 84-root run. The new native owners passed on the integrated
source as recorded above. Production code shrinks by 134 lines. State clone
storage falls from 1,120/1,120/5,384 to 832/832/3,208 bytes for 1/8/32 entries;
allocation counts remain 11/11/13, and joins remain allocation-free.

Complete Go1.27.3 SHA1 was attempted with that refactor and still fails at
private CFG address identity. The complete 55-file rerun also identifies P256's
nonzero-frame private helper, chacha8's register-ABI entry and chacha20poly1305's
ordinary FP operand in the shared GP-effect consumer. Four diagnostics changed;
thirteen did not. These are separate format/binding gaps, not four new opcodes.
The target took 99 seconds versus the older single observation of 68 seconds;
neither proves a stable performance change, and earlier rejection is not a
speedup. Do not add a library-specific relational interpreter.

The subsequent FP batch reuses the existing typed-slot reader for 174 legal GP
input forms and the existing ADX decoder in original-frame observation. It adds
18 production lines, not a new per-opcode proof. Current/minimum scopes pass
46 roots / 1,640 and 1,628 children; private-leaf boundary controls pass 3 / 22.
The full original chacha source now reaches native address subtraction
(`SUBQ SI,BX`), which still lacks its address-to-number transfer rule. Do not
count it as a whole-file pass or increase 38/55 without a complete rerun.

## Subsequent local repairs

The following changes are integrated locally through `eb81c31ee`, not yet
certified by a complete CI round:

- `303e7c718` separates defined native machine bits from address capabilities
  in the existing typed effects. Arithmetic can consume native pointer bits;
  an integer result does not automatically become a dereferenceable pointer.
  Private relocated addresses, readonly storage, undefined bytes and source
  lifetime retain their checks. The change adds 73 production lines without
  another opcode registry or proof state.
- Its first wider current-SDK run retained 113 roots, with two failures on old
  rejection expectations. `9dfc82d6e` keeps those original source bodies and
  corrects the expectations, adding actual objects and independent arithmetic,
  AF, vector-spill and multiply oracles. The final affected scope passes
  8 roots / 364 children. These receipts jointly cover 118 distinct roots;
  they are not a single 118-root test command. The final minimum-SDK run is a
  complete 97 roots / 4,397 children, zero failures/skips.
- `5dd3ea254` uses the existing private-call CFG for observed automatic helper
  frames. It saves/restores BP facts and invalidates frame-adjustment FLAGS;
  helpers cannot observe SP/FP or nest calls. Source/native frame identity and
  Windows SEH remain required. The isolated Go1.27.3 scope passes 17 roots /
  226 children, zero failures/skips. Complete original P256 now reaches a
  missing packed-compare effect entry; it still fails as a whole file.
- The full original Go1.27.2 chacha source now produces an LLVM22 object.
  Its two CLI regressions check an actual AMD64 relocatable ELF object and
  unchanged original SDK bytes. The integrated CLI qualification at
  `eb81c31ee` passes 2 roots / 2 children in the small CLI and 9 / 8 in the
  rich CLI, with zero failures/skips. It includes the explicit ABI-mode
  parameter cleanup requested in review. The documentation snapshot and
  dependency descriptions are also corrected.

New native execution remains pending. The prepared Linux scope includes the
253 machine-bit functions, 48 pointer programs and framed-helper oracles;
object compilation alone does not qualify their execution. Re-run the complete
55-file original inventory before changing its last recorded 38/55 result.

## Official rows and implementation

Each official join retains original ordered encoding rows and compares actual
Go bytes before three-API LLVM22 object compilation. High register banks force
narrow EVEX encodings. These are initialized object checks, not universal
runtime coverage or an architecture percentage.

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
| Local total | **847** | **2,738** |

Previous publication `55831306c` contains the first 730 rows / 2,402 witnesses.
This batch adds the final three joins, locally qualified on both endpoint SDKs.
Two PADDD alias witnesses are separate. MMX has independent
region/EMMS owners. Complete x86 row ownership, 386 applicability and raw/native
coverage remain unfinished.

The existing packed E4 pipeline already shares operand loading, mask handling
and writeback. Earlier measured parser refactors removed allocations and
production lines; those measurements do not establish whole-source speed.
The four-family pilot shares one typed spec/form across effects, memory-shape,
result routing and lowering, removing the replaced selectors. Its lowerers
consume the already parsed form; other analysis invocations still parse
independently. This is not whole-translation parse-once completion. Production
code shrinks by 65 lines while preserving all existing tests. Each endpoint SDK
passes 101 roots / 10,098 children; five related fallback roots add 728 children.
Adjacent ABBA measurements across all 20 names average about +1.1%, with large
per-pair variation under unrelated host load. Allocation gates pass, but neither
acceleration nor a precise performance bound has been established. Preserve
the earlier slower short measurement rather than discarding it.

The next experiment retained full forms in a function-local table for feature
inference, register discovery and lowering. It was rejected before integration:
the prototype added 73 production lines, retained about 488 bytes per matching
instruction, and increased allocation bytes in 18/20 cases. Its short adjacent
ABBA comparison was about 5.7% slower overall; this is evidence against that
prototype, not a universal performance bound. The patch and receipts remain
under ignored `_out/`, outside commit history.

Instead, local `ae2bda686` shares the existing four-family feature projection,
removing 11 production lines without another type, cache or lifetime. Current
and minimum SDKs each pass 19 roots / 122 children, with zero skips/failures,
on the isolated source-identical Go code. Short complete-translation benchmarks
show no material allocation increase; no speedup is claimed. Lifecycle and
error-precedence regressions remain. Future cross-stage reuse must justify its
retained data and complexity; it is not a prerequisite for sharing semantic
rules. See [architecture](architecture.md).

## Remaining work, in order

1. Watch LLGo's current CI and repair genuine failures. Qualify the complete
   local Plan9asm batch before publication; retain draft until its CI,
   coverage and external review pass.
2. Finish the first complete packed binary normalization group, with unchanged
   legal forms and meaningful rejection controls, independent runtime oracles,
   production line delta and separate parsing/whole-translation measurements.
3. Restore ordinary-source SHA1 and P256 through shared rules; qualify the
   new chacha20poly1305 object success in the complete inventory.
   Keep real ABI/layout requirements separate from extra source-safety analysis.
   Re-run the complete original SDK inventory before updating 38/55.
4. Keep the four implicit-entry cases, two GC-scan cases and eight
   system/runtime/test-hook failures separately diagnosed. These are not
   automatically all runtime exclusions. Integration continues in
   [issue47](https://github.com/xgo-dev/plan9asm/issues/47).
5. Complete remaining amd64 official-row joins and semantic validation before
   extending the architecture migration. Preserve ARM, ARM64, 386 and WASM
   gates; 271 ARM64 decoder forms, remaining WASM control/SIMD and native
   AVX512 hardware acceptance remain explicit gaps.
6. Publish the qualified batch through the allowed fork, update LLGo's exact
   dependency pin, and obtain review. Do not directly merge or push upstream.

External discovery, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run its
inventory locally or restore those jobs to translator CI. Preserve raw failed
receipts under ignored `_out/`; earlier detailed handoffs remain in git history.
