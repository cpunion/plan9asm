package plan9asm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Discover the whole namespace from the selected Go compiler, independently
// of the implementation's map and the official corpus's observed mnemonics.
func wasmFloatBinaryGoNames(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "wasm", "anames.go"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, match := range regexp.MustCompile(`"(F(?:32|64)(?:Eq|Ne|Lt|Gt|Le|Ge|Min|Max))"`).FindAllSubmatch(data, -1) {
		names = append(names, string(match[1]))
	}
	if len(names) != 16 {
		t.Fatalf("selected Go float comparison/minmax namespace changed: %v", names)
	}
	return names
}

func wasmFloatBinaryTestTypes(name string) (LLVMType, LLVMType, Reg, Reg) {
	typ, first, second := LLVMType("float"), Reg("F0"), Reg("F1")
	if strings.HasPrefix(name, "F64") {
		typ, first, second = "double", "F16", "F17"
	}
	result := I32
	if strings.HasSuffix(name, "Min") || strings.HasSuffix(name, "Max") {
		result = typ
	}
	return typ, result, first, second
}

func TestWASMFloatBinaryWholeGoNamespace(t *testing.T) {
	for _, name := range wasmFloatBinaryGoNames(t) {
		t.Run(name, func(t *testing.T) {
			typ, result, first, second := wasmFloatBinaryTestTypes(name)
			source := fmt.Sprintf("TEXT operation(SB),4,$0-0\nGet %s\nGet %s\n%s\nReturn\n", first, second, name)
			wasmDataGoObject(t, source, "js")
			file, err := Parse(ArchWASM, source)
			if err != nil {
				t.Fatal(err)
			}
			sig := FuncSig{Name: "operation", Args: []LLVMType{typ, typ}, Ret: result, ArgRegs: []Reg{first, second}}
			for _, route := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
				ir, err := wasmPackedTranslate(t, file, wasmPackedOptions(map[string]FuncSig{"operation": sig}, WASMABIDirect), route)
				if err != nil || ir == "" {
					t.Errorf("%s actual Go opcode not implemented: %v", route, err)
				}
			}
			inferred, err := InferWASMAssemblyFuncSig(file.Funcs[0], "operation")
			if err != nil || inferred.Ret != result || len(inferred.Args) != 2 || inferred.Args[0] != typ || inferred.Args[1] != typ {
				t.Errorf("actual Go stack signature not inferred: %+v, %v", inferred, err)
			}
		})
	}
}

func TestWASMFloatBinaryRejectsOperandsEvenWhenDead(t *testing.T) {
	for _, name := range wasmFloatBinaryGoNames(t) {
		for _, operand := range []string{" $0", " $1.25", " F0", " F0,F1", " 0(SP)", " symbol(SB)"} {
			for _, dead := range []bool{false, true} {
				prefix := ""
				if dead {
					prefix = "RET\n"
				}
				file, err := Parse(ArchWASM, "TEXT operation(SB),4,$0-0\n"+prefix+name+operand+"\nRET\n")
				if err != nil {
					continue
				}
				for _, route := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
					ir, err := wasmPackedTranslate(t, file, wasmPackedOptions(map[string]FuncSig{"operation": {Name: "operation", Ret: Void}}, WASMABIDirect), route)
					if err == nil || ir != "" {
						t.Errorf("%s accepted %s%s, dead=%t", route, name, operand, dead)
					}
				}
			}
		}
	}
}
