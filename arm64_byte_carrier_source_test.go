package plan9asm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// Reuse the native-width family, not Or32 or a fabricated allocation. All
// declarations and assembly calls are private, numeric and in one source file.
func arm64ByteCarrierFixture(t *testing.T) (string, string, []arm64NativeResultTailCase) {
	t.Helper()
	directory := t.TempDir()
	packagePath := "example.invalid/carrierfixture"
	cases := arm64NativeResultTailCases()
	var declarations, assembly strings.Builder
	declarations.WriteString("package carrierfixture\n\n")
	for _, test := range cases {
		name := strings.ToLower(test.name())
		declarations.WriteString(strings.ReplaceAll(test.declaration(), test.name(), name))
		fmt.Fprintf(&declarations, "func observe_%s(input uint64) (value,native,guard uint64)\n", name)
		fmt.Fprintf(&declarations, "func entry_%s(input uint64) (uint64,uint64,uint64) {\nreturn observe_%s(input)\n}\n", name, name)
		assembly.WriteString(strings.ReplaceAll(test.source(), test.name(), name))
		assembly.WriteString(strings.ReplaceAll(arm64NativeResultTailObserver(test), test.name(), name))
	}
	for name, data := range map[string]string{
		"go.mod": "module " + packagePath + "\ngo 1.20\n",
		"decl.go": declarations.String(), "original_arm64.s": assembly.String(),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory, packagePath, cases
}

func arm64ByteCarrierOriginal(t *testing.T, directory, packagePath string) ([]byte, []byte, []byte) {
	t.Helper()
	selection := arm64ABI0RegionCommand(t, "linux", directory, "list", "-json", "-compiled", "-export", "-mod=readonly", ".")
	var selected goAMD64FrameSelection
	decoder := json.NewDecoder(bytes.NewReader(selection))
	if err := decoder.Decode(&selected); err != nil {
		t.Fatal(err)
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || selected.ImportPath != packagePath ||
		len(selected.SFiles) != 1 || selected.SFiles[0] != "original_arm64.s" || selected.Export == "" {
		t.Fatalf("actual complete first-party source selection differs: %s", selection)
	}
	archive, err := os.ReadFile(selected.Export)
	if err != nil {
		t.Fatal(err)
	}
	arm64ABI0RegionCommand(t, "linux", directory, "tool", "asm", "-p", packagePath,
		"-o", "original.o", "original_arm64.s")
	object, err := os.ReadFile(filepath.Join(directory, "original.o"))
	if err != nil {
		t.Fatal(err)
	}
	builtins, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/goobj/builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	return archive, object, builtins
}

func TestARM64ByteCarrierActualArchiveObservation(t *testing.T) {
	directory, packagePath, cases := arm64ByteCarrierFixture(t)
	archive, object, builtins := arm64ByteCarrierOriginal(t, directory, packagePath)
	standalone, err := gostaticstorage.ObserveARM64AssemblyText(object, builtins, "linux/arm64")
	if err != nil {
		t.Fatalf("actual original Go caller/callee object is not observed: %v", err)
	}
	members, err := gostaticstorage.ObserveARM64AssemblyArchive(archive, builtins, "linux/arm64", []string{"original_arm64.s"})
	if err != nil {
		t.Fatalf("actual original Go complete caller/entry archive is not observed: %v", err)
	}
	if len(members) != 3 || members[1].Name != "_go_.o" || members[2].Name != "original_arm64.o" ||
		len(standalone) != 18 || len(members[2].Texts) != len(standalone) {
		t.Fatalf("incomplete original caller/callee/entry inventory: members=%d standalone=%d", len(members), len(standalone))
	}
	texts := make(map[string]gostaticstorage.TextRecord)
	for _, text := range members[2].Texts {
		texts[text.Name] = text
	}
	for _, original := range standalone {
		archived, exists := texts[original.Name]
		if !exists || original.ABI != archived.ABI || !bytes.Equal(original.Code, archived.Code) ||
			len(original.Relocations) != len(archived.Relocations) || len(archived.Positions) == 0 {
			t.Fatalf("standalone/archive original instruction join differs: %s", original.Name)
		}
	}
	for _, test := range cases {
		t.Run(test.name(), func(t *testing.T) {
			name := strings.ToLower(test.name())
			callee := texts[packagePath+"."+name]
			caller := texts[packagePath+".observe_"+name]
			if callee.ABI != 0 || caller.ABI != 0 || len(callee.Code) < 12 || len(caller.Code) == 0 {
				t.Fatal("original ABI0 caller/callee physical identity is missing")
			}
			// The actual native store keeps its whole width even when its Go
			// result is narrow. No padded FrameSlot is introduced by this test.
			store := binary.LittleEndian.Uint32(callee.Code[4:8])
			wanted := map[int]uint32{16: 0x790023e0, 32: 0xb90013e0, 64: 0xf9000be0}[test.storeBits]
			if store != wanted {
				t.Fatalf("native store width changed: got=%08x want=%08x", store, wanted)
			}
			calls := 0
			for _, relocation := range caller.Relocations {
				if relocation.Target.Name != callee.Name {
					continue
				}
				calls++
				pc := relocation.Offset
				if relocation.Size != 4 || pc < 0 || pc+8 > int64(len(caller.Code)) ||
					binary.LittleEndian.Uint32(caller.Code[pc:pc+4])&0xfc000000 != 0x94000000 ||
					!relocation.Target.ABIKnown || relocation.Target.ABI != callee.ABI {
					t.Fatalf("original direct call/continuation is not bound: %+v", relocation)
				}
			}
			if calls != 1 {
				t.Fatalf("same-file caller has %d original direct calls", calls)
			}
		})
	}
	for _, test := range []struct {
		name string
		data []byte
		target string
		sources []string
	}{
		{"truncated", archive[:len(archive)-1], "linux/arm64", []string{"original_arm64.s"}},
		{"trailing", append(append([]byte(nil), archive...), 0), "linux/arm64", []string{"original_arm64.s"}},
		{"wrong-target", archive, "linux/amd64", []string{"original_arm64.s"}},
		{"omitted-member", archive, "linux/arm64", nil},
		{"foreign-member", archive, "linux/arm64", []string{"foreign.s"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := gostaticstorage.ObserveARM64AssemblyArchive(test.data, builtins, test.target, test.sources); err == nil {
				t.Fatal("incomplete or mismatched original archive was accepted")
			}
		})
	}
}
