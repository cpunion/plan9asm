# Whole-source x86 direct-tail continuations

An exact same-file `JMP helper<>(SB)` is a source control-flow edge, not a
callable helper ABI. In particular, copying a root's signature or giving a
helper zero arguments loses the root's SI/DI, vector and FLAGS state. Shared
helpers can be reached by roots with different actual parameters and result
layouts; each invocation must retain its own root's state and RET contract.

`PrepareX86DirectTailSource` observes the complete parsed file and returns a
request-local copy with a private continuation graph. Original TEXTs,
instructions, operands, Raw spelling and source ordinals are not rewritten.
Each graph point retains its original TEXT/instruction coordinates. Lowering
places these points in the root's own CFG; it emits neither a callable helper
function nor manually replayed helper registers.

The graph uses the existing mode-bound instruction effects, numerical and
opaque-pointer domains, must-state joins, canonical bool proof and FLAGS/vector
effects. AMD64 explicit ABIInternal roots retain only the existing scalar
register contract. AMD64/386 ordinary roots retain only their own declared
logical ABI0 fields. The 386 graph does not inherit AMD64 register-entry facts.
Root layouts are never merged: identical helper spelling gives no result
padding, additional parameter bytes or caller/native-frame storage.

Only roots with a private-tail source edge receive graph views. Unrelated
functions retain their registered lowering, including supported framed/CALL
consumers. The complete unchanged file is checked through shared lowering
before KeepFunc or partitioning; an unsupported unrelated function cannot
silently disappear. All original helper references are audited even when a
root or instruction will later be dropped.

The first continuation class is call-free and zero-frame. It rejects missing,
duplicate, aliased, offset and indirect targets, helper CALL/address/DATA
escapes, unowned helper TEXTs, raw/native instruction roles and physical
SP/BP/PC/frame addresses. It adds no Go closure DX, g/thread/TLS, native entry,
caller extent, alignment, write or numerical pointer-bit permission. Existing
explicit X86TailGroups keep their independent source contract. Numerical/TLS
binding composition requires a separate exact original-node association and
remains Context here. Standalone Func queries without whole-source information
do not acquire graph authority.

Every request keeps independent original-source/signature snapshots and
verifies source, selected DATA/GLOBL/prelude, TEXT membership, mode and symbol
mapping before and after callbacks/consumption. A failed graph identity check
permanently invalidates that prepared graph; restoring public fields cannot
revive it. Concurrent owned LLVM contexts may read the same unchanged original
File without mutating it or sharing mutable graph state. The existing global
LLVM context ownership boundary is unchanged.

Combined root/helper work keeps the existing 4096-instruction,
16384-fixed-point-visit and 64-MiB logical byte-work envelopes. Shared helpers
count separately for each root's actual state; no per-helper or per-root extra
budget is introduced. These are work envelopes, not claims about peak RSS.

Independent tests retain whole original-Go files, differently typed roots,
renaming, chains and diamonds, unknown/partial/overwritten state, canonical
bool and frame/pointer controls, full-source filtering and mutation controls.
The object class covers AMD64 Linux/Darwin/Windows and 386 Linux/Windows,
all three public APIs and O0/O2. Required Linux AMD64 execution owners are
`TestCrossLinuxRuntimeMatrixAMD64DirectTailWholeSourceOriginalGoAndLLVM` and
`TestCrossLinuxRuntimeMatrixX86DirectTail386OriginalGoAndLLVM`; i386 uses the
kernel compatibility ISA and a required static cross compiler. Emulated local
execution is diagnostic and cannot replace these native counterparts.

This mechanism does not close the complete equal source's closure/entry
contract or authenticate an SDK numerical/native handoff. Offline Go objects
and immutable source graphs do not manufacture those separate permissions.
