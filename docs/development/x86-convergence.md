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

The first normalization batch covers four existing packed binary families:
average, word multiply, multiply-add and dword product (20 names). One typed
spec/form now feeds effects, memory queries, result routing and lowering;
the replaced consumer selectors are removed. Lowering retains the parsed form,
while separate analysis invocations still normalize independently. No cache is
attached to mutable instructions. Input/result widths, pair reductions and
even-lane products remain explicit: DQ arithmetic uses 32-bit inputs but 64-bit
memory granules; MADD retains full-tuple reads, distinct from E4 mask suppression.
Average and word multiply reuse the existing E4 pipeline. Production shrinks by
65 lines. Both endpoint SDKs pass 101 roots / 10,098 children without skips.
Paired benchmarks under unrelated host load do not establish acceleration.
Do not call this whole-translation parse-once completion or add a fifth family
before qualifying the shared function-representation lifetime.

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
The subsequent cleanup removes another 143 lines of unreachable packed
arithmetic from the residual vector dispatcher. Its original fixtures now call
the same typed owners as public lowering; 66 roots / 4,506 children passed with
no skips at `481291b93`. This is code reduction, not a measured speedup.

The selected Go 1.27.2 tables contain the following independent pilot rows:

| Family | Legacy XMM | AVX | Total |
|---|---:|---:|---:|
| Wrapping/saturating add | 8 | 64 | 72 |
| Wrapping/saturating subtract | 8 | 64 | 72 |
| D/Q low product | 1 | 14 | 15 |
| Pilot | 17 | 142 | 159 |

Seven PADD MMX rows and ten PMULLW/VPMULLW rows are outside this migration,
not unsupported or newly verified. PADDD's frontend alias adds no encoder row.
The initialized join now retains row identity, not deduplicated instruction
text. High AMD64 registers force EVEX X/Y encoding; the independent original
Go bytes must match the selected official row. The shared encoder and public
three-API harness use bounded batches of 32 functions. Current Go 1.27.3 and
minimum Go 1.20.14 both pass the following Linux/amd64 LLVM22 object scopes:

| Scope | Official rows | Concrete format witnesses |
|---|---:|---:|
| Add/subtract/D-Q low product | 159 | 480 |
| Signed/unsigned integer min/max | 132 | 432 |
| Rounded average, products and multiply-add | 90 | 258 |
| Integer and PS/PD logical | 132 | 504 |
| Integer and PS/PD unpack, low/high | 108 | 360 |
| Signed/unsigned saturating pack | 36 | 114 |
| Per-lane logical/arithmetic shifts | 64 | 236 |
| Conditional sign, B/W/D | 9 | 18 |
| Byte-indexed shuffle | 9 | 24 |
| Integer EQ/GT compare | 72 | 168 |
| Mask blend | 36 | 144 |
| Local total | 847 | 2,738 |

Published `55831306c` has the first 730 rows / 2,402 witnesses. The nine
PSHUFB/VPSHUFB rows are qualified at isolated `9c2923dd9` and integrated in
local `a5354ca1a`, whose combined current/minimum checks pass. The compare and
mask-blend joins each pass both endpoint SDKs and are integrated in local
`2c19ef128`. That complete local candidate still needs joint qualification
and publication. Published `55831306c` has since finished all 325 CI jobs
successfully; its evidence does not qualify later local changes. The shared
encoder consumes the selected legacy two-byte opcode payload; the existing
translator needs no additional instruction rule. Both endpoint SDKs pass all
nine shuffle rows and the complete rule reader: 38 roots / 432 children,
zero failures/skips. The later compare scope passes 42 roots / 482 children;
mask blend passes 46 / 514, both with zero failures/skips on each SDK.

Two PADDD alias witnesses are separate from the canonical counts. Min/max
includes 48 non-Q, unmasked, narrow EVEX forms; low-register VEX objects cannot
substitute for them. These are initialized named-source/object checks, not raw
decoder or executed-semantic coverage of every row. Complete legacy/AVX mode
applicability and remaining family joins are still required; do not divide this
pilot by unfiltered encoding-table rows to claim an architecture percentage.

The logical join shares the existing row grammar, original-byte verifier and
three-API object driver. Prefix, carrier width and VEX/EVEX identity are
orthogonal encoding axes; impossible VEX512 and conflicting prefix forms fail.
Its 32 names need 19 additional implementation lines, not 504 separate rules.
The seven average/product and four logical MMX rows keep separate region owners.
The residual dword-product dispatcher also loses 40 unreachable production lines;
its original direct fixtures now invoke the same typed owner as public lowering.

The unpack join covers 24 names through the same row grammar and driver, adding
ten scope-selection lines and no translator rule. All original Go bytes and
initialized objects pass on both supported endpoint SDKs. Six MMX rows retain
their independent region owner; this join does not expand native execution
claims or weaken low-MMX m32 versus high-MMX m64 memory-width distinctions.

Saturating pack reuses that same grammar and driver for eight names, 36 rows
and 114 witnesses, adding seven scope-selection lines. Dword inputs alone
admit its 18 broadcast witnesses; three MMX rows remain independent. Both
endpoint SDKs pass the complete initialized join and rule-reader regression
at `d17e46d5d`, without expanding the translator or native execution claims.

The pack parser at `25555b5f0` also shares physical normalization with the four
average/product families, removing 20 production lines. It preserves mode32
error classification, MMX operand ownership and E4NF memory behavior. Current
and minimum SDKs each pass 33 roots / 1,211 children; sixteen parser benchmarks
each remove one allocation and show 9–31% lower median time over five samples.
This measures parsing, not complete-source translation or native instruction speed.

The per-lane shift join at `376a9f4d0` adds nine names, 64 rows and 236 witnesses
through the existing binary grammar and object driver, with no translator rule.
Both endpoint SDKs pass 30 roots / 394 children. A shared test inventory removes
repeated family loops while preserving independent official membership, literal
encoding anchors, unique row IDs, width/encoding axes and separate MMX owners.
The complete batch removes 126 net lines. Its 54 broadcast witnesses and narrow
EVEX forms are object evidence, not new native execution acceptance.

Conditional sign at `2c2c98295` adds all six B/W/D names: three legacy and six
VEX rows, 18 register/memory witnesses and six independent byte anchors. Go has
no MMX, EVEX or broadcast row for this family. Sharing the anchor assertion loop
with UNPACK and per-lane shifts preserves their original 19 byte anchors. The
complete rule reader, new object owner and listing-identity negative pass on
both endpoint SDKs: 25 roots / 88 children each. This adds 17 net lines across
six files, with no translator rule or new native execution claim.

Fixed-frame GP analysis now consumes the existing typed byte/FLAGS effects
instead of a second arithmetic opcode list. Its bounded worklist intersects
numerical definedness across loops while retaining exact address identities.
All eight original SP residues and the independent final replay remain checked;
provisional loop accesses are published only after convergence. Numerical LEA
and sized NOP addresses are not mislabeled as actual memory accesses. Ordinary
typed-input base/displacement reads now compose with private scratch through
the existing exact-width faulting scalar path. Full-word affine operations can
now retain that input's origin without converting its representation to NUM or
granting a pointee extent. Indexed private scratch retains its separate limits.
BP/R14 entry bytes stay unknown; a defined value may use the complete GP bank.
Named LEA and scalar MOV share machine grammar with ordinary consumers, while
source-role filtering stays separate. An exact native-input access selects the
fault-preserving scalar emitter even through BP; it is never a private-frame
certificate. Full-source and native tests remain separate gates.

Fixed-local result consumers now share a call-local independently parsed
original file and complete ordered feature index. Public module entry/exit
still perform live source verification; standalone internal consumers retain
their original guards. This reduces repeated source/tool work without caching
mutable verification or removing per-function proof replay. The measured
one/four-function API fixtures keep exactly two verification rounds; complete
standard-library translation has not yet shown a speedup.

The new result-mask helper emits five vector IR instructions instead of a
lane-wise sequence. This is not a five-machine-instruction or constant-runtime
claim: LLVM may split its predicate vectors. Shared splat formatting has constant
allocation count but necessarily linear output size. Keep the independent
runtime oracles and measure generated-program performance on actual AMD64
hardware before claiming that whole-axis improvement.
