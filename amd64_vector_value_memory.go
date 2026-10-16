package plan9asm

import "fmt"

// cmd/asm accepts extended GP address spellings in mode32, but their source
// names do not supply physical 386 registers. Keep that missing role distinct
// from an invalid operand format; never create numerical/address facts or
// silently allocate an AMD64 register while translating a 386 function.
func x86VectorValueMemoryContext(ins Instr, arch string) error {
	if arch != "386" {
		return nil
	}
	for _, operand := range ins.Args {
		if operand.Kind != OpMem || x86PackedShiftMemoryValid(operand, arch) {
			continue
		}
		local := operand
		if local.Mem.Off == int64(uint32(local.Mem.Off)) {
			local.Mem.Off = int64(int32(local.Mem.Off))
		}
		if x86PackedShiftMemoryValid(local, "amd64") {
			return fmt.Errorf("%w: 386 %s address has no bound physical extended-GP role: %q", ErrProbeNeedsContext, ins.Op, ins.Raw)
		}
	}
	return nil
}
