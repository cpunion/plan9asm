package plan9asm

import (
	"fmt"
	"testing"
)

func TestARM64RNDRNamedAndRawStatusAgreement(t *testing.T) {
	for _, operation := range []struct {
		name string
		word uint32
	}{{"RNDR", 0xd53b2400}, {"RNDRRS", 0xd53b2420}} {
		for _, instruction := range []string{
			"MRS " + operation.name + ",R0",
			"MOVD " + operation.name + ",R0",
			fmt.Sprintf("WORD $%#08x", operation.word),
		} {
			source := "TEXT comparison(SB),516,$0-0\n" + instruction + "\nCSET EQ,R0\nRET\n"
			requireARM64GoAssemblerResult(t, source, true)
			for _, api := range []string{"text", "module", "public-module", "Go"} {
				t.Run(api+"/"+instruction, func(t *testing.T) {
					present, err := arm64FPControlTargetAPI(t, api, source, "aarch64-unknown-linux-gnu")
					if !present || err != nil {
						t.Fatalf("original-Go RNG defines all NZCV bits for every spelling: output=%t err=%v", present, err)
					}
				})
			}
		}
	}
}
