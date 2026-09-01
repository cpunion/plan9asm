# Plan 9 assembly instruction coverage

This document defines the path to complete support for the Go assembler's Plan
9 syntax. Coverage is measured by architecture, instruction family, opcode,
and operand form. An opcode name by itself is not a support claim.

The target architectures are 386, amd64, arm, and arm64. The 386 and amd64
lowerers share x86 implementation code, but their registers, address widths,
valid opcodes, and runtime ABI differ. ARM and ARM64 are separate instruction
sets. ARM64 also distinguishes scalar, NEON, and SVE families.

## Sources of truth

Coverage uses four increasingly strong layers:

1. Official opcode namespace
   - `cmd/internal/obj/x86/anames*.go`
   - `cmd/internal/obj/arm/anames*.go`
   - `cmd/internal/obj/arm64/anames*.go`
2. Official positive assembler testdata
   - `cmd/asm/internal/asm/testdata`
   - this supplies concrete, architecture-valid operand forms
3. Selected standard-library assembly
   - lower-case `.s` files selected by `go list -json std` for each
     GOOS/GOARCH
4. Executable semantic conformance cases
   - `testdata/conformance`
   - the same assembly is run once with the native Go assembler and once after
     plan9asm-to-LLVM translation

The x86 opcode table is shared by 386 and amd64 and is therefore a namespace,
not proof that every listed name is legal on both architectures. The separate
386 and amd64 assembler testdata supplies the valid operand forms.

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
- non-runnable: privileged, trapping, or environment-dependent instructions;
  these require parser, lowering, and object compilation checks but are never
  executed merely for coverage

The scanner reports these states separately. In particular, `XORB reg,reg`
does not imply `XORB reg,mem`.

## Definition of complete

A runnable user-mode operand form is complete only when:

1. the Go assembler accepts its conformance source;
2. plan9asm parses and lowers it;
3. LLVM compiles the emitted IR;
4. both native Go assembly and the LLVM object execute and produce the same
   checked result; and
5. the form is listed in `testdata/conformance/manifest.json`.

A context-dependent form must additionally pass the full standard-library or
package corpus compiler. A non-runnable form must have an explicit reason and
must pass object compilation where the target supports it.

Complete architecture support means zero unexpected parser failures and zero
unsupported runnable forms in the selected official Go versions. The complete
opcode catalog remains visible even where implementation is still pending.

## Commands

Generate a complete JSON opcode and operand-form report:

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

## Go 1.21 through Go 1.27 differences

The checked baseline stores a fingerprint for every Go minor version and
architecture. It detects any form changing among supported, context-required,
unsupported, or parser-failed states.

The Go 1.27 snapshot currently reports:

| GOARCH | official names | observed ops | operand forms | supported | context | unsupported | runtime verified | parse failures |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 386 | 1600 shared x86 names | 21 | 60 | 37 | 6 | 17 | 0 | 0 |
| amd64 | 1600 shared x86 names | 1456 | 6744 | 699 | 6 | 6039 | 23 | 0 |
| arm | 181 | 135 | 499 | 295 | 34 | 170 | 0 | 0 |
| arm64 | 1417 including SVE | 1268 | 1908 | 379 | 36 | 1493 | 0 | 93 |

These numbers describe current implementation progress, not completion. The
large amd64 gap is mostly the exhaustive legacy/SIMD/AVX test matrix. Go 1.27
adds the ARM64 SVE corpus and exposes the currently unsupported SVE parser and
lowering surface explicitly instead of hiding it.

The machine-readable cross-version snapshots are in
`testdata/coverage/go-asm-baseline.json`. A CI mismatch is blocking and
must be investigated at form level before the snapshot is intentionally
updated.

## Adding or changing an instruction

1. Locate the opcode in the official catalog and its positive assembler cases.
2. Classify the opcode family and every operand form being added.
3. Add parser and lowering unit tests.
4. Add or extend a runnable conformance routine and manifest entry.
5. Run the native Go and plan9asm/LLVM semantic checks.
6. Run the standard-library corpus gate.
7. Generate all four architecture reports and inspect changes.
8. Update the cross-version fingerprint only after the change is understood.

For third-party failures, first reduce the source to its official operand form.
If the form already exists in Go assembler testdata, fix the family once and
add a semantic conformance case rather than adding a project-specific
workaround.
