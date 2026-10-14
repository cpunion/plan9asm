# Canonical x86 bool results

The ordinary scalar-register lowering role accepts one logical `i1` result
only when the complete explicit ABIInternal source proves AX's physical low
byte is initialized and exactly zero or one on every path to RET. The actual
Go logical signature, scalar input register assignment and result field must
still agree. Bool inputs, tuples, spills, FP/AUTO backing, raw directives and
hidden/native entry state retain their independent Context boundaries.

Per-byte canonicality is a separate must-fact from initialization. Complete
typed SETcc, scalar moves (including source-only MOVD/MOVDQ2Q aliases),
extensions, conditional moves and AND/OR/XOR preserve only proven canonical
bytes; other writes revoke the affected facts. Reads, FLAGS, source shape and
memory are first validated by the independent typed transfer. Joins intersect
both domains, and the existing finite byte-work budget counts the additional
state without increasing its limit. No known low byte initializes unknown GP
upper bytes. Arbitrary initialized bytes 2, 3 or 255 cannot gain a logical bool
by truncation to `i1`.

Execution tests keep two independent axes. The complete actual Go assembler
ABIInternal instruction bytes execute through an explicit test-only AX/BX
adapter with raw-u64, stack and GP canaries. Standalone `-p runtime` grants
encoder-fixture selector permission only: it is not a captured SDK source,
native Go caller certificate or runtime-path masquerade. LLVM calls the
translated `i1` function through an explicit `i1` to `i32` shim, with C crossing
only integer signatures; this is not a C `_Bool` ABI assertion. A separate owned
Go package executes genuine ABI0 bool callers and observes their original raw
return byte and canaries, including noncanonical 2/3/255 raw-byte controls that
are never normalized through a bool call. Linux/amd64 is the required execution
counterpart when local AMD64/Rosetta execution is unavailable.

This role neither binds the linked SDK literal-minimum-alignment producer nor
exposes a linked address as a number. Full original maps functions remain
subject to their linked-low-bit, caller, pointer, loop and native-entry contracts.

## Bounded 386 ABI0 byte result

The independent first ABI0 consumer admits only a complete ordinary 386
no-argument, zero-frame function with exactly one declared bool result byte
at FP+0 and original argument size one. Other frame/input/result, direct-tail,
raw and native roles retain Context; the shape alone authenticates no source.
Each declared physical byte has a separate bounded canonical must-fact,
cloned and intersected across the complete source CFG. Numerical initializedness
remains independent: raw 2, 3 and 255 may be read or compared and then replaced
by SETcc or another proved zero/one before RET. Overwrites revoke canonicality,
and every predecessor must supply a canonical result. The existing shared
byte-work envelope counts this state without increasing its limit.

The cached whole-source proof is required even without an FP/SIMD probe.
Intermediate loads and stores retain the physical I8 value; only the final
proved result is truncated to logical I1. This does not initialize unknown GP
upper bytes, expand FP storage, permit pointer/frame escapes or establish
native Go entry. The required Linux AMD64 execution counterpart uses original
386 Go bool callers and independent raw-byte/canary observers, plus LLVM's
explicit I1-to-I32 shims; fixture inventory/IR tests alone are not execution.
