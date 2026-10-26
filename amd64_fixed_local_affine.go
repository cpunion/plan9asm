package plan9asm

import "math"

// Full-word affine operations preserve an actual typed input's identity, not
// numerical pointer bytes, pointee bounds or private-frame ownership. Grammar,
// widths and FLAGS come from the same typed forms as ordinary GP lowering.
func fixedLocalInputAddressStep(ins *Instr, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int) (bool, error) {
	if len(state.opaque) == 0 && len(state.opaqueCells) == 0 {
		return false, nil
	}
	fail := func() (bool, error) {
		return true, goAMD64FrameContext("fixed local affine input lacks a unique pointer and numerical offset: %s", ins.Op)
	}
	write := func(destination Operand, origin x86FixedLocalOrigin, killed, defined uint8) (bool, error) {
		root, valid := x86CallerGP(destination.Reg)
		if destination.Kind != OpReg || !valid || root != destination.Reg || root == SP {
			return fail()
		}
		writeFixedLocalOrigin(state, root, origin)
		if origin.literal.symbol != "" {
			defined = 0 // Relocated DATA addresses do not establish numerical FLAGS.
		}
		state.flags = state.flags&^killed | defined
		return true, nil
	}
	shift := func(origin x86FixedLocalOrigin, operand Operand, subtract bool) (x86FixedLocalOrigin, bool) {
		if origin.literal.symbol == "" {
			return origin, true // Existing native input/global offsets have no inferred extent.
		}
		delta, known := operand.Imm, operand.Kind == OpImm && operand.ImmRaw == "" && !operand.ImmIsFloat
		if operand.Kind == OpReg {
			delta, known = state.constants[operand.Reg]
		}
		if !known || subtract && delta == math.MinInt64 {
			return origin, false
		}
		if subtract {
			delta = -delta
		}
		return fixedLocalShiftLiteral(origin, delta, proof)
	}
	if !x86ResultNamedPredicateInstruction(*ins) {
		return false, nil
	}
	if form, known, err := x86LockScalarRMWProperties(ins.Op, *ins, "amd64"); known && err == nil && form.spec.bits == 64 {
		origin, pointer := fixedLocalOpaqueRegister(form.destination, state)
		sourceOrigin, sourcePointer := fixedLocalOpaqueRegister(form.source, state)
		switch form.spec.stem {
		case "ADD", "SUB":
			if !pointer && !sourcePointer {
				return false, nil
			}
			if pointer && !sourcePointer && fixedLocalNumericRead(form.source, 8, state) {
				var valid bool
				origin, valid = shift(origin, form.source, form.spec.stem == "SUB")
				if !valid {
					return fail()
				}
				return write(form.destination, origin, form.spec.flagsKilled, form.spec.flagsDefined)
			}
			if form.spec.stem == "ADD" && sourcePointer && !pointer && fixedLocalNumericRead(form.destination, 8, state) {
				var valid bool
				sourceOrigin, valid = shift(sourceOrigin, form.destination, false)
				if !valid {
					return fail()
				}
				return write(form.destination, sourceOrigin, form.spec.flagsKilled, form.spec.flagsDefined)
			}
			return fail()
		case "INC", "DEC":
			if pointer {
				var valid bool
				origin, valid = shift(origin, Operand{Kind: OpImm, Imm: 1}, form.spec.stem == "DEC")
				if !valid {
					return fail()
				}
				return write(form.destination, origin, form.spec.flagsKilled, form.spec.flagsDefined)
			}
		}
	}
	if effect, known, err := x86NamedLEAByteEffect(ins.Op, *ins, "amd64"); known && err == nil && x86LEAWidth(ins.Op, "amd64") == 8 {
		memory := ins.Args[0].Mem
		origin, pointer := state.opaque[memory.Base]
		if pointer {
			for _, input := range effect.operands {
				if input.read && input.operand.Reg != memory.Base && !fixedLocalNumericRead(input.operand, input.width, state) {
					return fail()
				}
			}
			// An aliased pointer index is still a second pointer, even if it
			// happens to share the base's spelling or original entry slot.
			if _, secondPointer := state.opaque[memory.Index]; secondPointer {
				return fail()
			}
			if origin.literal.symbol != "" {
				var valid bool
				origin, valid = fixedLocalShiftLiteral(origin, memory.Off, proof)
				if !valid || memory.Index != "" {
					return fail()
				}
			}
			return write(ins.Args[1], origin, effect.flagsKilled, effect.flagsDefined)
		}
	}
	form, known, err := x86ScalarCompareProperties(ins.Op, *ins, "amd64")
	if !known || err != nil || form.test || form.bits != 64 {
		return false, nil
	}
	literal := false
	opaque := func(index int) bool {
		operand := form.args[index]
		if origin, found := fixedLocalOpaqueRegister(operand, state); found {
			literal = literal || origin.literal.symbol != ""
			return true
		}
		if operand.Kind != OpMem || proof == nil {
			return false
		}
		address, err := fixedLocalMemoryAddress(*ins, index, proof.frame)
		if err != nil || address.Index != "" || address.Sym != "" || address.Segment != "" {
			return false
		}
		base, found := state.pointers[address.Base]
		if !found || address.Off < -proof.frame || address.Off > proof.frame {
			return false
		}
		origin, found := state.opaqueCells[base+address.Off]
		literal = literal || found && origin.literal.symbol != ""
		return found
	}
	pointers := [2]bool{opaque(0), opaque(1)}
	if !pointers[0] && !pointers[1] {
		return false, nil
	}
	if literal {
		return fail() // No independent proof of placement-invariant address comparisons.
	}
	for index, operand := range form.args {
		if !pointers[index] {
			if operand.Kind == OpImm && fixedLocalNumericRead(operand, 8, state) {
				continue
			}
			return fail()
		}
		if operand.Kind == OpMem {
			if _, err := fixedLocalMemory(ins, index, state, proof, residue, 8, 1, false, true); err != nil {
				return true, err
			}
		}
	}
	state.flags = x86ByteFlagsAll
	return true, nil
}

func fixedLocalOpaqueRegister(operand Operand, state *x86FixedLocalState) (x86FixedLocalOrigin, bool) {
	root, valid := x86CallerGP(operand.Reg)
	if operand.Kind != OpReg || !valid || root != operand.Reg {
		return x86FixedLocalOrigin{}, false
	}
	origin, found := state.opaque[root]
	return origin, found
}

func intersectFixedLocalOrigins[K, V comparable](previous, current map[K]V) bool {
	changed := false
	for key, value := range previous {
		if incoming, found := current[key]; !found || incoming != value {
			delete(previous, key)
			changed = true
		}
	}
	return changed
}
