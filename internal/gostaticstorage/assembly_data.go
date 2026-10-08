package gostaticstorage

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
	Layouts                        []ASMGlobalLayout
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
	return nil, contextError("assembly DATA source/object observation is not implemented")
}
