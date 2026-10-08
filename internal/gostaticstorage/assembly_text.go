package gostaticstorage

import (
	"path/filepath"
	"strconv"
	"strings"
)

// ObserveX86AssemblyText decodes complete original Go assembler object bytes.
// It is an offline observation: neither these bytes nor its return value create
// an Observation, a live source lease or a memory/entry permission.
func ObserveX86AssemblyText(data, builtinSource []byte, target string) ([]TextRecord, error) {
	if target != "linux/386" && target != "linux/amd64" {
		return nil, contextError("unobserved x86 assembly target")
	}
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
	if target != "linux/386" && target != "linux/amd64" {
		return nil, contextError("unobserved x86 assembly archive target")
	}
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
	result := []TextMember{{Name: members[0].name, Offset: members[0].offset, Size: int64(len(members[0].data))}}
	for index, member := range members[1:] {
		flags := uint32(0)
		if index != 0 {
			name := sources[index-1]
			if filepath.Base(name) != name || !strings.HasSuffix(name, ".s") || seenSources[name] || member.name != strings.TrimSuffix(name, ".s")+".o" {
				return nil, contextError("assembly source/member route differs")
			}
			seenSources[name] = true
			flags = 4
		}
		body, err := objectBody(member.data, target)
		if err != nil {
			return nil, err
		}
		records, err := readObjectRole(body, flags, false)
		if err != nil {
			return nil, err
		}
		for name, record := range records {
			key := name + "\x00" + strconv.Itoa(int(record.ABI))
			if definitions[key] || storage[name] || index != 0 && record.Kind != 1 {
				return nil, contextError("duplicate same-ABI/storage or assembly DATA definition")
			}
			definitions[key] = true
			storage[name] = record.Kind != 1
		}
		texts, err := readTextObject(body, target, flags, false, builtins)
		if err != nil {
			return nil, err
		}
		result = append(result, TextMember{Name: member.name, Offset: member.offset, Size: int64(len(member.data)), Flags: flags, Texts: texts})
	}
	return result, nil
}
