package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type x86RawLogicalResultCase struct {
	name string
	code []byte
	mask int
}

// Independent Go _yvandnpd/_yvaddsubpd/_yvblendmpd witnesses: twelve VEX
// names, sixteen EVEX names, every vector width, mask role and memory tuple.
// No production spec supplies this directory or its expected cardinality.
func x86RawLogicalResultCases() []x86RawLogicalResultCase {
	var rows []x86RawLogicalResultCase
	for _, operation := range []struct {
		opcode byte
		stem   string
	}{{0x54, "VAND"}, {0x55, "VANDN"}, {0x56, "VOR"}, {0x57, "VXOR"}} {
		for pp, suffix := range []string{"PS", "PD"} {
			for _, width := range []int{16, 32} {
				for _, vex3 := range []bool{false, true} {
					rows = append(rows, x86RawLogicalResultCase{
						name: fmt.Sprintf("%s%sVEX%d_%t", operation.stem, suffix, width, vex3),
						code: x86RawFloatLogicalLiteral(operation.opcode, byte(pp), vex3, width),
					})
				}
			}
			for _, width := range []int{16, 32, 64} {
				for _, mask := range []int{0, 3, 7} {
					for _, broadcast := range []bool{false, true} {
						rows = append(rows, x86RawLogicalResultCase{
							name: fmt.Sprintf("%s%sEVEX%d_%d_%t", operation.stem, suffix, width, mask, broadcast),
							code: x86RawEVEXFloatLogicalLiteral(operation.opcode, byte(pp), width, mask, mask == 7, broadcast),
							mask: mask,
						})
					}
				}
			}
		}
	}
	for _, op := range []Op{"VPAND", "VPANDN", "VPOR", "VPXOR"} {
		for _, width := range []int{16, 32} {
			for _, vex3 := range []bool{false, true} {
				code := x86RawPackedLogicalLiteral(op, false, width, 0, false, false)
				if !vex3 {
					code = append([]byte{0xc5, code[2]}, code[3:]...)
				}
				rows = append(rows, x86RawLogicalResultCase{
					name: fmt.Sprintf("%sVEX%d_%t", op, width, vex3), code: code,
				})
			}
		}
	}
	for _, op := range []Op{"VPANDD", "VPANDQ", "VPANDND", "VPANDNQ", "VPORD", "VPORQ", "VPXORD", "VPXORQ"} {
		for _, width := range []int{16, 32, 64} {
			for _, mask := range []int{0, 3, 7} {
				for _, broadcast := range []bool{false, true} {
					rows = append(rows, x86RawLogicalResultCase{
						name: fmt.Sprintf("%sEVEX%d_%d_%t", op, width, mask, broadcast),
						code: x86RawPackedLogicalLiteral(op, true, width, mask, mask == 7, broadcast),
						mask: mask,
					})
				}
			}
		}
	}
	return rows
}

func x86RawLogicalResultSource(row x86RawLogicalResultCase) string {
	var source strings.Builder
	fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n", row.name)
	source.WriteString("MOVL $5,AX\nKMOVD AX,K3\nKMOVD AX,K7\nVPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\n")
	for _, value := range row.code {
		fmt.Fprintf(&source, "BYTE $%#02x\n", value)
	}
	return source.String()
}

func x86RawLogicalLiteralObjects(t *testing.T, triple string) {
	t.Helper()
	rows := x86RawLogicalResultCases()
	if len(rows) != 336 {
		t.Fatal("complete independent logical literal directory changed", len(rows))
	}
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for start := 0; start < len(rows); start += 64 {
		end := start + 64
		if end > len(rows) {
			end = len(rows)
		}
		t.Run(fmt.Sprintf("batch-%d", start/64), func(t *testing.T) {
			var source, declarations strings.Builder
			for _, row := range rows[start:end] {
				source.WriteString(x86RawLogicalResultSource(row))
				fmt.Fprintf(&declarations, "func %s()\n", row.name)
			}
			x86LEAActualTargetGoObject(t, "amd64", triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), "amd64", triple, api)
					if err != nil {
						t.Fatal("exact logical literal/mask/source bytes lost", err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, triple, text, level)
						}
					}
				})
			}
		})
	}
}

func TestX86RawLogicalLiteralCompleteLinuxObjects(t *testing.T) {
	x86RawLogicalLiteralObjects(t, "x86_64-unknown-linux-gnu")
}

func TestX86RawLogicalLiteralCompleteDarwinObjects(t *testing.T) {
	x86RawLogicalLiteralObjects(t, "x86_64-apple-darwin")
}

func TestX86RawLogicalLiteralCompleteWindowsObjects(t *testing.T) {
	x86RawLogicalLiteralObjects(t, "x86_64-pc-windows-msvc")
}

func TestX86RawLogicalLiteralKeepsIndependentInputsAndSource(t *testing.T) {
	for _, row := range x86RawLogicalResultCases() {
		source := x86RawLogicalResultSource(row)
		declaration := "func " + row.name + "()"
		variants := []string{
			strings.Replace(source, "$0-0", "$16-0", 1),
			strings.Replace(source, "MOVL $5,AX", "MOVQ $0,SP\nMOVL $5,AX", 1),
		}
		if row.mask != 0 {
			variants = append(variants, strings.NewReplacer("KMOVD AX,K3\n", "", "KMOVD AX,K7\n", "").Replace(source))
		}
		for _, changed := range variants {
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, changed, declaration, "amd64", "x86_64-unknown-linux-gnu", api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("literal borrowed a missing mask or raw-return frame", row.name, api, err)
				}
			}
		}
		for _, change := range []string{"operand", "DATA"} {
			file, err := normalizeX86RawFile(mustParseX86RawReturn(t, source), "amd64")
			if err != nil || len(file.Data) != 1 {
				t.Fatal("original literal source lost its exact DATA", row.name, err)
			}
			if err := checkX86ScalarFPResultInitialization(file.Funcs[0], FuncSig{Name: row.name, Ret: Void}, "amd64"); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("declaration-only query acquired lowering DATA", row.name, err)
			}
			if change == "DATA" {
				file.Data[0].Payload[0] ^= 1
			} else {
				for index := range file.Funcs[0].Instrs {
					instruction := &file.Funcs[0].Instrs[index]
					if instruction.x86RIPLiteral {
						instruction.Args[0].Sym += "changed"
						break
					}
				}
			}
			options := Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu",
				Sigs:       map[string]FuncSig{row.name: {Name: row.name, Ret: Void}},
				ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") }}
			if ir, err := Translate(file, options); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("changed source borrowed original literal identity", row.name, change, err)
			}
		}
	}
}
