# PR40 closeout

Updated 2026-10-27. Finish **amd64 correctness, shared typed families and
translation efficiency first**, preserving all five architecture gates.
Read [architecture](architecture.md), [x86 convergence](x86-convergence.md)
and [validation](validation.md). Historical receipts remain in git history;
this page records the current candidate and its limits.

## Publication and review

PR40 remains draft. Published `1c54c530e` carries the qualified translation
repairs below; its [complete CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40505530452)
is running. Previous `ab68a6ecc` completed its
[full CI round](https://github.com/xgo-dev/plan9asm/actions/runs/40477411915)
naturally: **325/325 successful jobs**, no failed/cancelled/skipped jobs and
**88.0%** main statement coverage. This qualifies that revision, not the next
publication. Both inline threads are resolved; external approval is outstanding.

The next batch adds ABS and FMA3 official-row joins without changing translator
production code or native fixture inventories. Integrated `6ca5e4a9d` is
byte-identical to qualified `d21fb3c8f` except two documentation files. Its
fresh publication preflight and five-architecture gate pass. Batch the next
publication after the current CI finishes; full new-head CI remains required.
The failed `004da9f44` cache-policy round and earlier failures retain their
original outcomes. Its stale Node assertion was repaired without removing
the actual Go lifecycle, ownership or cleanup checks.

[LLGo PR2694](https://github.com/xgo-dev/llgo/pull/2694), head `9ccc923f8`, is
rebased onto `854634f22` and pins the downloaded public
`v0.0.0-20261027123924-1c54c530ed3d`. Adapter tests pass 41 roots / 131 children,
assembly-build checks 21 / 23, and complete CI policies 160/160, with zero
failures/skips. Rebase layout/initializer controls also pass. Current-head cloud
CI is running. A preparation timeout occurred before any steps and was retained;
rerunning that preparation and its dependents restored the Go workflow.
Previous `cdca51fc4` completed 13 workflows / 130 jobs with two release-condition
job skips; those do not qualify this pin. Its P1 dependency-provenance requirement remains
open until an official upstream version replaces the temporary fork pin.
Never push or merge upstream directly.

## Current implementation

- Native machine bits and dereference permissions remain separate in existing
  typed effects. Defined addresses can participate in arithmetic; numerical
  results gain no pointer extent, frame role or source lifetime.
- Private automatic frames use the existing CFG. Ordinary analysis runs first;
  one retry partitions compatible private-offset layouts using the remaining
  budget. Role keys, SP, saved BP, all eight SP residues, ordered accesses and
  final independent replay must agree. Limits remain 2^18 steps, 64MiB logical
  work and 1MiB retained payload. This adds 380 net production lines without
  another instruction interpreter or a larger budget.
- The retry charges variable clone, meet and candidate payloads. Pure GP work
  reuses the current instruction's typed descriptor at its original late
  fallback. Special/native/private paths keep their order. A full-form cache
  was measured and rejected; mutable source verification is never cached.
- Readonly logical dependencies reuse the packed logical family and complete
  original Go DATA bytes. AND/ANDN need the second input only where the literal
  has set bits; OR needs it where the literal is not all ones; XOR needs all
  bits. Full memory access, alignment, identity and live owner checks remain.
  This repairs SHA1 with 77 net production lines and preserves the old shifted
  literal rules. Zero-filled implementation storage supplies no source facts.
- Existing compare and GP/X transfers share typed byte effects. Terminal RET
  validates each incoming state before dropping dead private identities.
  Ordinary compatible dead private offsets now have a whole-source positive
  test; mismatched roles, saved BP, undefined bytes and restoration still reject.
- Byte-definedness uses a bitset. The four-family packed normalization pilot
  and shared feature projection remove 76 production lines. SDK verification
  reuses its fresh dependency query, reducing two subprocess queries to one.
  Test-owned compiler artifacts are shared with independent source leases and
  bounded cleanup; user-global caches are never cleared.

The final GP descriptor reuse avoids the rejected eager-query overhead.
Measured replay ABBA differences against the same partition implementation
before eager analysis are +4.0% time for register work and +2.4% for mixed work,
with allocation-byte differences +0.22% and +0.56%. Flat cases vary -0.05% to
+1.62%. These are residual local costs, not a whole-translator speedup.

## Qualification of this batch

| Scope | Result | Boundary |
|---|---|---|
| Joint Go1.27.3 / LLVM22 at `d2784db04` | **104 roots / 4,292 children PASS**, zero failures/skips | Original selected tests, unique complete terminals; frozen source/tools |
| Joint minimum Go1.20.14 | **73 roots / 3,938 children PASS**, zero failures/skips | Actual Go-selected files; 31 current-SDK source roles are absent, not runtime skips |
| Complete preflight at `d2784db04` | **573 Node policies; 26 Go roots / 299 children PASS** | Exact Node22.23.4 and unchanged CI selector |
| [Joint Linux execution](https://github.com/cpunion/plan9asm/actions/runs/40496018236) | **3 roots / 35 children PASS**, zero failures/skips | Workflow-only child `0a3eb36b2`; complete Git blobs, tools and raw terminals independently audited |
| [Private-frame Linux execution](https://github.com/cpunion/plan9asm/actions/runs/40489811812) | 5 roots / 36 children PASS | Earlier `89e53140f` source; physical FLAGS owner is original-Go-only |
| Complete original SDK objects | **41/55 PASS**, 14 failures, zero N/A/unattempted | Go1.27.2 Linux/amd64; identical 55 original hashes; SHA1 newly passes, no prior pass lost |
| Five-architecture official corpus at `d2784db04` | PASS | Go1.27.3; no baseline mismatch or parse error; classifications are not initialized/runtime coverage |
| Horizontal integer official join | 56 roots / 685 children PASS on each endpoint SDK | Independent rows, original Go bytes and LLVM22 O0 objects through three APIs |
| Packed ABS join and affected regressions at `b5ca7da71` | Go1.27.3: 60 roots / 1,048 children; Go1.20.14: 59 / 1,048 PASS | Full rule reader, every existing row join, original ABS/sign five-target objects and portable values; zero failures/skips |
| Combined ABS/FMA3 join at `d21fb3c8f` | Go1.27.3: **112 roots / 9,018 children**; Go1.20.14: **111 / 9,018 PASS** | All prior row joins, 178 initialized FMA axes, full named formats, raw/undefined-input controls and existing portable owners; zero failures/skips |
| Integrated `6ca5e4a9d` preflight | **573 Node policies; 26 Go roots / 299 children; five-architecture gate PASS** | Code identical to the qualified combined join; source/tools frozen |

The joint native scope executes 45 logical kernels over all 256 byte values
with separate and overlapping buffers, 36 original shifted-literal kernels
over the same values, and three layout kernels over seven inputs. It checks
**32,277 original Go observations and 193,662 LLVM comparisons**, through three
APIs at O0/O2. Logical outputs compare both complete 96-byte buffers and canaries.
Actual SSE2/AVX2 is mandatory. This is not full SHA1/P256 execution or AVX512
hardware acceptance. Earlier byte-state, compare, GP/X and pointer receipts
remain scoped to their original revisions.

The unfiltered SDK command correctly exits nonzero for its 14 failures. Its
103-second run includes driver build and is not a controlled performance A/B.
Complete P256 retains all 14 TEXTs, with ten public definitions and four private
helpers in the observed CFG. Complete SHA1 retains both TEXTs and 48,115 bytes.
Neither file is cropped or substituted with an instruction-only fixture.

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
| Packed absolute value | 33 | 108 |
| Byte-indexed shuffle | 9 | 24 |
| Integer EQ/GT compare | 72 | 168 |
| Mask blend | 36 | 144 |
| Horizontal integer | 18 | 36 |
| FMA3 | 360 | 2,028 |
| Total | **1,258** | **4,910** |

These joins retain independent ordered Go rows and actual Go bytes, then compile
initialized functions through three APIs with LLVM22. Two PADDD aliases are
separate. High banks force narrow EVEX; MMX has separate region/EMMS owners.
This is neither all x86 rows nor an architecture percentage. Context and zero
unsupported classifications never establish complete translation.

The ABS join adds only a unary format adapter and reuses the existing ABS/sign
initializer, original Go byte verifier and bounded three-API object driver.
Translator production code and native semantic fixtures are unchanged. The
follow-up `65ef2b9bf` only clarifies syntax metadata and strengthens its inventory
assertion: both endpoint SDKs pass its focused 4 roots / 43 children. All 865
previous rows / 2,774 witnesses and both PADDD aliases retain identical IDs,
instructions, bytes and axes on both SDKs. That follow-up also passes all 573
Node policies and the exact Go preflight's 26 roots / 299 children. These are
local object and existing portable-value receipts, not new original-native
execution or current-head CI acceptance.

FMA3 retains all 60 official members and every ordered row. Its 2,028 witnesses
include scalar N4/N8 tuples, 324 broadcasts, 720 register-only rounding controls
and 612 zeroing forms. Arithmetic always reads the old destination, even with
`.Z`. Existing 178 amd64 and 94 386 axis source/declaration pairs remain
byte-identical; both SDKs retain every previous 898-row / 2,882-witness field.
The shared format/encoding extension adds 39 net harness lines and the existing
initializer adds nine. No new translator rule or runtime matrix is introduced.
An ignored Go1.20 overlay snapshot probe hit a vet virtual-file defect; its
failed receipt is retained. Only that diagnostic comparison uses `-vet=off`;
the complete family and CI gates retain their original default vet behavior.

## Remaining work, in order

1. Finish both published revisions' CI/coverage and obtain external review.
   Publish the qualified row-join batch through the allowed fork after the
   active round completes. Retain draft status meanwhile.
2. Complete amd64 official-row ownership, 386 applicability and independent
   semantics. Extend existing family/object/native owners instead of creating
   another opcode registry or enumerating unrelated source/target products.
3. Keep the 14 implicit-entry, GC-scan and system/runtime/test-hook failures
   individually diagnosed in [issue47](https://github.com/xgo-dev/plan9asm/issues/47).
   They remain failed integration work, not automatic exclusions. Object
   success does not prove runtime ABI.
4. Preserve other architecture gates while finishing amd64. The 271 ARM64
   decoder forms, remaining WASM control/SIMD and native AVX512 acceptance remain
   explicit gaps before claiming complete architecture coverage.

External discovery, ledgers and compatibility batches belong to
[llgo-compat/plan9asm](https://github.com/llgo-compat/plan9asm). Do not run inventory
locally or restore those jobs here. Keep raw failed receipts under ignored
`_out/` with their original revision and outcome.
