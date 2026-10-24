package plan9asm

import (
	"bytes"
	"path/filepath"
	"strings"
)

type x86FixedLocalData struct {
	size, alignment int64
}

func fixedLocalDataSymbol(symbol string) (string, bool) {
	name := strings.TrimPrefix(symbol, "·")
	if !strings.HasSuffix(name, "<>") || len(name) <= 2 || strings.ContainsAny(name, "·/.") {
		return "", false
	}
	return "·" + name, true
}

// Called under the original source lock after its whole parsed file has been
// matched. Private literal definitions cannot be replaced by a Go variable,
// linker winner, mutable global, address initializer or caller-supplied extent.
func (source *GoAMD64FrameSource) fixedLocalData(file *File, expanded string, resolve func(string) string) (map[string]x86FixedLocalData, map[string]string, error) {
	if len(file.Data)+len(file.Globl) == 0 {
		return nil, nil, nil
	}
	object, err := source.read(filepath.Join(source.owned, source.request.AssemblyFile+".o"))
	if err != nil {
		return nil, nil, err
	}
	observed, err := source.frameDataObject(expanded, object)
	if err != nil {
		return nil, nil, err
	}
	if len(observed.Layouts) != len(file.Globl) {
		return nil, nil, goAMD64FrameContext("fixed local DATA lacks its whole original definition inventory")
	}
	data := make(map[string]x86FixedLocalData)
	symbols := make(map[string]string)
	var total int64
	for index, global := range file.Globl {
		layout := observed.Layouts[index]
		definition := layout.Definition
		name, private := fixedLocalDataSymbol(global.Sym)
		expected := source.selection.ImportPath + "." + strings.TrimPrefix(name, "·")
		if _, duplicate := symbols[name]; duplicate {
			return nil, nil, goAMD64FrameContext("fixed local DATA definitions collapse to the same translated identity")
		}
		total += int64(definition.Size)
		if !private || resolve(name) != expected || !layout.Static || layout.ABI != 65535 ||
			layout.Flags != 8 && layout.Flags != 24 || definition.Kind != 3 || definition.Flags != 0 ||
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
		symbols[name] = expected
	}
	return data, symbols, nil
}

// Readonly DATA never receives a frame offset or a native-frame certificate.
// The only consumer is an already registered numerical memory operand. The
// complete original file and immutable symbol resolution own its LLVM global.
func fixedLocalReadMemory(ins *Instr, index int, state *x86FixedLocalState, proof *x86FixedLocalProof,
	residue int, width, alignment int64) (offset int64, literal bool, err error) {
	if index >= 0 && index < len(ins.Args) && ins.Args[index].Kind == OpSym {
		if proof.binding == nil || width <= 0 || alignment <= 0 {
			return 0, false, goAMD64FrameContext("fixed local DATA read has no original literal owner")
		}
		name, offset, valid := parseSBRef(ins.Args[index].Sym)
		data, found := proof.binding.literalData[name]
		if !valid || !found || offset < 0 || width > data.size || offset > data.size-width ||
			alignment > data.alignment || offset%alignment != 0 {
			return 0, false, goAMD64FrameContext("fixed local DATA read exceeds original identity, extent or alignment")
		}
		return offset, true, nil
	}
	offset, err = fixedLocalMemory(ins, index, state, proof, residue, width, alignment, false, false)
	return offset, false, err
}
