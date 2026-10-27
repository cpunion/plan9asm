# PR40 closeout

Updated 2026-10-27. Finish **amd64 correctness, shared typed families and
translation efficiency first**. Retain all five architectures' regression
gates. Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). This page distinguishes published evidence
from local work; neither is a complete architecture coverage certificate.

## Publication and CI

PR40 remains draft. Published head `55831306c`
[finished its complete CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40402863391):
**325 jobs, all successful**, main statement coverage **88.0%**. The round
finished naturally, without cancellation or reruns. Its cross-runtime
completion artifact reports a passing aggregate with 16 owners and 562 selected
roots. Portable oracles in that scope do not establish native AVX512 acceptance.

Both review threads are resolved; no reviewer has approved the contribution.
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

Current integrated code is `2c19ef128`, not yet published. It adds pointer
choices, complete framed roots with observed zero-frame private leaves,
Windows leaf observation, broadcast boundaries, implicit MUL routing and
three initialized official-family joins. The whole candidate still needs
joint qualification and fresh complete CI.

| Scope | Evidence | Limit |
|---|---|---|
| Pointer/shuffle integration at `a5354ca1a` | Go1.27.3: 58 roots / 1,051 children; Go1.20.14: 55 / 998; zero skips/failures | Earlier integrated subset, not every later local change |
| Observed private leaves | Current and minimum SDK partitions passed; source/ownership and budget controls retained | A preceding single-command timeout remains failed; partitioned passes are separate receipts |
| Windows zero-frame leaf observation | Current root/profile: 13 / 161 and 6 / 109; historical Go1.23.12: 4 / 67 and 4 / 68 | Exact observed leaf only; framed-root SEH requirements remain |
| Implicit MUL/IMUL fixed-local adapter | Go1.27.3: 38 / 4,086; Go1.20.14: 36 / 4,043; zero skips/failures | Native owner is prepared but not executed |
| Conditional-pointer native owner at `6d0f167bc` | [Linux run](https://github.com/cpunion/plan9asm/actions/runs/40407417050): 23 programs, 736 observations per pipeline, six LLVM pipelines | Certifies that older source only, not the expanded domain or complete SHA1 |
| Complete Linux/amd64 SDK objects | Latest complete Go1.27.2 inventory remains **38/55**, 17 failures, zero N/A | No later single-file attempt changes this denominator |

An isolated native-address refactor removes the four-origin enumeration cap
and redundant dynamic DATA bounds analysis while retaining private/frame,
readonly-store, source-lifetime and known constant-bound checks. Its first
frozen related run completed 84 roots / 2,491 children: 82 roots passed,
two failed on six old rejection expectations. Those failures are retained.
After preserving the original source bodies, the corrected expectations and
independent address/FLAGS driver pass 6 roots / 94 children on the current SDK,
with zero skips. The full minimum-SDK qualification remains in progress.
The expanded 42-program and new FLAGS native owners have not executed.

Complete Go1.27.3 SHA1 was attempted with that refactor and still fails at
private CFG address identity. Clearing its earlier DATA-extent refusal is not
a whole-file pass. Do not add a library-specific relational interpreter.

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

Published `55831306c` contains the first 730 rows / 2,402 witnesses.
The final three joins are locally qualified on both endpoint SDKs; publication
is pending. Two PADDD alias witnesses are separate. MMX has independent
region/EMMS owners. Complete x86 row ownership, 386 applicability and raw/native
coverage remain unfinished.

The existing packed E4 pipeline already shares operand loading, mask handling
and writeback. Earlier measured parser refactors removed allocations and
production lines; those measurements do not establish whole-source speed.
The new architectural focus is to retain a normalized family form through
effects, memory-shape queries and lowering, removing the replaced selectors.
Do not add another dispatch table or just cache the same duplicated decisions.
See the normalization boundary in [architecture](architecture.md).

## Remaining work, in order

1. Watch LLGo's current CI and repair genuine failures. Qualify the complete
   local Plan9asm batch before publication; retain draft until its CI,
   coverage and external review pass.
2. Finish the first complete packed binary normalization group, with unchanged
   legal forms and meaningful rejection controls, independent runtime oracles,
   production line delta and separate parsing/whole-translation measurements.
3. Restore ordinary-source SHA1, P256 and chacha20poly1305 through shared rules.
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
