package gostaticstorage

import (
	"encoding/base64"
	"encoding/binary"
	"strings"
)

// WindowsAMD64UnwindRecord observes one complete original Go metadata record.
// It grants no source, frame, handler or native Windows unwind permission.
type WindowsAMD64UnwindRecord struct {
	Index                uint32
	Name                 string
	Data                 []byte
	PrologueEnd, PushEnd uint8
}

type WindowsAMD64AssemblyUnwind struct {
	Texts  []TextRecord
	Unwind []WindowsAMD64UnwindRecord
}

// ObserveWindowsAMD64AssemblyUnwind is a read-only, bounded object observation.
// A live source factory must independently bind original source, actual tools,
// the pinned producer/schema, full listing and physical prologue/source PCs.
func ObserveWindowsAMD64AssemblyUnwind(data, builtinSource []byte) (WindowsAMD64AssemblyUnwind, error) {
	var result WindowsAMD64AssemblyUnwind
	if len(data) == 0 || len(data) > archiveLimit {
		return result, contextError("Windows assembly object exceeds its byte bound")
	}
	builtins, err := textBuiltins(builtinSource)
	if err != nil {
		return result, err
	}
	body, err := objectBody(data, "windows/amd64")
	if err != nil {
		return result, err
	}
	object, err := checkedObject(body, 4, false)
	if err != nil {
		return result, err
	}
	texts, err := readTextObject(body, "windows/amd64", 4, false, builtins)
	if err != nil {
		return result, err
	}
	if len(texts) == 0 {
		return result, contextError("Windows metadata lacks its complete TEXT inventory")
	}
	block := object.block
	counts := [4]int{}
	var symbols []byte
	for index := range counts {
		counts[index] = len(block(index+3)) / 21
		symbols = append(symbols, block(index+3)...)
	}
	hashedStart, hashedEnd := counts[0]+counts[1], counts[0]+counts[1]+counts[2]
	byIndex := make(map[uint32]WindowsAMD64UnwindRecord)
	names := make(map[string]bool)
	for index := 0; index < object.count; index++ {
		symbol := symbols[index*21 : (index+1)*21]
		name, err := object.stringAt(symbol[:8])
		if err != nil || name != "" && names[name] {
			return result, contextError("duplicate or invalid Windows definition name")
		}
		if name != "" {
			names[name] = true
		}
		for n := object.indices[1][index]; n < object.indices[1][index+1]; n++ {
			if block(15)[int(n)*9] == 14 && symbol[10] != 1 {
				return result, contextError("Windows SEH auxiliary belongs to a non-TEXT definition")
			}
		}
		if symbol[10] != 25 {
			if strings.HasPrefix(name, "go:sehuw.") {
				return result, contextError("Windows unwind name lacks its actual metadata kind")
			}
			continue
		}
		start, end := object.indices[2][index], object.indices[2][index+1]
		if index < hashedStart || index >= hashedEnd || binary.LittleEndian.Uint16(symbol[8:]) != 0 ||
			symbol[11] != 3 || symbol[12] != 0 || binary.LittleEndian.Uint32(symbol[13:]) != 12 ||
			binary.LittleEndian.Uint32(symbol[17:]) != 4 || end-start != 12 ||
			object.indices[0][index] != object.indices[0][index+1] ||
			object.indices[1][index] != object.indices[1][index+1] {
			return result, contextError("Windows unwind symbol/block/extent/auxiliary/relocation role differs")
		}
		bytes := block(16)[start:end]
		if name != "go:sehuw.12."+base64.StdEncoding.EncodeToString(bytes) {
			return result, contextError("Windows unwind name does not bind its complete original bytes")
		}
		if bytes[0] != 1 {
			return result, contextError("Windows native exception-handler/version flags require a separate role")
		}
		if bytes[1] == 0 || bytes[2] != 2 || bytes[3] != 5 || bytes[4] != bytes[1] ||
			bytes[5] != 3 || bytes[6] == 0 || bytes[6] >= bytes[4] || bytes[7] != 0x50 ||
			binary.LittleEndian.Uint32(bytes[8:]) != 0 {
			return result, contextError("unknown Windows unwind header/node/offset/tail grammar")
		}
		value := WindowsAMD64UnwindRecord{uint32(index), name, append([]byte(nil), bytes...), bytes[1], bytes[6]}
		byIndex[value.Index] = value
		result.Unwind = append(result.Unwind, value)
	}
	used := make(map[uint32]bool)
	textNames := make(map[string]bool)
	for _, text := range texts {
		if text.Name == "" || textNames[text.Name] || text.Kind != 1 || text.ABI != 0 {
			return WindowsAMD64AssemblyUnwind{}, contextError("Windows closed-local metadata needs complete distinct ABI0 TEXTs")
		}
		textNames[text.Name] = true
		links := 0
		for _, aux := range text.Auxiliary {
			if aux.Type != 14 {
				continue
			}
			links++
			if aux.Target.PackageIndex != 2147483645 || !aux.Target.ABIKnown || aux.Target.ABI != 0 ||
				uint64(aux.Target.SymbolIndex) >= uint64(counts[2]) {
				return WindowsAMD64AssemblyUnwind{}, contextError("Windows TEXT SEH reference lacks its local hashed role")
			}
			index := aux.Target.SymbolIndex + uint32(hashedStart)
			unwind, exists := byIndex[index]
			if !exists || aux.Target.Name != unwind.Name || int(unwind.PrologueEnd) >= len(text.Code) {
				return WindowsAMD64AssemblyUnwind{}, contextError("Windows TEXT refers to missing/wrong/oversized unwind metadata")
			}
			used[index] = true
		}
		if links != 1 {
			return WindowsAMD64AssemblyUnwind{}, contextError("Windows TEXT has missing or duplicate SEH auxiliary ownership")
		}
	}
	if len(used) != len(byIndex) {
		return WindowsAMD64AssemblyUnwind{}, contextError("Windows unwind inventory has an orphan definition")
	}
	result.Texts = texts
	return result, nil
}
