package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func x86SignedPackedFloatRows(arch string) []string {
	var rows []string
	last := 15
	if arch == "386" {
		last = 7
	}
	for _, op := range []string{"CVTPL2PS", "CVTPL2PD"} {
		for _, source := range []string{"X0", "M0", "M7", "3(AX)", "16(AX)(CX*4)", "3(FS)", "16(GS)", "256"} {
			rows = append(rows, fmt.Sprintf("%s %s,X%d", op, source, last))
		}
	}
	for _, op := range []string{"VCVTDQ2PS", "VCVTDQ2PD"} {
		for _, bank := range []string{"X", "Y", "Z"} {
			src := bank
			if op == "VCVTDQ2PD" {
				src = "X"
				if bank == "Z" {
					src = "Y"
				}
			}
			for _, input := range []string{src + "0", "3(AX)", "16(AX)(CX*4)", "3(FS)", "16(GS)", "256"} {
				for _, mode := range []string{"", "K1,", "K7,"} {
					name := op
					if mode == "K7," {
						name += ".Z"
					}
					rows = append(rows, fmt.Sprintf("%s %s,%s%s7", name, input, mode, bank))
				}
			}
			for _, mask := range []string{"", "K1,", "K7,"} {
				name := op + ".BCST"
				if mask == "K7," {
					name += ".Z"
				}
				rows = append(rows, fmt.Sprintf("%s 3(AX),%s%s7", name, mask, bank))
			}
			if arch == "amd64" {
				rows = append(rows, fmt.Sprintf("%s %s31,%s31", op, src, bank), fmt.Sprintf("%s %s20,K3,%s21", op, src, bank))
			}
		}
		rows = append(rows, op+" X0,X0")
		if op == "VCVTDQ2PS" {
			rows = append(rows, op+" Y0,Y0")
		}
	}
	for _, rc := range []string{"RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
		for _, mask := range []string{"", "K1,", "K7,"} {
			name := "VCVTDQ2PS." + rc
			if mask == "K7," {
				name += ".Z"
			}
			rows = append(rows, name+" Z0,"+mask+"Z7")
		}
	}
	return rows
}

func TestX86SignedPackedIntegerFloatCompleteGoTables(t *testing.T) {
	dir := filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "x86")
	legacy, err := os.ReadFile(filepath.Join(dir, "asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	avx, err := os.ReadFile(filepath.Join(dir, "avx_optabs.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []string{
		"var yxcvm2 = []ytab{\n\t{Zm_r_xm, 2, argList{Yxm, Yxr}},\n\t{Zm_r_xm, 2, argList{Ymm, Yxr}},\n}",
		"{ACVTPL2PD, yxcvm2, Px, opBytes{Pf3, 0xe6, Pe, 0x2a}}", "{ACVTPL2PS, yxcvm2, Pm, opBytes{0x5b, 0, 0x2a, 0}}",
	} {
		if !strings.Contains(string(legacy), row) {
			t.Fatalf("actual Go legacy table changed: %s", row)
		}
	}
	for _, table := range []string{"_yvcvtdq2ps", "_yvcvtdq2pd"} {
		start := strings.Index(string(avx), "var "+table+" = []ytab{")
		if start < 0 {
			t.Fatal(table)
		}
		body := strings.SplitN(string(avx)[start:], "\n}", 2)[0]
		if strings.Count(body, "zcase:") != 8 {
			t.Fatalf("%s rows changed: %s", table, body)
		}
		for _, destination := range []string{"YxrEvex", "YyrEvex", "Yzr"} {
			source := "YxmEvex"
			if table == "_yvcvtdq2ps" && destination == "YyrEvex" || destination == "Yzr" {
				source = "YymEvex"
			}
			if table == "_yvcvtdq2ps" && destination == "Yzr" {
				source = "Yzm"
			}
			for _, row := range []string{fmt.Sprintf("args: argList{%s, %s}", source, destination), fmt.Sprintf("args: argList{%s, Yknot0, %s}", source, destination)} {
				if !strings.Contains(body, row) {
					t.Fatalf("actual %s row changed: %s", table, row)
				}
			}
		}
	}
	for _, name := range []string{"VCVTDQ2PS", "VCVTDQ2PD"} {
		start := strings.Index(string(avx), "{as: A"+name+",")
		if start < 0 {
			t.Fatal(name)
		}
		body := strings.SplitN(string(avx)[start:], "\n\t}},", 2)[0]
		if strings.Count(body, "avxEscape | vex") != 2 || strings.Count(body, "avxEscape | evex") != 3 || strings.Contains(body, "evexSaeEnabled") {
			t.Fatalf("actual encoding rows/SAE changed: %s", body)
		}
		for _, line := range strings.Split(body, "\n") {
			if !strings.Contains(line, "avxEscape | evex") {
				continue
			}
			width := 16
			if strings.Contains(line, "evex256") {
				width = 32
			}
			if strings.Contains(line, "evex512") {
				width = 64
			}
			tuple := width
			if name == "VCVTDQ2PD" {
				tuple /= 2
			}
			if !strings.Contains(line, fmt.Sprintf("evexN%d | evexBcstN4", tuple)) || !strings.Contains(line, "evexZeroingEnabled") ||
				strings.Contains(line, "evexRoundingEnabled") != (name == "VCVTDQ2PS" && width == 64) {
				t.Fatalf("actual tuple/ER changed: %s", line)
			}
		}
	}
	if len(x86SignedPackedFloatSpecs) != 4 {
		t.Fatal("signed dword family namespace incomplete")
	}
}

func TestX86SignedPackedIntegerFloatCompleteObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"}, {"amd64", "x86_64-unknown-linux-gnu"}, {"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"}, {"386", "i686-pc-windows-msvc"},
	} {
		t.Run(target.triple, func(t *testing.T) {
			var source strings.Builder
			sigs := map[string]FuncSig{}
			rows := x86SignedPackedFloatRows(target.arch)
			contextRows := 0
			for index, row := range rows {
				name := fmt.Sprintf("packed%d", index)
				one, err := Parse(ArchAMD64, "TEXT "+name+"(SB),4,$0-0\n"+row+"\nRET\n")
				if err != nil {
					t.Fatal(err)
				}
				form, _, err := x86SignedPackedFloatProperties(one.Funcs[0].Instrs[1], target.arch)
				if err != nil {
					t.Fatal(err)
				}
				if form.mmx {
					goAssemblerOracleOutput(t, target.arch, "TEXT "+name+"(SB),4,$0-0\n"+row+"\nRET\n", "", true)
					opt := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: map[string]FuncSig{name: {Ret: Void}}}
					for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
						if _, err := x86SignedPackedFloatTranslate(one, opt, api); !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("unclosed physical MMX caller was authorized", row, api, err)
						}
					}
					contextRows++
					continue
				}
				fmt.Fprintf(&source, "TEXT %s(SB),4,$0-0\n%s\nRET\n", name, row)
				sigs[name] = FuncSig{Ret: Void}
			}
			goAssemblerOracleOutput(t, target.arch, source.String(), "", true)
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			opt := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: sigs}
			for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
				ir, err := x86SignedPackedFloatTranslate(file, opt, api)
				if err != nil {
					t.Fatal(api, err)
				}
				if strings.Contains(ir, " = sitofp ") || strings.Contains(ir, " = call <16 x float> @llvm.experimental.constrained.sitofp") || strings.Count(ir, "asm sideeffect") != len(rows)-contextRows {
					t.Fatalf("native conversion dropped/scalarized: native=%d expected%d", strings.Count(ir, "asm sideeffect"), len(rows)-contextRows)
				}
				optimized := x86SignedIntegerFloatOptimize(t, ir)
				if strings.Count(optimized, "asm sideeffect") != len(rows)-contextRows {
					t.Fatal("O2 erased dead conversion side effects")
				}
				compileLLVMToObject(t, llc, target.triple, api+".ll", api+".o", optimized)
			}
			t.Logf("%d actual Go accepted forms: %d native LLVM x3 APIs, %d MMX Context x3 APIs", len(rows), len(rows)-contextRows, contextRows)
		})
	}
}

func x86SignedPackedFloatTranslate(file *File, opt Options, api string) (string, error) {
	if api == "Translate" {
		return Translate(file, opt)
	}
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	var module llvm.Module
	var err error
	if api == "TranslateModule" {
		module, err = TranslateModule(file, opt)
	} else {
		module, err = TranslateModuleInContext(ctx, file, opt)
	}
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		return "", err
	}
	return module.String(), nil
}

func TestX86SignedPackedIntegerFloatRetainsContextAndInvalidForms(t *testing.T) {
	for _, row := range []string{
		"CVTPL2PS AX,X0", "CVTPL2PD X0,Y1", "CVTPL2PS.Z X0,K1,X1", "VCVTDQ2PS Y0,X1", "VCVTDQ2PD Y0,Y1",
		"VCVTDQ2PD.SAE Y0,Z1", "VCVTDQ2PD.RN_SAE Y0,Z1", "VCVTDQ2PS.SAE Z0,Z1", "VCVTDQ2PS.RU_SAE X0,X1",
		"VCVTDQ2PS.RU_SAE (AX),Z1", "VCVTDQ2PS.BCST X0,X1", "VCVTDQ2PS.Z X0,X1", "VCVTDQ2PS X0,K0,X1",
		"VCVTDQ2PS.BCST.RU_SAE (AX),Z1", "VCVTDQ2PS X0,K1,(AX)",
	} {
		// Go's yxcvm2 also accepts a GP spelling that encodes an MMX
		// register. It is not a typed GP-source numerical contract.
		if row != "CVTPL2PS AX,X0" && row != "VCVTDQ2PS.Z X0,X1" {
			goAssemblerOracleOutput(t, "amd64", "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", "", false)
		}
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		if err != nil {
			continue
		}
		if _, err := Translate(file, Options{Goarch: "amd64", Sigs: map[string]FuncSig{"f": {Ret: Void}}}); err == nil {
			t.Fatal("accepted invalid", row)
		}
	}
	for _, row := range []string{"CVTPL2PS (SP),X1", "CVTPL2PD (BP),X1", "VCVTDQ2PS (AX)(BP*2),Y1", "VCVTDQ2PS.Z p+0(FP),K1,Z1", "VCVTDQ2PD p+0(FP),X1"} {
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-64\n"+row+"\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		frame := FrameLayout{}
		args := []LLVMType{}
		for byte := 0; byte < 64; byte++ {
			frame.Params = append(frame.Params, FrameSlot{Offset: int64(byte), Type: I8, Index: byte, Field: -1})
			args = append(args, I8)
		}
		if _, err := Translate(file, Options{Goarch: "amd64", Sigs: map[string]FuncSig{"f": {Ret: Void, Args: args, Frame: frame}}}); !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("fabricated native backing", row, err)
		}
	}
	for _, row := range []string{"VCVTDQ2PS X8,X0", "VCVTDQ2PS X0,X31", "VCVTDQ2PD X8,Y0", "VCVTDQ2PD Y0,Z8"} {
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Translate(file, Options{Goarch: "386", Sigs: map[string]FuncSig{"f": {Ret: Void}}}); !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("fabricated mode32 physical bank", row, err)
		}
	}
}

func TestX86SignedPackedIntegerFloatActualGoMasklessZeroBoundary(t *testing.T) {
	asm, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "x86", "asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	// The old encoder accepts EVEX.z without a nonzero writemask. Observe
	// its actual source/bytes, rather than guessing from runtime.Version.
	checksMask := strings.Contains(string(asm), "mask register must be specified for .Z instructions")
	const row = "VCVTDQ2PS.Z X0,X1"
	goAssemblerOracleOutput(t, "amd64", "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", "", !checksMask)
	if !checksMask {
		listing, err := x86ExtensionGoListing(t, "amd64", row)
		if err != nil {
			t.Fatal(err, listing)
		}
		var code []byte
		for _, line := range strings.Split(listing, "\n") {
			code = append(code, x86InsertExtractListedBytes(line)...)
		}
		if len(code) < 6 || code[0] != 0x62 || code[3]&0x87 != 0x80 {
			t.Fatal("historical maskless zero encoding changed", code)
		}
		if _, _, known, err := decodedX86SameWidthConversionInstruction(code, 64); !known || err == nil {
			t.Fatal("invalid maskless zero machine code accepted", code, known, err)
		}
	}
	file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Translate(file, Options{Goarch: "amd64", Sigs: map[string]FuncSig{"f": {Ret: Void}}}); err == nil {
		t.Fatal("maskless zero acquired a valid native contract")
	}
}

func TestX86SignedPackedIntegerFloatGoSuffixOrder(t *testing.T) {
	for _, row := range []string{
		"VCVTDQ2PS.Z.BCST (AX),K1,X1", "VCVTDQ2PD.Z.BCST (AX),K1,Y1",
		"VCVTDQ2PS.Z.RN_SAE Z0,K1,Z1", "VCVTDQ2PS.Z.RD_SAE Z0,K1,Z1",
		"VCVTDQ2PS.Z.RU_SAE Z0,K1,Z1", "VCVTDQ2PS.Z.RZ_SAE Z0,K1,Z1",
	} {
		goAssemblerOracleOutput(t, "amd64", "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", "", false)
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Translate(file, Options{Goarch: "amd64", Sigs: map[string]FuncSig{"f": {Ret: Void}}}); err == nil {
			t.Fatal("accepted invalid modifier order", row)
		}
	}
}

func TestX86SignedPackedIntegerFloatMMXUnclosedCallerContext(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		registers := []string{"M0", "M1", "M7", "AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"}
		if arch == "amd64" {
			registers = append(registers, "R8", "R15")
		}
		for _, op := range []string{"CVTPL2PS", "CVTPL2PD"} {
			for _, register := range registers {
				// The caller can observe physical MM0/M1/Tag/TOS after RET,
				// even when f itself contains only a conversion and a return.
				// No body whitelist can certify the absent caller contract.
				source := "TEXT f(SB),4,$0-0\n" + op + " " + register + ",X0\nRET\n" +
					"TEXT caller(SB),4,$0-0\nCALL f(SB)\nFXSAVE (AX)\nRET\n"
				goAssemblerOracleOutput(t, arch, source, "", true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				form, known, err := x86SignedPackedFloatProperties(file.Funcs[0].Instrs[1], arch)
				if !known || err != nil || !form.mmx || form.inputBytes != 8 {
					t.Fatal("actual Ymm class missing", arch, op, register, form, err)
				}
				opt := Options{Goarch: arch, Sigs: map[string]FuncSig{"f": {Ret: Void}, "caller": {Ret: Void}}}
				for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
					if _, err := x86SignedPackedFloatTranslate(file, opt, api); !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("ordinary API fabricated physical MMX caller closure", arch, op, register, api, err)
					}
				}
			}
		}
	}
}

func TestX86SignedPackedIntegerFloatIdentityAndISAContract(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, tc := range []struct{ row, features string }{
			{"CVTPL2PS X0,X1", "+sse2"}, {"CVTPL2PD X0,X1", "+sse2"},
			{"CVTPL2PS M0,X1", "+mmx,+sse"}, {"CVTPL2PD M0,X1", "+mmx,+sse,+sse2"},
			{"VCVTDQ2PS Y0,Y1", "+avx"}, {"VCVTDQ2PD X0,Y1", "+avx"},
			{"VCVTDQ2PS X0,K1,X1", "+avx512f,+avx512vl"}, {"VCVTDQ2PD Y0,Z1", "+avx512f"},
		} {
			file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+tc.row+"\nRET\n")
			if err != nil {
				t.Fatal(err)
			}
			fn := file.Funcs[0]
			if got := inferFuncTargetFeaturesForGOARCH(ArchAMD64, arch, fn); got != tc.features {
				t.Fatal("source ISA contract", arch, tc.row, got, tc.features)
			}
			ins := fn.Instrs[1]
			if ins.Args[0].Reg == "M0" {
				continue
			}
			var ir strings.Builder
			ctx := newX86Ctx(&ir, fn, FuncSig{Ret: Void}, testResolveSym("example"), nil, arch, "", false)
			if err := ctx.emitEntryAllocas(); err != nil {
				t.Fatal(err)
			}
			before := ir.Len()
			if _, _, err := ctx.lowerSignedPackedIntegerFloat("VCVTUDQ2PS", ins); err == nil || ir.Len() != before {
				t.Fatal("contradictory source opcode emitted IR", err)
			}
			ins.x86AddressBits = 32
			if _, _, err := ctx.lowerSignedPackedIntegerFloat(ins.Op, ins); !errors.Is(err, ErrProbeNeedsContext) || ir.Len() != before {
				t.Fatal("unbound address identity emitted IR", err)
			}
			ins.x86AddressBits = 0
			ins.Op = Op(strings.ToLower(string(ins.Op)))
			if _, _, err := ctx.lowerSignedPackedIntegerFloat(ins.Op, ins); err != nil {
				t.Fatal("equivalent lowercase helper identity rejected", err)
			}
		}
	}
	for _, row := range []string{"VCVTDQ2PS.SAE Z0,Z1", "VCVTDQ2PS.RU_SAE (AX),Z1", "VCVTDQ2PD.RN_SAE Y0,Z1"} {
		file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		if got := inferFuncTargetFeaturesForGOARCH(ArchAMD64, "amd64", file.Funcs[0]); got != "" {
			t.Fatal("invalid signed conversion acquired ISA attributes", row, got)
		}
	}
}
