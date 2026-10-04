package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// These are independent original Go corpus examples, not strings generated
// from the translator's form table. The native opcode needs original storage;
// a software SP/BP allocation must not impersonate its fault/alignment contract.
func TestX86SignedPackedFloatOfficialStackStorageContext(t *testing.T) {
	for _, row := range []string{
		"VCVTDQ2PD -17(BP), K7, Y6",
		"VCVTDQ2PS -17(BP)(SI*4), K4, X31",
		"VCVTDQ2PS 17(SP)(BP*2), K1, Y19",
		"VCVTDQ2PS 17(SP), K4, X31",
	} {
		t.Run(row, func(t *testing.T) {
			source := "TEXT f(SB),4,$0-0\n" + row + "\nRET\n"
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			options := Options{
				Goarch:       "amd64",
				TargetTriple: "x86_64-unknown-linux-gnu",
				Sigs:         map[string]FuncSig{"f": {Ret: Void}},
			}
			assertX86PackedFloatStorageContext(t, file, options)
		})
	}
}

func TestX86SignedPackedFloatAllWidthsZeroMaskStorageContext(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			var original strings.Builder
			index := 0
			for _, op := range []string{"CVTPL2PS", "CVTPL2PD", "VCVTDQ2PS", "VCVTDQ2PD"} {
				banks, modes := []string{"X"}, []string{""}
				if strings.HasPrefix(op, "V") {
					banks = []string{"X", "Y", "Z"}
					modes = []string{"", "K1,", ".Z K1,", ".BCST.Z K1,"}
				}
				for _, bank := range banks {
					for _, address := range []string{"16(SP)", "16(BP)", "16(AX)(BP*2)"} {
						for _, mode := range modes {
							mnemonic, mask := op, mode
							if split := strings.IndexByte(mode, ' '); split >= 0 {
								mnemonic += mode[:split]
								mask = mode[split+1:]
							}
							for _, maskValue := range []int{0, 1} {
								row := fmt.Sprintf("%s %s,%s%s1", mnemonic, address, mask, bank)
								name := fmt.Sprintf("storage%d", index)
								index++
								source := fmt.Sprintf("TEXT %s(SB),4,$0-0\nMOVL $%d,AX\nKMOVW AX,K1\n%s\nRET\n", name, maskValue, row)
								original.WriteString(source)
								t.Run(fmt.Sprintf("%s/mask=%d", row, maskValue), func(t *testing.T) {
									file, err := Parse(ArchAMD64, source)
									if err != nil {
										t.Fatal(err)
									}
									options := Options{
										Goarch: arch,
										Sigs:   map[string]FuncSig{name: {Ret: Void}},
									}
									assertX86PackedFloatStorageContext(t, file, options)
								})
							}
						}
					}
				}
			}
			if index != 156 {
				t.Fatalf("storage boundary denominator changed: %d", index)
			}
			goAssemblerOracleOutput(t, arch, original.String(), "", true)
			t.Logf("%d original Go accepted stack-boundary forms; all three APIs retain Context", index)
		})
	}
}

func assertX86PackedFloatStorageContext(t *testing.T, file *File, options Options) {
	t.Helper()
	for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
		var err error
		if api == "Translate" {
			_, err = Translate(file, options)
		} else {
			context := llvm.NewContext()
			var module llvm.Module
			if api == "TranslateModule" {
				module, err = TranslateModule(file, options)
			} else {
				module, err = TranslateModuleInContext(context, file, options)
			}
			if !module.IsNil() {
				module.Dispose()
			}
			context.Dispose()
		}
		if !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatalf("%s authorized unbound native stack storage: %v", api, err)
		}
	}
}
