//go:build go1.27

package plan9asm

import (
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestX86DirectionOriginalSDKIndexByte386ThreeAPIsNextContext(t *testing.T) {
	root := runtime.GOROOT()
	dir := filepath.Join(root, "src", "internal", "bytealg")
	declaration, err := os.ReadFile(filepath.Join(dir, "indexbyte_native.go"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := ReadGoAssemblySourceAtGOROOT(filepath.Join(dir, "indexbyte_386.s"), dir, root)
	if err != nil {
		t.Fatal(err)
	}
	pkg := mustGoPackage(t, "internal/bytealg", string(declaration))
	file, err := Parse(ArchAMD64, string(source))
	if err != nil || len(file.Funcs) != 2 {
		t.Fatal("complete original SDK member changed", err)
	}
	for _, function := range file.Funcs {
		name := strings.TrimPrefix(function.Sym, "·")
		declaration := pkg.Types.Scope().Lookup(name)
		if declaration == nil {
			t.Fatal("complete original declaration disappeared", name)
		}
		sig, err := DeriveGoFuncABI0Signature(name, declaration.(*types.Func), "386", types.SizesFor("gc", "386"))
		if err != nil {
			t.Fatal(err)
		}
		for _, query := range []func(Func, FuncSig, string) error{CheckX86ABI0ResultInitialization, checkX86ScalarFPResultInitialization} {
			err := query(function, sig, "386")
			if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "instruction 5: REPN") {
				t.Fatal("original TEXT has not retained the independent repeat contract", name, err)
			}
		}
	}
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewPackageAPI(t, string(source), pkg, "386", "i386-unknown-linux-gnu", api)
			// This is complete unchanged source/declaration numerical observation,
			// not a compiler graph/native-memory grant or a whole-file pass.
			if !errors.Is(err, ErrProbeNeedsContext) || ir != "" || strings.Contains(err.Error(), "instruction 4: CLD") || !strings.Contains(err.Error(), "REPN") {
				t.Fatal("DF control has not advanced to the independent repeated-memory contract", err)
			}
			t.Log(err)
		})
	}
}

func TestX86DirectionOriginalGo127CompleteYnoneRows(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "x86", "asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	table := string(data)
	if !strings.Contains(table, "var ynone = []ytab{\n\t{Zlit, 1, argList{}},\n}") {
		t.Fatal("complete Go ynone operand table changed")
	}
	rows := map[string]string{
		"CLD": "Px, opBytes{0xfc}", "STD": "Px, opBytes{0xfd}",
		"REP": "Px, opBytes{0xf3}", "REPN": "Px, opBytes{0xf2}",
	}
	for _, family := range []struct {
		name                   string
		byteOpcode, wideOpcode int
	}{
		{"MOVS", 0xa4, 0xa5}, {"STOS", 0xaa, 0xab},
		{"SCAS", 0xae, 0xaf}, {"CMPS", 0xa6, 0xa7}, {"LODS", 0xac, 0xad},
	} {
		for _, width := range []struct{ suffix, prefix string }{{"B", "Pb"}, {"W", "Pe"}, {"L", "Px"}, {"Q", "Pw"}} {
			opcode := family.wideOpcode
			if width.suffix == "B" {
				opcode = family.byteOpcode
			}
			rows[family.name+width.suffix] = fmt.Sprintf("%s, opBytes{0x%x}", width.prefix, opcode)
		}
	}
	if len(rows) != 24 || len(x86DirectionControlSpecs) != 2 || len(x86StringSpecs) != 20 {
		t.Fatal("complete family denominator changed")
	}
	for name, row := range rows {
		matches := regexp.MustCompile(`(?m)^\s*\{A`+name+`,[^\n]+`).FindAllString(table, -1)
		want := "{A" + name + ", ynone, " + row + "},"
		if len(matches) != 1 || strings.TrimSpace(matches[0]) != want {
			t.Fatal("complete original source instruction row changed", name, matches, want)
		}
	}
}
