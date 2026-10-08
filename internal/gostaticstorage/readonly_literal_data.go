package gostaticstorage

// ReadonlyLiteralDataImage is a pure initializer plan, never a live owner,
// pointer, numerical-read or native-body capability. Its object-local identity
// is retained instead of choosing a global by name or hash.
type ReadonlyLiteralDataImage struct {
	SourceOrdinal, MemberOrdinal int
	SourceName, MemberName       string
	DeclaredSize                 int64
	Definition                   ASMDataDefinition
	Fragments                    []ASMDataFragment
	Initial                      []byte
}

// ReadonlyLiteralDataInitializers examines already observed complete DATA
// records. Only the separate live SDK factory can create an initializer owner.
func ReadonlyLiteralDataInitializers(observation *SDKASMDataObservation) ([]ReadonlyLiteralDataImage, error) {
	return nil, contextError("private readonly literal DATA initializer is not implemented")
}
