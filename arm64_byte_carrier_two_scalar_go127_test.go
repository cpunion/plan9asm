//go:build go1.27

package plan9asm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The second parameter is a real declared uint32, not permission to access its
// following gap. The assembly caller initializes that separate gap explicitly.
func arm64ByteCarrierTwoScalarFixture(t *testing.T, widths []arm64NativeResultTailCase) (string, string) {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	packagePath := "example.invalid/twoscalarcarrier"
	var declarations, assembly strings.Builder
	declarations.WriteString("package twoscalarcarrier\n\n")
	for _, width := range widths {
		name := strings.ToLower(width.name())
		fmt.Fprintf(&declarations, "func %s(input uint64, spare uint32) (result uint%d)\n", name, width.resultBits)
		fmt.Fprintf(&declarations, "func observe_%s(input uint64) (value,native,guard,paramgap uint64)\n", name)
		fmt.Fprintf(&assembly, "TEXT ·%s(SB),4,$0-%d\nMOVD input+0(FP),R0\n%s R0,result+16(FP)\nRET\n",
			name, 16+width.resultBits/8, width.op)
		fmt.Fprintf(&assembly, `
TEXT ·observe_%s(SB),4,$48-40
    MOVD input+0(FP),R0
    MOVD R0,8(RSP)
    MOVD $0xabad1dea00123456,R1
    MOVD R1,16(RSP)
    MOVD $0xfeedfacecafebeef,R1
    MOVD R1,24(RSP)
    MOVD $0x7ac35ef012345678,R1
    MOVD R1,32(RSP)
    CALL ·%s(SB)
    MOVD 24(RSP),R0
    MOVD R0,native+16(FP)
    %s 24(RSP),R0
    MOVD R0,value+8(FP)
    MOVD 32(RSP),R0
    MOVD R0,guard+24(FP)
    MOVD 16(RSP),R0
    MOVD R0,paramgap+32(FP)
    RET
`, name, name, map[int]string{8: "MOVBU", 16: "MOVHU", 32: "MOVWU"}[width.resultBits])
	}
	for name, data := range map[string]string{
		"go.mod":  "module " + packagePath + "\ngo 1.20\n",
		"decl.go": declarations.String(), "original_arm64.s": assembly.String(),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory, packagePath
}

func TestARM64ByteCarrierTwoScalarActualSourceTargetMatrix(t *testing.T) {
	widths := arm64NativeResultTailCases()
	if len(widths) != 9 {
		t.Fatal("native-width family is incomplete")
	}
	directory, packagePath := arm64ByteCarrierTwoScalarFixture(t, widths)
	request := arm64ByteCarrierRequest(t, directory, packagePath)
	lease, release := runtimeOracleStageContext(t)
	defer release()
	for _, goos := range []string{"darwin", "linux", "windows", "freebsd", "openbsd"} {
		t.Run(goos, func(t *testing.T) {
			request := request
			request.Target = goos + "/arm64"
			request.Environment = append([]string(nil), request.Environment...)
			for index, pair := range request.Environment {
				if strings.HasPrefix(pair, "GOOS=") {
					request.Environment[index] = "GOOS=" + goos
				}
			}
			source, err := CaptureGoARM64ByteCarrierSource(lease, request)
			if err != nil {
				t.Fatalf("producer precondition: %v", err)
			}
			defer source.Close()
			if len(source.report.Calls) != 9 || len(source.report.Standalone) != 18 {
				t.Fatal("original complete call/body inventory differs")
			}
			for _, width := range widths {
				name := strings.ToLower(width.name())
				declaration := source.declarations[name]
				if len(declaration.Frame.Params) != 2 || declaration.Frame.Params[0].Offset != 0 ||
					declaration.Frame.Params[1].Offset != 8 || declaration.Frame.Params[1].Type != I32 ||
					len(declaration.Frame.Results) != 1 || declaration.Frame.Results[0].Offset != 16 {
					t.Fatalf("actual two-scalar fields differ: %+v", declaration.Frame)
				}
			}
			if err := source.Verify(lease); err != nil {
				t.Fatal(err)
			}
			t.Log("actual producer complete: nine callees, nine same-file CALLs, FP0/8 parameters and FP16 result")
		})
	}
}

func TestARM64ByteCarrierTwoScalarNativeWidthConsumers(t *testing.T) {
	for _, width := range arm64NativeResultTailCases() {
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(width.name()+"/"+api, func(t *testing.T) {
				// A failed binding revokes its owner. Every case therefore captures
				// its own actual source; no earlier failure is reused as a RED.
				directory, packagePath := arm64ByteCarrierTwoScalarFixture(t, []arm64NativeResultTailCase{width})
				lease, release := runtimeOracleStageContext(t)
				defer release()
				source, err := CaptureGoARM64ByteCarrierSource(lease, arm64ByteCarrierRequest(t, directory, packagePath))
				if err != nil {
					t.Fatalf("producer precondition, not consumer RED: %v", err)
				}
				defer source.Close()
				if len(source.report.Calls) != 1 {
					t.Fatal("original same-file native CALL is missing")
				}
				t.Log("actual producer reached P before consumer")
				goSource, err := os.ReadFile(filepath.Join(directory, "decl.go"))
				if err != nil {
					t.Fatal(err)
				}
				assembly, err := os.ReadFile(filepath.Join(directory, "original_arm64.s"))
				if err != nil {
					t.Fatal(err)
				}
				pkg := mustGoPackage(t, packagePath, string(goSource))
				ir, err := arm64ByteCarrierTranslateAPI(t, api, source, pkg, string(assembly), "aarch64-unknown-linux-gnu", nil)
				if err != nil {
					t.Fatalf("observed two-scalar native-width consumer failed: %v", err)
				}
				if !strings.Contains(ir, "(i64 %arg0, i32 %arg1)") || !strings.Contains(ir, "store volatile i32") {
					t.Fatal("typed second-parameter transport was omitted")
				}
			})
		}
	}
}
