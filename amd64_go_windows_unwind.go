package plan9asm

import (
	"bytes"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"golang.org/x/arch/x86/x86asm"
)

func (source *GoAMD64FrameSource) observeWindowsUnwind(object []byte) (gostaticstorage.WindowsAMD64AssemblyUnwind, error) {
	var empty gostaticstorage.WindowsAMD64AssemblyUnwind
	for _, declaration := range source.declarations {
		if !fixedLocalTransportRole(declaration.transport) {
			return empty, goAMD64FrameContext("Windows unwind join requires the closed typed ABI0 transport scope")
		}
	}
	builtin, err := gostaticstorage.WindowsAMD64UnwindSources(source.env["GOVERSION"], func(name string) ([]byte, error) {
		return source.read(filepath.Join(source.env["GOROOT"], filepath.FromSlash(name)))
	})
	if err != nil {
		return empty, goAMD64FrameContext("Windows unwind actual SDK sources: %v", err)
	}
	observed, err := gostaticstorage.ObserveWindowsAMD64AssemblyUnwind(object, builtin)
	if err != nil {
		return empty, goAMD64FrameContext("Windows unwind complete object metadata: %v", err)
	}
	if observed.Version != source.env["GOVERSION"] {
		return empty, goAMD64FrameContext("Windows unwind object differs from its actual producer")
	}
	return observed, nil
}

func goAMD64WindowsUnwindByteRow(line string, offset int, expected []byte) ([]byte, error) {
	fields := strings.Fields(line)
	if len(expected) != 12 || offset != 0 || len(fields) != 14 || fields[0] != "0x0000" {
		return nil, goAMD64FrameContext("unknown Windows unwind listing byte/relocation/extent record")
	}
	actual := make([]byte, 12)
	var ascii strings.Builder
	for index := range actual {
		if len(fields[index+1]) != 2 {
			return nil, goAMD64FrameContext("Windows unwind listing has an incomplete byte record")
		}
		value, err := strconv.ParseUint(fields[index+1], 16, 8)
		if err != nil {
			return nil, goAMD64FrameContext("Windows unwind listing byte is invalid")
		}
		actual[index] = byte(value)
		printed := byte('.')
		if actual[index] >= ' ' && actual[index] <= 0x7e {
			printed = actual[index]
		}
		ascii.WriteByte(printed)
	}
	if !bytes.Equal(actual, expected) || fields[13] != ascii.String() {
		return nil, goAMD64FrameContext("Windows unwind listing differs from complete original binary bytes")
	}
	return actual, nil
}

// This private join consumes the factory's independently read whole object.
// A public offline observation is never itself an authorization input.
func goAMD64FrameWindowsUnwindListing(listing []byte, file *File, pkg GoPackage, origins map[string]int, directory string, observed gostaticstorage.WindowsAMD64AssemblyUnwind) (map[string]goAMD64FrameFunction, error) {
	functions, err := goAMD64FrameListingWithWindowsUnwind(listing, file, pkg, origins, &observed)
	if err != nil {
		return nil, err
	}
	for _, text := range observed.Texts {
		if len(text.Relocations) != 0 {
			return nil, goAMD64FrameContext("Windows object/listing TEXT relocation requires its separate complete observer")
		}
	}
	if err := goAMD64FrameWindowsUnwindFunctions(functions, origins, directory, observed); err != nil {
		return nil, err
	}
	return functions, nil
}

func goAMD64FrameWindowsUnwindFunctions(functions map[string]goAMD64FrameFunction, origins map[string]int, directory string, observed gostaticstorage.WindowsAMD64AssemblyUnwind) error {
	if len(observed.Texts) != len(functions) {
		return goAMD64FrameContext("Windows object/listing complete TEXT inventories differ")
	}
	unwind := make(map[string]gostaticstorage.WindowsAMD64UnwindRecord)
	for _, record := range observed.Unwind {
		unwind[record.Name] = record
	}
	seen := make(map[string]bool)
	for _, text := range observed.Texts {
		fn, exists := functions[text.Name]
		if !exists || seen[text.Name] || text.ABI != 0 || text.Kind != 1 || text.Flags != 16 || text.Flags2 != observed.TextFlags2 ||
			!bytes.Equal(text.Code, fn.code) {
			return goAMD64FrameContext("Windows object/listing TEXT name/ABI/code role differs")
		}
		seen[text.Name] = true
		var metadata gostaticstorage.WindowsAMD64UnwindRecord
		for _, aux := range text.Auxiliary {
			if aux.Type == observed.AuxType {
				metadata = unwind[aux.Target.Name]
			}
		}
		first, err := x86asm.Decode(fn.code, 64)
		if err != nil || first.Op != x86asm.PUSH || first.Args[0] != x86asm.RBP || first.Len != int(metadata.PushEnd) {
			return goAMD64FrameContext("Windows unwind push end differs from actual automatic prologue")
		}
		second, err := x86asm.Decode(fn.code[first.Len:], 64)
		if err != nil || second.Op != x86asm.MOV || second.Args[0] != x86asm.RBP || second.Args[1] != x86asm.RSP ||
			first.Len+second.Len != int(metadata.PrologueEnd) {
			return goAMD64FrameContext("Windows unwind frame end differs from actual automatic prologue")
		}
		if err := goAMD64WindowsUnwindSourcePCs(text, fn, directory, origins); err != nil {
			return err
		}
	}
	return nil
}

func goAMD64WindowsUnwindSourcePCs(text gostaticstorage.TextRecord, fn goAMD64FrameFunction, directory string, origins map[string]int) error {
	normalize := func(name string) string {
		if !filepath.IsAbs(name) {
			name = filepath.Join(directory, filepath.FromSlash(name))
		}
		return filepath.Clean(name)
	}
	// Go object file tables use slash paths on Windows, while captured
	// origins retain host paths and the command's relative source spelling.
	// Compare their already-bound paths consistently with the listing join;
	// aliases must still agree on the complete original source extent.
	linesByPath := make(map[string]int, len(origins))
	for name, lines := range origins {
		path := normalize(name)
		if previous, found := linesByPath[path]; found && previous != lines {
			return goAMD64FrameContext("Windows source origin aliases disagree on their original extent")
		}
		linesByPath[path] = lines
	}
	var previous int64
	for _, span := range text.Positions {
		if span.File == "" || span.Line <= 0 || int(span.Line) > linesByPath[normalize(span.File)] || span.Start < 0 ||
			span.Start != previous || span.End <= span.Start || span.End > int64(len(fn.code)) {
			return goAMD64FrameContext("Windows object source-PC interval lacks its original file/line/extent")
		}
		previous = span.End
	}
	if previous != int64(len(fn.code)) {
		return goAMD64FrameContext("Windows source-PC intervals do not cover the complete physical TEXT")
	}
	for pc, positions := range fn.sourcePositions {
		matched := false
		for _, span := range text.Positions {
			if span.Start > int64(pc) || span.End <= int64(pc) {
				continue
			}
			for _, position := range positions {
				if normalize(position.file) == normalize(span.File) && position.line == int(span.Line) {
					matched = true
				}
			}
		}
		if !matched {
			return goAMD64FrameContext("Windows object/listing source-PC file/line witness differs")
		}
	}
	return nil
}
