package gostaticstorage

import (
	"path/filepath"
	"reflect"
	"strings"
)

// ASMDataProtocol describes offline source/object observations only. Neither
// a fragment nor a successful join is a storage, initializer, read or native
// body certificate. Live SDK source provenance is a separate producer role.
const ASMDataProtocol = "actual_go_asm_data_observation_v1"

type ASMDataValueKind string

const (
	ASMDataInteger ASMDataValueKind = "integer"
	ASMDataFloat   ASMDataValueKind = "float"
	ASMDataString  ASMDataValueKind = "string"
	ASMDataAddress ASMDataValueKind = "address"
)

type ASMDataAddressValue struct {
	Name   string
	Static bool
	ABI    uint16
	Addend int64
}

// ASMDataFragment retains the source value and its actual encoded bytes. Bits
// are the width-truncated integer or IEEE float bit pattern, not a Go type.
type ASMDataFragment struct {
	Ordinal, Line int
	Offset, Width int64
	Kind          ASMDataValueKind
	Bits          uint64
	String        []byte
	Address       *ASMDataAddressValue
}

type ASMGlobalLayout struct {
	Ordinal, Line int
	Name          string
	Static        bool
	ABI           uint16
	Flags         uint64
	DeclaredSize  int64
	Fragments     []ASMDataFragment
	Definition    ASMDataDefinition
}

// ASMDataDefinition is one exact object-local record. Repeated names, hashes
// or bytes do not merge identities or select a linker winner.
type ASMDataDefinition struct {
	Namespace           TextSymbolNamespace
	Relative, Index     uint32
	Name                string
	ABI                 uint16
	Kind, Flags, Flags2 uint8
	Size, Align         uint32
	Hash                []byte
	DataOffset          uint32
	Data                []byte
	Relocations         []TextRelocation
}

type ASMDataMember struct {
	Name         string
	NameBytes    []byte
	Ordinal      int
	Offset, Size int64
	ObjectOffset int64
	Flags        uint32
	Definitions  []ASMDataDefinition
	Texts        []TextRecord
}

// SDKASMDataSource contains whole, separately retained inputs. The offline
// API does not authenticate a CPP invocation, generated header or SDK origin.
// Original and Expanded must each describe the observed directive grammar;
// unknown macros and generated-header roles remain Context.
type SDKASMDataSource struct {
	Name                           string
	Original, Expanded             []byte
	OriginalObject, ExpandedObject []byte
}

type SDKASMDataObservationRequest struct {
	PackagePath   string
	Target        string
	Archive       []byte
	BuiltinSource []byte
	Sources       []SDKASMDataSource
	Referenced    map[string]SDKReferencedArchive
}

type ASMDataSourceObservation struct {
	Name                           string
	OriginalSHA256, ExpandedSHA256 string
	OriginalObjectSHA256           string
	ExpandedObjectSHA256           string
	OriginalProgramSHA256          string
	ExpandedProgramSHA256          string
	Layouts                        []ASMGlobalLayout
	ExpandedLayouts                []ASMGlobalLayout
	Original, Expanded             ASMDataMember
}

type SDKASMDataObservation struct {
	Protocol, PackagePath, Target string
	ArchiveSHA256                 string
	Members                       []ASMDataMember
	Sources                       []ASMDataSourceObservation
}

// ObserveX86SDKASMData is intentionally separate from the existing owned and
// SDK CPU readers. Successful offline parsing cannot mint their permissions.
func ObserveX86SDKASMData(request SDKASMDataObservationRequest) (*SDKASMDataObservation, error) {
	if request.Target != "linux/386" && request.Target != "linux/amd64" || request.PackagePath == "" ||
		strings.ContainsAny(request.PackagePath, "\x00\r\n") || len(request.PackagePath) > 4096 {
		return nil, contextError("unobserved SDK DATA target or package identity")
	}
	if len(request.Sources) == 0 || len(request.Sources) > 62 {
		return nil, contextError("SDK DATA source inventory violates the archive member bound")
	}
	total := int64(len(request.Archive))
	names := make([]string, len(request.Sources))
	for index, source := range request.Sources {
		if filepath.Base(source.Name) != source.Name || !strings.HasSuffix(source.Name, ".s") {
			return nil, contextError("SDK DATA source name differs from its exact selected route")
		}
		for _, input := range [][]byte{source.Original, source.Expanded, source.OriginalObject, source.ExpandedObject} {
			if len(input) == 0 || len(input) > archiveLimit {
				return nil, contextError("SDK DATA source/object input violates its complete byte bound")
			}
			total += int64(len(input))
		}
		names[index] = source.Name
	}
	for _, dependency := range request.Referenced {
		total += int64(len(dependency.Archive))
	}
	if total > 64<<20 {
		return nil, contextError("SDK DATA whole input graph exceeds its existing byte bound")
	}
	memberNames, err := sdkAssemblyMemberNames(names)
	if err != nil {
		return nil, err
	}
	referenced, err := sdkReferencedArchiveSymbols(request.Referenced, request.Target)
	if err != nil {
		return nil, err
	}
	members, err := archiveMembers(request.Archive)
	if err != nil {
		return nil, err
	}
	if len(members) != len(names)+2 || members[0].name != "__.PKGDEF" || members[1].name != "_go_.o" {
		return nil, contextError("SDK DATA archive differs from the complete selected inventory")
	}
	observation := &SDKASMDataObservation{Protocol: ASMDataProtocol, PackagePath: request.PackagePath,
		Target: request.Target, ArchiveSHA256: digest(request.Archive)}
	for ordinal, member := range members {
		value := ASMDataMember{Name: member.name, NameBytes: []byte(member.name), Ordinal: ordinal,
			Offset: member.offset, Size: int64(len(member.data))}
		if ordinal != 0 {
			flags := uint32(16)
			if ordinal > 1 {
				if member.name != memberNames[ordinal-2] {
					return nil, contextError("SDK DATA archive/source ordinal or exact byte-truncated member name differs")
				}
				flags |= 4
			}
			value, err = observeASMDataObject(member.data, request.BuiltinSource, request.Target, flags, referenced)
			if err != nil {
				return nil, err
			}
			value.Name, value.NameBytes, value.Ordinal, value.Offset, value.Size = member.name, []byte(member.name), ordinal, member.offset, int64(len(member.data))
			value.ObjectOffset += member.offset
		}
		observation.Members = append(observation.Members, value)
	}
	for index, source := range request.Sources {
		originalProgram, err := asmDataProgramFingerprint(source.Original)
		if err != nil {
			return nil, err
		}
		expandedProgram, err := asmDataProgramFingerprint(source.Expanded)
		if err != nil || originalProgram != expandedProgram {
			return nil, contextError("whole original and expanded ASM token programs differ")
		}
		originalStates, err := parseASMDataSource(source.Original, request.PackagePath, request.Target)
		if err != nil {
			return nil, err
		}
		expandedStates, err := parseASMDataSource(source.Expanded, request.PackagePath, request.Target)
		if err != nil {
			return nil, err
		}
		original, err := observeASMDataObject(source.OriginalObject, request.BuiltinSource, request.Target, 20, referenced)
		if err != nil {
			return nil, err
		}
		expanded, err := observeASMDataObject(source.ExpandedObject, request.BuiltinSource, request.Target, 20, referenced)
		if err != nil {
			return nil, err
		}
		if err := joinASMSourceTexts(source.Original, request.PackagePath, original); err != nil {
			return nil, err
		}
		if err := joinASMSourceTexts(source.Expanded, request.PackagePath, expanded); err != nil {
			return nil, err
		}
		layouts, err := joinASMDataLayouts(originalStates, original)
		if err != nil {
			return nil, err
		}
		expandedLayouts, err := joinASMDataLayouts(expandedStates, expanded)
		if err != nil {
			return nil, err
		}
		if !sameASMDataObjects(original, observation.Members[index+2]) || !sameASMDataObjects(original, expanded) {
			return nil, contextError("whole original/expanded/compiler ASM DATA or TEXT object identities differ")
		}
		// Ordinals and physical source lines stay in both retained source records;
		// compare the directive grammar separately from those original positions.
		normalize := func(values []ASMGlobalLayout) []ASMGlobalLayout {
			result := append([]ASMGlobalLayout(nil), values...)
			for index := range result {
				result[index].Ordinal, result[index].Line = 0, 0
				result[index].Definition.DataOffset = 0
				result[index].Fragments = append([]ASMDataFragment(nil), result[index].Fragments...)
				for ordinal := range result[index].Fragments {
					result[index].Fragments[ordinal].Ordinal, result[index].Fragments[ordinal].Line = 0, 0
				}
			}
			return result
		}
		if !reflect.DeepEqual(normalize(layouts), normalize(expandedLayouts)) {
			return nil, contextError("original and expanded DATA/GLOBL grammar differs")
		}
		observation.Sources = append(observation.Sources, ASMDataSourceObservation{Name: source.Name,
			OriginalSHA256: digest(source.Original), ExpandedSHA256: digest(source.Expanded),
			OriginalObjectSHA256: digest(source.OriginalObject), ExpandedObjectSHA256: digest(source.ExpandedObject),
			OriginalProgramSHA256: originalProgram, ExpandedProgramSHA256: expandedProgram,
			Layouts: layouts, ExpandedLayouts: expandedLayouts, Original: original, Expanded: expanded})
	}
	return observation, nil
}
