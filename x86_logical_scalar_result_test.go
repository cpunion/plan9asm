package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// A zero-argument manual signature uses the existing logical AX return
// transport. It is not an assertion about Go ABI0, ABIInternal or native entry.
func TestX86Mode32LogicalScalarReturnMustBeInitialized(t *testing.T) {
	for _, row := range []struct {
		typ   LLVMType
		store string
	}{
		{I8, "MOVB"}, {I16, "MOVW"}, {I32, "MOVL"},
	} {
		for _, sample := range []struct {
			name, value string
			pass        bool
		}{
			{"initialized", row.store + " $7,AX\n", true},
			{"missing", "", false},
			{"wrong-register", row.store + " $7,BX\n", false},
			{"one-path-only", "MOVL $1,DX\nCMPL DX,$0\nJE done\n" + row.store + " $7,AX\ndone:\n", false},
		} {
			// A private bit-preserving FLAGS copy does not expose its unknown
			// bits. It also cannot manufacture missing AX result bytes.
			source := "TEXT ·logical(SB),4,$0-0\nPUSHFL\nPOPL CX\n" + sample.value + "RET\n"
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			sig := FuncSig{Name: "logical", Ret: row.typ}
			for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
				t.Run(fmt.Sprintf("%s/%s/%s", row.typ, sample.name, api), func(t *testing.T) {
					options := Options{
						Goarch: "386", TargetTriple: "i386-unknown-linux-gnu",
						Sigs:       map[string]FuncSig{"logical": sig},
						ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") },
					}
					ir, err := x86PackedFloatStorageAPI(api, file, options)
					if sample.pass && (err != nil || ir == "") ||
						!sample.pass && (!errors.Is(err, ErrProbeNeedsContext) || ir != "") {
						t.Fatal("logical AX result proof disagrees with actual reaching bytes", sample.pass, err)
					}
				})
			}
			if !errors.Is(CheckX86ABI0ResultInitialization(file.Funcs[0], sig, "386"), ErrProbeNeedsContext) {
				t.Fatal("manual logical return invented a Go ABI0 declaration")
			}
		}
	}
}

func TestX86Mode32LogicalScalarReturnCannotBypassProof(t *testing.T) {
	for _, body := range []string{"", "MOVL $7,BX\n", "MOVB $7,AX\n", "MOVW $7,AX\n"} {
		file, err := Parse(ArchAMD64, "TEXT ·logical(SB),4,$0-0\n"+body+"RET\n")
		if err != nil {
			t.Fatal(err)
		}
		for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
			t.Run(fmt.Sprintf("%q/%s", body, api), func(t *testing.T) {
				options := Options{
					Goarch: "386", TargetTriple: "i386-unknown-linux-gnu",
					Sigs:       map[string]FuncSig{"logical": {Name: "logical", Ret: I32}},
					ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") },
				}
				ir, err := x86PackedFloatStorageAPI(api, file, options)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("missing AX bytes escaped through a cost-only proof selector", err)
				}
			})
		}
	}
}
