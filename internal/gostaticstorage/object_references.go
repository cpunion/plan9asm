package gostaticstorage

import (
	"path/filepath"
	"sort"
	"strings"
)

// Each slice is exactly the compiler object's SymbolDefs block. Hashed and
// nonpackage definitions cannot be reached by a foreign package SymbolIndex.
// The live SDK source producer separately authenticates the map's package,
// export path and whole archive bytes; these records alone grant no role.
type sdkReferencedSymbols map[string][]indexedObjectSymbol

func resolveSDKIndexedReference(ref TextReference, diagnostic string, present bool, referenced sdkReferencedSymbols) (TextReference, error) {
	symbols, exists := referenced[ref.Package]
	if !exists || uint64(ref.SymbolIndex) >= uint64(len(symbols)) {
		return ref, contextError("TEXT foreign reference lacks actual indexed compiler definition")
	}
	symbol := symbols[ref.SymbolIndex]
	if symbol.Namespace != TextSymbolSelf || symbol.Relative != ref.SymbolIndex || symbol.Index != ref.SymbolIndex ||
		symbol.Name == "" || symbol.ABI != 0 && symbol.ABI != 1 || symbol.Kind == 0 ||
		present && diagnostic != symbol.Name {
		return ref, contextError("TEXT foreign indexed definition or diagnostic identity differs")
	}
	ref.Name, ref.ABI, ref.ABIKnown = symbol.Name, symbol.ABI, true
	ref.Kind, ref.KindKnown = symbol.Kind, true
	return ref, nil
}

func sdkReferencedArchiveSymbols(archives map[string]SDKReferencedArchive, target string) (sdkReferencedSymbols, error) {
	if len(archives) > 4096 {
		return nil, contextError("excessive SDK referenced archive inventory")
	}
	paths := make([]string, 0, len(archives))
	for path := range archives {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	result := make(sdkReferencedSymbols, len(archives))
	var totalBytes int
	for _, path := range paths {
		archive := archives[path]
		if path == "" || path != archive.PackagePath || !filepath.IsLocal(filepath.FromSlash(path)) ||
			filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))) != path {
			return nil, contextError("SDK referenced package key/path differs")
		}
		totalBytes += len(archive.Archive)
		if totalBytes > 64<<20 {
			return nil, contextError("complete SDK referenced archives exceed their byte bound")
		}
		members, err := archiveMembers(archive.Archive)
		if err != nil {
			return nil, err
		}
		if len(members) != len(archive.Sources)+2 || members[0].name != "__.PKGDEF" || members[1].name != "_go_.o" {
			return nil, contextError("SDK referenced archive differs from selected member inventory")
		}
		seen := make(map[string]bool)
		for index, member := range members[1:] {
			flags := uint32(16)
			if index != 0 {
				name := archive.Sources[index-1]
				if filepath.Base(name) != name || !strings.HasSuffix(name, ".s") || seen[name] ||
					member.name != strings.TrimSuffix(name, ".s")+".o" {
					return nil, contextError("SDK referenced source/member route differs")
				}
				seen[name] = true
				flags |= 4
			}
			body, err := objectBody(member.data, target)
			if err != nil {
				return nil, err
			}
			object, err := checkedObject(body, flags, index != 0)
			if err != nil {
				return nil, err
			}
			symbols, err := object.indexedSymbols()
			if err != nil {
				return nil, err
			}
			if index == 0 {
				count := len(object.block(3)) / 21
				result[path] = append([]indexedObjectSymbol(nil), symbols[:count]...)
			}
		}
	}
	return result, nil
}
