# Instruction-family development

Read the repository [rules](../../AGENTS.md) first. This guide describes the
workflow, not a list of already completed instruction families.

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

Treat an external-library failure as a counterexample to the systematic
coverage mechanism, not only as a package-specific regression. Identify where
the independent Go rules, typed grammar, lowering/effect consumers or executed
tests diverged; then enumerate and repair the entire affected rule class.
Add a regression that still finds the defect without that external library.
Rerun both the complete affected class and the original discovering sample.

Keep the Go-table oracle independent of the implementation spec. Retain each
encoding row's prefix, zcase, byte offset and opcode-specific attributes as well
as its operand classes. A shared ytab or a name/row count is not a complete
format check. Concrete syntax witnesses do not prove which overlapping encoder
row Go selected, LLVM object success, runtime semantics or an ABI contract.

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

Prefer typed, table-driven grammars over mnemonic-specific branches. Separate
orthogonal axes such as operation, lane/stack width, operand class, encoding
generation, architecture mode, masks/broadcast/rounding, implicit registers,
and semantic effect. The parser/validator, register-liveness scan, lowering,
supported-op extraction, and completeness test should consume the same spec
where practical. A completeness test must compare the whole spec table with
the current Go encoder family so adding one observed spelling cannot leave
sibling forms unmodeled.

Macro expansion uses complete Go assembler identifier tokens, including middle
dot and division slash, never prefixes or text inside string/character literals.
Unresolved constants need context for an isolated instruction probe, but must
fail full translation before raw decoding. Preserve resolved named stack
addresses; never replace missing constants or malformed addresses with zero.

ARM64's fresh virtual SP allocation is disjoint from incoming pointers only
when the complete original and normalized source excludes escaping its address
and entry transport cannot replace SP. A source CFG proof may admit exact
affine GP temporaries used only for bounded, non-writeback local reads and
fully overwritten before calls or returns. Mixed/different-offset joins keep
may-frame taint; unknown effects, indexing, address stores and live call/result
transport cannot wash it away. Under that bounded proof,
ordinary pointer stores and pure escaped code-address values cannot corrupt
saved SP cells. Actual source arithmetic/transforms on a code address must
retain derived-code taint even after unknown CFG joins; filtering ordinary
data cannot turn a computed native PC into a pure-code alias proof.
Explicit/aliased/indexed SP writes, FP contents, unknown callouts,
raw/native effects and frame-address escape remain conservative. Audited
same-file register helpers retain the original proof across local-call
rewriting; a bare source-local call never acquires it. Scalar and paired atomic
stores must account for their memory operand as well as GP outputs.

Raw-containing framed ARM64 source-leaf functions have a separate bounded
physical-SP proof, checked on both original and normalized CFGs. Constant SP
movement must balance on every source RET; different-delta joins fail. R29 and
the automatic saved-FP cell start as unknown may-frame values. Known scalar and
GP/FP pair memory forms preserve private eight-byte pointer spills and reject
partial overlap, unknown indexing/writeback and SIMD pointer transport. Only
the actual Go framed-leaf RET epilogue consumes R29; continuation/LR checks stay
independent. External stores, live result aliases, calls, hardware RETs and
unregistered effects cannot borrow this proof. Go's physical autosize is not
the LLVM private-allocation size. Runtime oracles must execute original raw
Go objects as well as LLVM objects with independent references and canaries.

ARM64 VMOV/VDUP availability records the validated source lane span separately
from implementation loads that merely preserve other destination bytes. Lane
insertion defines only the written bits; F/V aliases and CFG joins retain the
unknown remainder, and a full vector consumer still needs all of its bits.
Scalar/arranged non-preserving transfers clear their architectural inactive
bits. Typed calls and source scratch invalidation revoke prior lane definitions.
This is numerical definedness, not caller, pointer or private-frame permission.
VDUP's arranged case79 derives both source width and index bounds from its
destination; the source suffix identifies only the ELEM class. VMOV and scalar
VDUP retain their separate suffix-selected bounds.
Lane syntax shares Go's single decimal-token rule after CPP expansion: trim
lexical whitespace, reject signs/expressions/nondecimal or unresolved tokens,
and retain every instruction's independent physical width/index bound.

The eight mask/vector B/W/D/Q moves share a register-only EVEX grammar with
X/Y/Z widths and K0-K7 data operands, never writemasking or suffixes. Caller
effects require initialized mask/vector inputs; forward writes clear every
wider alias, while reverse writes zero-extend the mask remainder. Use the
encoding-aware packed writeback even for X destinations: legacy `storeX`
preserves wider bytes. Independent whole-byte runtime checks must observe both
directions separately rather than relying only on a round trip.

The eight immediate shuffles share their integer-only source grammar and exact
byte selectors with lowering, caller definedness and memory shapes. Bind the
complete legacy and VEX/EVEX rows independently, including signed MMX Yi8,
unsigned legacy X Yu8, mode32 int32 source classification and raw imm8 signed
rendering. Float/unresolved operands fail before every consumer. Legacy X
writes preserve upper aliases; VEX/EVEX clears them. E4NF reads retain the full
tuple even with dead results or zero masks; dword broadcast reads m32.
Legacy m128 retains the original observable SSE2 memory opcode, not an align-1
replacement. Virtual FP/SP/BP and unbound address-size roles remain Context,
as does caller alignment without a real witness. Execute independent original
Go opcode bytes and LLVM live/dead boundary controls. Label IR-prepass axes
separately from backend optimization; MMX objects and Rosetta diagnostics do
not certify physical x87/MMX state or native AVX512 execution.
Go 1.20's maskless `.Z` acceptance is a historical encoder defect, not another
legal format. Observe its exact reserved EVEX.z/aaa=0 bytes and still require
both named and raw rejection through all APIs; current Go must reject source.

An x86 raw-return caller-region producer is a separate, bounded source proof,
not a pointer-type or `noescape` nonalias promise. It requires the same actual
selected Go/ASM inventory, typed syntax, actual numerical/pointer sizes and
source bytes, closes every use of
an unexported bodyless ABI0 entry, and accepts only direct slice-data inputs
and independent numerical result/length locals. Missing sources, indirect or
assembly callers, linknames and local-address escape remain Context. A region
witness alone never admits raw RET: the complete source stack/CFG/effect proof
must separately consume the exact slots and write envelope. Compiler loaders
must supply the actual complete selection; an AST subset or self-reported hash
is not authenticated package-selection evidence.

The bounded AMD64 caller-region consumer covers scalar MOVB/W/L/Q and the
Go-table MOVSS/SD, VMOVSS/SD and KMOVB/W/D/Q move forms. It tracks exact
full-word pointer imports/copies, constant affine offsets and numerically
initialized GP/SIMD bytes. Mask values must be source-defined, not pointer bits;
legacy lane preservation and VEX/EVEX merge/zero rules remain distinct.
Floating literals follow the actual Go frontend rewrite: original positive
zero in X0-X15 becomes legacy XORPS, preserving wider aliases; negative zero,
float32 rounding/underflow and high-register constants follow their pool-memory
encoding. Integer/unresolved literals and invalid mask/merge forms remain
rejected. Legacy scalar ADDSS/ADDSD/SUBSS/SUBSD and MULSS/MULSD/DIVSS/DIVSD
use their own finite FLOAT constant-pool rewrite, including both signs of zero;
they do not inherit MOV's positive-zero XORPS rewrite. Their FLOAT leaf may
have unary signs and parentheses, but not integer or binary-expression source
syntax. The destination's low scalar bytes must already be numerically defined.
Raw FP16, packed moves and arithmetic need separate complete effect
families and do not inherit this permission.
Inputs and local lengths are read-only;
only the independently fresh scalar result has a bounded write envelope.
Instruction-local memory certificates bind the entire operands, direction and
native width. Every source return must still restore SP/BP, and unknown joins,
private-stack escape, g-register use, unregistered vector/native effects and stale source or
signature metadata remain Context. Rosetta fixture execution is diagnostic,
not evidence that an external AVX/AVX512 library ran on native AMD64 hardware.

Signed packed dword-to-float conversion uses one validated whole native
CVTDQ2PS/PD opcode, retaining ambient MXCSR rounding, precision, trapping,
masking and embedded rounding. Its caller effects require every potential
input byte and merge byte; masks do not authorize a missing memory envelope.
Packed FP fields and virtual SP/BP lack original native backing and remain
Context. Mode32 high banks likewise cannot borrow a software register view.
Go's MMX-to-X forms (including GP spellings admitted to Ymm) remain Context:
no current producer closes physical MMX identity, x87 Tag/TOS, pending
exceptions and observations by the caller after RET. They are Go-object and
format coverage, not successful LLVM objects. The native Linux AMD64 fault
oracle is a required cross-runtime root; SSE numerical Rosetta observations
and five-target AVX512 object compilation do not replace it or AVX512 execution.

The CLI captures a private inventory at its actual package-load boundary and
rechecks complete compiled Go sources with that load's dependency types and
target sizes only for raw-return candidates. All selected assembly and CPP
inputs are included, with source/expansion checks through the final object.
Effective overlays, generated/cgo mappings and external package drivers require
their actual compiled-source mapping; original files cannot impersonate them.
This private load snapshot cannot be restored from JSON or supplied by an old
standalone `compileOne` caller. Candidate detection is only a cost predicate.

ARM32's specialized raw-data lowerers and NZCV source proof share one typed
registration. Only those decoded data forms preserve source-defined APSR
flags; conditional forms still require their predicate on every CFG path.
Their GP outputs conservatively kill source-value witnesses, including memory
writeback. Unknown WORDs, calls and unproved FPSCR-to-APSR transfers do not
retain flags. Check every registered family and condition against LLVM-MC,
Go objects and the required ARM Go-WORD/LLVM four-flag execution oracle.

For a newly discovered instruction, update all four layers before calling it
complete: the family-specific lowerer, positive/negative Go-table form tests,
LLVM 22 object compilation for every affected supported architecture, and the
architecture's supported-op extraction test in `cmd/plan9asmll/main_test.go`.
The external corpus uses that extraction result in its diagnostics, so omitting
the last layer can make a supported instruction look unsupported.

ARM64 named system registers share the raw encoding grammar. Regenerate their
encoding/access metadata with current Go using `go generate arm64_sysregs.go`;
do not edit `arm64_sysregs_generated.go` manually. The completeness regression
compares the full current Go table and checks every readable register against
LLVM 22. Use physical system-register names to avoid feature-gated LLVM aliases,
never fabricated constants for CPU-ID reads. Keep text/direct-module/CFG/raw
paths observable and require the Linux MRS runtime oracle: only the kernel's
HWCAP_CPUID-advertised ID-register space is safe for unprivileged execution.

The seven ARM64 scalar GP memory moves share `parseARM64ScalarMemoryForm`:
signed/unsigned widths, direct/indexed addressing, pre/post writeback, AUTO/FP
and SB roles. Go's long-offset expansion reads the old effective address,
materializes R27, then reads the possibly aliased transfer register. Lowering,
caller-link/code provenance, private-frame bounds, availability and raw-PC width
must consume that same ordered model. Only actual WORD decoding supplies a
physical-instruction identity. AUTO/FP classes require the complete source TEXT;
LLVM shadow slots are not native SP or linked-page identity. Differing old/new
Go encoders require source context when their scratch value is observed, never
a guess from the translator's build version. Preserve the independent Go-table,
byte/state, initialized numerical/canary and complete target-object tests in
`arm64_scalar_memory_*_test.go`.

Isolated ARM64 coverage probes parse their own explicit zero-frame NOFRAME
TEXT fixture and retain a real faulting exit. This validates concrete AUTO
forms without borrowing the original function's frame or encoder identity.
Foreign TEXT, unbound FP, unresolved offsets and unsafe continuations remain
rejected; a compile-only probe is never executable semantic coverage.

Compiler/executable/emulator subprocesses in native root oracles should use
`runtimeOracleCombinedOutput` with their actual `CommandContext` and deadline.
The helper retains output and failures, bounds pipe waits and cleans its owned
Unix process group after cancellation or ordinary parent exit. Non-Unix owns
only the direct child. Extend the lifecycle AST regression when adding a new
consumer; do not introduce another direct `CombinedOutput` call that bypasses
the helper. The reusable LLVM compile, C link and executable-run stages each
receive a fresh five-minute budget, limited by the existing test deadline minus
three seconds for cleanup; this is not a five-minute whole-suite budget. An
exhausted budget fails without launching another command. Release LLVM modules
before their contexts and before each fixture's temporary files are removed.

A typical x86 investigation starts with the Go 1.27 tables and an assembler
probe:

```sh
goroot=$(go env GOROOT)
rg -n 'A?VADDPS|yvaddps' "$goroot/src/cmd/internal/obj/x86"
sed -n '<start>,<end>p' "$goroot/src/cmd/internal/obj/x86/avx_optabs.go"
GOOS=linux GOARCH=amd64 go tool asm -o _out/form.o _out/form_amd64.s
GOOS=linux GOARCH=386 go tool asm -o _out/form.o _out/form_386.s
```

Record the initial failure in test output before adding the lowering. Focused
form tests normally belong in `amd64_ecosystem_test.go`,
`amd64_operand_forms_test.go`, or the corresponding ARM file. Add the opcode to
`TestExtractSupportedOpsFindsCompleteAddedInstructionFamilies` so external
corpus diagnostics recognize it. Semantic fixtures live under
`testdata/conformance/<arch>` and must be checked against both native Go and the
LLVM runtime oracle where supported.

Raw x86 far/interrupt returns must select the existing `IRETW/L/Q` and
`RETFW/L/Q` grammar from the decoder's actual mode and operand size. A bare
`CB` remains a 32-bit far return even in amd64 mode; widthless `LRET` is not
a source-level Go opcode or evidence for a 64-bit return. Preserve encoded
imm16 operands, reject unregistered prefixes, and compare original Go and
LLVM object bytes for all five supported x86 targets. Those object checks
do not prove a privileged kernel's far-return stack or runtime behavior.

Named x86 addresses retain the pseudo-base role separately from subsequent
physical register groups: `slot(FP)(SI*1)` supplies an index, while
`slot(FP)(SI)` supplies Reg, which Go rejects on amd64 and ignores on 386.
Scale spellings are Go's single decimal tokens 1/2/4/8, not arbitrary constant
expressions. A parsed indexed FP/AUTO operand does not establish a contiguous
frame, dynamic bounds, or address provenance and remains Context without those
contracts. Ordinary amd64 SB memory cannot combine RIP with a base/index.
For 386 SB memory/LEA, Reg is ignored but Index is used; a non-shared relocated
SB address immediate ignores both. Keep these effects and the decoder's
physical SP operands distinct. Alternate pseudo-role overwrite forms remain
explicit Context rather than being misclassified as Go-rejected syntax.
The complete Go x86 register-name inventory is checked before discarding
unused 386 SB annotations; this never grants a segment/SIMD register input.
TLS memory progedit and segment Index operands remain Context without their
native address/source-encoder contracts. In particular TYPE_ADDR segment Index
prefixes differ in the actual Go 1.20 and 1.27 encoders; no build-version guess
selects their meaning.
Grouped SB CALL/JMP/RET operands have Go's TYPE_BRANCH role, not a native
indexed read; they remain Context without a bound branch/entry contract.

Unnamed x86 source register groups preserve a missing second-group scale as
zero. This is distinct from an explicit invalid multiplier token. MOVL/Q
address sources become LEAL/Q in Go's progedit before operand classification;
those derived LEAs and ordinary physical memory must reject unencodable SIB
indexes/scales. Audited TYPE_CONST rows instead ignore Reg/Index annotations.
Do not apply the ordinary SIB guard to named, segment/TLS, branch or decoded
physical roles, or let source normalization repair an original CFG certificate.
Unnamed address-immediate register groups retain the source `g` spelling until
the actual source mode is known. Go registers this alias only for amd64; 386
still accepts explicit `R14` annotations that its TYPE_CONST rows ignore.
Qualifying the alias grants neither a register input nor a native address.

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

For named ARM64 private data tables, the explicit typed-call transport and
fresh-frame proof may authorize bounded indexed accesses and counted-loop
writeback. Track numeric bounds separately from may-frame provenance: joins,
stale comparisons, unknown transforms and hardware-masked shift counts must
never restore an erased address or narrow an unknown index incorrectly.
Certify each store's complete byte envelope within private DATA `[8, frame+8)`;
the saved caller link at SP+0 and overlapping cells remain protected. This is
an instruction-local certificate, not a function-wide unknown-store exemption.
Reject address escape through calls, results, FP, memory or vector banks, raw
effects, mixed origins and exhausted proof budgets. Exercise the shared scalar
and GP/FP pair grammar with Go-accepted widths, signed loads, zero aliases,
indexed and offset/pre/post forms, LLVM objects and initialized runtime
canaries. An old fixture with undefined data establishes only its explicit
object contract; never report it as successful runtime coverage.

For ARM64 GP data-processing proof, use the lowerer's typed operand grammar:
explicit Rd is a full overwrite, but two-operand arithmetic/shift forms and
BFI/BFXIL read it. W results clear the upper GP half; extracts and extensions
still consume their actual source. Unknown, aliased or joined frame inputs
cannot be washed away. Preserve transformed code-address taint and require
independent exact LR/SP/BP/g/OS-register restoration, even for typed void
returns. Go optab rows do not automatically establish frontend spellings:
single-operand NEG is rejected by stock Go, and a cross-bank W bitfield
accepted by its alias encoder is not an overwrite proof.

Scalar GP moves likewise distinguish their source from the write-only Rd.
Narrow register copies retain derived code-address bits; resolved MOVW/MOVD
literals and the seven literal-zero aliases do not read old Rd. MOVK remains
read-modify-write. Go's inherited nonzero MOVWU literal rows can emit zero
machine words; stock syntax acceptance alone grants no overwrite witness.

For ARM64 immediate arithmetic, logical and compare forms, follow the complete
Go constant classes (including MBCON), register classes and reserved row size.
R27 may be written before an aliased source is read; direct/split forms need
not write it. Raw decoded instructions must not acquire source-pseudo effects.
Go-accepted fixed-width expansion can contain a reachable UDF, and different
Go encoders can produce non-equivalent fallthrough. Preserve exact byte/fault
oracles and require source-encoder Context when that identity is unavailable.

The bounded raw zero-frame proof may also produce the local-control plan,
but only after every original and exact normalized source context passes.
Different helper invocations retain separate incoming/private pointer states;
their return labels are combined only after exact SP/LR/fixed-register and
escape checks. No partial certificate survives a failed component, unknown
effect or exhausted bound. Ordinary/native/unproved functions keep the coarse
control analysis. Runtime fixtures must execute both helper pointer contexts,
not just compile a synthetic return dispatcher.

An authenticated current-PC raw BL in a single-instruction block is a terminal
sink: LLVM still writes its continuation to LR and branches to itself. The
source proof checks all preceding effects, joins and other returns, and emits
no return certificate for the sink. Ordinary recursion, multi-instruction
helpers and actual dead source retain the existing bounds and guards; only an
empty block-splitter continuation after a final self BL has no source to check.
Floating immediates never acquire an integer GP MOV overwrite witness, including
through the legacy MOVD transfer fallback.
The seven scalar GP MOV lowerers also reject Go's distinct floating-constant
class before register, memory, symbol or writeback dispatch; FMOV is unchanged.

An ARM64 ABI0 native-width read may overlap the alignment gap after the final
direct narrow integer parameter only with a complete actual typed declaration
and independent original/normalized whole-source bit-use proofs. Every use of
the unspecified bits must remain unobservable across GP and SIMD aliases, CFG
joins, comparisons, addresses, stores and results. The bounded first contract
is call-free and rejects unknown effects. A proved read may choose an explicit
representative for those dead bits; it does not assert that native padding was
initialized, create a padding slot, or authorize any address or write. Preserve
the source MOV width, encoder scratch effects and PC identity. Manual sparse
frames and changed type layouts retain their original missing-slot failure.

A native-width store beyond a narrow declared result is different: its tail
bytes are observable by an assembly caller after RET. A callee declaration,
call-free body, or typed numerical oracle cannot authorize discarding them.
Dropping that caller with `KeepFunc` does not close the original boundary;
code-address and DATA references likewise grant no typed-entry permission.
Keep the missing-slot failure until a real caller/backing contract closes the
complete native effect. Check the independent nine-width native observer and
all three translation APIs when changing this boundary.

`TestARM64ABI0ResultTailCompilerCallerObservation` independently compiles an
ordinary Go caller and its ABI wrapper with the original eight-byte ASM store.
Both SDKs allocate outgoing storage but read only the declared uint32 result.
This observation is not proof that every caller ignores the tail: the separate
nine-width ASM observer reads it, and the current SDK's atomic entry has public
linkname and cross-package ASM tail-call roles. Run the current-source audit as
well as the cross-SDK fixtures. ABI0's rounded allocation size is not a writable
FrameSlot and must not erase the missing native caller/backing contract.

The framed raw-leaf proof registers PC-relative WORD branches through the same
exact source-boundary checker as translation. It independently checks both
original and normalized CFGs, including dead components; named normalized
edges must not reclassify skipped executable WORDs as anonymous data. Unknown
widths, native address/data islands, calls, escapes and unbalanced/fixed-register
returns remain rejected. This CFG registration supplies neither a callee ABI
nor permission to clobber Go's reserved R18/g entry identities.

The x/arch decoder corpus is not a Go-accepted instruction list. Its
`MOVHU.P 107(R13), R13` and `MOVHU.W 192(R2), R2` encodings violate Go's
base/data writeback-overlap check. The mixed MOVHU decoder form is therefore
unsupported, not a missing valid Go format. The independent
`TestARM64DecoderWritebackOverlapRejectedLikeGo` regression checks both
assemblers' rejection; `check-arm64-plan9-corpus.sh` separately requires every
Go-accepted selected-family case to remain lowerable.

The decoder also prints reserved ADD/SUB extended-register shifts of five to
seven. `TestARM64DecoderExtendedShiftBoundariesMatchGo` checks every W/X and
flags alias, all eight extensions and shifts zero through seven against the
actual invoked Go assembler, including the older compatibility toolchain.
Only shifts zero through four are accepted. Native register CALL/JMP forms
retain function-context classification: a decoder entry supplies neither a
callee ABI nor its native frame, continuation and result transport.

## Semantic and raw-encoding checks

Go's original `reflect.makeFuncStub` and `methodValueCall` have empty source
declarations but physical dynamic callers; their native TEXT retains
`ArgsSizeUnknown`. Local outgoing-call storage, the argument-only `argLen`,
and the GC pointer bitmap cannot bound the complete argument/result frame.
Keep these failures explicit. A real producer would need the actual function
type/layout, closure-context register and entry identity, caller/result
transport, lifetime, and the runtime's dynamic stack maps/`retValid` protocol.
A manual LLVM carrier fixture or ordinary `func()` signature is not that
producer and cannot make the default benchmark pass.

The Advanced SIMD SM3 family is WORD-only in Go 1.27. Its exact seven forms
follow [Arm DDI 0602](https://documentation-service.arm.com/static/67e40f3398aa3c3b6eea6a85)
and LLVM 22's crypto encoder: PARTW1/PARTW2, SS1 and TT1A/B/TT2A/B, with
four-word vectors and TT indices 0–3. The shared raw-vector registry records no
GP/SP or memory effects; destructive destination reads and all aliases remain
explicit. Portable LLVM arithmetic is checked separately from original Go
WORD execution, which is required on Linux/QEMU CPU=max. Host-only arithmetic
execution is not evidence that the host implements FEAT_SM3.

- Distinguish physical encoding rules from Go frontend acceptance, especially
  for 386 VEX/EVEX registers and ignored VEX.W bits. Check primary architecture
  specifications in addition to Go tables.
- Validate operand order, implicit inputs/outputs, destination aliases, flags,
  width extension and lane boundaries. Keep every overlapping X/Y/Z view
  coherent; preserve legacy upper lanes and clear VEX/EVEX upper lanes.
- Test dynamic masks, irrelevant high mask bits, unaligned addresses, null
  inactive sources and guard pages. Consume results so optimization cannot
  remove the memory access under test.
- Check count masking versus saturating shifts, signedness, NaNs and
  out-of-range conversions. Avoid LLVM poison and undefined arithmetic.
- Native access width is not internal register-slot width. A 386 Q spelling
  can still require a four-byte memory read.
- Validate independent frame slots and spans; do not relax ABI checks just to
  compile an external library.
- Unknown raw VEX/EVEX instructions must fail closed, never become unrelated
  legacy instructions. Test truncated and reserved encodings.
- Demonstrate that the old implementation fails the new runtime oracle where
  practical. Use an isolated checkout/overlay, not a running frozen tree.

Use independent-axis tests for shared operand grammars rather than repeating
the same huge Cartesian product for every mnemonic. See
[validation](validation.md) for the required gates.

The packed unpack and logical families share only their binary operand grammar;
operation effects remain separate. Unpack interleaves bits, including floating
NaN payloads, within each 128-bit lane and keeps legacy upper aliases distinct
from VEX/EVEX upper-zeroing. Intel SDM Vol. 2A Tables 2-44/45 classify EVEX
unpack memory as E4NF: even a zero mask retains the complete native access.
MMX low unpack reads four bytes; MMX high unpack reads eight. FP spans must
cover all independent fields before emission. Go-accepted 386 names above
X7/Y7/Z7 need a real physical mode32 binding; keep their actual Go encodings
and Context regressions rather than interpreting them as ordinary virtual
registers. Consult the [Intel SDM](https://www.intel.com/content/www/us/en/developer/articles/technical/intel-sdm.html)
and preserve zero-mask/guard-page and complete-byte runtime oracles.

The zero-operand VZEROUPPER/VZEROALL family has implicit SIMD writes, not no-op
semantics. Its shared grammar/lowerer/caller effect uses registers 0-15 on
amd64 and 0-7 on 386, preserves registers 16-31, and distinguishes preserved
low X bytes from known-zero wider bytes. Raw VEX.L selects the two operations;
reserved pp/vvvv and forbidden prefix combinations must not inherit the pinned
decoder's incorrect VZEROUPPER spelling. Keep full-bank alias, flags, MXCSR and
mask canaries. See Intel SDM Vol. 2C VZEROALL/VZEROUPPER in the linked manual;
portable LLVM execution does not prove native AVX execution.

The six unsigned packed widening ratios have a separate PMOVZX/VPMOVZX typed
grammar and caller byte effect; their existence does not authorize signed
PMOVSX callers. Bind legacy `yxm_q4` and both complete AVX operand tables, the
shrinking source tuple, K1-K7 and `.Z`; low-bank masked 386 forms are accepted
by Go, while high named X/Y registers still need a physical mode32 contract.
Read independent FP fields byte by byte, rejecting any missing byte before
emission. Legacy writes preserve coherent wider aliases; VEX/EVEX writes zero
above the destination width. Intel SDM PMOVSX/PMOVZX distinguishes VEX WIG
from EVEX's DQ-only W0 requirement. Keep complete-byte, NULL/guard-page and
flags/MXCSR oracles: E5 inactive lanes suppress source reads, but a known mask
does not itself grant caller memory bounds. Original legacy execution and
portable LLVM VEX/EVEX semantics are separate from native AVX512 execution.

`CaptureGoAMD64FrameSource` is an initial, sealed **observation**, not a frame
permission. It independently selects and exports a flat main module's complete
ordinary Go sources, retaining only bodyless integer, finite pointer and slice
assembly declarations. Stock-SDK imports bind the actual compiler-selected
source/export graph; bodies, values and imports themselves grant no native
permission. Foreign module dependencies still need their own source handoff.
The observer then binds
all selected ASM, headers,
actual Go tool/source routes, original object bytes and complete source positions.
The caller selects the trusted stock-Go installation and explicit private command
caches; a public selection record, manual signature or JSON cannot recreate its
token. Both conventional BP producers are recognized from actual machine bytes,
not from a Go version guess. Unknown generated/foreign/test/conditional/unwind
roles remain Context, including current Windows auxiliary SEH records. Every
Verify rechecks bytes, complete inventory and actual export/selection before and
after commands; failure or Close revokes the token. The observer alone has no
allocator, caller certificate, native-memory grant or compiler adapter.
Pointer/slice observations retain the independent language type/model and
object-bound ABI0 data/len/cap slots, including pointee, element, named/alias RHS
and field identity. A go/types pointee model is not an observed native extent:
Go1.20 StdSizes omits compiler tail padding/zero-tail storage in some structs.
These observations grant no input extent, alias or GC contract. The numeric-only
fixed-local consumer still rejects composite transport. Its separate Binder
re-typechecks the complete original numeric declaration and retains that
language identity; each consumer checks it against the live factory declaration
and object role/signature before and after translation. By-value aggregates,
recursive/generic types, noninteger elements and nonordinary roles remain Context.
That fixed-local contract accepts either one initialized numeric result or a
typed void return with no result slots. Void does not exempt any original or
normalized DAG, local-byte initialization, native alignment, SP restoration,
BP/g, escape or source-lifetime check, and cannot authorize an FP input store.
Compile-only void object tests establish no numerical return or external-library
runtime coverage; pointer/slice observations still have no local-memory permission.
Factory artifacts use the request's effective explicit TMPDIR, not the parent's
temporary environment. Verify and Close check the created directory identity;
replaced or redirected directories are never recursively cleaned. The package
root is fixed to its resolved canonical source at capture: a later change to
the caller's original alias does not change that source. Explicit cache roots
remain caller-owned; capture does not authenticate ownership or delete them.
