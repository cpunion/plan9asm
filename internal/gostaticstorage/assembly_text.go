package gostaticstorage

import (
	"encoding/binary"
	"path/filepath"
	"strconv"
	"strings"
)

// ObserveARM64AssemblyText is an offline original-object observation. It does
// not construct a live source, caller carrier, entry adapter or frame slot.
func ObserveARM64AssemblyText(data, builtinSource []byte, target string) ([]TextRecord, error) {
	if !arm64AssemblyObservationTarget(target) {
		return nil, contextError("unobserved ARM64 assembly target")
	}
	return observeAssemblyText(data, builtinSource, target)
}

// ObserveARM64AssemblyArchive retains original caller/callee records only.
// Neither selected source names nor a successful decode grant LLVM authority.
func ObserveARM64AssemblyArchive(data, builtinSource []byte, target string, sources []string) ([]TextMember, error) {
	if !arm64AssemblyObservationTarget(target) {
		return nil, contextError("unobserved ARM64 assembly archive target")
	}
	return observeAssemblyArchive(data, builtinSource, target, sources, false)
}

func arm64AssemblyObservationTarget(target string) bool {
	switch target {
	case "darwin/arm64", "linux/arm64", "windows/arm64", "freebsd/arm64", "openbsd/arm64":
		return true
	}
	return false
}

// ObserveX86AssemblyText decodes complete original Go assembler object bytes.
// It is an offline observation: neither these bytes nor its return value create
// an Observation, a live source lease or a memory/entry permission.
func ObserveX86AssemblyText(data, builtinSource []byte, target string) ([]TextRecord, error) {
	if target != "linux/386" && target != "linux/amd64" {
		return nil, contextError("unobserved x86 assembly target")
	}
	return observeAssemblyText(data, builtinSource, target)
}

func observeAssemblyText(data, builtinSource []byte, target string) ([]TextRecord, error) {
	builtins, err := textBuiltins(builtinSource)
	if err != nil {
		return nil, err
	}
	body, err := objectBody(data, target)
	if err != nil {
		return nil, err
	}
	return readTextObject(body, target, 4, false, builtins)
}

// ObserveX86AssemblyArchive retains every member of a bounded owned-main Go
// archive. sources is selection metadata, not authority. A live factory must
// independently bind that actual selection, the SDK work/gc.go member route,
// original source and standalone assembler output before using any record.
func ObserveX86AssemblyArchive(data, builtinSource []byte, target string, sources []string) ([]TextMember, error) {
	return observeX86AssemblyArchive(data, builtinSource, target, sources, false)
}

func observeX86AssemblyArchive(data, builtinSource []byte, target string, sources []string, sdk bool) ([]TextMember, error) {
	if target != "linux/386" && target != "linux/amd64" {
		return nil, contextError("unobserved x86 assembly archive target")
	}
	return observeAssemblyArchive(data, builtinSource, target, sources, sdk)
}

func observeAssemblyArchive(data, builtinSource []byte, target string, sources []string, sdk bool) ([]TextMember, error) {
	builtins, err := textBuiltins(builtinSource)
	if err != nil {
		return nil, err
	}
	members, err := archiveMembers(data)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 || len(members) != len(sources)+2 || members[0].name != "__.PKGDEF" || members[1].name != "_go_.o" {
		return nil, contextError("assembly archive differs from selected member inventory")
	}
	seenSources := make(map[string]bool)
	definitions := make(map[string]bool)
	storage := make(map[string]bool)
	sdkDefinitions := make(sdkAssemblyNamedDefinitions)
	// Go 1.27.1 work/init.go passes -shared for default PIE on Darwin.
	// Windows default PIE does not pass it. This exact archive role is not
	// inherited by direct tool asm objects or the Linux x86 observation APIs.
	compilerFlags := uint32(0)
	if target == "darwin/arm64" {
		compilerFlags = 1
	}
	result := []TextMember{{Name: members[0].name, Offset: members[0].offset, Size: int64(len(members[0].data))}}
	for index, member := range members[1:] {
		flags := compilerFlags
		if sdk {
			flags |= 16 // ObjFlagStd in the independently pinned Go 1.27.1 format.
		}
		if index != 0 {
			name := sources[index-1]
			if filepath.Base(name) != name || !strings.HasSuffix(name, ".s") || seenSources[name] || member.name != strings.TrimSuffix(name, ".s")+".o" {
				return nil, contextError("assembly source/member route differs")
			}
			seenSources[name] = true
			flags |= 4
		}
		body, err := objectBody(member.data, target)
		if err != nil {
			return nil, err
		}
		if !sdk {
			records, err := assemblyArchiveDefinitions(body, flags, false, index != 0)
			if err != nil {
				return nil, err
			}
			for _, record := range records {
				name := record.Name
				key := name + "\x00" + strconv.Itoa(int(record.ABI))
				if definitions[key] || storage[name] || index != 0 && record.Kind != 1 {
					return nil, contextError("duplicate same-ABI/storage or assembly DATA definition")
				}
				definitions[key] = true
				storage[name] = record.Kind != 1
			}
		} else {
			object, err := checkedObject(body, flags, index != 0)
			if err != nil {
				return nil, err
			}
			symbols, err := object.indexedSymbols()
			if err != nil {
				return nil, err
			}
			for _, symbol := range symbols {
				if err := sdkDefinitions.observe(symbol, index != 0); err != nil {
					return nil, err
				}
			}
		}
		texts, err := readTextObject(body, target, flags, sdk && index != 0, builtins)
		if err != nil {
			return nil, err
		}
		result = append(result, TextMember{Name: member.name, Offset: member.offset, Size: int64(len(member.data)), Flags: flags, Texts: texts})
	}
	return result, nil
}

// The storage reader intentionally has a name-only view. SDK compiler TEXT
// wrappers can share their assembly name with a different ABI, so this
// observation enumerates each independently checked definition instead.
// It returns no DATA bytes, initialization, storage or native entry proof.
func assemblyArchiveDefinitions(data []byte, flags uint32, sdk, assembly bool) ([]dataRecord, error) {
	if !sdk {
		records, err := readObjectRole(data, flags, false)
		if err != nil {
			return nil, err
		}
		values := make([]dataRecord, 0, len(records))
		for _, record := range records {
			values = append(values, record)
		}
		return values, nil
	}
	object, err := checkedObject(data, flags, assembly)
	if err != nil {
		return nil, err
	}
	var records []dataRecord
	for index := 3; index <= 6; index++ {
		block := object.block(index)
		for offset := 0; offset < len(block); offset += 21 {
			symbol := block[offset : offset+21]
			name, err := object.stringAt(symbol[:8])
			if err != nil {
				return nil, err
			}
			if name != "" {
				records = append(records, dataRecord{Name: name, ABI: binary.LittleEndian.Uint16(symbol[8:]), Kind: symbol[10]})
			}
		}
	}
	return records, nil
}

// ObserveX86SDKAssemblyArchive observes an original SDK archive. This offline
// decoder does not establish SDK source provenance, storage or native entry
// authority; its caller must independently authenticate those roles.
func ObserveX86SDKAssemblyArchive(data, builtinSource []byte, target string, sources []string) ([]TextMember, error) {
	return observeX86AssemblyArchive(data, builtinSource, target, sources, true)
}
