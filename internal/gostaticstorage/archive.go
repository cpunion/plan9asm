// Package gostaticstorage observes bounded actual Go static storage. An object
// record is not a pointer, caller, frame, initialized result or read permission.
package gostaticstorage

import "errors"

var ErrContext = errors.New("static storage needs actual source context")

type relocation struct {
	Offset int64
	Size   int64
	Type   uint16
}

type dataRecord struct {
	Name          string
	Kind          uint8
	ABI           uint16
	Flags, Flags2 uint8
	Size, Align   int64
	Data          []byte
	Relocations   []relocation
}

// readArchive is intentionally a missing-proof boundary until the actual
// binary object/source producer is implemented. This scaffold changes no
// existing translation or declaration-only initialization consumer.
func readArchive(data []byte, target string) (map[string]dataRecord, error) {
	return nil, ErrContext
}
