# ARM64 raw zero-frame ownership proof

Zero-frame and `NOFRAME` functions do not receive a Go LR/FP prologue merely
because a raw WORD later decodes to BL. A private raw BL witness records its
exact same-source PC target; diagnostic comments cannot provide that origin.

The additive whole-source proof tracks a bounded negative-SP envelope, may-frame
GP aliases and spills, and independent BP/LR reaching values. Every root return
restores entry SP/BP/LR; a bounded local raw call returns with its call-site SP
and exact continuation. Unknown/native calls, unallocated stack cells, differing
SP joins, partial pointer spills, unknown indexing and address escapes fail.
Dead components are checked too. Existing framed-source and continuation guards
remain independent. Proof registration is not an object or runtime pass.

`TestARM64RawZeroFrame*` retains original Go-accepted minimum RED cases, explicit
escape/overlap/call-context negatives, and physical-leaf source-origin checks.
The shared MC matrix adds 132 raw scalar/GP-pair/FP-pair width/sign/offset and
writeback/local-call variants plus 48 actual Go memory spellings. Required
ARM64 native/cross runtime tests compare original Go WORD objects and LLVM22
against independent numeric/full-canary oracles, including two arithmetic
helper calls whose effects must actually execute. These fixtures do not prove
that an entire external module translates.
