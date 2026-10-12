# Current work: consolidate and repair PR 40

External scanning, ledgers and corpus verification have moved to Draft
[llgo-compat/plan9asm#1](https://github.com/llgo-compat/plan9asm/pull/1).
Its initial ledger import was compared byte-for-byte with the pre-split
`c8724add` snapshot. Historical pending/failing outcomes were not promoted.
The original source and failed reports remain recoverable; the new repository
owns subsequent inventory imports, cursors, external failures and ledger updates.

## Translator scope

- Keep all instruction-family repairs through `c8724add`, official Go coverage,
  standard-library compilation, cross-runtime tests and benchmarks.
- `gotoolprofile` is now public so the independent harness and translator CLI
  share the same guarded observations without copying the implementation.
- Remove external commands, ledgers, exception manifests, scripts and jobs.
  Core CI depends on its policy preflight, not the external corpus.
- The exhaustive module runner assigns owner 1 to `gotoolprofile` rather than
  the moved corpus driver. Every actual package still has exactly one owner;
  root partitions and the strict complete-module aggregate remain mandatory.

## Consolidated repair checkpoint

`3dc36d71` contains the instruction repairs and external-corpus separation.
Continue from this integrated source, not an older family worktree. The latest
review's SDK, environment, tool provisioning, fixture and cache findings have
focused repairs below; broader CI gates still need current-head completion.

The first repair batch fixes the ARM memory-source ADD and the ARM64 branch
label that collided with a vector register. Runtime and portable preflight
tests share their four-architecture fixtures. Their actual Go objects and
LLVM 22 objects pass on Go 1.27.2 and Go 1.20.14; native Linux execution still
requires the cross-runtime lane. Ordinary, race and coverage jobs now install
the required IA32 compiler/libc, with an independent ordering regression.
Workflow policy tests and actionlint pass; the split CLI import formatting is
also corrected. RED evidence is retained locally, not converted into a pass.

Go 1.27.2 observation support now retains the audited format pins, separately
registers its two changed native-header source files and binds object/archive
headers to the producer's exact patch. ARM64 actual-source, CRC32 three-API
LLVM 22 objects, static-storage and marker mutation guards pass; Go 1.27.1 and
Go 1.20.14 focused compatibility checks pass. Unknown patches remain rejected.

The ordinary CLI now resolves its actual cache paths and explicit defaults
before package loading, then hands that same environment to source observers.
Explicit nonordinary options remain visible, not silently normalized away.
Go's actual target-specific CGO default is retained, so native package inventories
do not shrink. Pure-Go source proofs still require a compatible explicit profile.
Implicit workspaces, saved Go flags and automatically discovered package drivers
are preserved too; ordinary proofs cannot disable those selection inputs.
Default-cache, explicit-profile, overlay and source-selection tests pass.
The twelve ARM64 undefined-byte cases now share only a test-owned build cache;
their source/temp/observation state remains independent. Local elapsed time
was 32.52 seconds before and 12.72 seconds after, without removing any cases.

CI run `38131270471` finished with failures. Its standard-library benchmark
compiled 163 of 185 files; its 68-second target total is not a complete-success
performance result. Keep the PR draft and preserve failed evidence.

Both complete CLI suites and every non-root main-module package passed on
`fa9cbfee`. The translator CLI took 1431.87 seconds and `gotoolprofile` took
1294.84 seconds. That snapshot's local cross-target benchmark compiled 163/184
files with 21 real context/ABI failures in 78 target-seconds; its inventory is
not the native Linux CI inventory. Neither result establishes CI completion.

## Remaining delivery gates

The next local repair batch rejects unresolved x86 SB displacements before
all three translation APIs can emit a wrong symbol. The 386 standard-library
scan now reports zero unsupported forms, but its translation gate still fails
on the independent runtime storage/ABI contracts; it is not a corpus pass.

Runtime oracle objects now use PIC to match Linux's PIE links. Their standalone
C processes disable Linux core-pipe collection before fault children run,
without changing host settings. Focused Linux-container link/process tests pass.
Legacy AES keeps its native alignment exception and rejects unbacked logical
FP addresses; logical numerical tests use an explicit unaligned move. Shared
128-bit FP reads now assemble every scalar field instead of losing adjacent
float32 fields. AES/legacy-move object and diagnostic execution tests pass.

VZEROUPPER's slot+16 tail stores now explicitly retain 16-byte alignment.
Previously LLVM inferred 64 for a 48-byte store and emitted misaligned YMM
stores in the packed double/dword fixture. RED/GREEN alignment tests, five-target
objects and portable execution pass. Conversion-fault assertions now require
one surviving native conversion in each function, independently of output
stores. Three-target objects, the five-architecture official classification
gate, focused Go 1.20 checks and vet pass. Current-Go Windows downloaded-frame
object tests also pass. Native Linux precision/alignment and AVX-512 execution
still need their required owners: Rosetta does not reproduce the precision
trap, and QEMU's diagnostic AVX-512 attempt failed its ISA execution. Neither
failure is reclassified as a pass or a source rejection.

The independent exported SSE fault program passed on native Linux x86_64;
this is not a complete native CI-owner receipt. The seven legacy packed moves
now distinguish logical FP fields from native alignment on both reads and
writes. Packed FP stores use complete byte-field writes too: an executed
four-uint32 return regression previously produced 3 instead of 2 in field one.
The complete move-family tests pass on Go 1.27.2 and focused Go 1.20.14 tests,
including three APIs, five-target objects and O0/O2 tuple-result execution.

The old monolithic `41d50b0d` root run exhausted its 45-minute budget; cascading
deadline errors remain infrastructure failures, not instruction rejections.
Two stale authenticated-fixture paths after the profile migration are repaired.
The exhaustive 32-part `fa9cbfee` run completed all 5897 roots: 5857 passed,
39 host-inapplicable skips, and one TLS classification failure. `db371c8c`
repairs the whole scalar arithmetic TLS class without inventing native backing;
615 actual-Go address rows, malformed controls, three-API/five-target objects,
Go 1.20 checks, the official classification gate and vet pass. Replaying the
entire original failed shard on that revision retained all 204 roots: 203
passed, one host-inapplicable skip, no failure or missing root. The original
failed report remains failed; this replay is not a new complete-suite receipt.

Generated logical and insert/extract numerical oracles no longer return large
case IDs as process status: an injected wrong result previously returned 256
and appeared successful on Unix. Both mutation regressions pass. The logical
oracle retains all 472 formats and 7552 inputs in eight bounded batches, with
non-inlined independent checks. Local whole-test time fell from 88.49 to 15.59
seconds; Linux-container execution passed in 37.41 seconds. Go 1.20 and the
complete 278-format insert/extract numerical matrix also pass. Container and
Rosetta observations do not certify native AVX512 execution.

The verified repairs through `575c5962` are now published to the allowed fork;
PR 40's new CI run is `38182744838`. Its fmt and build jobs pass, but the
standard-library gates still fail and the complete run is not yet a pass.

The AMD64 repair now separates explicit scalar ABIInternal register results
from ABI0 FP initialization while sharing the existing all-path byte analysis.
The eight insertion names, all 60 lane formats, three APIs, three-target O0/O2
objects and portable byte-reference execution pass. Go 1.20 focused checks and
the shared result regressions pass too. Real SDK maps replay still fails on
immediate-shuffle result analysis and the independent AES global-storage role;
neither source file is promoted to a corpus pass.

The next verified local batch fixes empty LLVM_CONFIG handling in isolated
module fixtures, Go 1.26's unconditional wasm feature tags and Linux C oracle
header/identifier collisions. Process setup now uses a separate translation
unit; its dumpability check and all eight ARM64 MRS execution combinations pass
in the Linux container. Complete Linux root owners now install pinned QEMU;
the actual installer and original-Go/LLVM 386 mutable-read matrix pass locally.
These changes do not establish native AVX512 execution or a complete CI pass.

Windows repairs share escaped local proxy URLs and host executable suffixes,
keep shell checkouts LF, and move shared fixed-local definitions out of Unix-only
files. The Windows helper-selection RED/GREEN, Go 1.20/current-Go source and
object matrices, portable transport execution, CLI regressions and vet pass.
Windows native execution still requires CI. Continue with historical Go Windows
SEH schemas and Windows canonical input-path failures. Keep their existing
failures; do not loosen source guards.
The local repair batch has not been pushed while run `38182744838` is active.

ARM64 bytealg private fixture signatures now describe explicit register
transport without fabricated ABI0 frames. Go 1.20/current-Go objects pass;
the missing R26 closure entry remains a negative. PCALIGN/PCALIGNMAX now use
the complete Go operand grammar, retain function alignment and permit NOP
padding in tail-only continuations whose whole-file audit excludes native
PC/raw/address observers. CALL continuation guards remain separate. The 27
format cases pass all three APIs and five target objects; historical SDK Count,
source mutation guards, Go 1.20 checks and vet pass. PCALIGN loop execution
passes original-Go and independent LLVM/C references under ARM64 QEMU.

The remaining 21 standard-library failures require actual runtime/reflect
entry, TLS, hidden-register, caller-storage or SDK source contracts. Do not
erase them, clip observable result padding, or fabricate frames to pass the
benchmark. Keep runtime/compiler integration explicit and separately tested;
retain the contribution draft until the required native CI and these gates pass.

The migration PR is draft. Its historical assembly ledger remains incomplete;
this does not change translator coverage or certify any external module.
The split is already published to PR 40. Verify the new repairs with focused
RED/GREEN cases, the shared package, both CLIs and affected full gates, then
publish a verified batch to the allowed fork. Do not merge or close upstream PRs.

Historical runs, including `fe553bff` and fork run `38116495799`, certify only
their own revisions. Retain their evidence when retiring inactive worktrees;
never promote old or incomplete reports to current-head success.

The five-architecture official Go classification gate passes after the split.
The supplementary ARM64 decoder baseline now records the audited rejection of
four decoder-only spellings: SCVTFDD, SCVTFSS, UCVTFDD and UCVTFSS. Both Go 1.20
and 1.27 assemblers reject them. Reclassifying only those four forms reproduces
the entire prior fingerprint; no accepted source form was removed. The separate
actual-Go-accepted corpus gate remains mandatory.

CRC32/PCLMUL feature inference now has a small independent Go-accepted fixture
on all Go versions and hosts, replacing the raw-SDK test with handwritten frames
and unresolved generated constants. Complete current-SDK CRC32 coverage remains
in its three-API source/header/CPU/literal-object oracle, including the feature
assertions. Historical SDK lanes test the portable instruction contract, not
full SDK CRC32 compatibility. Both use bounded LLVM 22 object emission; neither
test establishes native CRC execution.

Use [instruction-family development](instructions.md) and
[validation](validation.md). Keep old detailed progress narratives in Git
history instead of growing this checkpoint again.
