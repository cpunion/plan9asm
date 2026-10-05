package plan9asm

import (
	"reflect"
	"strings"
	"testing"
)

const arm64PrivateCallCopySource = `TEXT ·Copy(SB),4,$0-24
MOVD dst+0(FP),R8
MOVD src+8(FP),R1
MOVD n+16(FP),R3
CMP $32,R3
BLS bounded
MOVD $32,R3
bounded:
CALL copy<>(SB)
RET
TEXT copy<>(SB),516,$0-0
CBZ R3,copy0
CMP $16,R3
BLE copy16
LDP (R1),(R21,R22)
ADD R1,R3,R15
LDP -16(R15),(R23,R24)
STP (R21,R22),(R8)
ADD R8,R3,R26
STP (R23,R24),-16(R26)
RET
copy16:
ADD R1,R3,R15
ADD R8,R3,R26
CMP $8,R3
BLT copy7
MOVD (R1),R21
MOVD -8(R15),R22
MOVD R21,(R8)
MOVD R22,-8(R26)
RET
copy7:
TBZ $2,R3,copy3
MOVWU (R1),R21
MOVWU -4(R15),R22
MOVW R21,(R8)
MOVW R22,-4(R26)
RET
copy3:
TBZ $1,R3,copy1
MOVHU (R1),R21
MOVHU -2(R15),R22
MOVH R21,(R8)
MOVH R22,-2(R26)
RET
copy1:
MOVBU (R1),R21
MOVBU R21,(R8)
copy0:
RET
`

func TestARM64PrivateCallCFGSourcePreparation(t *testing.T) {
	register := strings.Replace(arm64PrivateRegisterHelperSource("516,$0-0", "CALL mix<>(SB)"),
		"ADC $0,R9,R20\nMUL R19,R20,R20\nRET", `CBZ R19,zero
ADC $0,R9,R20
MUL R19,R20,R20
RET
zero:
MOVD $7,R20
RET`, 1)
	for _, test := range []struct {
		name, source string
	}{
		{"register_multiple_returns", register},
		{"typed_scalar_pair_copy", arm64PrivateCallCopySource},
		{"two_static_calls", strings.Replace(arm64PrivateCallCopySource,
			"CALL copy<>(SB)", "CALL copy<>(SB)\nCALL copy<>(SB)", 1)},
		{"caller_label_collision", strings.Replace(arm64PrivateCallCopySource,
			"CALL copy<>(SB)", "__arm64_private_0:\ncopy0:\nCALL copy<>(SB)", 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			requireARM64GoAssemblerResult(t, test.source, true)
			file, err := Parse(ArchARM64, test.source)
			if err != nil {
				t.Fatal(err)
			}
			original, err := Parse(ArchARM64, test.source)
			if err != nil {
				t.Fatal(err)
			}
			prepared, err := CoalesceARM64PrivateRegisterHelpers(file)
			if err != nil {
				t.Fatal("closed private call CFG did not preserve its source continuation:", err)
			}
			if prepared == file || len(prepared.Funcs) != 1 || !reflect.DeepEqual(file, original) {
				t.Fatal("continuation preparation changed the original file or invented a callable helper")
			}
			repeated, err := CoalesceARM64PrivateRegisterHelpers(prepared)
			if err != nil || !reflect.DeepEqual(prepared, repeated) {
				t.Fatalf("repeated continuation preparation changed its source: %v", err)
			}
		})
	}
}
