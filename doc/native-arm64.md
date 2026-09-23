# Native Darwin/ARM64 backend

`TranslateNativeARM64Source` translates a restricted Plan 9 source file to
Darwin/ARM64 assembly. A native assembler such as LLVM's integrated assembler
creates the final Mach-O object. There is no `go tool asm` invocation, dependency
on a Go object layout, instruction-byte copying, or guessed LLVM function type.

The contract is **physical register preservation**, not inferred C typing. A
caller of an entry point must arrange exactly the registers expected by its
source. The source must implement the platform ABI itself, including stack
alignment, saving/restoring the link register, callee-saved registers and any
required native frame. This backend supplies no Go ABI wrappers, stack growth,
GC stack maps, exception/unwind metadata or transitions into the Go runtime.

`ForeignARM64Functions` is only a routing hint: it reports a file composed entirely
of file-local TEXT definitions. Local linkage does not prove a C ABI. A driver
must explicitly choose native translation in a foreign-call context and report
native translation errors rather than retrying with invented signatures.
Package-visible Go functions remain on the typed LLVM path. An unresolved local
function on that path now requires `ManualSig` or a supported inferred signature;
it no longer silently receives `void()`.

## Source contract

- The output target is Darwin/ARM64 only, independent of the build host.
- Every `TEXT` is file-local (`name<>`), has `NOSPLIT`, and declares `$0` or `$0-0`.
  `NOFRAME` is also required if the body uses `BL` or `CALL`. No implicit Go
  prologue or epilogue is generated, including for zero-frame functions.
- Other TEXT flags are rejected. The source manages any native frame explicitly.
- Only `#include "textflag.h"` is accepted. Other includes, macros, conditional
  preprocessing, and block comments are currently rejected. Line comments and
  constant integer expressions are accepted.
- Integer registers are `R0`–`R30` except Darwin's reserved `R18`, plus `ZR`.
  `RSP` is accepted in supported stack-pointer forms; `SP`, `FP`, `g`, `R31`,
  `Wn` aliases, and platform aliases are rejected. Floating registers are `F0`–`F31`.
- Local function/branch/import identifiers use ASCII letters, digits and `_`,
  beginning with a letter or `_`. Data definitions use `·name` for a package global.
- Local labels and function addresses are renamed per assembly file. Foreign
  symbols are resolved only through the driver's explicit import map; no Go
  package/runtime call, symbol addend, computed branch or indirect call is accepted.
- `DATA` supports constant integers of 1/2/4/8 bytes and 8-byte addresses of selected
  local functions, defined data or declared imports. Each initializer must fit its
  `GLOBL` allocation and must not overlap another initializer. Gaps are zero-filled.
- `GLOBL` requires a constant size from 1 byte to 64 MiB. Supported flags are `0`,
  `RODATA` and `NOPTR`. Definitions are 8-byte aligned. `RODATA` is placed in
  `__DATA_CONST,__const`, allowing address fixups before becoming read-only; mutable
  definitions use `.data`. The returned DATA metadata lets the driver verify the
  Go global size and bind its definition to the native object.

## Instruction and operand forms

| Plan 9 operation | Supported forms and native semantics |
| --- | --- |
| `MOVD` | Integer register/register, integer constant/register, register/base+offset memory in either direction. Constants expand to `movz`/`movk` using only the destination and preserving flags. |
| `MOVW`, `MOVWU` | Register/register or register/base+offset memory. `MOVW` register/load sign-extends to 64 bits; `MOVWU` zero-extends. Stores write 32 bits. Immediate forms are rejected. |
| `FMOVD` | Integer/FP register bit transfers or FP/FP register copies. No numeric conversion or memory form. |
| `ADD`, `SUB` | Two or three operands: register or nonnegative 12-bit immediate, source, destination. Register forms involving SP are rejected. No hidden scratch-register expansion. |
| `CMP`, `CMPW` | Register or nonnegative 12-bit immediate against a register; 64-bit and 32-bit comparisons respectively. |
| `LSL` | Two or three operands, with a register or immediate shift from 0 through 63. |
| `B`, `JMP` | A defined local label, selected local function or declared import. |
| `BL`, `CALL` | A defined local label, selected local function or declared import; the containing function must have `NOFRAME`. |
| `BEQ`, `BNE`, `BLT`, `BLE`, `BGT`, `BGE`, `BHS`, `BLO`, `BHI`, `BLS`, `BMI`, `BPL`, `BVS`, `BVC` | Defined local labels, preserving the corresponding hardware condition. |
| `CBZ`, `CBNZ` | 64-bit integer register and defined local label. |
| `RET` | No operands; returns through the native link register. |

Memory operands use a single base register and constant byte offset. An offset
must fit the instruction's unsigned scaled 12-bit or signed unscaled 9-bit field.
Index registers, symbolic memory operands, pre/post-indexing and larger offsets
are rejected. Raw instruction directives (`WORD`, `BYTE`) and all unlisted
instructions are rejected. Validation errors return no partial assembly or DATA.

## Validation

Tests compile generated assembly to Mach-O on all hosts with Clang, check data
section placement, and execute C harnesses on Darwin/ARM64. The harnesses cover
foreign-call parameters/results, integer-to-FP trampoline arguments, physical
register copies, signed and unsigned loads, constant expansion, negative and
unaligned memory offsets, and local control flow. Rejection tests exercise the
unsupported forms above. The parser retains address initializers explicitly;
they are never inferred by decoding an object relocation.
