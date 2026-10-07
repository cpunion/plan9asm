package plan9asm

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
	"golang.org/x/arch/x86/x86asm"
)

// Compare source operands with independently assembled physical operands.
// The initial strict comparison is retained in the test-first RED commit.
func x86MinimumPositionOperandsMatchSource(decoded, source []Operand) bool {
	return reflect.DeepEqual(decoded, source)
}

func x86MinimumPositionOriginalGoBytes(t *testing.T, arch, source string) []byte {
	t.Helper()
	listing, err := x86ScalarLiteralGoListing(t, arch, source)
	if err != nil {
		t.Fatal(err)
	}
	var code []byte
	for _, line := range strings.Split(listing, "\n") {
		text := strings.TrimLeft(line, "\t ")
		fields := strings.Fields(text)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "0x") || len(fields[1]) != 2 {
			continue
		}
		offset, err := strconv.ParseUint(fields[0][2:], 16, 32)
		start := strings.IndexByte(text, ' ')
		if err != nil || int(offset) != len(code) || start < 0 || len(text) < start+48 {
			t.Fatalf("incomplete/noncontiguous stock Go hex row: %s", line)
		}
		gap := false
		for column := 0; column < 16; column++ {
			field := text[start+3*column : start+3*(column+1)]
			if field == "   " {
				gap = true
				continue
			}
			value, err := strconv.ParseUint(field[1:], 16, 8)
			if gap || field[0] != ' ' || err != nil {
				t.Fatalf("invalid stock Go hex column: %q", field)
			}
			code = append(code, byte(value))
		}
	}
	if len(code) == 0 {
		t.Fatal("stock Go omitted the original instruction bytes")
	}
	return code
}

func TestX86RawMinimumPositionSourceOperandComparison(t *testing.T) {
	source := []Operand{{Kind: OpMem, Mem: MemRef{Base: SP, Off: 8192}}, {Kind: OpReg, Reg: "X0"}}
	before := fmt.Sprintf("%#v", source)
	for _, scale := range []int64{0, 1, 2, 4, 8} {
		decoded := append([]Operand(nil), source...)
		decoded[0].Mem.Scale = scale
		if !x86MinimumPositionOperandsMatchSource(decoded, source) {
			t.Errorf("inactive SIB scale %d changed address semantics: decoded=%#v source=%#v", scale, decoded, source)
		}
	}
	mutations := []struct {
		name string
		edit func(*Operand)
	}{
		{"invalid scale", func(a *Operand) { a.Mem.Scale = 3 }},
		{"base", func(a *Operand) { a.Mem.Base = BP }},
		{"index", func(a *Operand) { a.Mem.Index = CX }},
		{"offset", func(a *Operand) { a.Mem.Off++ }},
		{"segment", func(a *Operand) { a.Mem.Segment = FS }},
		{"symbol", func(a *Operand) { a.Mem.Sym = "slot" }},
		{"unresolved offset", func(a *Operand) { a.Mem.OffRaw = "unknown" }},
		{"index extension", func(a *Operand) { a.Mem.IndexExt = ExtendUXTW }},
		{"operand role", func(a *Operand) { a.Kind = OpFP }},
		{"unrelated field", func(a *Operand) { a.Imm = 1 }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			decoded := append([]Operand(nil), source...)
			mutation.edit(&decoded[0])
			if x86MinimumPositionOperandsMatchSource(decoded, source) {
				t.Fatalf("ignored a real operand difference: %#v", decoded)
			}
		})
	}
	indexed := append([]Operand(nil), source...)
	indexed[0].Mem.Index, indexed[0].Mem.Scale = CX, 4
	for _, scale := range []int64{0, 1, 2, 3, 8} {
		decoded := append([]Operand(nil), indexed...)
		decoded[0].Mem.Scale = scale
		if x86MinimumPositionOperandsMatchSource(decoded, indexed) {
			t.Fatalf("ignored active index scale %d", scale)
		}
	}
	if !x86MinimumPositionOperandsMatchSource(indexed, indexed) || before != fmt.Sprintf("%#v", source) {
		t.Fatal("comparison lost exact indexed equality or mutated its source")
	}
}

func TestX86RawMinimumPositionInactiveSIBScalePhysicalWitness(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		mode := 64
		if arch == "386" {
			mode = 32
		}
		for shift := byte(0); shift < 4; shift++ {
			// Original legacy PHMINPOSUW 8192(SP),X0. SIB.index=4
			// suppresses the index; all four scale bit patterns are ignored.
			code := []byte{0x66, 0x0f, 0x38, 0x41, 0x84, 0x24 | shift<<6, 0, 0x20, 0, 0}
			x86HandoffOriginalBytes(t, arch, code)
			physical, err := x86asm.Decode(code, mode)
			memory, ok := physical.Args[1].(x86asm.Mem)
			base := x86asm.RSP
			if mode == 32 {
				base = x86asm.ESP
			}
			if err != nil || physical.Len != len(code) || physical.Op != x86asm.PHMINPOSUW ||
				!ok || memory.Base != base || memory.Index != 0 || memory.Scale != 1<<shift || memory.Disp != 8192 {
				t.Fatalf("independent full SIB witness: %+v %v", physical, err)
			}
			decoded, err := decodeX86RawDirectives(rawX86Function(code), arch)
			want := Operand{Kind: OpMem, Mem: MemRef{Base: SP, Scale: 1 << shift, Off: 8192}}
			if err != nil || len(decoded.Instrs) != 1 || len(decoded.Instrs[0].Args) != 2 ||
				!reflect.DeepEqual(decoded.Instrs[0].Args[0], want) {
				t.Fatalf("decoder discarded its physical SIB metadata: %+v %v", decoded, err)
			}
			original := append([]byte(nil), code...)
			// A synthetic register valuation independently confirms the
			// address has no index term; this is not native stack permission.
			for _, sp := range []uint64{0, 1, 0x7fffffff, 0x123456789abcdef0} {
				actual := sp + uint64(memory.Disp)
				wantAddress := sp + 8192
				if mode == 32 {
					actual, wantAddress = uint64(uint32(actual)), uint64(uint32(wantAddress))
				}
				if actual != wantAddress || !bytes.Equal(code, original) {
					t.Fatal("inactive scale changed effective address or input bytes")
				}
			}
		}
	}
}

func TestX86RawMinimumPositionGoAddressFormsAllAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			registers := []string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"}
			if arch == "amd64" {
				for i := 8; i < 16; i++ {
					registers = append(registers, fmt.Sprintf("R%d", i))
				}
			}
			var addresses []string
			for _, base := range registers {
				for _, offset := range []int64{0, -128, 127, -129, 128, 8192, -2147483648, 2147483647} {
					addresses = append(addresses, fmt.Sprintf("%d(%s)", offset, base))
				}
			}
			for _, index := range registers {
				if index == "SP" {
					continue // SIB.index=4 is the absent-index witness above.
				}
				for _, scale := range []int{1, 2, 4, 8} {
					addresses = append(addresses, fmt.Sprintf("-7(BP)(%s*%d)", index, scale),
						fmt.Sprintf("8192(%s*%d)", index, scale))
				}
			}
			var source strings.Builder
			source.WriteString("TEXT addresses(SB),4,$0-0\n")
			for _, op := range []string{"PHMINPOSUW", "VPHMINPOSUW"} {
				for _, address := range addresses {
					fmt.Fprintf(&source, "%s %s,X0\n", op, address)
				}
			}
			source.WriteString("RET\n")
			code := x86MinimumPositionOriginalGoBytes(t, arch, source.String())
			parsed, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := decodeX86RawDirectives(rawX86Function(code), arch)
			if err != nil || len(decoded.Instrs) != len(parsed.Funcs[0].Instrs)-1 {
				t.Fatalf("complete original instruction sequence: %+v %v", decoded, err)
			}
			for i, instruction := range decoded.Instrs {
				want := parsed.Funcs[0].Instrs[i+1]
				if instruction.Op != want.Op || !x86MinimumPositionOperandsMatchSource(instruction.Args, want.Args) {
					t.Fatalf("original Go address %d %s: physical=%#v source=%#v", i, want.Raw, instruction.Args, want.Args)
				}
			}
			file := x86RawStackObservingFormProbeFile(t, arch, "addresses", code)
			for _, target := range x86RawReturnTargets {
				if target.arch != arch {
					continue
				}
				for _, api := range []string{"text", "module", "context"} {
					t.Run(target.triple+"/"+api, func(t *testing.T) {
						options := Options{Goarch: arch, TargetTriple: target.triple,
							Sigs: map[string]FuncSig{"addresses": {Name: "addresses", Ret: Void}}}
						var ir string
						switch api {
						case "text":
							ir, err = Translate(file, options)
						case "module":
							module, callErr := TranslateModule(file, options)
							err = callErr
							if err == nil {
								ir = module.String()
								module.Dispose()
							}
						case "context":
							ctx := llvm.NewContext()
							defer ctx.Dispose()
							module, callErr := TranslateModuleInContext(ctx, file, options)
							err = callErr
							if err == nil {
								ir = module.String()
								module.Dispose()
							}
						}
						if err != nil {
							t.Fatal(err)
						}
						compileLLVMToObject(t, llc, target.triple, "addresses.ll", "addresses.o", ir)
					})
				}
			}
			t.Logf("actual Go address encodings=%d, compile-only faulting probes; no native stack/return permission", 2*len(addresses))
		})
	}
}
