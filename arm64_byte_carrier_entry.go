package plan9asm

import "github.com/xgo-dev/plan9asm/internal/gostaticstorage"

// This offline, private observation is not a live caller region or authority.
// The consumer must independently bind its complete source, allocation,
// continuation, initialized bytes and native effects before using a carrier.
type arm64ByteCarrierEntryObservation struct {
	allocation, dataStart, dataBytes int64
	callPC, continuationPC           int64
	parameterStores, resultLoads     []arm64ByteCarrierMemory
}

func arm64ByteCarrierTypedEntry(text gostaticstorage.TextRecord, declaration FuncSig) (arm64ByteCarrierEntryObservation, error) {
	return arm64ByteCarrierEntryObservation{}, arm64ByteCarrierContext("actual typed entry allocation/effect observation is not implemented")
}
