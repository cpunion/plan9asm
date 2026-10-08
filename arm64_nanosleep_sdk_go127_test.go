//go:build go1.27

package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// This is an independent full original SDK/archive observation, not a source,
// ABI or memory token. Success here never certifies the whole SDK file's IR.
func TestARM64NanosleepSDKOriginalObjectAndBody(t *testing.T) {
	selection := arm64ABI0RegionCommand(t, "linux", ".", "list", "-json", "-compiled", "-export", "runtime")
	var selected goAMD64FrameSelection
	decoder := json.NewDecoder(bytes.NewReader(selection))
	if err := decoder.Decode(&selected); err != nil {
		t.Fatal(err)
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || selected.ImportPath != "runtime" || selected.Export == "" {
		t.Fatal("actual complete runtime selection is missing or ambiguous")
	}
	const sourceName = "sys_linux_arm64.s"
	present := false
	for _, name := range selected.SFiles {
		present = present || name == sourceName
	}
	if !present {
		t.Fatal("actual Linux/ARM64 Go selection omitted the original SDK source")
	}
	sourcePath := filepath.Join(selected.Dir, sourceName)
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	file, err := Parse(ArchARM64, string(data))
	if err != nil {
		t.Fatal("complete original SDK source", err)
	}
	var original strings.Builder
	count := 0
	for _, function := range file.Funcs {
		if function.Sym == "runtime·usleep" {
			count++
			for _, instruction := range function.Instrs {
				original.WriteString(instruction.Raw + "\n")
			}
		}
	}
	if count != 1 {
		t.Fatal("original SDK usleep TEXT membership differs")
	}
	archive, err := os.ReadFile(selected.Export)
	if err != nil {
		t.Fatal(err)
	}
	builtins, err := os.ReadFile(filepath.Join(testGOROOT(t), "src/cmd/internal/goobj/builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	// Runtime's compiler member is not an ordinary user-package observer role.
	// Do not broaden that observer: let the actual Go archive tool enumerate
	// every member and extract only this original assembler object. The bytes
	// still supply observation only, not a source capability or an ABI grant.
	inventory := arm64ABI0RegionCommand(t, "linux", ".", "tool", "pack", "t", selected.Export)
	memberName := strings.TrimSuffix(sourceName, ".s") + ".o"
	if len(memberName) > 16 {
		memberName = memberName[:16] // Go's actual ar member-name field.
	}
	members := strings.Fields(string(inventory))
	seen := make(map[string]bool)
	for _, name := range members {
		if seen[name] {
			t.Fatal("ambiguous actual Go archive member inventory")
		}
		seen[name] = true
	}
	if len(members) < 3 || !seen[memberName] {
		t.Fatal("actual original assembler member absent")
	}
	object := arm64ABI0RegionCommand(t, "linux", ".", "tool", "pack", "p", selected.Export, memberName)
	texts, err := gostaticstorage.ObserveARM64AssemblyText(object, builtins, "linux/arm64")
	if err != nil {
		t.Fatal("complete original runtime ASM member observation", err)
	}
	count = 0
	for _, text := range texts {
		if text.Name != "runtime.usleep" || text.ABI != 0 {
			continue
		}
		count++
		if len(text.Code) != 96 || len(text.Relocations) != 0 ||
			binary.LittleEndian.Uint32(text.Code[32:]) != 0xf90007e3 ||
			binary.LittleEndian.Uint32(text.Code[52:]) != 0xf9000be5 ||
			binary.LittleEndian.Uint32(text.Code[56:]) != 0x910023e0 ||
			binary.LittleEndian.Uint32(text.Code[60:]) != 0xaa1f03e1 ||
			binary.LittleEndian.Uint32(text.Code[64:]) != 0xd2800ca8 ||
			binary.LittleEndian.Uint32(text.Code[68:]) != 0xd4000001 {
			t.Fatal("original private timespec/NULL/R8/SVC0 object evidence differs")
		}
		t.Logf("SDK_ORIGINAL_BODY sourceSHA256=%x archiveSHA256=%x member=%s codeSHA256=%x bytes=%d relocations=%d",
			sha256.Sum256(data), sha256.Sum256(archive), memberName, sha256.Sum256(text.Code), len(text.Code), len(text.Relocations))
	}
	if count != 1 {
		t.Fatal("ambiguous original ABI0 body; a compiler wrapper is not the ASM body")
	}
	declarationPath := filepath.Join(selected.Dir, "stubs2.go")
	declarationBytes, err := os.ReadFile(declarationPath)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	declarations, err := parser.ParseFile(fset, declarationPath, declarationBytes, 0)
	if err != nil {
		t.Fatal(err)
	}
	var declaration bytes.Buffer
	for _, node := range declarations.Decls {
		if function, ok := node.(*ast.FuncDecl); ok && function.Name.Name == "usleep" && function.Body == nil {
			if err := format.Node(&declaration, fset, function); err != nil {
				t.Fatal(err)
			}
		}
	}
	if declaration.Len() == 0 {
		t.Fatal("actual ordinary SDK declaration absent")
	}
	pkg := mustGoPackage(t, "runtime", "package runtime\n"+declaration.String()+"\n")
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			if _, err := arm64NanosleepPackageAPI(t, api, "aarch64-unknown-linux-gnu", original.String(), pkg); err != nil {
				t.Fatal("exact original SDK usleep body (not a whole-file pass)", err)
			}
		})
	}
}
