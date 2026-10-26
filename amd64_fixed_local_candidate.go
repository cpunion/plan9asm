package plan9asm

// GoAMD64FixedLocalFrameBindingNeeded is a source-role query, not a frame or
// memory certificate. The complete actual-object factory and fixed-local
// proof must independently succeed before any instruction gains backing.
func GoAMD64FixedLocalFrameBindingNeeded(file *File, goarch string) bool {
	if file == nil || file.Arch != ArchAMD64 || goarch != "amd64" {
		return false
	}
	for _, fn := range file.Funcs {
		if fn.FrameSize <= 0 {
			continue
		}
		for _, ins := range fn.Instrs {
			// The final candidate role requires original SP/BP memory. Avoid
			// parsing allocating effects for register-only or unrelated inputs;
			// this negative filter grants neither grammar nor storage permission.
			localMemory := false
			for _, operand := range ins.Args {
				if operand.Kind == OpMem && (operand.Mem.Base == SP || operand.Mem.Base == BP ||
					operand.Mem.Index == SP || operand.Mem.Index == BP) {
					localMemory = true
					break
				}
			}
			if !localMemory {
				continue
			}
			memory, eligible := -1, false
			physicalOnly := false
			if _, known, err := x86ScalarRegisterLoadProperties(ins.Op, ins, goarch); known {
				memory, eligible, physicalOnly = 0, err == nil, true
			} else if _, known, err := x86ScalarRegisterStoreProperties(ins.Op, ins, goarch); known {
				memory, eligible, physicalOnly = 1, err == nil, true
			} else if form, known, err := fixedLocalConversionProperties(ins); known {
				memory, eligible = form.memory, err == nil && !form.mmx
			} else if form, known, err := fixedLocalScalarLaneProperties(ins); known {
				memory, eligible = form.memory, err == nil
			} else if form, known, err := amd64PackedVectorMoveProperties(ins, goarch); known {
				memory, eligible = form.memory, err == nil && !form.undefinedOpcode()
				physicalOnly = !form.spec.aligned
			} else if form, known, err := x86NonTemporalVectorProperties(ins.Op, ins, goarch); known {
				memory, eligible = form.move.memory, err == nil && !form.trap
			} else if form, known, err := x86LegacyPackedMoveProperties(ins.Op, ins, goarch); known {
				memory, eligible = form.memory, err == nil
				physicalOnly = !form.aligned
			} else if form, known, err := x86AESProperties(ins.Op, ins, goarch); known {
				memory, eligible = form.memory, err == nil && form.byteWidth == 16
			} else if effect, known, err := fixedLocalVectorByteEffect(ins); known {
				eligible = err == nil
				for _, operand := range effect.operands {
					if operand.operand.Kind == OpMem {
						memory = operand.index
					}
				}
			} else if form, known, err := x86FloatCompareProperties(ins, goarch); known {
				memory, eligible = form.memory, err == nil && form.legacy && !form.scalar && form.byteWidth == 16
			} else if form, known, err := x86ConditionalMoveProperties(ins.Op, ins, goarch); known {
				memory, eligible = 0, err == nil && form.source.Kind == OpMem
			} else if effect, known, err := fixedLocalGPByteEffect(ins); known {
				eligible, physicalOnly = err == nil && !effect.dynamicMemory, true
				for _, operand := range effect.operands {
					if operand.operand.Kind == OpMem && (operand.read || operand.write) {
						memory = operand.index
					}
				}
			}
			if eligible && memory >= 0 && memory < len(ins.Args) {
				source := ins.Args[memory]
				// Static SB and incoming native GP addresses keep their existing
				// contracts. Virtual FP is not original local native storage.
				if source.Kind == OpMem && (source.Mem.Base == SP || source.Mem.Base == BP ||
					source.Mem.Index == SP || source.Mem.Index == BP) {
					if physicalOnly {
						// Scalar and unaligned transport can use logical AUTO. Only
						// physical SP/BP needs this additional source observer; the
						// existing parser retains the distinct named displacement.
						address, err := fixedLocalMemoryAddress(ins, memory, fn.FrameSize)
						if err != nil || address != source.Mem {
							continue
						}
					}
					return true
				}
			}
		}
	}
	return false
}
