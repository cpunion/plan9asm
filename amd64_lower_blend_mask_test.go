package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestTranslateX86MaskBlendCompleteGoAssemblerForms(t *testing.T) {
	// Go 1.27 shares _yvblendmpd across VPBLENDM{B,W,D,Q}: matching X/Y/Z
	// sources and destination, optional K1-K7 selection mask, and .Z. D/Q
	// additionally enable scalar-memory broadcast on the first Plan 9 source.
	for _, target := range []struct {
		name   string
		goarch string
		triple string
	}{
		{name: "darwin-amd64", goarch: "amd64", triple: "x86_64-apple-darwin"},
		{name: "linux-amd64", goarch: "amd64", triple: "x86_64-unknown-linux-gnu"},
		{name: "windows-amd64", goarch: "amd64", triple: "x86_64-pc-windows-msvc"},
		{name: "linux-386", goarch: "386", triple: "i386-unknown-linux-gnu"},
		{name: "windows-386", goarch: "386", triple: "i686-pc-windows-msvc"},
	} {
		t.Run(target.name, func(t *testing.T) {
			lastZ := 20
			if target.goarch == "386" {
				lastZ = 7
			}
			var source strings.Builder
			source.WriteString("TEXT maskblendforms(SB),NOSPLIT,$0-0\n")
			for _, op := range []string{"VPBLENDMB", "VPBLENDMW", "VPBLENDMD", "VPBLENDMQ"} {
				fmt.Fprintf(&source, "\t%s X1, X20, X21\n", op)
				fmt.Fprintf(&source, "\t%s 8(AX), Y20, Y21\n", op)
				fmt.Fprintf(&source, "\t%s Z1, Z2, Z%d\n", op, lastZ)
				if target.goarch == "amd64" {
					fmt.Fprintf(&source, "\t%s X1, X20, K1, X21\n", op)
					fmt.Fprintf(&source, "\t%s.Z Y1, Y20, K2, Y21\n", op)
					fmt.Fprintf(&source, "\t%s.Z 40(AX), Z20, K3, Z21\n", op)
				}
				if strings.HasSuffix(op, "D") || strings.HasSuffix(op, "Q") {
					fmt.Fprintf(&source, "\t%s.BCST 104(AX), X20, X21\n", op)
					fmt.Fprintf(&source, "\t%s.BCST 112(AX), Y20, Y21\n", op)
					fmt.Fprintf(&source, "\t%s.BCST 120(AX), Z2, Z%d\n", op, lastZ)
					if target.goarch == "amd64" {
						fmt.Fprintf(&source, "\t%s.BCST.Z 128(AX), Z20, K4, Z21\n", op)
					}
				}
			}
			source.WriteString("\tRET\n")

			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			ll, err := Translate(file, Options{
				TargetTriple: target.triple,
				Goarch:       target.goarch,
				Sigs:         map[string]FuncSig{"maskblendforms": {Name: "maskblendforms", Ret: Void}},
			})
			if err != nil {
				t.Fatal(err)
			}
			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Skip("LLVM 22 llc not found")
			}
			compileLLVMToObject(t, llc, target.triple, "mask-blend-"+target.name+".ll", "mask-blend-"+target.name+".o", ll)
		})
	}
}

func TestTranslateX86MaskBlendRejectsFormsOutsideGoAssemblerTable(t *testing.T) {
	for _, instruction := range []string{
		"VPBLENDMQ X0, X1",
		"VPBLENDMQ X0, Y1, Y2",
		"VPBLENDMQ X0, (AX), X2",
		"VPBLENDMQ X0, X1, K0, X2",
		"VPBLENDMQ X0, X1, AX",
		"VPBLENDMQ.Z X0, X1, X2",
		"VPBLENDMQ.BCST X0, X1, X2",
		"VPBLENDMB.BCST (AX), X1, X2",
		"VPBLENDMW.BCST (AX), X1, X2",
		"VPBLENDMQ.Z.BCST (AX), X1, K1, X2",
		"VPBLENDMQ.RN_SAE X0, X1, X2",
	} {
		t.Run(strings.NewReplacer(" ", "_", "(", "_", ")", "_").Replace(instruction), func(t *testing.T) {
			assertX86MaskBlendRejected(t, "amd64", "x86_64-unknown-linux-gnu", instruction)
		})
	}
	assertX86MaskBlendRejected(t, "386", "i386-unknown-linux-gnu", "VPBLENDMQ X0, X1, K1, X2")
	assertX86MaskBlendRejected(t, "386", "i386-unknown-linux-gnu", "VPBLENDMQ Z0, Z1, Z8")
}

func assertX86MaskBlendRejected(t *testing.T, goarch, triple, instruction string) {
	t.Helper()
	file, err := Parse(ArchAMD64, "TEXT bad(SB),NOSPLIT,$0-0\n\t"+instruction+"\n\tRET\n")
	if err != nil {
		return
	}
	if _, err := Translate(file, Options{
		TargetTriple: triple,
		Goarch:       goarch,
		Sigs:         map[string]FuncSig{"bad": {Name: "bad", Ret: Void}},
	}); err == nil {
		t.Fatalf("Translate accepted %q outside Go 1.27's VPBLENDM forms", instruction)
	}
}
