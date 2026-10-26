# Translator architecture

This is a navigation guide, not a coverage certificate. Start with the
[family workflow](instructions.md) and [validation commands](validation.md).

The current development order is [amd64 architecture convergence](x86-convergence.md):
finish one architecture's implementation, efficiency and independent coverage
before expanding another. Existing cross-architecture regression gates remain.

## Translation path

1. Parsing/preprocessing retains assembly instructions, labels, DATA/GLOBL and
   the source architecture. GoModule additionally derives signatures and frame
   fields from Go declarations. Already-preprocessed input is not expanded again.
2. Source and ABI binding establish any extra facts needed by complete functions:
   parameter/result roles, private continuations, external storage and special
   runtime entries. An opcode spelling alone cannot establish these facts.
3. Architecture-local typed parsers validate complete operand forms. Named and
   raw-encoded forms should converge on the same numerical lowerer and effects.
   Effects describe reads, writes, widths, flags and preserved register aliases.
4. Lowerers emit LLVM IR. ARM64 declarations are selected from an immutable
   catalog by the generated module's actual global references, not a second
   opcode-to-intrinsic table. Comments and string data do not select declarations.
5. The shared module return path checks backend restrictions for every API,
   including special early returns, and disposes rejected modules. The caller
   owns a successful module and must dispose it before its LLVM context.

`Goarch` selects source encoding/ABI rules. `TargetTriple` selects the LLVM
backend. They need not be identical for portable operations. Backend-specific
requirements remain explicit: SVE needs AArch64; wasm cannot use blockaddress.
Used LLVM `llvm.x86.*` intrinsics require a matching x86 source mode and target,
except MXCSR load/store, whose i32 memory operand works in either x86 mode.
The guard checks LLVM intrinsic IDs, not an opcode list; unused declarations,
ordinary functions and generic portable intrinsics impose no such restriction.

## Terms used in family notes

| Term | Meaning and limit |
|---|---|
| Typed grammar | A validated instruction-family form, including operand roles and widths; not proof that arbitrary inputs are initialized. |
| Numerical value / NUM | Defined bytes usable as a number. A pointer's frame occupancy alone does not grant this role. |
| Raw continuation | Lowering of instructions following original encoded bytes, with their exact operands and physical-mode effects retained. Decoding alone does not prove caller, frame or memory contracts. |
| Producer replay | Checking that the source/tool observation which supplied extra facts still matches the consumed input. It is separate from numerical instruction semantics. |
| Native backing | Concrete storage with the required extent and access contract. A logical frame field or a pointer-shaped value alone does not establish contiguous storage. |
| Source owner / lease | The bounded lifetime of a captured source/tool observation. Expiry, mutation or closure invalidates its facts. |
| Context | A probe lacks facts required to decide or translate the form. This is neither support nor successful validation. |

An immutable declaration table may be shared. Mutable source verification and
caller facts must not be cached merely to hide repeated subprocess cost. Move
capture outside repeated translation when its owner and lifetime can be made
explicit; measure complete-source performance and retain invalidation tests.

## Evidence levels

| Evidence | What it establishes | What it does not establish |
|---|---|---|
| Go opcode/encoding table | Names, forms and machine encodings | Translator implementation |
| Official-corpus classification | Outcome for observed probe inputs | Unobserved forms or initialized-context success |
| IR generation and verification | Well-formed LLVM module for supplied inputs | Backend acceptance or correct execution |
| LLVM 22 object compilation | Backend accepts that module/target | Correct calling convention or values at runtime |
| Executed differential test | Observed values/effects agree with an independent oracle | Every unexecuted form, target or hardware feature |

Never promote `context`, a hardware skip, or zero observed `unsupported` forms
to object/runtime coverage. Re-test context forms with justified initialized
inputs, and retain whole-file regressions independently of instruction probes.
