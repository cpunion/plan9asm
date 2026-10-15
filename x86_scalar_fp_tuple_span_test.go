package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// These byte tuples have no padding. Their independently specified little-
// endian bytes describe physical FP storage, not a coercion to its first field.
type x86ScalarFPTupleSpanCase struct {
	name, source, declaration string
	want                      []uint64
}

func x86ScalarFPTupleSpanCases(arch string, read bool) []x86ScalarFPTupleSpanCase {
	widths := []int{2, 4}
	if arch == "amd64" {
		widths = append(widths, 8)
	}
	var cases []x86ScalarFPTupleSpanCase
	for _, width := range widths {
		for _, boolean := range []int{0, width - 1, -1} {
			position := "numeric"
			if boolean == 0 {
				position = "bool-first"
			} else if boolean == width-1 {
				position = "bool-last"
			}
			var fields []string
			var values []uint64
			var packed uint64
			for index := 0; index < width; index++ {
				typ, value := "uint8", uint64(index+2)
				if boolean == 0 {
					value--
				}
				if index == boolean {
					typ, value = "bool", 1
				}
				fields = append(fields, typ)
				values = append(values, value)
				packed |= value << uint(8*index)
			}
			move := map[int]string{2: "MOVW", 4: "MOVL", 8: "MOVQ"}[width]
			var body strings.Builder
			role, extent := "store", width
			if read {
				role, extent = "read", 2*width
				for index, value := range values {
					fmt.Fprintf(&body, "MOVB $%d,ret+%d(FP)\n", value, index)
				}
				fmt.Fprintf(&body, "%s ret+0(FP),AX\n%s AX,ret+%d(FP)\n", move, move, width)
				fields = append(fields, fmt.Sprintf("uint%d", 8*width))
				values = append(values, packed)
			} else {
				// Register transport keeps Q's full-width literal legal: x86 has
				// no arbitrary imm64-to-memory encoding.
				fmt.Fprintf(&body, "%s $0x%x,AX\n%s AX,ret+0(FP)\n", move, packed, move)
			}
			body.WriteString("RET\n")
			cases = append(cases, x86ScalarFPTupleSpanCase{
				name:        fmt.Sprintf("%s/%s/%s", role, move, position),
				declaration: "func Value() (" + strings.Join(fields, ",") + ")",
				source:      fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", extent, body.String()),
				want:        values,
			})
			if read && boolean >= 0 {
				// The Go bool return must be restored to one, but its earlier
				// physical byte two must survive the wide numerical read.
				physical := strings.Replace(body.String(), fmt.Sprintf("MOVB $1,ret+%d(FP)", boolean), fmt.Sprintf("MOVB $2,ret+%d(FP)", boolean), 1)
				physical = strings.Replace(physical, "RET\n", fmt.Sprintf("MOVB $1,ret+%d(FP)\nRET\n", boolean), 1)
				expected := append([]uint64(nil), values...)
				expected[len(expected)-1] += uint64(1) << uint(8*boolean)
				cases = append(cases, x86ScalarFPTupleSpanCase{
					name:        fmt.Sprintf("%s/%s/%s-physical-two", role, move, position),
					declaration: "func Value() (" + strings.Join(fields, ",") + ")",
					source:      fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", extent, physical),
					want:        expected,
				})
			}
		}
	}
	return cases
}

func x86ScalarFPTupleSpanConstantResult(t *testing.T, ir string, want []uint64) {
	t.Helper()
	context := llvm.NewContext()
	defer context.Dispose()
	module, err := parseIRModuleInContext(context, ir)
	if !module.IsNil() {
		defer module.Dispose()
	}
	if err != nil {
		t.Fatal(err)
	}
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	if err := module.RunPasses("default<O2>", llvm.TargetMachine{}, options); err != nil {
		t.Fatal(err)
	}
	function := module.NamedFunction("Value")
	if function.IsNil() {
		t.Fatal("missing exact translated Value entry")
	}
	returns := 0
	for block := function.FirstBasicBlock(); !block.IsNil(); block = llvm.NextBasicBlock(block) {
		for instruction := block.FirstInstruction(); !instruction.IsNil(); instruction = llvm.NextInstruction(instruction) {
			if instruction.IsAReturnInst().IsNil() {
				continue
			}
			returns++
			if instruction.OperandsCount() != 1 {
				t.Fatalf("missing physical-byte tuple result: %s", function.String())
			}
			result := instruction.Operand(0)
			if result.IsAConstantStruct().IsNil() || result.OperandsCount() != len(want) {
				t.Fatalf("whole constant fixture did not produce an inspectable tuple: %s", function.String())
			}
			for index, expected := range want {
				field := result.Operand(index)
				if field.IsAConstantInt().IsNil() || field.ZExtValue() != expected {
					t.Errorf("FP span lost field %d: got %s, want %d; optimized actual fixture:\n%s", index, field.String(), expected, function.String())
				}
			}
		}
	}
	if returns != 1 {
		t.Fatalf("constant fixture has %d returns, want one", returns)
	}
}

func x86ScalarFPTupleSpanThreeAPIs(t *testing.T, read bool) {
	t.Helper()
	for _, target := range x86RawReturnTargets {
		for _, row := range x86ScalarFPTupleSpanCases(target.arch, read) {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, row.source, row.declaration, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatalf("complete initialized physical FP span rejected: %v", err)
					}
					x86ScalarFPTupleSpanConstantResult(t, ir, row.want)
				})
			}
		}
	}
}

func TestX86ScalarFPWideTupleStoresPreserveEveryField(t *testing.T) {
	x86ScalarFPTupleSpanThreeAPIs(t, false)
}

func TestX86ScalarFPWideTupleReadsPreserveEveryField(t *testing.T) {
	x86ScalarFPTupleSpanThreeAPIs(t, true)
}

func TestX86ScalarFPWideTupleSpanOriginalGoAssembly(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			var source strings.Builder
			for _, read := range []bool{false, true} {
				for index, row := range x86ScalarFPTupleSpanCases(arch, read) {
					name := fmt.Sprintf("Span%dRead%t", index, read)
					source.WriteString(strings.ReplaceAll(row.source, "Value", name))
				}
			}
			goAssemblerOracleOutput(t, arch, source.String(), "", true)
		})
	}
}
