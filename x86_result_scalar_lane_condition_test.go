package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// This contains no shifts or literal masks. The same numeric counter controls
// both the conditional byte definition and its eventual observable store.
func x86ScalarLaneConditionalSource() string {
	return `TEXT ·view(SB),4,$0-32
MOVQ out+0(FP),DI
MOVQ input+8(FP),SI
MOVQ n+16(FP),BX
MOVQ other+24(FP),CX
CMPQ BX,$2
JB one
PINSRB $1,1(SI),X0
one:
PINSRB $0,(SI),X0
CMPQ BX,$2
JB last
PEXTRB $1,X0,1(DI)
last:
PEXTRB $0,X0,(DI)
RET
`
}

func TestX86ScalarLaneConditionalPrefixSameCounter(t *testing.T) {
	const declaration = "func view(out,input *byte,n,other uint64)\n"
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, x86ScalarLaneConditionalSource(), declaration,
				"amd64", "x86_64-unknown-linux-gnu", api)
			if err != nil || ir == "" {
				t.Fatalf("matching counter conditions lost their defined byte: output=%t err=%v", ir != "", err)
			}
		})
	}
}

func TestX86ScalarLaneConditionalPrefixRejectsUnrelatedConditions(t *testing.T) {
	const declaration = "func view(out,input *byte,n,other uint64)\n"
	source := x86ScalarLaneConditionalSource()
	for _, row := range []struct{ name, source string }{
		{"missing-byte", strings.Replace(source, "PINSRB $1,1(SI),X0\n", "", 1)},
		{"different-counter", strings.Replace(source, "PINSRB $0,(SI),X0\nCMPQ BX", "PINSRB $0,(SI),X0\nCMPQ CX", 1)},
		{"counter-overwritten", strings.Replace(source, "PINSRB $0,(SI),X0\n", "PINSRB $0,(SI),X0\nMOVQ $2,BX\n", 1)},
		{"different-bound", strings.Replace(source, "PINSRB $0,(SI),X0\nCMPQ BX,$2", "PINSRB $0,(SI),X0\nCMPQ BX,$1", 1)},
		{"signed-input-branch", strings.Replace(source, "JB one", "JLT one", 1)},
		{"overwritten-flags", strings.Replace(source, "JB last", "CMPQ CX,$2\nJB last", 1)},
	} {
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(fmt.Sprintf("%s/%s", row.name, api), func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, row.source, declaration,
					"amd64", "x86_64-unknown-linux-gnu", api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("unrelated predicate supplied a missing byte: output=%t err=%v", ir != "", err)
				}
			})
		}
	}
}
