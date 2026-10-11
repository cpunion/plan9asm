package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"golang.org/x/arch/x86/x86asm"
)

// These are the complete original Go 1.27.1 metadata producer/schema inputs.
// The factory reads the actual SDK files itself; caller observations cannot
// replace them or grant a Source, handler or native unwind role.
var goAMD64WindowsUnwindSources = map[string]string{
	"src/cmd/internal/obj/x86/seh.go":       "3e7b9a27ed35bb507c39685d3aff193c758566baa4a392816cdf4d19e8419e67",
	"src/cmd/internal/goobj/objfile.go":     "caf835a17e23aa02cae9bda6e8723d8be359846fc4453bd53f3a3947744b938a",
	"src/cmd/internal/obj/objfile.go":       "803546227c02cf46bfa4158be3dfd8067e7457f4e4f78349cb6d42c700b2fcb7",
	"src/cmd/internal/obj/sym.go":           "2bb2646a6ccff23aa42d8c4960be7cf82af5570ae5a83a1a355b7b822ee2149d",
	"src/cmd/internal/obj/plist.go":         "50c121d6c72c356c605176e02bce549dd6e0e8107af6c0eed467f7b3adb418e3",
	"src/cmd/internal/objabi/symkind.go":    "de287864b7ff9a3614becc538cbabdccdecc030a848aad5d5bbef6ec7ec5c225",
	"src/cmd/internal/objabi/reloctype.go":  "49bebd30a27d05a24de77398ab524dfbeb0630c8b88c06646c1e4734e4f1fc9d",
	"src/cmd/internal/goobj/builtinlist.go": "8f38ef757cdebb27963ee9befe24dd8041b86d217937db30f13e4d3d91207311",
}

func (source *GoAMD64FrameSource) observeWindowsUnwind(object []byte) (gostaticstorage.WindowsAMD64AssemblyUnwind, error) {
	var empty gostaticstorage.WindowsAMD64AssemblyUnwind
	for _, declaration := range source.declarations {
		if !fixedLocalTransportRole(declaration.transport) {
			return empty, goAMD64FrameContext("Windows unwind join requires the closed typed ABI0 transport scope")
		}
	}
	var builtin []byte
	for name, expected := range goAMD64WindowsUnwindSources {
		data, err := source.read(filepath.Join(source.env["GOROOT"], filepath.FromSlash(name)))
		if err != nil {
			return empty, err
		}
		if digest := sha256.Sum256(data); hex.EncodeToString(digest[:]) != expected {
			return empty, goAMD64FrameContext("Windows unwind producer/schema differs from actual pinned SDK: %s", name)
		}
		if name == "src/cmd/internal/goobj/builtinlist.go" {
			builtin = data
		}
	}
	observed, err := gostaticstorage.ObserveWindowsAMD64AssemblyUnwind(object, builtin)
	if err != nil {
		return empty, goAMD64FrameContext("Windows unwind complete object metadata: %v", err)
	}
	return observed, nil
}

var goAMD64WindowsUnwindHeader = regexp.MustCompile(`^(\S+) SSEHUNWINDINFO dupok size=12 align=0x4$`)

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
		if !exists || seen[text.Name] || text.ABI != 0 || text.Kind != 1 || text.Flags != 16 || text.Flags2 != 0 ||
			!bytes.Equal(text.Code, fn.code) {
			return goAMD64FrameContext("Windows object/listing TEXT name/ABI/code role differs")
		}
		seen[text.Name] = true
		var metadata gostaticstorage.WindowsAMD64UnwindRecord
		for _, aux := range text.Auxiliary {
			if aux.Type == 14 {
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
	var previous int64
	for _, span := range text.Positions {
		if span.File == "" || span.Line <= 0 || int(span.Line) > origins[span.File] || span.Start < 0 ||
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
