package plan9asm

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
		delete(state.pointers, root)
		delete(state.constants, root)
		state.opaque[root], state.known[root], state.booleans[root] = origin, 0, 0
		state.flags = state.flags&^killed | defined
		return true, nil
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
				return write(form.destination, origin, form.spec.flagsKilled, form.spec.flagsDefined)
			}
			if form.spec.stem == "ADD" && sourcePointer && !pointer && fixedLocalNumericRead(form.destination, 8, state) {
				return write(form.destination, sourceOrigin, form.spec.flagsKilled, form.spec.flagsDefined)
			}
			return fail()
		case "INC", "DEC":
			if pointer {
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
			return write(ins.Args[1], origin, effect.flagsKilled, effect.flagsDefined)
		}
	}
	form, known, err := x86ScalarCompareProperties(ins.Op, *ins, "amd64")
	if !known || err != nil || form.test || form.bits != 64 {
		return false, nil
	}
	opaque := func(index int) bool {
		operand := form.args[index]
		if _, found := fixedLocalOpaqueRegister(operand, state); found {
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
		_, found = state.opaqueCells[base+address.Off]
		return found
	}
	if !opaque(0) && !opaque(1) {
		return false, nil
	}
	for index, operand := range form.args {
		if !opaque(index) {
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
