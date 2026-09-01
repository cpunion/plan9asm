# Plan 9 assembly instruction coverage

This document defines the path to complete support for the Go assembler's Plan
9 syntax. Coverage is measured by architecture, instruction family, opcode,
and operand form. An opcode name by itself is not a support claim.

The target architectures are 386, amd64, arm, and arm64. The 386 and amd64
lowerers share x86 implementation code, but their registers, address widths,
valid opcodes, and runtime ABI differ. ARM and ARM64 are separate instruction
sets. ARM64 also distinguishes scalar, NEON, and SVE families.

## Sources of truth

There is no version-independent "Plan 9 instruction set" that can be imported
as a complete oracle. Go's assembler uses Plan 9 syntax for a distinct,
semi-abstract instruction set, and its accepted language evolves with the Go
toolchain. Coverage therefore uses the union of Go 1.20 through the latest
supported release, through these increasingly strong layers:

1. Official opcode namespace
   - `cmd/internal/obj/x86/anames*.go`
   - `cmd/internal/obj/arm/anames*.go`
   - `cmd/internal/obj/arm64/anames*.go`
2. Official architecture encoder tables
   - x86 `optab`, `ytab`, `ymovtab`, and generated AVX/EVEX tables
   - ARM and ARM64 `optab` rows and their alias mappings
   - generated ARM64 instruction encoders, including SVE
   - these tables are the authority for legal abstract operand classes
3. Official positive assembler testdata
   - `cmd/asm/internal/asm/testdata`
   - this supplies concrete, architecture-valid operand forms
4. Selected standard-library assembly
   - lower-case `.s` files selected by `go list -json std` for each
     GOOS/GOARCH
5. Reduced real-world issue and pull-request regressions
   - each case links back to its report in the conformance manifest
   - the reduced instruction must first be accepted by the native Go assembler
6. Executable semantic conformance cases
   - `testdata/conformance`
   - the same assembly is run once with the native Go assembler and once after
     plan9asm-to-LLVM translation

The x86 opcode and encoder tables are shared by 386 and amd64 and are therefore
a common inventory, not proof that every listed form is legal in both modes.
Native assembly of generated concrete cases supplies that mode check. Positive
testdata is useful but not complete: for example, the Go 1.27 386 corpus
observes only 21 opcodes from the shared 1,600-name x86 namespace.

## Coverage states

- name claimed: a lowerer contains the opcode name; this is informational only
- supported form: an instruction with this operand shape passes the real
  lowering pipeline
- context required: the instruction depends on a frame slot, a preceding flag
  write, a REP successor, or a function-local PC target and must be judged by
  full-file compilation
- unsupported form: parsing succeeded but the real lowerer rejected it
- runtime verified: a manifest case is compiled and run through both the Go
  assembler and plan9asm/LLVM with identical expected results
- compile-only policy: privileged, trapping, or environment-dependent
  instructions; these require parser, lowering, and object compilation checks
  but are never executed merely for coverage

The scanner reports name, encoder-table, form-lowering, context-required,
unsupported, and runtime-verified coverage separately. Compile-only is an
execution policy recorded by conformance metadata rather than a lowering
status. In particular, `XORB reg,reg` does not imply `XORB reg,mem`.

## Definition of complete

A runnable user-mode operand form is complete only when:

1. the Go assembler accepts its conformance source;
2. plan9asm parses and lowers it;
3. LLVM compiles the emitted IR;
4. both native Go assembly and the LLVM object execute and produce the same
   checked result; and
5. the form is listed in `testdata/conformance/manifest.json`.

A context-dependent form must additionally pass the full standard-library or
package corpus compiler. A compile-only form must have an explicit reason and
must pass object compilation where the target supports it.

Complete architecture support means every encoder-table operand class in the
Go 1.20-to-latest union has at least one native-Go-accepted concrete case, zero
unexpected parser failures, and zero unsupported runnable forms. Every safe
user-mode form must compile, link, and execute against the native Go oracle;
compile-only forms must carry an explicit reason and still compile and link.
The complete opcode and encoder-form catalogs remain visible even where
implementation or concrete-case generation is still pending.

## Commands

Generate a complete JSON opcode, encoder-table form, and observed concrete-form
report:

    go run ./cmd/plan9asmscan \
      -corpus go-asm \
      -goroot "$(go env GOROOT)" \
      -goos linux \
      -goarch amd64 \
      -repo-root . \
      -format json \
      -out /tmp/go-asm-amd64.json

Generate the human-readable family summary by changing `-format` to
`md`. Repeat with `386`, `arm`, and `arm64`.

Run the cross-version regression gate:

    scripts/check-go-asm-coverage.sh

Run the selected standard-library corpus through parser, form-level lowering,
whole-file translation, and LLVM object compilation:

    scripts/check-stdlib-corpus.sh

Run the executable semantic cases:

    go test . -run 'Test.*Conformance' -count=1

## Go 1.20 through Go 1.27 differences

The checked baseline stores a fingerprint for every Go minor version and
architecture. It detects any form changing among supported, context-required,
unsupported, or parser-failed states.

This version range describes the union of official namespaces, encoder tables,
and assembler source corpora, not the minimum compiler version of every command
submodule. The executable conformance cases are shared when Go versions accept
the same form; they are not duplicated once per Go release. Linux/amd64 CI
compiles, links, and runs the root conformance suite with every Go version from
1.20 through 1.27. `cmd/plan9asm` and `cmd/plan9asmll` retain their own Go 1.24
module requirement, so Go 1.20 validation intentionally does not enter those
nested modules.

The encoder-table union across Go 1.20 through Go 1.27 is:

| GOARCH | encoder opcodes | encoder operand forms |
|---|---:|---:|
| 386 | 1603 shared x86 opcodes | 4998 shared x86 forms |
| amd64 | 1603 shared x86 opcodes | 4998 shared x86 forms |
| arm | 144 | 528 |
| arm64 | 1255 | 4396 |

These are abstract encoder forms, not runtime support claims. The larger ARM64
union is intentional: operand-class names and generated SVE encoders evolve
between releases, so using only the newest tree would lose forms that existed
in earlier supported releases.

The Go 1.27 snapshot currently reports:

| GOARCH | official names | encoder forms | observed ops | observed forms | supported | context | unsupported | runtime verified | parse failures |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 386 | 1600 shared x86 names | 4997 shared x86 forms | 21 | 60 | 37 | 6 | 17 | 0 | 0 |
| amd64 | 1600 shared x86 names | 4997 shared x86 forms | 1456 | 6744 | 699 | 6 | 6039 | 23 | 0 |
| arm | 181 | 528 | 135 | 499 | 295 | 34 | 170 | 0 | 0 |
| arm64 | 1417 including SVE | 2964 | 1268 | 1908 | 379 | 36 | 1493 | 0 | 93 |

These numbers describe current implementation progress, not completion.
Encoder forms use Go's internal operand classes and are a complete machine-
readable inventory of the encoder rows; observed forms use plan9asm's concrete
shape classification and can outnumber encoder rows because generated testdata
varies registers, address shapes, and concrete encodings. The large amd64 gap
is mostly the exhaustive legacy/SIMD/AVX test matrix. Go 1.27 adds the ARM64 SVE
corpus and encoder tables and exposes the currently unsupported SVE parser and
lowering surface explicitly instead of hiding it.

The machine-readable cross-version snapshots are in
`testdata/coverage/go-asm-baseline.json`. A CI mismatch is blocking and
must be investigated at form level before the snapshot is intentionally
updated.

## Adding or changing an instruction

1. Locate the opcode and operand classes in the official encoder inventory.
2. Locate or generate native-Go-accepted concrete cases for every relevant
   operand class and classify the opcode family.
3. Add parser and lowering unit tests.
4. Add or extend a runnable conformance routine and manifest entry.
5. Run the native Go and plan9asm/LLVM semantic checks.
6. Run the standard-library corpus gate.
7. Generate all four architecture reports and inspect changes.
8. Update the cross-version fingerprint only after the change is understood.

For third-party failures, first reduce the source to its official operand form,
record the issue URL in the conformance manifest, and verify that the native Go
assembler accepts it. Fix the instruction family once and add a semantic
conformance case rather than adding a project-specific workaround. The
`xgo-dev/llgo#2464` regression for `XORB reg,mem` and `PUNPCKLQDQ` is the first
case tracked this way.
