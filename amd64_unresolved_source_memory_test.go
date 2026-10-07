package plan9asm

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// A source inventory must retain every later instruction when a generated
// header has not supplied an earlier displacement. Retention is not permission
// to translate that displacement, even when a raw RET follows it.
func TestX86UnresolvedSourceMemoryPreservesSyntax(t *testing.T) {
	for _, row := range []struct {
		address string
		raw     string
		base    Reg
		index   Reg
		scale   int64
	}{
		{"UNKNOWN(AX)", "UNKNOWN", AX, "", 0},
		{"g_stackguard0(BP)", "g_stackguard0", BP, "", 0},
		{"g(CX)", "g", CX, "", 0},
		{"(UNKNOWN+4)(AX)", "(UNKNOWN+4)", AX, "", 0},
		{"UNKNOWN+4(AX)(CX*2)", "UNKNOWN+4", AX, CX, 2},
		{"UNKNOWN(AX)(CX)", "UNKNOWN", AX, CX, 0},
	} {
		t.Run(row.address, func(t *testing.T) {
			source := "TEXT one(SB),4,$0-0\nMOVL " + row.address + ",BX\nADDL $7,BX\nRET\n"
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			memory := file.Funcs[0].Instrs[0].Args[0]
			if memory.Kind != OpMem || memory.Mem.OffRaw != row.raw || memory.Mem.Off != 0 ||
				memory.Mem.Base != row.base || memory.Mem.Index != row.index || memory.Mem.Scale != row.scale {
				t.Fatalf("lost unresolved syntax: %#v", memory)
			}
			for _, arch := range []string{"386", "amd64"} {
				if err := ProbeInstruction(ArchAMD64, arch, file.Funcs[0].Instrs[0]); !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("%s unresolved memory acquired a lowering witness: %v", arch, err)
				}
			}
			if len(file.Funcs[0].Instrs) != 3 || file.Funcs[0].Instrs[1].Op != "ADDL" {
				t.Fatal("source inventory lost the instruction after the unresolved operand")
			}
		})
	}
}

func TestX86UnresolvedSourceMemoryThreeAPIsRejectBeforeRawDecoding(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, opcode := range []string{"MOVL", "VMOVD", "PXOR", "LEAL"} {
			for _, raw := range []bool{false, true} {
				t.Run(target.triple+"/"+opcode+"/raw="+strconv.FormatBool(raw), func(t *testing.T) {
					destination := "BX"
					if opcode == "VMOVD" || opcode == "PXOR" {
						destination = "X0"
					}
					ending := "RET"
					if raw {
						ending = "BYTE $0xc3"
					}
					source := "TEXT ·Value(SB),4,$0-0\n" + opcode + " UNKNOWN(AX)," + destination + "\n" + ending + "\n"
					goAssemblerOracleOutput(t, target.arch, source, "", false)
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					resolve := func(name string) string { return strings.TrimPrefix(name, "·") }
					options := Options{Goarch: target.arch, TargetTriple: target.triple, ResolveSym: resolve,
						Sigs: map[string]FuncSig{"Value": {Name: "Value", Ret: Void}}}
					pkg := mustGoPackage(t, "memory", "package memory\nfunc Value()\n")
					for _, api := range []string{"text", "module", "GoModule"} {
						t.Run(api, func(t *testing.T) {
							ctx := llvm.NewContext()
							defer ctx.Dispose()
							var err error
							switch api {
							case "text":
								var ir string
								ir, err = Translate(file, options)
								if ir != "" {
									t.Fatal("unresolved memory emitted IR")
								}
							case "module":
								var module llvm.Module
								module, err = TranslateModuleInContext(ctx, file, options)
								if !module.IsNil() {
									module.Dispose()
									t.Fatal("unresolved memory emitted a module")
								}
							case "GoModule":
								var result *GoModuleTranslation
								result, err = translateGoModuleInContext(ctx, pkg, []byte(source), GoModuleOptions{
									GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: resolve})
								if result != nil && !result.Module.IsNil() {
									result.Module.Dispose()
									t.Fatal("unresolved memory emitted a Go module")
								}
							}
							if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "UNKNOWN") {
								t.Fatalf("%s lost the unresolved displacement diagnostic: %v", api, err)
							}
						})
					}
				})
			}
		}
	}
}
