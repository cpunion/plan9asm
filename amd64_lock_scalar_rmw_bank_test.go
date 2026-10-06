package plan9asm

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func TestX86LockScalarRMWByteSourcePhysicalBank(t *testing.T) {
	for _, stem := range []string{"ADC", "ADD", "AND", "OR", "SBB", "SUB", "XOR"} {
		for _, source := range []string{"AH", "CH", "DH", "BH"} {
			for _, base := range []string{"SI", "R8", "R15"} {
				t.Run(stem+"/"+source+"/"+base, func(t *testing.T) {
					assembly := x86NamedLockOracleSource("amd64", "CLC\nLOCK\n"+stem+"B "+source+",("+base+")")
					code := x86LockScalarOriginalInstruction(t, "amd64", assembly)
					physical, err := x86asm.Decode(code, 64)
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("actual original Go bytes=%x operands=%s", code, x86asm.GoSyntax(physical, 0, nil))
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86OrderingHintTranslateAPI(t, api, assembly,
								"package ordering\nfunc locked(p *[16]byte)\n", "amd64", "")
							if base == "SI" {
								if err != nil || !strings.Contains(ir, "lock; "+strings.ToLower(stem)+"b") {
									t.Fatal("native high byte without REX lost its actual source bank", err)
								}
							} else if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
								t.Fatal("source high-byte spelling borrowed a different REX physical bank", err)
							}
						})
					}
				})
			}
		}
	}
}
