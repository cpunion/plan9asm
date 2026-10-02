package plan9asm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// obj6.progedit tests the original float64 bits for positive zero before
// rounding an SS constant to float32. In particular, underflow is a pool load,
// not the legacy XORPS rewrite. Keep independent expected IEEE bit patterns.
type x86ScalarLiteralValue struct {
	name, text string
	bits32     uint32
	bits64     uint64
	zero       bool
}

var x86ScalarLiteralValues = []x86ScalarLiteralValue{
	{name: "positiveZero", text: "0.0", zero: true},
	{name: "negativeZero", text: "-0.0", bits32: 0x80000000, bits64: 0x8000000000000000},
	{name: "fraction", text: "1.25", bits32: 0x3fa00000, bits64: 0x3ff4000000000000},
	{name: "negative", text: "-2.5", bits32: 0xc0200000, bits64: 0xc004000000000000},
	{name: "roundEven", text: "1.000000059604644775390625", bits32: 0x3f800000, bits64: 0x3ff0000010000000},
	{name: "underflow", text: "1e-50", bits64: 0x358dee7a4ad4b81f},
	{name: "overflow32", text: "1e40", bits32: 0x7f800000, bits64: 0x483d6329f1c35ca5},
}

func x86ScalarLiteralGoListing(t *testing.T, arch, source string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	asm := filepath.Join(dir, "literal.s")
	if err := os.WriteFile(asm, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-S", "-p", "literal", "-o", filepath.Join(dir, "literal.o"), asm)
	command.Env = append(os.Environ(), "GOENV=off", "GOWORK=off", "GOTOOLCHAIN=local", "GOOS=linux", "GOARCH="+arch)
	command.WaitDelay = 2 * time.Second
	listing, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("actual Go literal assembler timeout: %v", ctx.Err())
	}
	return string(listing), err
}

func TestX86ScalarLiteralActualGoRewriteInventory(t *testing.T) {
	if !goToolchainAtLeast(runtime.Version(), 1, 27) {
		t.Skip("current Go 1.27 optab inventory; historical lane checks the implementation separately")
	}
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"MOVSS", "MOVSD", "VMOVSS", "VMOVSD"} {
			legacy := !strings.HasPrefix(op, "V")
			for _, value := range x86ScalarLiteralValues {
				for _, register := range []int{0, 7, 8, 15, 16, 31} {
					zeroRewrite := value.zero && register < 16
					accepted := !legacy || register < 16 && (arch != "386" || register < 8)
					if arch == "386" && zeroRewrite && register >= 8 {
						accepted = false // the rewritten legacy XORPS is illegal in 32-bit mode
					}
					instruction := fmt.Sprintf("%s $%s,X%d", op, value.text, register)
					source := "TEXT literal(SB),4,$0-0\n" + instruction + "\nRET\n"
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					form, recognized, err := amd64ScalarVectorMoveProperties(file.Funcs[0].Instrs[1], arch)
					if !recognized || (err == nil) != accepted || accepted && (!form.literal || form.legacyZero != zeroRewrite) {
						t.Fatalf("%s %s literal rewrite axes = %+v / %v", arch, instruction, form, err)
					}
					listing, err := x86ScalarLiteralGoListing(t, arch, source)
					if (err == nil) != accepted {
						t.Fatalf("%s %s expected accepted=%t: %v\n%s", arch, instruction, accepted, err, listing)
					}
					if !accepted {
						continue
					}
					// The same actual Go object establishes the generation that
					// owns upper-lane semantics; a VMOV spelling alone is not it.
					var prefix byte
					found := false
					for _, line := range strings.Split(listing, "\n") {
						fields := strings.Fields(line)
						if len(fields) < 2 || fields[0] != "0x0000" || len(fields[1]) != 2 {
							continue
						}
						first, err := strconv.ParseUint(fields[1], 16, 8)
						if err == nil {
							prefix, found = byte(first), true
							break
						}
					}
					if !found || (legacy || zeroRewrite) && (prefix == 0xc4 || prefix == 0xc5 || prefix == 0x62) ||
						!legacy && !zeroRewrite && (register < 16 && prefix != 0xc5 || register >= 16 && prefix != 0x62) {
						t.Fatalf("%s %s actual encoding prefix=%02x does not match the final rewrite\n%s", arch, instruction, prefix, listing)
					}
					if zeroRewrite {
						if !strings.Contains(listing, "XORPS") || strings.Contains(listing, "$f32.") || strings.Contains(listing, "$f64.") {
							t.Fatalf("positive zero did not become legacy XORPS: %s", listing)
						}
					} else {
						pool := fmt.Sprintf("$f64.%016x", value.bits64)
						if strings.HasSuffix(op, "SS") {
							pool = fmt.Sprintf("$f32.%08x", value.bits32)
						}
						if !strings.Contains(listing, pool) {
							t.Fatalf("constant-pool rounding/sign differs: expected %s\n%s", pool, listing)
						}
					}
				}
			}
		}
	}
}

func TestX86ScalarLiteralGoOperandGrammar(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"VMOVSS", "VMOVSD"} {
			for _, value := range x86ScalarLiteralValues[:3] {
				for _, register := range []int{0, 8, 31} {
					for _, shape := range []struct {
						name, suffix, middle string
						mask                 int
					}{
						{name: "two", mask: -1},
						{name: "mask1", middle: "K1,", mask: 1},
						{name: "mask7zero", suffix: ".Z", middle: "K7,", mask: 1},
						{name: "three merge", middle: "X1,", mask: -1},
						{name: "four merge", middle: "X1,K1,", mask: 2},
						{name: "four zero", suffix: ".Z", middle: "X1,K7,", mask: 2},
					} {
						zeroRewrite := value.zero && register < 16
						accepted := shape.name == "two" || shape.mask == 1 && !zeroRewrite
						if arch == "386" && zeroRewrite && register >= 8 {
							accepted = false
						}
						instruction := fmt.Sprintf("%s%s $%s,%sX%d", op, shape.suffix, value.text, shape.middle, register)
						source := "TEXT literal(SB),4,$0-0\n" + instruction + "\nRET\n"
						t.Run(arch+"/"+instruction, func(t *testing.T) {
							requireX86GoAssemblerResult(t, arch, source, accepted)
							file, err := Parse(ArchAMD64, source)
							if err != nil {
								t.Fatal(err)
							}
							form, recognized, err := amd64ScalarVectorMoveProperties(file.Funcs[0].Instrs[1], arch)
							if !recognized || (err == nil) != accepted || accepted && (!form.literal || form.mask != shape.mask || form.upper != -1 || form.memory != -1) {
								t.Fatalf("actual Go literal grammar expected accepted=%t: %+v / %v", accepted, form, err)
							}
						})
					}
				}
			}
		}
	}
}

func TestX86ScalarLiteralKnownByteEffectsFollowFinalEncoding(t *testing.T) {
	for _, example := range []struct {
		instruction string
		old, want   uint64
		maskKnown   bool
	}{
		{instruction: "MOVSS $1.25,X0", want: 0xffff},
		{instruction: "MOVSD $-0.0,X0", old: 0xffff00000000, want: 0xffff0000ffff},
		{instruction: "VMOVSS $0.0,X0", want: 0xffff},
		{instruction: "VMOVSD $0.0,X0", old: 0xffff00000000, want: 0xffff0000ffff},
		{instruction: "VMOVSS $1e-50,X0", want: ^uint64(0)},
		{instruction: "VMOVSD $-0.0,X0", want: ^uint64(0)},
		{instruction: "VMOVSS $0.0,X31", want: ^uint64(0)},
		{instruction: "VMOVSS $1.25,K1,X0", want: ^uint64(0xf), maskKnown: true},
		{instruction: "VMOVSD.Z $1.25,K1,X0", want: ^uint64(0), maskKnown: true},
	} {
		file, err := Parse(ArchAMD64, "TEXT literal(SB),4,$0-0\n"+example.instruction+"\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		instruction := file.Funcs[0].Instrs[1]
		destination := x86CallerVectorRoot(instruction.Args[len(instruction.Args)-1].Reg)
		state := x86SourceStackState{
			vectors: map[Reg]uint64{destination: example.old}, masks: map[Reg]bool{"K1": example.maskKnown},
		}
		if !x86CallerVectorMoveKnown(instruction) || !x86CallerVectorMoveStep(&instruction, &state, nil) {
			t.Fatalf("literal transfer rejected %s", example.instruction)
		}
		if got := state.vectors[destination]; got != example.want {
			t.Fatalf("%s byte initialization got=%016x want=%016x", example.instruction, got, example.want)
		}
		if len(instruction.x86CallerMemory) != 0 {
			t.Fatal("literal constant pool acquired a user-memory access certificate")
		}
	}
}

func TestX86ScalarLiteralCallerRegionFreshWrites(t *testing.T) {
	for _, op := range []string{"MOVSS", "MOVSD", "VMOVSS", "VMOVSD"} {
		typ := "float32"
		if strings.HasSuffix(op, "SD") {
			typ = "float64"
		}
		for _, value := range x86ScalarLiteralValues {
			body := fmt.Sprintf("%s $%s,X0\n%s X0,(DX)", op, value.text, op)
			source := strings.ReplaceAll(x86CallerFixtureSource, "float32", typ)
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			requireX86GoAssemblerResult(t, "amd64", assembly, true)
			pkg, _ := x86CallerCoreFixture(t, source, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatalf("%s %s exact caller-region literal rejected: %v", op, value.text, err)
			}
		}
	}
}

func TestX86ScalarLiteralRejectsIntegerUnresolvedAndInvalidMaskAxes(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, instruction := range []string{
			"MOVSS $0,X0", "MOVSD $1,X0", "MOVSD $7,X0", "VMOVSS $0,X0", "VMOVSD $1,X0",
			"MOVSS $1.25,K1,X0", "MOVSD.Z $1.25,X0", "VMOVSS.Z $1.25,X0",
			"VMOVSS $1.25,K0,X0", "VMOVSD $-0.0,K8,X0", "VMOVSS $1.25,(AX)",
			"VMOVSD $1.25,X1,K1,X0", "VMOVSS $0.0,K1,X0", "VMOVSD.Z $0.0,K1,X15",
		} {
			source := "TEXT literal(SB),4,$0-0\n" + instruction + "\nRET\n"
			requireX86GoAssemblerResult(t, arch, source, false)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				continue
			}
			if _, _, err := amd64ScalarVectorMoveProperties(file.Funcs[0].Instrs[1], arch); err == nil {
				t.Fatalf("literal grammar accepted %s %s", arch, instruction)
			}
			triple := "x86_64-unknown-linux-gnu"
			if arch == "386" {
				triple = "i386-unknown-linux-gnu"
			}
			if _, err := Translate(file, Options{Goarch: arch, TargetTriple: triple,
				Sigs: map[string]FuncSig{"literal": {Name: "literal", Ret: Void}},
			}); err == nil {
				t.Fatalf("lowerer accepted %s %s", arch, instruction)
			}
		}
	}
	for _, op := range []string{"MOVSS", "MOVSD", "VMOVSS", "VMOVSD"} {
		file, err := Parse(ArchAMD64, "TEXT literal(SB),4,$0-0\n"+op+" $1.25,X0\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		file.Funcs[0].Instrs[1].Args[0].ImmRaw = "unresolved"
		if _, _, err := amd64ScalarVectorMoveProperties(file.Funcs[0].Instrs[1], "amd64"); err == nil {
			t.Fatal("unresolved immediate inherited a literal-pool rewrite")
		}
	}
}
