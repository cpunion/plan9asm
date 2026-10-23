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

func wasmIntegerCountGoNames(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/wasm/anames.go"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, match := range regexp.MustCompile(`"(I(?:32|64)(?:Clz|Ctz|Popcnt))"`).FindAllSubmatch(data, -1) {
		names = append(names, string(match[1]))
	}
	if len(names) != 6 {
		t.Fatal("complete Go unary integer-count namespace changed", names)
	}
	return names
}

func wasmIntegerCountBody(name, reg string) string {
	body := "Get " + reg + "\n"
	if strings.HasPrefix(name, "I32") {
		return body + "I32WrapI64\n" + name + "\nI64ExtendI32U\n"
	}
	return body + name + "\n"
}

func TestWASMIntegerCountCompleteGoNamespace(t *testing.T) {
	var source strings.Builder
	sigs := make(map[string]FuncSig)
	for _, name := range wasmIntegerCountGoNames(t) {
		for bank := 0; bank < 16; bank++ {
			function, reg := fmt.Sprintf("count%d", len(sigs)), fmt.Sprintf("R%d", bank)
			fmt.Fprintf(&source, "TEXT %s(SB),4,$0-0\n%sReturn\n", function, wasmIntegerCountBody(name, reg))
			sigs[function] = FuncSig{Name: function, Args: []LLVMType{I64}, ArgRegs: []Reg{Reg(reg)}, Ret: I64}
		}
	}
	for _, goos := range wasmPackedGoPlatforms {
		wasmDataGoObject(t, source.String(), goos)
	}
	file, err := Parse(ArchWASM, source.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, fn := range file.Funcs {
		sig, err := InferWASMAssemblyFuncSig(fn, fn.Sym)
		if err != nil || sig.Ret != I64 || len(sig.Args) != 1 || sig.Args[0] != I64 {
			t.Errorf("complete unary stack signature: %+v, %v", sig, err)
		}
	}
	for _, route := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
		t.Run(route, func(t *testing.T) {
			ir, err := wasmPackedTranslate(t, file, wasmPackedOptions(sigs, WASMABIDirect), route)
			if err != nil || ir == "" {
				t.Fatal("Go unary integer-count family missing", err)
			}
			if strings.Contains(ir, "blockaddress") {
				t.Fatal("Wasm scalar operations introduced blockaddress")
			}
			compileLLVMToObject(t, wasmDataTool(t, "llc"), "wasm32-unknown-unknown", "counts.ll", "counts.o", ir)
		})
	}
}

func TestWASMIntegerCountRequiresTypedStackAndCompleteFormat(t *testing.T) {
	for _, name := range wasmIntegerCountGoNames(t) {
		other, result := "I64Const $1\n", I32
		if strings.HasPrefix(name, "I64") {
			other, result = "I32Const $1\n", I64
		}
		bodies := []string{name, other + name, "Get F0\n" + name, "Get F16\n" + name}
		for _, operand := range []string{" $0", " R0", " R0,R1", " 0(SP)", ".invalid"} {
			bodies = append(bodies, name+operand, "RET\n"+name+operand)
		}
		for index, body := range bodies {
			t.Run(fmt.Sprintf("%s/%d", name, index), func(t *testing.T) {
				file, err := Parse(ArchWASM, "TEXT count(SB),4,$0-0\n"+body+"\nReturn\n")
				if err != nil {
					return // Original source syntax can reject before stack typing.
				}
				if sig, err := InferWASMAssemblyFuncSig(file.Funcs[0], "count"); err == nil {
					t.Errorf("inference accepted invalid unary stack as %+v", sig)
				}
				for _, route := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
					ir, err := wasmPackedTranslate(t, file, wasmPackedOptions(map[string]FuncSig{"count": {Name: "count", Ret: result}}, WASMABIDirect), route)
					if err == nil || ir != "" {
						t.Errorf("%s accepted invalid unary stack/format: %s", route, body)
					}
				}
			})
		}
	}
}
