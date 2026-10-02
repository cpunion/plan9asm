package plan9asm

import (
	"strconv"
	"strings"
)

// The private-frame and continuation proofs consume the same scalar and
// GP/FP-pair grammar. A narrow/FP store never creates an exact saved pointer;
// its actual byte range can still invalidate an overlapping eight-byte cell.
func (state *arm64ControlState) transferTypedMemory(ins Instr) bool {
	form, valid := arm64RawPrivateMemory(ins)
	if !valid {
		return false
	}
	memory := form.memory
	if form.post && memory.Kind == OpMem {
		memory.Mem.Off = 0
	}
	var values [2]arm64ControlValue
	for index, reg := range form.registers {
		if form.store {
			values[index] = state.source(Operand{Kind: OpReg, Reg: reg}, false)
		} else {
			cell := arm64ControlElement(memory, int64(index)*form.elementBytes)
			values[index] = state.readMemoryWidth(cell, form.elementBytes, form.gp)
		}
	}
	for index, reg := range form.registers {
		cell := arm64ControlElement(memory, int64(index)*form.elementBytes)
		if form.store {
			state.writeMemoryWidth(cell, values[index], form.elementBytes, form.gp)
		} else {
			state.write(Operand{Kind: OpReg, Reg: reg}, values[index], false)
		}
	}
	if form.pre || form.post {
		base := arm64ControlBase(form.memory.Mem.Base)
		state.regs[base] = arm64ControlOffset(arm64ControlRead(state.regs, base), form.memory.Mem.Off)
	}
	return true
}

func (state *arm64ControlState) readMemoryWidth(memory Operand, width int64, gp bool) arm64ControlValue {
	if gp && width == 8 {
		return state.source(memory, false)
	}
	value := arm64ControlUnion(arm64ControlExternal(), arm64ControlLabels(state.source(memory, false)))
	if key := state.memoryKey(memory, false); strings.HasPrefix(key, "sp:") || strings.HasPrefix(key, "fp:") {
		// Partial or floating loads can carry bits of a code pointer, but
		// cannot restore its exact control identity. Inspect every overlap,
		// including a load beginning inside a saved cell rather than at it.
		for cell, stored := range state.memory {
			if arm64ControlMemoryRangesOverlap(key, width, cell, 8) {
				value = arm64ControlUnion(value, arm64ControlLabels(stored))
			}
		}
	}
	if len(arm64ControlLabels(value)) != 0 {
		value["code-derived:"] = true
	}
	return value
}

func (state *arm64ControlState) writeMemoryWidth(memory Operand, value arm64ControlValue, width int64, gp bool) {
	if gp && width == 8 {
		state.write(memory, value, false)
		return
	}
	key := state.memoryKey(memory, false)
	if key == "" {
		state.invalidateUnknownStore(memory, value)
		return
	}
	state.invalidateOverlappingControlCells(key, width)
	// Retain transformed code bits without inventing an exact pointer cell.
	// Unknown ordinary data need not create a new tracked continuation cell.
	taint := arm64ControlLabels(value)
	if len(taint) != 0 {
		taint = arm64ControlUnion(taint, arm64ControlExternal())
		taint["code-derived:"] = true
		for offset := int64(0); offset < width; offset += 8 {
			cell := arm64ControlElement(memory, offset)
			part := state.memoryKey(cell, false)
			if part == "" {
				state.invalidateUnknownStore(cell, taint)
				continue
			}
			state.memory[part] = arm64ControlUnion(state.memory[part], taint)
		}
	}
}

func arm64ControlMemoryRangesOverlap(first string, firstWidth int64, second string, secondWidth int64) bool {
	if len(first) < 4 || len(second) < 4 || first[:3] != second[:3] || firstWidth <= 0 || secondWidth <= 0 {
		return false
	}
	left, err := strconv.ParseInt(first[3:], 10, 64)
	if err != nil {
		return false
	}
	right, err := strconv.ParseInt(second[3:], 10, 64)
	if err != nil {
		return false
	}
	// Unsigned distance avoids signed overflow at either int64 boundary.
	if left <= right {
		return uint64(right)-uint64(left) < uint64(firstWidth)
	}
	return uint64(left)-uint64(right) < uint64(secondWidth)
}
