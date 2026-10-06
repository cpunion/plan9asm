package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestX86LockScalarRMWUnknownIncomingCarryNeedsContext(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		widths := "BWL"
		if arch == "amd64" {
			widths += "Q"
		}
		for _, stem := range []string{"ADC", "SBB"} {
			for _, width := range widths {
				register := "CX"
				if width == 'B' {
					register = "CL"
				}
				body := "LOCK\n\t" + stem + string(width) + " " + register + ",(SI)"
				t.Run(arch+"/"+stem+string(width), func(t *testing.T) {
					source := x86NamedLockOracleSource(arch, body)
					machine := x86LockScalarOriginalInstruction(t, arch, source)
					t.Logf("actual Go accepts original instruction=%x; incoming CF remains undefined by the source", machine)
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86OrderingHintTranslateAPI(t, api, source,
								"package ordering\nfunc locked(p *[16]byte)\n", arch, "")
							if !errors.Is(err, ErrProbeNeedsContext) || ir != "" ||
								!strings.Contains(err.Error(), "carry") {
								t.Fatalf("native carry-dependent memory effect borrowed an undefined entry flag: %v", err)
							}
						})
					}
				})
			}
		}
	}
}
