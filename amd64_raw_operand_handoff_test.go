package plan9asm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
	"golang.org/x/arch/x86/x86asm"
)

// This oracle assembles literal BYTE directives, not GoSyntax's synthetic
// memory spelling. Its complete contiguous text bytes independently anchor
// the private decoder input on both supported SDK test lanes.
func x86HandoffOriginalBytes(t *testing.T, arch string, code []byte) {
	t.Helper()
	source := "TEXT handoff(SB),4,$0-0\n" + x86RawDisplacementDirectives(code)
	listing, err := x86ScalarLiteralGoListing(t, arch, source)
	if err != nil {
		t.Fatal(err)
	}
	var actual []byte
	for _, line := range strings.Split(listing, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "0x") || len(fields[1]) != 2 {
			continue
		}
		offset, err := strconv.ParseUint(fields[0][2:], 16, 32)
		if err != nil || int(offset) != len(actual) {
			t.Fatalf("noncontiguous stock Go bytes: %s", line)
		}
		for _, field := range fields[1:] {
			if len(field) != 2 {
				break
			}
			value, err := strconv.ParseUint(field, 16, 8)
			if err != nil {
				t.Fatal(err)
			}
			actual = append(actual, byte(value))
		}
	}
	if !bytes.Equal(actual, code) {
		t.Fatalf("stock Go raw text = %x, want complete %x", actual, code)
	}
}

func TestX86PrivateDecoderX87PhysicalMemoryHandoff(t *testing.T) {
	forms := []struct {
		op   Op
		code []byte
	}{
		{"FSAVE", []byte{0xdd, 0x74, 0x8b, 0xf9}},
		{"FSTCW", []byte{0xd9, 0x7c, 0x8b, 0xf9}},
		{"FSTENV", []byte{0xd9, 0x74, 0x8b, 0xf9}},
		{"FSTSW", []byte{0xdd, 0x7c, 0x8b, 0xf9}},
		{"FLDCW", []byte{0xd9, 0x6c, 0x8b, 0xf9}},
		{"FLDENV", []byte{0xd9, 0x64, 0x8b, 0xf9}},
		{"FRSTOR", []byte{0xdd, 0x64, 0x8b, 0xf9}},
	}
	segments := []struct {
		prefix byte
		reg    Reg
	}{{0x26, ES}, {0x2e, CS}, {0x36, SS}, {0x3e, DS}, {0x64, FS}, {0x65, GS}}
	for _, arch := range []string{"386", "amd64"} {
		var complete []byte
		for _, form := range forms {
			for _, segment := range segments {
				complete = append(complete, segment.prefix)
				complete = append(complete, form.code...)
			}
		}
		complete = append(complete, 0xc3)
		x86HandoffOriginalBytes(t, arch, complete)
		for _, form := range forms {
			for _, segment := range segments {
				t.Run(arch+"/"+string(form.op)+"/"+string(segment.reg), func(t *testing.T) {
					code := append([]byte{segment.prefix}, form.code...)
					decoded, err := decodeX86RawDirectives(rawX86Function(code), arch)
					if err != nil {
						t.Fatal(err)
					}
					want := Operand{Kind: OpMem, Mem: MemRef{Base: BX, Index: CX, Scale: 4, Off: -7, Segment: segment.reg}}
					if len(decoded.Instrs) != 1 || decoded.Instrs[0].Op != form.op || len(decoded.Instrs[0].Args) != 1 || !reflect.DeepEqual(decoded.Instrs[0].Args[0], want) {
						t.Fatalf("physical memory changed: %+v, want %s %+v", decoded.Instrs, form.op, want)
					}
				})
			}
		}
	}
}

func TestX86PrivateDecoderImplicitPhysicalSegments(t *testing.T) {
	forms := []struct {
		name   string
		opcode byte
		source bool
		dest   bool
	}{
		{"MOVS", 0xa4, true, true}, {"CMPS", 0xa6, true, true},
		{"STOS", 0xaa, false, true}, {"LODS", 0xac, true, false},
		{"SCAS", 0xae, false, true}, {"INS", 0x6c, false, true},
		{"OUTS", 0x6e, true, false},
	}
	for _, mode := range []int{32, 64} {
		arch := "amd64"
		if mode == 32 {
			arch = "386"
		}
		var complete []byte
		for _, form := range forms {
			for _, segment := range []byte{0, 0x64, 0x65} {
				for _, repeat := range []byte{0, 0xf2, 0xf3} {
					code := []byte{0x67}
					if repeat != 0 {
						code = append(code, repeat)
					}
					if segment != 0 {
						code = append(code, segment)
					}
					code = append(code, form.opcode)
					complete = append(complete, code...)
				}
			}
		}
		complete = append(complete, 0xc3)
		x86HandoffOriginalBytes(t, arch, complete)
		for _, form := range forms {
			for _, segment := range []byte{0, 0x64, 0x65} {
				for _, repeat := range []byte{0, 0xf2, 0xf3} {
					t.Run(fmt.Sprintf("mode%d/%s/seg%02x/rep%02x", mode, form.name, segment, repeat), func(t *testing.T) {
						code := []byte{0x67}
						if repeat != 0 {
							code = append(code, repeat)
						}
						if segment != 0 {
							code = append(code, segment)
						}
						code = append(code, form.opcode)
						physical, err := x86asm.Decode(code, mode)
						if err != nil || physical.Len != len(code) {
							t.Fatalf("complete physical decode: %+v %v", physical, err)
						}
						instructions, err := x86AddressOverrideDecode(mode, code)
						if err != nil {
							t.Fatal(err)
						}
						var memories []MemRef
						for _, ins := range instructions {
							if ins.Op == OpTEXT || ins.Op == OpRET {
								continue
							}
							if !ins.x86UnboundAddressOverride || ins.x86AddressBits != mode/2 {
								t.Fatalf("lost physical address restriction: %+v", ins)
							}
							for _, arg := range ins.Args {
								if arg.Kind == OpMem {
									memories = append(memories, arg.Mem)
								}
							}
							ctx, out := newAMD64CtxWithFuncForTest(t, Func{Instrs: []Instr{ins}}, FuncSig{Ret: Void}, nil)
							ctx.goarch = arch
							before := out.String()
							if terminated, err := ctx.lowerInstr(0, 0, ins, nil, nil); terminated || !errors.Is(err, ErrProbeNeedsContext) || out.String() != before {
								t.Fatalf("unbound physical memory emitted IR: %t %v", terminated, err)
							}
						}
						wantSource := DS
						if segment == 0x64 {
							wantSource = FS
						} else if segment == 0x65 {
							wantSource = GS
						}
						seenSource, seenDest := false, false
						for _, memory := range memories {
							if memory.Base == SI && memory.Segment == wantSource {
								seenSource = true
							} else if memory.Base == DI && memory.Segment == ES {
								seenDest = true
							} else {
								t.Fatalf("unexpected implicit physical memory: %+v", memory)
							}
						}
						if seenSource != form.source || seenDest != form.dest {
							t.Fatalf("lost source/destination segments: %+v", memories)
						}
					})
				}
			}
		}
	}
}

func TestX86PrivateDecoderRIPPoolAllPublicAPIs(t *testing.T) {
	code := []byte{0x48, 0x8d, 0x05, 1, 0, 0, 0, 0xc3, 0x41}
	x86HandoffOriginalBytes(t, "amd64", code)
	file := mustParseX86RawReturn(t, x86StaticReturnSource("Bound", code))
	for _, target := range x86RawReturnTargets {
		if target.arch != "amd64" {
			continue // RIP is a mode64-only machine addressing form.
		}
		opt := Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: map[string]FuncSig{"Bound": {Name: "Bound", Ret: Void}}}
		for _, api := range []string{"text", "module", "owned-module"} {
			t.Run(target.triple+"/"+api, func(t *testing.T) {
				var ir string
				var err error
				switch api {
				case "text":
					ir, err = Translate(file, opt)
				case "module":
					module, callErr := TranslateModule(file, opt)
					err = callErr
					if err == nil {
						ir = module.String()
						module.Dispose()
					}
				default:
					ctx := llvm.NewContext()
					module, callErr := TranslateModuleInContext(ctx, file, opt)
					err = callErr
					if err == nil {
						ir = module.String()
						module.Dispose()
					}
					ctx.Dispose()
				}
				if err != nil {
					t.Fatal(err)
				}
				llc := findLLVM22Tool("llc")
				if llc == "" {
					t.Fatal("LLVM 22 llc required")
				}
				compileLLVMToObject(t, llc, target.triple, "handoff.ll", "handoff.o", ir)
			})
		}
	}
}
