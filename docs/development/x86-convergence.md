# x86 architecture convergence

The current priority is to finish **amd64** architecture, correctness and
efficiency before expanding another architecture. Changes to shared x86 code
must retain 386 regressions. Existing ARM, ARM64 and WASM gates remain enabled;
this ordering does not remove supported targets or their tests.

## Acceptance, not an opcode-count target

The review of `728520def` found 33/55 selected Linux/amd64 standard-library
files produced LLVM 22 objects, versus 55/55 at `e2164ac0`. These are historical
Go 1.27.2 observations, not execution certificates. Retain the original failures
and repair ordinary-source regressions before declaring this work complete.
Runtime entry contracts remain a separate integration responsibility; a source
failure must not be assigned there without identifying its actual cause.

Completion requires all of the following:

- An independent inventory of the selected Go assembler's legacy, AVX and
  special encoding rows, with their order, attributes and mode restrictions.
  Opcode names, operand-class buckets and corpus occurrences are different
  denominators. Neither the production registry nor a changed baseline may
  define away an unimplemented official row.
- Explicit ownership of every row: a validated concrete form, an independently
  observed frontend refusal for that mode, or an unresolved gap. Preserve
  overlapping rows. A syntax witness alone does not identify its actual VEX or
  EVEX encoding; compare original bytes when claiming that encoding's coverage.
- Initialized complete functions through the public APIs, LLVM 22 objects and
  independent semantic tests. Keep those evidence levels separate. Context,
  unavailable hardware and unfinished tests never count as successful execution.
- Whole-source standard-library and independent regression gates alongside
  family tests. External libraries remain independent checks, not the source
  from which we enumerate instruction formats.
- Measured code reduction and translation efficiency for each migrated family,
  without decreasing its retained format, target or regression inventory.

## Architecture boundary

Normalize named and decoded instructions into architecture-local typed forms.
Keep encoding identity and original operand roles. Reuse the validated form
for numerical dependencies, register/FLAGS effects and lowering; do not make
each consumer rediscover the family and parse its operands again.

Separate the following axes, constraining their legal combinations with the
official rules rather than constructing an unrestricted Cartesian product:

- Operation and arithmetic mode, including signedness and saturation.
- Register carrier, active vector, element, source/result and memory widths.
- Explicit/implicit input roles and old-destination dependencies.
- Encoding, masks, broadcast and preserved/zeroed register aliases.
- Memory alignment, access/fault suppression, ordering and machine-state effects.
- Source mode and the features required by the chosen lowering route.

Portable arithmetic does not automatically require the source opcode's ISA.
Native memory and floating-point paths may need exact instructions to preserve
faults, rounding or exceptions. Keep those deliberate exceptions visible.

Grammar and numerical effects describe requirements; they do not establish
pointer extent, initialized bytes, frame ownership or caller authority. Evolve
the existing analyses toward shared typed effects with separate numerical and
address-origin domains. Do not add another instruction whitelist to each proof.
Physical indexed SP scratch cannot be relabeled as named logical AUTO storage.

Prefer small concrete descriptors and direct dispatch. Avoid a runtime grammar
interpreter, reflection-based dispatch, or a universal struct of unrelated
optional fields. Common lowering pipelines retain family-specific operations
and explicit memory contracts. Delete replaced code in the same coherent batch.

## Work order and measurements

1. Reproduce and repair ordinary-source regressions with independent tests.
2. Pilot shared packed add/subtract/low-multiply lowering, retaining the complete
   widths, masks, memory behavior and aliases. Extend only after the pilot
   demonstrates a smaller implementation and unchanged observations.
3. Complete the independent rule inventory and initialized-form joins; migrate
   registration, feature inference and effect consumers family by family.
4. Move repeated source/tool capture out of numerical lowering using explicitly
   owned immutable inputs. Do not cache mutable verification results or weaken
   mutation, expiry, cancellation and Close checks.
5. Finish all amd64 families and whole-source gates before generalizing the
   architecture to ARM, ARM64 or WASM.

For each batch record the exact revision, SDK, LLVM version, source fixture,
successful/failed counts, production/test line delta and retained case IDs.
Measure parsing, analysis, feature inference, IR generation and public
translation separately, including allocations and resource release. Keep
external commands out of microbenchmarks; measure their cost in the complete
source/object workload as well. Never report early rejection as a speedup.

Benchmark identical inputs before and after, with repeated samples and unchanged
tools. Preserve vector IR and independently check generated-program performance.
Do not clear global compiler caches. Batch LLVM tests with bounded modules and
prompt disposal; batching must preserve every test case, not sample the suite.

## Independent rule inventory

The official scanner includes the selected SDK's ordered `asm6.go` `optab` and
`ymovtab` under `legacy_encoder` for both amd64 and 386. It retains duplicate
operand rows, nil-ytab special entries, sentinels, full opcode payloads and
symbolic attributes separately from the existing AVX inventory. This is a
declaration inventory, not proof of mode applicability or selected encoding.

```sh
go run ./cmd/plan9asmscan -corpus=go-asm -goos=linux -goarch=amd64 \
  -format=json -out=_out/amd64-official-rules.json
```

`table_row_fingerprint` hashes normalized retained table rows. It deliberately
does not claim to hash constant definitions or all assembler selection logic.
No original corpus classifications or AVX fingerprints change when this
inventory is added. Row-to-initialized-form joins and original-byte selection
witnesses remain separate work; do not count inventory rows as passing tests.

## First pipeline pilot and next gate

The shared E4 pipeline retains family-specific parsing and storage checks, then
handles operand reads, arithmetic, masking and encoding-aware writeback once.
Legacy memory still uses its original faulting opcode. MMX and word low-product
multiplication retain separate owners. The three migrated lowering files plus
their helper shrink from 403 to 346 lines; this is not the whole PR's line delta.

The selected Go 1.27.2 tables contain the following independent pilot rows:

| Family | Legacy XMM | AVX | Total |
|---|---:|---:|---:|
| Wrapping/saturating add | 8 | 64 | 72 |
| Wrapping/saturating subtract | 8 | 64 | 72 |
| D/Q low product | 1 | 14 | 15 |
| Pilot | 17 | 142 | 159 |

Seven PADD MMX rows and ten PMULLW/VPMULLW rows are outside this migration,
not unsupported or newly verified. PADDD's frontend alias adds no encoder row.
The next gate must retain row identity rather than deduplicate instruction text.
Force AMD64 EVEX X/Y encoding with high registers, and independently check the
actual Go instruction bytes before associating a witness with that row. Reuse
the existing witness grammar and public three-API object harness in bounded
32-function batches. Initialized source, matching encoding, public IR, objects
and runtime remain separate outcomes; no inventory count is a passing outcome.

The new result-mask helper emits five vector IR instructions instead of a
lane-wise sequence. This is not a five-machine-instruction or constant-runtime
claim: LLVM may split its predicate vectors. Shared splat formatting has constant
allocation count but necessarily linear output size. Keep the independent
runtime oracles and measure generated-program performance on actual AMD64
hardware before claiming that whole-axis improvement.
