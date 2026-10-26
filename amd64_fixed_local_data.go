package plan9asm

import (
	"bytes"
	"strings"
)

type x86FixedLocalData struct {
	size, alignment int64
}

type x86FixedLocalMemoryRole uint8

const (
	x86FixedLocalPrivateMemory x86FixedLocalMemoryRole = iota
	x86FixedLocalLiteralMemory
	x86FixedLocalNativeInput
)

func fixedLocalDataSymbol(symbol string) (string, bool) {
	name := strings.TrimPrefix(symbol, "·")
	if !strings.HasSuffix(name, "<>") || len(name) <= 2 || strings.ContainsAny(name, "·/.") {
		return "", false
	}
	return "·" + name, true
}

// Library callers preserve the private <> spelling; the ordinary CLIs render
// it as $local. Both describe the same source-local LLVM namespace, not the
// original Go object's name/ABI. Never accept an arbitrary caller renaming.
func fixedLocalDataEmissionMatches(symbol, packagePath, emitted string) bool {
	canonical := numericalNativeName(symbol, packagePath)
	return emitted == canonical || emitted == strings.TrimSuffix(canonical, "<>")+"$local"
}

// Called under the original source lock after its whole parsed file has been
// matched. Private literal definitions cannot be replaced by a Go variable,
// linker winner, mutable global, address initializer or caller-supplied extent.
func (source *GoAMD64FrameSource) fixedLocalData(file *File, expanded string, resolve func(string) string) (map[string]x86FixedLocalData, map[string]string, error) {
	if len(file.Data)+len(file.Globl) == 0 {
		return nil, nil, nil
	}
	observed, err := source.fixedLocalDataObservation(source.request.AssemblyFile, expanded)
	if err != nil {
		return nil, nil, err
	}
	if len(observed.Layouts) != len(file.Globl) {
		return nil, nil, goAMD64FrameContext("fixed local DATA lacks its whole original definition inventory")
	}
	data := make(map[string]x86FixedLocalData)
	symbols := make(map[string]string)
	// BindFixedLocal has already frozen the resolver outside the source lock.
	// Keep every emitted definition injective, including the complete TEXT set.
	used := make(map[string]bool, len(file.Funcs)+len(file.Globl))
	for _, fn := range file.Funcs {
		name := resolve(fn.Sym)
		if name == "" || used[name] {
			return nil, nil, goAMD64FrameContext("fixed local emitted function identities are empty or collapsed")
		}
		used[name] = true
	}
	var total int64
	for index, global := range file.Globl {
		layout := observed.Layouts[index]
		definition := layout.Definition
		name, private := fixedLocalDataSymbol(global.Sym)
		emitted := resolve(name)
		if _, duplicate := symbols[name]; duplicate {
			return nil, nil, goAMD64FrameContext("fixed local DATA definitions collapse to the same translated identity")
		}
		if !private || !fixedLocalDataEmissionMatches(name, source.selection.ImportPath, emitted) || used[emitted] {
			return nil, nil, goAMD64FrameContext("fixed local DATA resolver changed or collapsed its private emitted identity")
		}
		used[emitted] = true
		total += int64(definition.Size)
		if !layout.Static || layout.ABI != 65535 ||
			layout.Flags != 8 && layout.Flags != 24 || definition.Kind != 3 && definition.Kind != 4 || definition.Flags != 0 ||
			definition.Flags2 != 0 || len(definition.Relocations) != 0 || global.SizeRaw != "" ||
			definition.Size == 0 || definition.Size > 64<<10 || total > 1<<20 || len(definition.Data) > int(definition.Size) {
			return nil, nil, goAMD64FrameContext("fixed local DATA requires bounded private relocation-free readonly storage")
		}
		// Compare the translator's complete initializer with independent actual
		// Go bytes, including sparse zero gaps and trailing allocation. A GLOBL
		// that shortens preceding DATA must not acquire a larger LLVM extent.
		image := make([]byte, definition.Size)
		size := global.Size
		for _, fragment := range file.Data {
			if fragment.Sym != global.Sym {
				continue
			}
			payload, err := dataStmtPayload(fragment)
			if err != nil || fragment.Addr != "" || fragment.Off < 0 || fragment.Off > int64(len(image))-int64(len(payload)) {
				return nil, nil, goAMD64FrameContext("fixed local DATA initializer has an unbound source byte extent")
			}
			copy(image[fragment.Off:], payload)
			if end := fragment.Off + int64(len(payload)); end > size {
				size = end
			}
		}
		if size != int64(definition.Size) || !bytes.Equal(image[:len(definition.Data)], definition.Data) ||
			!bytes.Equal(image[len(definition.Data):], make([]byte, len(image)-len(definition.Data))) {
			return nil, nil, goAMD64FrameContext("fixed local translated DATA differs from the original Go initialized extent")
		}
		data[global.Sym] = x86FixedLocalData{size, goAssemblyDataAlignment(ArchAMD64, size)}
		symbols[name] = emitted
	}
	return data, symbols, nil
}

// Readonly DATA never receives a frame offset or a native-frame certificate.
// The only consumer is an already registered numerical memory operand. The
// complete original file and immutable symbol resolution own its LLVM global.
func fixedLocalReadMemory(ins *Instr, index int, state *x86FixedLocalState, proof *x86FixedLocalProof,
	residue int, width, alignment int64) (offset int64, role x86FixedLocalMemoryRole, err error) {
	if index >= 0 && index < len(ins.Args) && ins.Args[index].Kind == OpSym {
		if proof.binding == nil || width <= 0 || alignment <= 0 {
			return 0, x86FixedLocalLiteralMemory, goAMD64FrameContext("fixed local DATA read has no original literal owner")
		}
		name, offset, valid := parseSBRef(ins.Args[index].Sym)
		if !valid || !fixedLocalLiteralRange(proof.binding, x86StaticSBAddress{name, offset}, width, alignment) {
			return 0, x86FixedLocalLiteralMemory, goAMD64FrameContext("fixed local DATA read exceeds original identity, extent or alignment")
		}
		return offset, x86FixedLocalLiteralMemory, nil
	}
	return fixedLocalMemoryAccess(ins, index, state, proof, residue, width, alignment, false, false)
}

func fixedLocalLiteralRange(binding *x86FixedLocalBinding, address x86StaticSBAddress, width, alignment int64) bool {
	if binding == nil || width <= 0 || alignment <= 0 {
		return false
	}
	data, found := binding.literalData[address.symbol]
	return found && address.offset >= 0 && width <= data.size && address.offset <= data.size-width &&
		alignment <= data.alignment && address.offset%alignment == 0
}

// A literal retains its exact source-local identity and a bounded offset, not
// an arbitrary native pointee. LLVM may place separate constants differently;
// crossing into another object's address must never acquire its bytes.
func fixedLocalShiftLiteral(origin x86FixedLocalOrigin, delta int64, proof *x86FixedLocalProof) (x86FixedLocalOrigin, bool) {
	if proof == nil || proof.binding == nil {
		return origin, false
	}
	data, found := proof.binding.literalData[origin.literal.symbol]
	if !found || origin.literal.offset < 0 || origin.literal.offset > data.size ||
		delta < -origin.literal.offset || delta > data.size-origin.literal.offset {
		return origin, false
	}
	origin.literal.offset += delta
	return origin, true
}

func fixedLocalStaticAddressStep(ins Instr, state *x86FixedLocalState, proof *x86FixedLocalProof) (bool, error) {
	address, valid := fixedLocalStaticAddress(ins)
	if !valid {
		return false, nil
	}
	fail := func() (bool, error) {
		return true, goAMD64FrameContext("fixed local address has no matching original storage: %s", ins.Op)
	}
	if proof == nil || proof.binding == nil {
		return fail()
	}
	var origin x86FixedLocalOrigin
	if _, literal := proof.binding.literalData[address.symbol]; literal {
		origin.literal = address
		if _, valid := fixedLocalShiftLiteral(origin, 0, proof); !valid {
			return fail()
		}
	} else {
		object, found := proof.binding.globals[fixedLocalGlobalSymbol(address.symbol)]
		if ins.Op == "LEAQ" || !found || object.variable == nil || object.native == "" {
			return fail()
		}
		origin.global = object.native
	}
	writeFixedLocalOrigin(state, ins.Args[1].Reg, origin)
	return true, nil
}

func writeFixedLocalOrigin(state *x86FixedLocalState, destination Reg, origin x86FixedLocalOrigin) {
	state.revokeRange(destination)
	delete(state.pointers, destination)
	delete(state.constants, destination)
	state.opaque[destination], state.known[destination], state.booleans[destination] = origin, 0, 0
}
