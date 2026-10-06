package plan9asm

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// Select the semantic family by name, but derive its entire vector syntax
// denominator from Go's independent encoder reader, never the lowerer spec.
func x86OfficialPackedNarrowRows(t *testing.T, arch string) []string {
	t.Helper()
	byOpcode := x86OfficialPackedBinaryWitnesses(t)
	var rows []string
	for _, opcode := range []string{"VPACKSSDW", "VPACKSSWB", "VPACKUSDW", "VPACKUSWB"} {
		witnesses := byOpcode[opcode]
		if len(witnesses) == 0 {
			t.Fatal("missing independent official pack family", opcode)
		}
		for _, witness := range witnesses {
			if arch == "386" && strings.Count(witness, ",") == 3 {
				continue // Actual mode32 frontend refusals are a separate gate.
			}
			rows = append(rows, witness)
		}
	}
	return rows
}

func TestX86PackedNarrowOfficialRowsThreeAPIsFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"},
		{"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			declarations.WriteString("package narrow\n")
			rows := x86OfficialPackedNarrowRows(t, target.arch)
			for index, row := range rows {
				fmt.Fprintf(&source, "TEXT ·F%d(SB),4,$0-0\n%s\nRET\n", index, row)
				fmt.Fprintf(&declarations, "func F%d()\n", index)
			}
			goAssemblerOracleOutput(t, target.arch, source.String(), "", true)
			pkg := mustGoPackage(t, "example.invalid/narrow", declarations.String())
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			sigs := map[string]FuncSig{}
			for index := range rows {
				name := fmt.Sprintf("F%d", index)
				sig, err := DeriveGoFuncABI0Signature(name, pkg.Types.Scope().Lookup(name).(*types.Func),
					target.arch, types.SizesFor("gc", target.arch))
				if err != nil {
					t.Fatal(err)
				}
				sigs[name] = sig
			}
			options := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: sigs,
				ResolveSym: func(symbol string) string { return strings.TrimPrefix(symbol, "·") }}
			before := fmt.Sprintf("%#v %#v", file, sigs)
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "narrow-text.ll", "narrow-text.o", ir)
			context := llvm.NewContext()
			defer context.Dispose()
			module, err := TranslateModuleInContext(context, file, options)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "narrow-module.ll", "narrow-module.o", module.String())
			translated, err := translateGoModuleInContext(context, pkg, []byte(source.String()), GoModuleOptions{
				GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: options.ResolveSym,
			})
			if translated != nil && !translated.Module.IsNil() {
				defer translated.Module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "narrow-go.ll", "narrow-go.o", translated.Module.String())
			if before != fmt.Sprintf("%#v %#v", file, sigs) {
				t.Fatal("object coverage changed original source/signatures")
			}
			t.Logf("%d independent Go-accepted vector forms: three APIs and LLVM 22 objects, not runtime/caller coverage", len(rows))
		})
	}
}

func TestX86PackedNarrowMode32ActualGoRawForms(t *testing.T) {
	rows := x86OfficialPackedNarrowRows(t, "386")
	source := "TEXT rawpack(SB),4,$0-0\n" + strings.Join(rows, "\n") + "\nRET\n"
	goAssemblerOracleOutput(t, "386", source, "", true)
	code := assembleX87ControlBytes(t, "386", source)
	decoded, err := decodeX86RawDirectives(rawX86Function(code), "386")
	if err != nil {
		t.Fatal("actual Go mode32 pack bytes lost physical decoding", err)
	}
	named, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	want := named.Funcs[0].Instrs[1:]
	if len(decoded.Instrs) != len(want) {
		t.Fatal("actual byte decoder lost instructions", len(decoded.Instrs), len(want))
	}
	for index, instruction := range decoded.Instrs {
		if instruction.Op != want[index].Op || !reflect.DeepEqual(instruction.Args, want[index].Args) {
			t.Fatalf("mode32 instruction %d: %+v != %+v", index, instruction, want[index])
		}
	}
}

func TestX86PackedNarrowE4NFMemoryRemainsObservable(t *testing.T) {
	for _, opcode := range []string{"VPACKSSDW", "VPACKSSWB", "VPACKUSDW", "VPACKUSWB"} {
		for _, bank := range []string{"X", "Y", "Z"} {
			rows := []string{opcode + ".Z (AX)," + bank + "0,K1," + bank + "1"}
			if opcode == "VPACKSSDW" || opcode == "VPACKUSDW" {
				rows = append(rows, opcode+".BCST.Z (AX),"+bank+"0,K1,"+bank+"1")
			}
			for _, row := range rows {
				t.Run(row, func(t *testing.T) {
					source := "TEXT pack(SB),4,$0-0\nMOVL $0,CX\nKMOVQ CX,K1\n" + row + "\nRET\n"
					goAssemblerOracleOutput(t, "amd64", source, "", true)
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					ir, err := Translate(file, Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu",
						Sigs: map[string]FuncSig{"pack": {Name: "pack", Ret: Void}}})
					if err != nil {
						t.Fatal(err)
					}
					if !strings.Contains(ir, "load volatile") ||
						!strings.Contains(x86MulDivRoundOptimize(t, ir), "load volatile") {
						t.Fatal("E4NF memory read disappeared for a zero mask/dead output")
					}
				})
			}
		}
	}
}
