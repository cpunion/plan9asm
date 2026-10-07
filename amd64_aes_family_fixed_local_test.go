package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86AESFamilyActualBoundFixedLocalConsumers(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 required")
	}
	var assembly, declarations strings.Builder
	count := 0
	for _, name := range x86AESOracleNames[:6] {
		for _, source := range []string{"16(SP)", "tmp-144(SP)"} {
			row, _ := x86AESOracleRow(name, source, "X")
			if strings.HasPrefix(name, "V") && !strings.Contains(name, "IMC") && !strings.Contains(name, "KEYGEN") {
				row = name + " " + source + ",X0,X0"
			} else if strings.Contains(name, "KEYGEN") {
				row = name + " $27," + source + ",X0"
			} else {
				row = name + " " + source + ",X0"
			}
			body, _, _ := x86PrivateAlignedFrameAuditFixture(t, row)
			fn := fmt.Sprintf("F%d", count)
			count++
			assembly.WriteString(strings.Replace(body, "·Value", "·"+fn, 1))
			fmt.Fprintf(&declarations, "func %s(a,b,c,d uint64) uint64\n", fn)
		}
	}
	for _, target := range []struct{ name, triple string }{{"linux/amd64", "x86_64-unknown-linux-gnu"}, {"darwin/amd64", "x86_64-apple-darwin"}} {
		t.Run(target.name, func(t *testing.T) {
			pkg, request, file, options := x86CMOVFixedLocalFixture(t, assembly.String(), declarations.String(), target.name, target.triple)
			observer := requireGoAMD64FrameCapture(t, pkg, request)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			bound, err := observer.BindFixedLocal(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("AES consumer of independently sealed fixed local native bytes", err)
			}
			fixedLocalAPIs(t, pkg, file, bound, func(t *testing.T, ir string, err error) {
				if err != nil {
					t.Fatal(err)
				}
				compileLLVMToObject(t, llc, target.triple, "aes-native-local.ll", "aes-native-local.o", ir)
			})
		})
	}
}
