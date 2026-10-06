package plan9asm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// Portable compile control for the unchanged native CI reference, not an
// execution result. It prevents a platform-only driver from going unchecked
// on development hosts that cannot directly run IA32 executables.
func TestX86Timestamp386OriginalNativeReferenceCompiles(t *testing.T) {
	var source, declarations strings.Builder
	declarations.WriteString("package main\n")
	for _, op := range []Op{"RDTSC", "RDTSCP"} {
		declaration, assembly := x86Timestamp386OriginalFixture(op)
		declarations.WriteString(declaration)
		source.WriteString(assembly)
	}
	dir := t.TempDir()
	for name, contents := range map[string]string{
		"go.mod": "module timestamp386\n\ngo 1.20\n", "decl.go": declarations.String(),
		"main.go": x86Timestamp386OriginalDriver, "timestamp_386.s": source.String(),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(dir, "original-go")
	command := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-p=2", "-o", executable, ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOOS=linux", "GOARCH=386", "CGO_ENABLED=0",
		"GOENV=off", "GOWORK=off", "GOTOOLCHAIN=local", "GOFLAGS=", "GOEXPERIMENT=")
	x86FloatCompareCaptureCommand(t, command)
	sealed, err := x87IA32StaticExecutable(executable)
	if err != nil || sealed.verify() != nil {
		t.Fatalf("actual original native IA32 reference did not compile as stable ELF32: %v", err)
	}
	t.Log("actual original Go static IA32 executable compiled; not executed on this portable axis")
}

func TestX86TimestampCompleteOfficialRulesAndActualEncodedIdentity(t *testing.T) {
	table, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(table), "var ynone = []ytab{\n\t{Zlit, 1, argList{}},\n}") {
		t.Fatal("Go timestamp zero-operand ordered row changed")
	}
	if len(x86TimestampSpecs) != 2 {
		t.Fatal("timestamp family must have exactly two canonical opcodes")
	}
	for _, row := range []struct {
		op               Op
		registration     string
		literal, machine []byte
		outputs          amd64ImplicitRegisterSet
	}{
		{"RDTSC", "{ARDTSC, ynone, Pm, opBytes{0x31}}",
			[]byte{0x31}, []byte{0x0f, 0x31}, amd64ImplicitAX | amd64ImplicitDX},
		{"RDTSCP", "{ARDTSCP, ynone, Pm, opBytes{0x01, 0xf9, 0}}",
			[]byte{0x01, 0xf9, 0}, []byte{0x0f, 0x01, 0xf9},
			amd64ImplicitAX | amd64ImplicitDX | amd64ImplicitCX},
	} {
		if !strings.Contains(string(table), row.registration) {
			t.Fatal("complete actual Go prefix/literal row changed", row.registration)
		}
		spec := x86TimestampSpecs[row.op]
		if !bytes.Equal(spec.opBytes, row.literal) || spec.outputs != row.outputs {
			t.Fatal("typed timestamp metadata differs from independent complete row", row.op, spec)
		}
		for _, arch := range []string{"386", "amd64"} {
			t.Run(arch+"/"+string(row.op), func(t *testing.T) {
				source := "TEXT clock(SB),4,$0-0\n" + string(row.op) + "\nRET\n"
				actual := assembleX87ControlBytes(t, arch, source)
				expected := append(append([]byte(nil), row.machine...), 0xc3)
				if !bytes.Equal(actual, expected) {
					t.Fatalf("actual Go canonical machine code: %x want %x", actual, expected)
				}
				mode := 64
				if arch == "386" {
					mode = 32
				}
				decoded, err := decodeX86RawDirectiveGroup(actual[:len(actual)-1], mode, 0, source, map[string]bool{})
				if err != nil || len(decoded) != 1 {
					t.Fatalf("actual Go timestamp raw consumer: %#v/%v", decoded, err)
				}
				before := fmt.Sprintf("%#v", decoded[0])
				got, known, err := x86TimestampProperties(row.op, decoded[0], arch)
				if !known || err != nil || !reflect.DeepEqual(got, spec) || before != fmt.Sprintf("%#v", decoded[0]) {
					t.Fatal("raw timestamp identity changed effects/input", decoded, got, err)
				}
			})
		}
	}
}

func TestX86TimestampDispatchIdentityAndImplicitOutputLiveness(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range []Op{"RDTSC", "RDTSCP"} {
			file, err := Parse(ArchAMD64, "TEXT clock(SB),4,$0-0\n"+string(op)+"\nRET\n")
			if err != nil {
				t.Fatal(err)
			}
			for _, identity := range []Op{"", op, Op(strings.ToLower(string(op))), "MOVQ", Op(string(op) + ".BAD")} {
				t.Run(arch+"/"+string(op)+"/identity="+string(identity), func(t *testing.T) {
					instruction := file.Funcs[0].Instrs[1]
					instruction.Op = identity
					input := fmt.Sprintf("%#v", instruction)
					ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
					ctx.goarch = arch
					ir, temporary := output.String(), ctx.tmp
					known, terminated, err := ctx.lowerX86Timestamp(op, instruction)
					invalid := identity != "" && !strings.EqualFold(string(identity), string(op))
					if !known || terminated || (err != nil) != invalid ||
						invalid && (ir != output.String() || temporary != ctx.tmp) ||
						input != fmt.Sprintf("%#v", instruction) {
						t.Fatal("timestamp identity emitted before rejection or mutated its input", identity, err)
					}
				})
			}
			for _, invalidArch := range []string{"", "arm", "arm64", "wasm"} {
				ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
				ctx.goarch = invalidArch
				before, temporary := output.String(), ctx.tmp
				if known, terminated, err := ctx.lowerX86Timestamp(op, file.Funcs[0].Instrs[1]); !known || terminated || err == nil || before != output.String() || temporary != ctx.tmp {
					t.Fatal("unrelated target admitted timestamp grammar", invalidArch, err)
				}
			}
			for _, suffix := range []string{".Z", ".SAE", ".BCST"} {
				ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
				ctx.goarch = arch
				before, temporary := output.String(), ctx.tmp
				if known, terminated, err := ctx.lowerX86Timestamp(Op(string(op)+suffix), file.Funcs[0].Instrs[1]); !known || terminated || err == nil || before != output.String() || temporary != ctx.tmp {
					t.Fatal("timestamp suffix emitted before grammar rejection", suffix, err)
				}
			}
			ctx, _ := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
			for _, reg := range []Reg{AX, DX, CX} {
				required := reg != CX || op == "RDTSCP"
				if ctx.usedRegs[reg] != required || (ctx.regSlot[reg] != "") != required {
					t.Fatal("implicit architectural result omitted without explicit operand", op, reg)
				}
			}
		}
	}
}

func TestX86TimestampNamedAndRawThreeAPIsFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"}, {"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"}, {"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		for _, raw := range []bool{false, true} {
			var source strings.Builder
			declarations := "package timestamp\n"
			pointerMove, pointerSize := "MOVQ", 8
			if target.arch == "386" {
				pointerMove, pointerSize = "MOVL", 4
			}
			for _, op := range []Op{"RDTSC", "RDTSCP"} {
				if target.arch == "386" {
					declaration, original := x86Timestamp386OriginalFixture(op)
					declarations += declaration
					if raw {
						actual := assembleX87ControlBytes(t, "386",
							"TEXT clock(SB),4,$0-0\n"+string(op)+"\nRET\n")
						var literal strings.Builder
						for _, value := range actual[:len(actual)-1] {
							fmt.Fprintf(&literal, "BYTE $%d\n", value)
						}
						original = strings.Replace(original, "    "+string(op)+"\n", literal.String(), 1)
					}
					source.WriteString(original)
					continue
				}
				name := strings.ToLower(string(op))
				declarations += "func " + name + "(out *[4]uint32)\n"
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s out+0(FP),DI\nMOVL $0x7e6d5c4b,CX\n", name, pointerSize, pointerMove)
				if raw {
					bytes := assembleX87ControlBytes(t, target.arch,
						"TEXT clock(SB),4,$0-0\n"+string(op)+"\nRET\n")
					for _, value := range bytes[:len(bytes)-1] {
						fmt.Fprintf(&source, "BYTE $%d\n", value)
					}
				} else {
					fmt.Fprintln(&source, op)
				}
				source.WriteString("MOVL AX,0(DI)\nMOVL DX,4(DI)\nMOVL CX,8(DI)\nRET\n")
			}
			goAssemblerOracleOutput(t, target.arch, source.String(), "", true)
			pkg := mustGoPackage(t, "timestamp", declarations)
			resolve := func(symbol string) string { return strings.TrimPrefix(symbol, "·") }
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(fmt.Sprintf("%s/raw=%t/%s", target.triple, raw, api), func(t *testing.T) {
					ctx := llvm.NewContext()
					defer ctx.Dispose()
					var ir string
					if api == "Go" {
						translated, err := translateGoModuleInContext(ctx, pkg, []byte(source.String()), GoModuleOptions{
							GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: resolve,
						})
						if translated != nil && !translated.Module.IsNil() {
							defer translated.Module.Dispose()
						}
						if err != nil {
							t.Fatal(err)
						}
						ir = translated.Module.String()
					} else {
						file, err := Parse(ArchAMD64, source.String())
						if err != nil {
							t.Fatal(err)
						}
						sigs, err := goSigsForAsmFile(pkg, file, resolve, target.arch, nil, nil, nil)
						if err != nil {
							t.Fatal(err)
						}
						options := Options{Goarch: target.arch, TargetTriple: target.triple, ResolveSym: resolve, Sigs: sigs}
						if api == "text" {
							ir, err = Translate(file, options)
						} else {
							module, failure := TranslateModuleInContext(ctx, file, options)
							if !module.IsNil() {
								defer module.Dispose()
								ir = module.String()
							}
							err = failure
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					for _, opcode := range []string{"rdtsc", "rdtscp"} {
						if !strings.Contains(ir, "asm sideeffect \""+opcode+"\"") {
							t.Fatal("object consumer lost the actual architectural timestamp", opcode)
						}
					}
					compileLLVMToObject(t, llc, target.triple, "timestamp.ll", "timestamp.o", ir)
				})
			}
		}
	}
}
