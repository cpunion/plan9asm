package plan9asm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/llvm"
)

var x86ExtensionTestOperations = []string{
	"MOVBWSX", "MOVBWZX", "MOVBLSX", "MOVBLZX", "MOVBQSX", "MOVBQZX",
	"MOVWLSX", "MOVWLZX", "MOVWQSX", "MOVWQZX", "MOVLQSX", "MOVLQZX",
	"MOVSWW", "MOVZWW",
}

func TestX86ExtensionCompletePrimaryTables(t *testing.T) {
	path := filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "x86", "asm6.go")
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	table := string(bytes)
	for _, name := range []string{"ymb_rl", "yml_rl"} {
		pattern := regexp.MustCompile(`(?s)var ` + name + ` = \[\]ytab\{\s*\{(Zmb_r|Zm_r), 1, argList\{(Ymb|Yml), Yrl\}\},\s*\}`)
		if !pattern.MatchString(table) {
			t.Fatalf("actual Go table changed: %s", name)
		}
	}
	rows := regexp.MustCompile(`\{A(MOV[A-Z]+), (ymb_rl|yml_rl), (Pm|Pw|Pq|Pe|Px),`).FindAllStringSubmatch(table, -1)
	var names []string
	for _, row := range rows {
		names = append(names, row[1])
	}
	want := append([]string(nil), x86ExtensionTestOperations...)
	sort.Strings(names)
	sort.Strings(want)
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("complete extension rows changed: got %v, want %v", names, want)
	}
}

func x86ExtensionGoListing(t *testing.T, arch, instruction string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "extension.s")
	source := "TEXT ·extension(SB),4,$0-0\n" + instruction + "\nRET\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-S", "-p", "extension",
		"-o", filepath.Join(dir, "extension.o"), path)
	command.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+arch)
	output, err := runtimeOracleCombinedOutput(ctx, command)
	return string(output), err
}

func TestX86ExtensionActualGoAliasAudit(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, instruction := range []string{
			"MOVSWW AL,AX", "MOVSWW AH,AX", "MOVSWW BP,AX",
			"MOVZWW AL,AX", "MOVZWW AH,AX", "MOVZWW BP,AX",
			"MOVBLSX AH,AX", "MOVBLSX BP,AX", "MOVLQZX AX,BX",
		} {
			t.Run(arch+"/"+strings.ReplaceAll(instruction, " ", "_"), func(t *testing.T) {
				listing, err := x86ExtensionGoListing(t, arch, instruction)
				if err != nil {
					t.Fatalf("actual %s assembler: %v\n%s", runtime.Version(), err, listing)
				}
				t.Log(listing)
			})
		}
	}
}

func TestX86ExtensionGoPhysicalRegisterViews(t *testing.T) {
	for _, test := range []struct {
		instruction, bytes string
		source             Reg
	}{
		{"MOVSWW AL,AX", "66 0f bf c0 c3", AX},
		{"MOVSWW AH,AX", "66 0f bf c4 c3", SP},
		{"MOVZWW CH,AX", "66 0f b7 c5 c3", BP},
		{"MOVSWW DH,AX", "66 0f bf c6 c3", SI},
		{"MOVZWW BH,AX", "66 0f b7 c7 c3", DI},
		{"MOVBLSX AH,AX", "0f be c4 c3", AH},
		{"MOVBLSX AH,R8", "44 0f be c4 c3", SP},
		{"MOVBQSX BH,AX", "48 0f be c7 c3", DI},
	} {
		t.Run(test.instruction, func(t *testing.T) {
			listing, err := x86ExtensionGoListing(t, "amd64", test.instruction)
			if err != nil || !strings.Contains(listing, test.bytes) {
				t.Fatalf("actual Go machine view: %v\n%s", err, listing)
			}
			file, err := Parse(ArchAMD64, "TEXT v(SB),4,$0-0\n"+test.instruction+"\nRET\n")
			if err != nil {
				t.Fatal(err)
			}
			ins := file.Funcs[0].Instrs[1]
			before := append([]Operand(nil), ins.Args...)
			form, recognized, err := x86ScalarExtensionProperties(ins.Op, ins, "amd64")
			if !recognized || err != nil || form.source.Reg != test.source || !reflect.DeepEqual(before, ins.Args) {
				t.Fatalf("Go ModRM view/source mutation: %+v recognized=%t err=%v", form, recognized, err)
			}
		})
	}
}

func TestX86ExtensionCompleteFormatsThreeAPIFiveObjects(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var body strings.Builder
			for _, operation := range x86ExtensionTestOperations {
				if target.arch == "386" && (operation == "MOVBQSX" || operation == "MOVBQZX" ||
					operation == "MOVWQSX" || operation == "MOVWQZX" || operation == "MOVLQSX") {
					continue
				}
				for _, source := range []string{"AX", "8(BX)", "-1(BX)(SI*2)", "x+0(FP)", "@WORD(SP)", "·payload(SB)"} {
					fmt.Fprintf(&body, "%s %s,CX\n", operation, source)
				}
				if strings.HasPrefix(operation, "MOVB") || operation == "MOVSWW" || operation == "MOVZWW" {
					fmt.Fprintf(&body, "%s BL,CX\n", operation)
					if target.arch == "386" {
						fmt.Fprintf(&body, "%s DI,CX\n", operation)
					}
				}
			}
			body.WriteString("@MOV CX,ret+@WORD(FP)\nRET")
			source, sig, pkg := x86SourceStackNumericalFixture(target.arch, body.String())
			source += "GLOBL ·payload(SB),8,$16\nDATA ·payload+0(SB)/8,$0x1122334455667788\n"
			requireX86GoAssemblerResult(t, target.arch, source, true)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			options := x86SourceStackOptions(target.arch, target.triple, sig)
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Fatal("LLVM22 llc required")
			}
			compileLLVMToObject(t, llc, target.triple, "extension-form.ll", "extension-form.o", ir)
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, options)
			if err != nil {
				t.Fatal(err)
			}
			defer module.Dispose()
			compileLLVMToObject(t, llc, target.triple, "extension-form-module.ll", "extension-form-module.o", module.String())
			ctx = llvm.NewContext()
			defer ctx.Dispose()
			translation, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{
				GOARCH: target.arch, TargetTriple: target.triple,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer translation.Module.Dispose()
			compileLLVMToObject(t, llc, target.triple, "extension-form-go.ll", "extension-form-go.o", translation.Module.String())
		})
	}
}

func TestX86ExtensionSourceAndPhysicalDecoderShareEffects(t *testing.T) {
	for _, operation := range x86ExtensionTestOperations {
		for _, body := range []string{operation + " BX,AX", x86ExtensionNumericalRawBytes(operation)} {
			for _, defined := range []bool{false, true} {
				prefix := ""
				if defined {
					prefix = "MOVQ $0xff80,BX\n"
				}
				assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", prefix+body+"\nMOVL AX,(DX)", 1)
				requireX86GoAssemblerResult(t, "amd64", assembly, true)
				pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
				err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu")
				if defined && err != nil || !defined && !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("source/physical input availability defined=%t: %v\n%s", defined, err, body)
				}
			}
		}
	}
}

func TestX86ExtensionSharedFormIsOrderedAndImmutable(t *testing.T) {
	for _, operation := range x86ExtensionTestOperations {
		file, err := Parse(ArchAMD64, "TEXT v(SB),4,$0-0\n"+operation+" AX,BX\nRET\n")
		if err != nil {
			t.Fatal(err)
		}
		original := file.Funcs[0].Instrs[1]
		for _, spelling := range []Op{original.Op, "", "MOVQ"} {
			instruction := original
			instruction.Op = spelling
			before := append([]Operand(nil), instruction.Args...)
			_, recognized, err := x86ScalarExtensionProperties(original.Op, instruction, "amd64")
			if !recognized || (err == nil) != (spelling != "MOVQ") || !reflect.DeepEqual(before, instruction.Args) {
				t.Fatalf("%s spelling %s: recognized=%t err=%v source changed=%t", operation, spelling, recognized, err, !reflect.DeepEqual(before, instruction.Args))
			}
		}
	}
}

func TestX86ExtensionActualGoRejectedFormats(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, operation := range x86ExtensionTestOperations {
			if target.arch == "386" && amd64ScalarExtensionMoveSpecs[Op(operation)].rejectedOn386 {
				requireX86GoAssemblerResult(t, target.arch, "TEXT ·extension(SB),4,$0-0\n"+operation+" AX,BX\nRET\n", false)
				continue
			}
			for _, operands := range []string{"$1,AX", "AX,(BX)", "X1,AX", "AX,AL", "8(BAD),AX"} {
				source, sig, pkg := x86SourceStackNumericalFixture(target.arch, operation+" "+operands+"\nRET")
				requireX86GoAssemblerResult(t, target.arch, source, false)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					continue // The public parser itself rejected the original format.
				}
				options := x86SourceStackOptions(target.arch, target.triple, sig)
				if _, err := Translate(file, options); err == nil {
					t.Fatalf("Translate accepted Go-rejected %s %s", operation, operands)
				}
				ctx := llvm.NewContext()
				module, err := TranslateModuleInContext(ctx, file, options)
				if !module.IsNil() {
					module.Dispose()
				}
				ctx.Dispose()
				if err == nil {
					t.Fatalf("TranslateModule accepted Go-rejected %s %s", operation, operands)
				}
				ctx = llvm.NewContext()
				translation, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{GOARCH: target.arch, TargetTriple: target.triple})
				if translation != nil {
					translation.Module.Dispose()
				}
				ctx.Dispose()
				if err == nil {
					t.Fatalf("TranslateGoModule accepted Go-rejected %s %s", operation, operands)
				}
			}
		}
	}
}

func TestX86ExtensionRawMOVSXDRequiresKnownNativeWidth(t *testing.T) {
	// Register and default-32 memory widths have independent native contracts.
	// The 66 memory footprint still differs across vendor descriptions.
	for _, encoding := range []string{"BYTE $0x66\nWORD $0x0763", "BYTE $0x66\nWORD $0x0f63"} {
		assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", "MOVQ $0x80000000,BX\n"+encoding+"\nMOVL AX,(DX)", 1)
		requireX86GoAssemblerResult(t, "amd64", assembly, true)
		pkg, file := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
		resolve := func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) }
		sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc"} {
			options := Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs, ResolveSym: resolve}
			if _, err := Translate(file, options); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unknown raw MOVSXD width Translate: %v", err)
			}
			ctx := llvm.NewContext()
			module, err := TranslateModuleInContext(ctx, file, options)
			if !module.IsNil() {
				module.Dispose()
			}
			ctx.Dispose()
			if !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unknown raw MOVSXD width module: %v", err)
			}
			if err := x86CallerCoreTranslate(t, pkg, assembly, triple); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unknown raw MOVSXD width Go module: %v", err)
			}
		}
	}
}

func TestX86GoCallerRegionExtensionSourceInputs(t *testing.T) {
	for _, operation := range x86ExtensionTestOperations {
		for _, input := range []string{"register", "memory"} {
			t.Run(operation+"/"+input, func(t *testing.T) {
				operand := "AX"
				if input == "memory" {
					operand = "(DI)"
				}
				body := fmt.Sprintf("%s %s,AX\nMOVL AX,(DX)", operation, operand)
				assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
				requireX86GoAssemblerResult(t, "amd64", assembly, true)
				pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
				if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestX86GoCallerRegionExtensionDefinedBytesAndPointerNegatives(t *testing.T) {
	for _, test := range []struct {
		name, body string
		valid      bool
	}{
		{"one initialized low byte", "MOVB $128,BX\nMOVBLZX BL,AX", true},
		{"one initialized high byte", "MOVB $128,BH\nMOVBLSX BH,AX", true},
		{"complete low word via aliases", "MOVB $128,BL\nMOVB $255,BH\nMOVSWW BL,AX", true},
		{"word low byte incomplete", "MOVB $128,BL\nMOVSWW BL,AX", false},
		{"unknown source", "MOVLQSX BX,AX", false},
		{"partial source long", "MOVW $7,BX\nMOVLQSX BX,AX", false},
		{"caller pointer sign extension", "MOVLQSX DI,AX", false},
		{"caller pointer byte extension", "MOVBQZX DI,AX", false},
		{"caller pointer apparent high-byte REX", "MOVBQZX BH,AX", false},
		{"implicit SP source", "MOVSWW AH,AX", false},
		{"implicit BP source", "MOVZWW CH,AX", false},
		{"partial pointer destination", "MOVQ DI,BX\nMOVBWZX AL,BX", false},
		{"partial pointer replaced completely", "MOVQ DI,BX\nMOVBLZX AL,BX", true},
		{"unaligned pointer slot", "MOVLQSX a+1(FP),AX", false},
		{"partial pointer FP slot", "MOVLQSX a+0(FP),AX", false},
		{"return PC load", "MOVLQSX 8(BP),AX", false},
		{"saved BP load", "MOVWLZX 0(BP),AX", false},
		{"different defined bytes join", "CMPQ AX,$0\nJE other\nMOVB $7,BL\nother:\nMOVBLZX BL,AX", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", test.body+"\nMOVL AX,(DX)", 1)
			requireX86GoAssemblerResult(t, "amd64", assembly, true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu")
			if test.valid && err != nil || !test.valid && !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("valid=%t: %v", test.valid, err)
			}
		})
	}
}

func TestX86GoCallerRegionExtensionThreeAPIs(t *testing.T) {
	assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", "MOVLQSX (DI),AX\nMOVL AX,(DX)", 1)
	pkg, file := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
	resolve := func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) }
	sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			options := Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs, ResolveSym: resolve}
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, findLLVM22Tool("llc"), triple, "extension.ll", "extension.o", ir)
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, options)
			if err != nil {
				t.Fatal(err)
			}
			defer module.Dispose()
			compileLLVMToObject(t, findLLVM22Tool("llc"), triple, "extension-module.ll", "extension-module.o", module.String())
			if err := x86CallerCoreTranslate(t, pkg, assembly, triple); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestX86SourceStackExtensionRequiresDefinedLowBytes(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, operation := range x86ExtensionTestOperations {
			if target.arch == "386" && amd64ScalarExtensionMoveSpecs[Op(operation)].rejectedOn386 {
				continue // Covered separately as actual Go-rejected formats, not Context forms.
			}
			for _, prefix := range []string{"", "MOVB $7,BX\n", "@MOV x+0(FP),BX\n"} {
				body := prefix + operation + " BX,AX\n@MOV AX,ret+@WORD(FP)\nBYTE $0xc3"
				source, sig, pkg := x86SourceStackNumericalFixture(target.arch, body)
				requireX86GoAssemblerResult(t, target.arch, source, true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				options := x86SourceStackOptions(target.arch, target.triple, sig)
				if _, err := Translate(file, options); !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("noncaller source acquired extension proof: %v\n%s", err, source)
				}
				ctx := llvm.NewContext()
				module, err := TranslateModuleInContext(ctx, file, options)
				if !module.IsNil() {
					module.Dispose()
				}
				ctx.Dispose()
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("noncaller source acquired module proof: %v\n%s", err, source)
				}
				ctx = llvm.NewContext()
				translation, err := translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{GOARCH: target.arch, TargetTriple: target.triple})
				if translation != nil {
					translation.Module.Dispose()
				}
				ctx.Dispose()
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("noncaller source acquired Go module proof: %v\n%s", err, source)
				}
			}
		}
	}
}

func TestX86ExtensionCannotBorrowStaleCallerMemoryEffects(t *testing.T) {
	assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", "MOVLQSX (DI),AX\nMOVL AX,(DX)", 1)
	for _, change := range []string{"width", "opcode", "source", "missing caller"} {
		t.Run(change, func(t *testing.T) {
			pkg, file := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			resolve := func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) }
			sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			options := Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", Sigs: sigs, ResolveSym: resolve}
			normalized, err := normalizeX86RawFile(file, "amd64", options)
			if err != nil {
				t.Fatal(err)
			}
			for index := range normalized.Funcs[0].Instrs {
				instruction := &normalized.Funcs[0].Instrs[index]
				if instruction.Op != "MOVLQSX" {
					continue
				}
				switch change {
				case "width":
					access := instruction.x86CallerMemory[0]
					access.width = 8 // Actual read width is 4, never the result width.
					instruction.x86CallerMemory[0] = access
				case "opcode":
					instruction.Op = "MOVWQSX"
				case "source":
					instruction.Args[0].Mem.Base = DX
				}
			}
			if change == "missing caller" {
				sig := sigs["caller.kernel"]
				sig.x86CallerRegions = nil
				sigs["caller.kernel"] = sig
			}
			if change == "width" {
				if x86RawReturnMemoryBound(normalized.Funcs[0], normalized, options, true) {
					t.Fatal("extension accepted a stale destination-width memory certificate")
				}
				return // A final consumer legitimately recomputes unchanged source effects.
			}
			if _, err := Translate(normalized, options); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("changed %s borrowed an extension caller proof: %v", change, err)
			}
		})
	}
}
