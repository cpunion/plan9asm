package plan9asm

import "math"

// A full-word LEA keeps the same private-frame identity with a checked offset.
// It does not read memory or initialize numerical bytes; the later access must
// still establish its width, bounds, alignment and reaching initialized bytes.
func fixedLocalPrivateAddressStep(ins Instr, state *x86FixedLocalState, proof *x86FixedLocalProof) (bool, error) {
	if ins.Op != "LEAQ" || len(ins.Args) != 2 || ins.Args[0].Kind != OpMem {
		return false, nil
	}
	if _, private := state.pointers[ins.Args[0].Mem.Base]; !private {
		return false, nil
	}
	if proof == nil || ins.x86AddressBits != 0 {
		return true, goAMD64FrameContext("fixed local LEA requires the original native address width and frame")
	}
	width, err := x86LEAOperandWidth(ins, "amd64")
	address, addressErr := fixedLocalMemoryAddress(ins, 0, proof.frame)
	offset, known := fixedLocalPrivateOffset(address, state)
	destination := ins.Args[1].Reg
	if err != nil || width != 8 || addressErr != nil || !known || destination == SP {
		return true, goAMD64FrameContext("fixed local LEA requires a full private address and checked numerical displacement")
	}
	delete(state.opaque, destination)
	delete(state.constants, destination)
	delete(state.booleans, destination)
	state.revokeRange(destination)
	state.known[destination] = 0
	state.pointers[destination] = offset
	return true, nil
}

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
	write := func(destination Operand, origin x86FixedLocalOrigins, killed, defined uint8) (bool, error) {
		root, valid := x86CallerGP(destination.Reg)
		if destination.Kind != OpReg || !valid || root != destination.Reg || root == SP {
			return fail()
		}
		writeFixedLocalOrigin(state, root, origin)
		if origin.hasLiteral() {
			defined = 0 // Relocated DATA addresses do not establish numerical FLAGS.
		}
		state.flags = state.flags&^killed | defined
		if killed|defined != 0 {
			state.comparison = x86FixedLocalComparison{}
		}
		return true, nil
	}
	shift := func(origins x86FixedLocalOrigins, operand Operand, subtract bool) (x86FixedLocalOrigins, bool) {
		if !origins.hasLiteral() {
			return origins, true // Existing native input/global offsets have no inferred extent.
		}
		delta, known := operand.Imm, operand.Kind == OpImm && operand.ImmRaw == "" && !operand.ImmIsFloat
		if operand.Kind == OpReg {
			delta, known = state.constants[operand.Reg]
		}
		if !known || subtract && delta == math.MinInt64 {
			return origins, false
		}
		if subtract {
			delta = -delta
		}
		if delta == 0 {
			return origins, fixedLocalOriginOwners(origins, proof)
		}
		var entries [4]x86FixedLocalOrigin
		for index := uint8(0); index < origins.count; index++ {
			origin := origins.at(index)
			entries[index] = origin
			if origin.literal.symbol == "" {
				continue
			}
			shifted, valid := fixedLocalShiftLiteral(origin, delta, proof)
			if !valid {
				return origins, false
			}
			entries[index] = shifted
		}
		return fixedLocalOriginSet(entries, origins.count), true
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
		if pointer && origin.valid() {
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
			if origin.hasLiteral() {
				var valid bool
				origin, valid = shift(origin, Operand{Kind: OpImm, Imm: memory.Off}, false)
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
	values := [2]x86FixedLocalOrigins{}
	opaque := func(index int) bool {
		operand := form.args[index]
		if origin, found := fixedLocalOpaqueRegister(operand, state); found {
			values[index] = origin
			return true
		}
		if operand.Kind != OpMem || proof == nil {
			return false
		}
		address, err := fixedLocalMemoryAddress(*ins, index, proof.frame)
		if err != nil {
			return false
		}
		offset, found := fixedLocalPrivateOffset(address, state)
		if !found {
			return false
		}
		origin, found := state.opaqueCells[offset]
		values[index] = origin
		return found && origin.valid()
	}
	pointers := [2]bool{opaque(0), opaque(1)}
	if !pointers[0] && !pointers[1] {
		return false, nil
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
	flags := uint8(x86ByteFlagsAll)
	comparison := x86FixedLocalComparison{}
	if values[0].hasLiteral() || values[1].hasLiteral() {
		if !pointers[0] || !pointers[1] || !fixedLocalOriginOwners(values[0], proof) || !fixedLocalOriginOwners(values[1], proof) {
			return fail()
		}
		// Equality observes actual address identity, never placement-dependent
		// ordering. Keep both native/literal alias possibilities at this CMP.
		flags = x86ByteFlagZF
		for index, value := range values {
			if literal, single := value.literal(); single && form.args[1-index].Kind == OpReg {
				comparison = x86FixedLocalComparison{register: form.args[1-index].Reg, literal: literal}
				break
			}
		}
	}
	state.flags, state.comparison = flags, comparison
	return true, nil
}

func fixedLocalOpaqueRegister(operand Operand, state *x86FixedLocalState) (x86FixedLocalOrigins, bool) {
	root, valid := x86CallerGP(operand.Reg)
	if operand.Kind != OpReg || !valid || root != operand.Reg {
		return x86FixedLocalOrigins{}, false
	}
	origin, found := state.opaque[root]
	return origin, found && origin.valid()
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
