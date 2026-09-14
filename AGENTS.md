# plan9asm development guide

This repository translates Go Plan 9 assembly to LLVM IR. Changes are complete
only when the accepted operand forms match the Go assembler, the generated IR
is accepted by LLVM 22, and the relevant runtime semantics are tested.

## Active work checkpoint

The current ecosystem-coverage work lives in the worktree
`/private/tmp/plan9asm-reported-libraries-20260913`, branch
`codex/expand-ecosystem-corpus-20260913`, and draft PR
`https://github.com/xgo-dev/plan9asm/pull/40`. Preserve the worktree and inspect
it with `git status --short` before editing. The contribution history is based
directly on `xgo-dev/main`; obsolete per-run ledgers and compressed reports must
not reappear in any commit when the branch is updated.

The current ledger checkpoint contains 320,400 Go index entries, 117,891
successfully inspected exact module versions, 58,807 unique module paths, 661
assembly-bearing exact versions, and 4,180 failures retained for retry. Its next
index cursor is `2026-06-15T14:33:17.568398Z` in
`testdata/discovery/ledger/manifest.json`. The old multi-run ledger is
consolidated into the single `records/` layout, including in the rewritten
contribution history.

Local work has added broad x86/ARM/wasm instruction coverage. In addition to
the x86 binary-floating and horizontal-floating work, the complete Go 1.27 x87
move, arithmetic, comparison, constant, transcendental, environment, and
control families now translate and compile with LLVM 22. ARM has complete
scalar `MOVF`/`MOVD`, float unary/conversion, integer `DIV`/`DIVU`/`MOD`/`MODU`,
and `SLL`/`SRL`/`SRA` families, including FP-addressed classic stack slots.
Shard 3 added the whole ARM AADDF-derived scalar arithmetic family (F/D,
two-/three-register, negative, accumulate, fused, and negative-fused forms),
all `MOVWF`/`MOVWD`/`MOVFW`/`MOVDW` signed/unsigned register conversions, and
the complete `PLD C_SOREG` form. It also added `CMPF`/`CMPD`, decoded raw
`VCMP`/`VCMPE` plus `VMRS APSR_nzcv,FPSCR` sequences (including split S-register
D/M bits), and completed `MOVW` general/F-register bit transfers. ARM now
always uses the architecture-aware CFG lowerer in both translation entry
points; do not restore the old straight-line fallback because it silently
accepted forms absent from the Go tables. Wasm has the complete F32/F64 unary
family and accepts Go metadata directives. Legacy `TEXT name+0(SB)`
normalization, safe `GOROOT/pkg/include` expansion, Go TLS pseudo-register
parsing, and arithmetic constant displacements such as `0+(1*16)(BP)` are also
present.

Shard 2 exposed and now has focused red/green coverage for all four legacy x86
prefetch hints and the Plan 9 x86 `MOVD` alias. Go's x86 frontend maps `MOVD`
to `AMOVQ`; do not lower its XMM memory forms as Intel's 32-bit MOVD. Shard 3
also completed the legacy integer-to-float conversion family, scalar
`VBROADCASTSS`/`VBROADCASTSD`, and signed/unsigned dword-to-qword multiply
families. Their accepted 386 and amd64 operand classes are validated separately
against Go 1.27 and both triples are compiled by LLVM 22. `PMULULQ`'s MMX form
is amd64-only even though its XMM form is valid on 386. Shard 3 also completed
ARM64's eight signed/unsigned 32-/64-bit integer-to-float S/D conversions and
all eighteen scalar floating binary S/D operations, including correct
NaN-propagating `FMAX`/`FMIN` versus number-selecting `FMAXNM`/`FMINNM`
intrinsics. Those families compile for Darwin, Linux, and Windows ARM64 with
LLVM 22.

Shard 4 completed three more mechanisms. ARM64 tailcalls now adapt the full
`i1`/`i8`/`i16`/`i32`/`i64` return-width matrix with `trunc`/`zext`, plus
`ptr`/`i64`; keep the matrix test rather than special-casing `bool`. X86 now
implements the complete Go `yin`/`ynone` port-I/O families: `INB`/`INW`/`INL`
and `OUTB`/`OUTW`/`OUTL` with no operand or one 32-bit immediate, and
`INSB`/`INSW`/`INSL` plus `OUTSB`/`OUTSW`/`OUTSL` with no operands. Their
fixed-register effects, memory clobbers, updated string index/count registers,
and `REP`/`REPN` forms are expressed with constrained inline assembly and
compiled for 386 and amd64 by LLVM 22. The related ordinary string families
`MOVS`/`STOS`/`SCAS` are now complete for B/W/L on 386 and B/W/L/Q on amd64,
including bare, `REP`, `REPN`, forward, and backward forms. Do not restore the
old amd64 behavior that silently discarded `REP` and executed `MOVSB` once.

The parser accepts both Go-supported global declaration grammars,
`GLOBL sym(SB), $size` and `GLOBL sym(SB), flags, $size`; Go 1.27 assembler
acceptance of the historical two-operand form was verified against
`github.com/tonnerre/golang-go.crypto`. `FrameSlot.Name` preserves source-level
Go result names. On 386, a unique named result can correct a stale numeric FP
offset in legacy assembly; anonymous results and aggregate results with several
physical slots retain strict numeric-offset matching. This is required by
`github.com/sigma/vmw-guestinfo` and must not be weakened into accepting
arbitrary unknown frame writes.

Discovery now proves package applicability with an exact-package current-Go
build and applies `go vet -asmdecl` only to concrete FP/argument ABI
mismatches; generic vet diagnostics must still reach the translator. The
`plan9asmll` Go-package driver selects `WASMABIGo` for `GOARCH=wasm`, because
Go package assembly uses the resumable linear-memory stack ABI rather than the
direct LLGo ABI. ARM and ARM64 tailcalls adapt all integer result widths with LLVM
`trunc`/`zext`; this is needed when a local assembly helper has only a fallback
signature but the public Go declaration returns `bool`.

The current 661-candidate checkpoint has all 32 canonical reports green: 510
passing candidates, 151 structured source N/A candidates, zero failures, 7,800
LLVM 22 translations, and 283 target-level N/A translations. These values come
from `_out/discovered-library-corpus/shard-{0..31}.json`; ignore historical
files with suffixes such as `-fixed`. A targeted `linux/arm64` replay also
proved the saved-path query path: it selected 316 of the 661 assembly-bearing
versions without reading the module index or revisiting no-assembly records.
Shard 18 itself was green with 22 passing
candidates, 2 source N/A candidates, 492 translations, and 26 target-level N/A
translations. It independently exposed and now passes extended amd64 byte
register aliases in `github.com/bronze1man/AesCtr` and
`github.com/issuj/gofaster`, all `JCXZW/JCXZL/JCXZQ` counter-zero branches in
`github.com/clmul/checksum`, and the complete PALIGNR/VPALIGNR,
PCLMULQDQ/VPCLMULQDQ, and PSLLO/PSLLDQ/PSRLO/PSRLDQ/VPSLLDQ/VPSRLDQ families
needed by `github.com/nspcc-dev/tzhash`. Those families have Go 1.27
positive/negative form tests, all applicable x86 platform object tests, LLVM
22 semantic runtime tests, and supported-op extraction assertions.

Discovery records can also contain `_test_<goarch>.s` files. When an exact file
set contains test assembly, `plan9asmll` loads Go package test variants and
prefers their type scope for translation; otherwise declarations that exist
only in `_test.go` incorrectly degrade to the integer fallback ABI. Keep
`cmd/plan9asmll/testdata/testsignature` as the regression for a `float32`
result written by `MOVSS`. Do not fix this class of failure by accepting an
`i64`/float frame mismatch.

Shard 19 is green with 15 passing candidates, 5 source N/A candidates, 138
translations, and 21 target-level N/A translations. TDD failures independently
exposed `RCLQ/RCRQ` memory forms in `github.com/kilic/fp256`, `PSHUFLW` in
`github.com/ledao/arrgo`, and `VAESENC` in `github.com/mengzhuo/nabhash`. The
fixes cover the complete Go 1.27 RCL/RCR B/W/L/Q family, the legacy/VEX/EVEX
PSHUFHW/PSHUFLW family, and the AES/VAES round, inverse-mix-column, and
keygen-assist family. Its report remains useful as the per-family TDD evidence;
the aggregate checkpoint immediately above now also includes shard 20.

Shard 20 is green with 12 passing candidates, 9 source N/A candidates, 122
translations, and no target-level N/A translations. Its only initial failure
was a module requiring Go 1.27.1 while the scanner runs Go 1.27.0 with
`GOTOOLCHAIN=local`: `go mod download -json` had already cached a verified ZIP
but returned nonzero without a module directory. The corpus runner now parses
metadata even on nonzero exit and safely materializes an available module ZIP;
the later Go build version rejection is retained as explicit per-target source
N/A evidence.

Shards 21 through 23 are green. Together they add 43 passing candidates, 12
source N/A candidates, 360 translations, and no target-level N/A translations;
none exposed a new opcode or form gap. Shard 24 is green with 17 passing
candidates, 4 source N/A candidates, 735 translations, and 46 target-level N/A
translations. Its TDD failure in `github.com/chewxy/math32` exposed
`CVTSS2SL`; the fix covers the complete Go 1.27 scalar float-to-integer family:
eight legacy signed forms, eight VEX/EVEX signed forms, and eight EVEX unsigned
forms, including memory sources and the register-only embedded rounding/SAE
forms. Shard 25 is green with 22 passing candidates, 1 source N/A candidate,
242 translations, no target-level N/A translations, and no new form gap.
Shard 26 is green with 10 passing candidates, 4 source N/A candidates, 177
translations, no target-level N/A translations, and no new form gap. It also
proves that the generic ledger/discovery path independently selects and passes
`github.com/vmware/vmw-guestinfo`; it is not a hard-coded issue fixture.
Shard 27 is green with 12 passing candidates, 5 source N/A candidates, 100
translations, no target-level N/A translations, and no new form gap. Shard 28
is green with 18 passing candidates, 5 source N/A candidates, 422 translations,
and 35 target-level N/A translations. TDD failures independently exposed
`ROUNDSD` in `github.com/mdempsky/go` and `BOUNDL` in `rsc.io/tmp`; the fixes
cover the complete Go 1.27 legacy/VEX ROUND family and the 386-only BOUNDW/L
family. ROUND has a five-platform form matrix and LLVM runtime semantics for
all rounding modes plus scalar upper-lane preservation. BOUND is emitted as
the exact 62 /r encoding with fixed AX/CX constraints because LLVM 22 rejects
the mnemonic itself. Shard 29 is green with 18 passing candidates, 10 source
N/A candidates, 166 translations, 12 target-level N/A translations, and no new
form gap. Shard 30 is green with 17 passing candidates, 3 source N/A candidates, 292
translations, 23 target-level N/A translations, and no new form gap. Shard 0
and shard 31 have both been regenerated from this worktree.

Shard 31 is green with 22 passing candidates, 4 source N/A candidates, 423
translations, and 27 target-level N/A translations. `github.com/cloudflare/circl`
independently exposed the complete x86 CLC/STC/CMC carry-control family, the
complete ADC/SBB B/W/L/Q family and its memory destinations, and ARM64's
NEG/NEGW/NEGS/NEGSW plus NGC/NGCW/NGCS/NGCSW families. Their legal operand
rows come from Go 1.27, all applicable OS/architecture targets compile with
LLVM 22, and runtime tests cover carry/borrow, overflow, and NZCV semantics.
`github.com/notti/nocgo` exposed a signature-inference gap: a declaration-free
direct tail-forwarding TEXT now inherits the complete ABI only when all of its
external tail targets have declared, identical signatures. This fixes 386 C-ABI
trampolines without changing the conservative fallback for genuine undeclared
register-return assembly.

Shard 5 added two x86 mechanisms. Scalar `MOVL` may store either a GP register
or immediate into an FP result slot but rejects memory-to-memory forms. The
complete Go 1.27 `ydivb`, `ydivl`, `yimul`, and `yimul3` families now cover
`MUL`/`IMUL`/`DIV`/`IDIV` B/W/L/Q and `IMUL3` W/L/Q, with implicit AX or DX:AX
results, all GP/memory sources, IMUL immediate/register/memory forms, correct
386 Q-width rejection, multiply CF/OF behavior, and LLVM 22 object compilation
on both i386 and x86_64. Keep the full opcode switch in
`amd64_lower_muldiv_scalar.go`: both supported-op extractors rely on literal
case labels and must not silently omit a lowerable opcode. This fixes all
`github.com/bjwbell/gensimd` failures; one of its nine files is a structured
target N/A because its explicit TEXT arg size disagrees with the Go declaration.

Shard 6 hardened applicability evidence. An `unexpected EOF` accompanied by a
local `.s` diagnostic and `asm: assembly of ... failed` is deterministic source
rejection, not a transient proxy EOF. Also, old package tests can fail type
checking before `go vet -asmdecl` analyzes otherwise buildable production
assembly. On a generic vet failure, Discovery now copies the owning module out
of `GOMODCACHE`, replaces its test files with empty same-package files via a
temporary modfile, and reruns asmdecl. This is necessary because the Go command
forbids overlays under `GOMODCACHE`. The mechanism gives
`github.com/neclepsio/qml` concrete invalid-FP-offset evidence, while
`github.com/v2pro/plz` excludes only its truncated 386/ARM sources and still
translates its valid amd64/ARM64 files. Keep the focused tests in
`cmd/plan9asmcorpus/discovery_test.go`; do not collapse local assembler EOFs
back into infrastructure retry records.

Use final JSON aggregation for the PR funnel rather than recomputing this
prose. All 32 discovery shards are green for the current 661 matched versions.
The committed official Go form baselines cover Go 1.20 through Go 1.27 for
386, amd64, ARM, ARM64, and wasm; all 40 reports have zero supported-form
regressions and zero parse errors. The Go 1.27 standard-library target/feature
matrix also passes without unsupported instructions or operand forms. The full
root module, nested command modules, formatting gates, conformance suites, and
curated corpus must still be rerun immediately before publishing.

The next agent should:

1. Run every gate below, review the final coverage reports, and calculate the
   final funnel from the committed manifest and canonical shard reports.
2. Update PR 40 only through the `cpunion` fork, monitor every CI job, address
   review feedback, and keep it draft until local checks, CI, review, and
   coverage are green. Never merge the upstream PR.

This checkpoint is temporary handoff state. Update or remove it when the work is
finished; the remaining sections are durable repository procedure.

## Repository and contribution safety

- `xgo-dev/plan9asm` is upstream and is outside the direct-push authority
  boundary. Never push to `origin` or `xgo-dev`, and never merge the upstream
  PR. Push the contribution branch only to the allowed `cpunion` fork.
- Fetch upstream before final validation and inspect divergence. Do not rebase
  or rewrite while tests are running, and do not discard unrelated dirty-tree
  changes.
- Use `apply_patch` for source edits. Use `gofmt` for mechanical formatting.
- Generated reports belong under `_out/` or a temporary directory. Only the
  normalized Discovery ledger is committed; never commit ZIPs, module caches,
  object files, raw run directories, or compressed reports.

## Sources of truth

- Use the supported Go toolchain's assembler tables under
  `$(go env GOROOT)/src/cmd/internal/obj/<arch>` as the authority for opcode,
  operand, register, suffix, masking, broadcast, and rounding forms.
- For x86 AVX/EVEX instructions, inspect both the referenced `ytab` and every
  opcode attribute in `avx_optabs.go`. A shared `ytab` does not imply that
  rounding, SAE, broadcast, or zeroing flags are identical.
- Confirm uncertain forms with the Go assembler. Do not infer a form solely
  from Intel syntax or from a third-party source file.
- Treat `386`, `amd64`, `arm`, `arm64`, and `wasm` as supported architectures.
  A form must not be classified as unsupported merely because the host has a
  different architecture.

## Instruction-family workflow

Use test-driven development for every instruction change:

1. Reproduce the unsupported instruction or incorrect operand-form behavior in
   a focused test and run it to observe the expected failure.
2. Before implementing the opcode, enumerate its complete family and all forms
   in the current Go assembler tables. Add positive and negative form tests.
3. Implement a coherent instruction family, not only the single spelling that
   exposed the gap. Cover every X/Y/Z width, register and memory source,
   architecture restriction, mask/zero form, broadcast form, and rounding/SAE
   form present in the Go tables.
4. Compile generated IR with LLVM 22 for every affected target. Add runtime
   conformance when an instruction's lane preservation, operand order, flags,
   masking, NaN behavior, or memory effects are not established by compilation.
5. Rerun the discovering external-library shard or official corpus that exposed
   the problem. One fixed mnemonic must not hide another form from the same
   family.

For a newly discovered instruction, update all four layers before calling it
complete: the family-specific lowerer, positive/negative Go-table form tests,
LLVM 22 object compilation for every affected supported architecture, and the
architecture's supported-op extraction test in `cmd/plan9asmll/main_test.go`.
The external corpus uses that extraction result in its diagnostics, so omitting
the last layer can make a supported instruction look unsupported.

A typical x86 investigation starts with the Go 1.27 tables and an assembler
probe:

```sh
goroot=$(go env GOROOT)
rg -n 'A?VADDPS|yvaddps' "$goroot/src/cmd/internal/obj/x86"
sed -n '<start>,<end>p' "$goroot/src/cmd/internal/obj/x86/avx_optabs.go"
GOOS=linux GOARCH=amd64 go tool asm -o /tmp/form.o /tmp/form_amd64.s
GOOS=linux GOARCH=386 go tool asm -o /tmp/form.o /tmp/form_386.s
```

Record the initial failure in test output before adding the lowering. Focused
form tests normally belong in `amd64_ecosystem_test.go`,
`amd64_operand_forms_test.go`, or the corresponding ARM file. Add the opcode to
`TestExtractSupportedOpsFindsCompleteAddedInstructionFamilies` so external
corpus diagnostics recognize it. Semantic fixtures live under
`testdata/conformance/<arch>` and must be checked against both native Go and the
LLVM runtime oracle where supported.

The main implementation entry points are:

- x86/386/amd64: `amd64_blocks.go`, `amd64_needed.go`, `amd64_translate.go`,
  then the family-specific `amd64_lower_*.go` file;
- ARM: `arm_blocks.go`, `arm_needed.go`, `arm_translate_cfg.go`, then
  `arm_lower_*.go` and frame/register handling in `arm_eval.go`;
- ARM64: `arm64_translate.go`, `arm64_eval.go`, and `arm64_lower_*.go`;
- wasm: opcode families in `wasm_ops.go`, lowering in `wasm_translate.go`, and
  intrinsic declarations in `translate_prelude.go`;
- command-side package signature, ABI, include, and target handling:
  `cmd/plan9asmll/main.go` with tests in its nested module.

Supported-opcode reporting scans opcode-keyed package maps as well as explicit
switches. Every newly introduced opcode map must have an architecture-specific
extraction regression in `cmd/plan9asmll/main_test.go`; otherwise a real
translator failure can be misreported as an unknown instruction.

For x86 form tests, cover at least these targets:

- `darwin/amd64` — `x86_64-apple-darwin`
- `linux/amd64` — `x86_64-unknown-linux-gnu`
- `windows/amd64` — `x86_64-pc-windows-msvc`
- `linux/386` — `i386-unknown-linux-gnu`
- `windows/386` — `i686-pc-windows-msvc`

Do not add an `unsupported_forms` skip for a supported target. Context-dependent
forms may be classified separately only when translation genuinely requires
information unavailable to the scanner.

## Toolchain requirements

- Use Go 1.27 for new instruction development, Discovery, and external-library
  corpus validation. The official standard-library/toolchain corpus separately
  retains the Go 1.20 through Go 1.27 compatibility matrix.
- Use LLVM 22 exactly. Never fall back to LLVM 23 or another installed release;
  compatibility with another release does not prove LLVM 22 compatibility.
- `scripts/llvm22.sh` accepts `llc-22`, an unversioned `llc` reporting version
  22, or `LLVM_CONFIG` pointing to LLVM 22. On macOS, Homebrew LLVM can be
  selected with `LLVM_CONFIG="$(brew --prefix llvm@22)/bin/llvm-config"`.
- Prefer repository scripts because they validate the Go and LLVM versions
  before running a corpus.

Check the active tools before a substantial validation run:

```sh
go version
source scripts/llvm22.sh
llc_cmd=$(find_llvm22_llc)
"$llc_cmd" --version
```

## Test and validation commands

Run focused red/green tests while developing, then run all applicable gates.

```sh
# Root module and the two nested command modules.
go test ./... -count=1
(cd cmd/plan9asm && go test ./... -count=1)
(cd cmd/plan9asmll && go test ./... -count=1)

# Formatting and build checks.
test -z "$(gofmt -l .)"
go build ./...
(cd cmd/plan9asm && go build ./...)
(cd cmd/plan9asmll && go build ./...)

# Semantic runtime checks. Cross execution requires the CI cross compilers and
# QEMU; native checks should still be run locally when applicable.
go test . -run '^(TestAMD64ConformanceNativeGo|TestAMD64ConformanceLLVMRuntime)$' -count=1 -v
PLAN9ASM_CROSS_EXEC=1 go test . -run '^(TestCrossLinuxRuntimeMatrix|TestARM64Conformance)' -count=1 -v

# Official Go assembler and standard-library corpora.
scripts/check-go-asm-coverage.sh
scripts/check-stdlib-corpus.sh
scripts/check-arm64-plan9-corpus.sh

# Curated issue/reported libraries and all discovered assembly candidates.
scripts/check-reported-library-corpus.sh all
for shard in $(seq 0 31); do
  scripts/check-discovered-library-corpus.sh "$shard" 32
done
```

The official corpus gate requires every form to be classified, no parse errors,
and no unsupported forms. Inspect form-level reports before intentionally
updating `testdata/coverage/go-asm-baseline.json`; never update a fingerprint
merely to make CI green.

## External-library discovery

Discovery reads the official Go module index, resolves each unique module as
`module@latest`, and inspects that exact ZIP. It must independently rediscover
libraries mentioned in issues; the curated manifest is an additional regression
suite, not an input whitelist for Discovery.

The committed checkpoint is one diff-friendly ledger:

```text
testdata/discovery/ledger/
  manifest.json
  records/00.jsonl ... records/ff.jsonl
```

- Select a shard with the first byte of `sha256(module path)`.
- Keep every version of a module in the same shard.
- Sort each shard by module path, Go semantic version, then record kind.
- Always resolve `@latest`. Add a newly resolved version while retaining older
  versions unless a retained version is proven incompatible with Go 1.27.
- Record completed exact versions even when they contain no assembly so later
  scans do not download them again.
- Matched records retain every discovered `.s` relative path, including files
  for architectures plan9asm does not support yet. Their `architectures` field
  is rebuilt from all GOARCH filename suffixes known to the current Go toolchain,
  not only architectures plan9asm already supports. When adding a platform,
  select exact `module@version` candidates from these paths and rerun only
  applicability/translation for that target; never rescan all index entries or
  all no-assembly modules. Source build constraints and declarations are
  reevaluated after fetching only those matched exact versions.
- Retain transient failures for retry; a successful retry removes the failure
  record for that exact version.
- Keep only `manifest.json` and `records/*.jsonl`. Do not commit run directories,
  recursive ledgers, generated archives, or `.gz` files.

Continue a scan by updating the ledger in place:

```sh
next_since=$(jq -r .next_since testdata/discovery/ledger/manifest.json)
go run ./cmd/plan9asmdiscover \
  -since "$next_since" \
  -limit 20000 \
  -workers 24 \
  -out-dir testdata/discovery/ledger
```

`plan9asmdiscover` is the only normal writer for committed records. With
`-out-dir` it first reads the existing ledger, resolves every unique module in
the requested index window through `module@latest`, avoids downloading an exact
version already recorded, merges the result, removes a same-version failure
after success, deduplicates exact module versions and assembly paths, sorts Go
versions with `golang.org/x/mod/semver`, rewrites all 256 hash shards in a
temporary directory, and replaces the ledger atomically. It records successful
versions even when no assembly exists. Do not hand-append JSONL lines.

The writer is deliberately idempotent. Running the same index window again
does not duplicate modules, versions, assembly paths, or failures; it rewrites
the canonical shards in the same order. There is no manual sorting step and no
separate per-run directory to merge. If writer behavior changes, add a fixture
to `cmd/plan9asmdiscover/main_test.go` that runs the merge twice and compares
the resulting files byte-for-byte.

The three record kinds are:

- `scanned`: exact `module@version` inspected successfully, with or without
  assembly;
- `matched`: the same exact version has assembly, plus normalized architecture
  hints and module-relative `.s` paths;
- `failure`: the exact version, or `@latest` if resolution failed, must be
  retried and is not considered scanned.

Validate record integrity and summarize the checkpoint after every update:

```sh
go test ./cmd/plan9asmdiscover ./cmd/plan9asmcorpus -count=1
jq '{since,next_since,index_entries,unique_modules,skipped_previously_scanned,scanned_records,matched_records,failure_records}' \
  testdata/discovery/ledger/manifest.json
test "$(find testdata/discovery/ledger/records -type f -name '*.jsonl' | wc -l | tr -d ' ')" = 256
find testdata/discovery -type f \( -name '*.gz' -o -name '*.zip' \) -print
```

The last command must print nothing. `go test` re-reads every shard and rejects
wrong hashes, duplicates, unsafe paths, invalid successful semvers, unexpected
files, or module/string-version/kind ordering. A legacy report can be imported
once with `-convert-report <path> -out-dir testdata/discovery/ledger`; do not
create another `runs/` hierarchy. Older exact versions remain in scope unless
there is concrete evidence that they are incompatible with Go 1.27; record that
evidence in the PR before removing them.

Discovery corpus sharding uses `sha256(module@version) % 32`, independently of
the ledger file shard. Run and retain one auditable report per shard:

```sh
mkdir -p _out/discovered-library-corpus
for shard in $(seq 0 31); do
  scripts/check-discovered-library-corpus.sh "$shard" 32 \
    "_out/discovered-library-corpus/shard-$shard.json"
done

jq -s '{
  index_shards: length,
  selected: (map(.selected) | add),
  passed: (map(.passed) | add),
  not_applicable: (map(.not_applicable) | add),
  failed: (map(.failed) | add),
  translations: (map(.translations) | add),
  not_applicable_translations: (map(.not_applicable_translations) | add)
}' _out/discovered-library-corpus/shard-*.json
```

When implementing a new supported platform, replay only records whose saved
assembly paths may match it. The filename filter excludes only paths that
conclusively name another GOOS or GOARCH; unsuffixed files and unknown/custom
suffixes remain eligible and are checked after the exact version is fetched:

```sh
# After linux/riscv64 has been added to plan9asm's supported-target validation:
for shard in $(seq 0 31); do
  PLAN9ASM_DISCOVERY_TARGETS=linux/riscv64 \
    scripts/check-discovered-library-corpus.sh "$shard" 32 \
    "_out/discovered-library-corpus/riscv64-shard-$shard.json"
done
```

This command reads only `records/*.jsonl`, downloads the recorded matched exact
versions, derives package paths from the saved `.s` paths, and performs current
Go applicability, translation, and LLVM 22 compilation for the requested
target. It does not read the Go module index or revisit scanned no-assembly
versions. The report records both the full ledger candidate count and the
target-eligible count so the filter is auditable.

Every report must satisfy `selected = passed + not_applicable + failed`, every
target must satisfy `total_asm = success + not_applicable + failed`, and final
`failed` must be zero. A candidate with at least one successful target is
`passed`; a candidate is globally `not_applicable` only when no selected source
can be translated on any supported target.

Source applicability must follow Go's rules: use file-name GOOS/GOARCH suffixes
and `go/build.Context.MatchFile` for constraints. For unsuffixed assembly whose
source is architecture-specific, probe it with the Go assembler for the target
matrix. Every applicable discovered `.s` file must enter translation and LLVM
22 compilation, or the candidate fails; do not silently skip an applicable file.

For each target/tag/package group, Discovery then runs `go build` against that
exact package path, never `package/...`. A compiler rejection on Go 1.27 is
structured source N/A evidence for that package/target only; it does not exclude
another package, target, or retained version. Network, proxy, timeout, process,
and filesystem failures are infrastructure failures and must fail the shard for
retry, never become N/A. A package that builds must proceed to `go vet -asmdecl`,
translation, and LLVM 22 compilation. Only asmdecl's wrong argument size,
invalid FP offset, or invalid FP-width diagnostics are ABI N/A evidence; unknown
variables, missing declarations, dependency errors, unsupported instructions,
parse errors, and LLVM errors are not.

An unsuffixed file may still encode one architecture's ABI in its `TEXT`
argument size. Classify a file/target pair as not applicable when its explicit
`TEXT $frame-args` size disagrees with the size derived from that target's
current Go declaration (the same condition reported by `go vet -asmdecl`). Do
not treat an omitted `-args` suffix as zero: current, valid packages use that
form. Emit the structured symbol, declared size, expected size, and target in
the corpus report.
Unsupported instructions, parse failures, LLVM failures, and generic source
errors are never valid not-applicable evidence.
If the current Go assembler conclusively rejects an unsuffixed source for every
target selected by Go's file/build constraints, retain a structured source-level
N/A item containing the file, attempted targets, and rejection kind. This is
appropriate for historical toolchain fixtures or generated invalid test input;
it must not be used when even one supported target accepts the source.

The final report validator requires every source N/A item to name an allowed
kind, one or more discovered assembly files, attempted `GOOS/GOARCH` targets,
and a nonempty diagnostic. `applicable_asm_files` is a unique file set even when
the same file is exercised on several targets; `translations` carries the
per-target count.

## Pull-request completion

- Keep broad corpus work as a draft until local tests, CI, review, and coverage
  checks pass.
- The PR body must report a funnel with index entries, exact module versions,
  unique module paths, assembly-bearing versions, retry failures, candidates
  passing/N/A/failing, and newly covered instruction families/forms.
- Calculate funnel values from the committed ledger manifest and the 32 saved
  shard reports, not from terminal recollection. Include a Markdown table in
  the PR body only; do not duplicate transient counts in the Discovery README.
- Verify that libraries named by user-reported issues occur in the Discovery
  ledger independently of the curated manifest, and state their exact versions
  and shard outcomes in the PR body.
- When replacing an obsolete generated ledger layout, rewrite the contribution
  branch before publishing so deleted run files and compressed artifacts do not
  remain in the PR's commit history. Use `--force-with-lease`, never overwrite
  an unrelated remote update.
